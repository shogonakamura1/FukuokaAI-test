package usecase

import (
	"fmt"

	"fukuoka-ai-api/infra/service"
	"fukuoka-ai-api/models"
	"fukuoka-ai-api/repositories"

	"github.com/google/uuid"
)

type TripUsecase interface {
	CreateTrip(userID string, mustPlaces []string, interestTags []string, freeText string) (*CreateTripOutput, error)
	RecomputeTrip(userID, tripID string, orderedPlaceIDs []string, stayMinutesMap map[string]int) (*RecomputeTripOutput, error)
	GetShare(shareID string) (*GetShareOutput, error)
}

type tripUsecase struct {
	tripRepo repositories.ITripRepository
	mlService service.IMLService
}

func NewTripUsecase(tripRepo repositories.ITripRepository, mlService service.IMLService) TripUsecase {
	return &tripUsecase{
		tripRepo:  tripRepo,
		mlService: mlService,
	}
}

type CreateTripOutput struct {
	TripID     string
	ShareID    string
	Itinerary  []models.TripPlace
	Candidates []models.TripPlace
	Route      *models.Route
}

func (u *tripUsecase) CreateTrip(userID string, mustPlaces []string, interestTags []string, freeText string) (*CreateTripOutput, error) {
	// ユーザーが存在しない場合は作成
	if err := u.tripRepo.EnsureUser(userID); err != nil {
		return nil, fmt.Errorf("ユーザー作成に失敗しました: %w", err)
	}

	// MLサービスに推薦依頼
	mlResp, err := u.mlService.Recommend(mustPlaces, interestTags, freeText)
	if err != nil {
		return nil, fmt.Errorf("推薦サービスエラー: %w", err)
	}

	// 旅程をDBに保存
	tripID := uuid.New().String()
	shareID := uuid.New().String()

	trip := &models.Trip{
		ID:        tripID,
		UserID:    userID,
		Title:     fmt.Sprintf("福岡観光 %s", tripID[:8]),
		StartTime: "10:00",
	}

	if err := u.tripRepo.CreateTrip(trip); err != nil {
		return nil, fmt.Errorf("旅程の保存に失敗しました: %w", err)
	}

	// スポットを保存
	itineraryData, ok := mlResp["initial_itinerary"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("初期旅程データの形式が不正です")
	}
	for i, item := range itineraryData {
		placeData, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		place := &models.TripPlace{
			ID:          uuid.New().String(),
			TripID:      tripID,
			PlaceID:     getString(placeData, "place_id"),
			Name:        getString(placeData, "name"),
			Lat:         getFloat64(placeData, "lat"),
			Lng:         getFloat64(placeData, "lng"),
			Kind:        "must",
			StayMinutes: 60,
			OrderIndex:  i,
		}
		if reason, ok := placeData["reason"].(string); ok {
			place.Reason = reason
		}
		if reviewSummary, ok := placeData["review_summary"].(string); ok {
			place.ReviewSummary = reviewSummary
		}
		if photoURL, ok := placeData["photo_url"].(string); ok {
			place.PhotoURL = photoURL
		}
		if err := u.tripRepo.CreateTripPlace(place); err != nil {
			return nil, fmt.Errorf("スポットの保存に失敗しました: %w", err)
		}
	}

	// 共有情報を保存
	if err := u.tripRepo.CreateShare(shareID, tripID); err != nil {
		return nil, fmt.Errorf("共有情報の保存に失敗しました: %w", err)
	}

	// レスポンス作成
	// DBから保存したplacesを取得してTimeRangeを計算
	savedPlaces, err := u.tripRepo.GetTripPlaces(tripID)
	if err != nil {
		return nil, fmt.Errorf("スポットの取得に失敗しました: %w", err)
	}
	itinerary := convertItineraryFromPlaces(savedPlaces, trip.StartTime)

	candidates := []models.TripPlace{}
	if cands, ok := mlResp["candidates"].([]interface{}); ok {
		for _, item := range cands {
			placeData, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			candidates = append(candidates, models.TripPlace{
				PlaceID:       getString(placeData, "place_id"),
				Name:          getString(placeData, "name"),
				Lat:           getFloat64(placeData, "lat"),
				Lng:           getFloat64(placeData, "lng"),
				Category:      getString(placeData, "category"),
				Reason:        getString(placeData, "reason"),
				ReviewSummary: getString(placeData, "review_summary"),
				PhotoURL:      getString(placeData, "photo_url"),
			})
		}
	}

	output := &CreateTripOutput{
		TripID:     tripID,
		ShareID:    shareID,
		Itinerary:  itinerary,
		Candidates: candidates,
	}

	if route, ok := mlResp["route"].(map[string]interface{}); ok {
		if polyline, ok := route["polyline"].(string); ok {
			output.Route = &models.Route{Polyline: polyline}
		}
	}

	return output, nil
}

type RecomputeTripOutput struct {
	Itinerary []models.TripPlace
	Route     *models.Route
}

func (u *tripUsecase) RecomputeTrip(userID, tripID string, orderedPlaceIDs []string, stayMinutesMap map[string]int) (*RecomputeTripOutput, error) {
	// 旅程を取得
	trip, err := u.tripRepo.GetTrip(tripID)
	if err != nil {
		return nil, fmt.Errorf("旅程が見つかりません: %w", err)
	}

	if trip.UserID != userID {
		return nil, fmt.Errorf("この旅程へのアクセス権限がありません")
	}

	// スポットを順序通りに取得
	places, err := u.tripRepo.GetTripPlacesByIDs(tripID, orderedPlaceIDs)
	if err != nil {
		return nil, fmt.Errorf("スポットの取得に失敗しました: %w", err)
	}

	// 滞在時間を更新
	for i, place := range places {
		place.OrderIndex = i
		if minutes, ok := stayMinutesMap[place.ID]; ok {
			place.StayMinutes = minutes
		}
		if err := u.tripRepo.UpdateTripPlace(place); err != nil {
			return nil, fmt.Errorf("スポットの更新に失敗しました: %w", err)
		}
	}

	// MLサービスで再計算
	waypoints := []map[string]float64{}
	for _, place := range places {
		waypoints = append(waypoints, map[string]float64{
			"lat": place.Lat,
			"lng": place.Lng,
		})
	}

	mlResp, err := u.mlService.RecomputeRoute(waypoints)
	if err != nil {
		return nil, fmt.Errorf("ルート再計算エラー: %w", err)
	}

	// タイムライン計算
	itinerary := calculateTimeline(places, trip.StartTime, mlResp)

	output := &RecomputeTripOutput{
		Itinerary: itinerary,
	}

	if route, ok := mlResp["route"].(map[string]interface{}); ok {
		if polyline, ok := route["polyline"].(string); ok {
			output.Route = &models.Route{Polyline: polyline}
		}
	}

	return output, nil
}

type GetShareOutput struct {
	Trip      *models.Trip
	Itinerary []models.TripPlace
	Route     *models.Route
}

func (u *tripUsecase) GetShare(shareID string) (*GetShareOutput, error) {
	trip, err := u.tripRepo.GetTripByShareID(shareID)
	if err != nil {
		return nil, fmt.Errorf("共有された旅程が見つかりません: %w", err)
	}

	places, err := u.tripRepo.GetTripPlaces(trip.ID)
	if err != nil {
		return nil, fmt.Errorf("スポットの取得に失敗しました: %w", err)
	}

	itinerary := convertItineraryFromPlaces(places, trip.StartTime)

	// ルートを再計算
	waypoints := []map[string]float64{}
	for _, place := range places {
		waypoints = append(waypoints, map[string]float64{
			"lat": place.Lat,
			"lng": place.Lng,
		})
	}

	var route *models.Route
	if len(waypoints) > 0 {
		mlResp, err := u.mlService.RecomputeRoute(waypoints)
		if err == nil {
			if routeData, ok := mlResp["route"].(map[string]interface{}); ok {
				if polyline, ok := routeData["polyline"].(string); ok {
					route = &models.Route{Polyline: polyline}
				}
			}
		}
	}

	return &GetShareOutput{
		Trip:      trip,
		Itinerary: itinerary,
		Route:     route,
	}, nil
}

