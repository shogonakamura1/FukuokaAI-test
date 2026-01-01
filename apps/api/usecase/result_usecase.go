package usecase

import (
	"fmt"
	"fukuoka-ai-api/infra/service"
	"fukuoka-ai-api/models"
	"time"
)

// IResultUsecase ルート提案機能のユースケースインターフェース
type IResultUsecase interface {
	ComputeOptimizedRoute(req *models.ResultRequest) (*models.ResultResponse, error)
}

// ResultUsecase ルート提案機能のユースケース実装
type ResultUsecase struct {
	geocodingService service.IGeocodingService
	placeDetailsService service.IPlaceDetailsService
	routeService     service.IRouteService
}

// NewResultUsecase 新しいResultUsecaseを作成
func NewResultUsecase(
	geocodingService service.IGeocodingService,
	placeDetailsService service.IPlaceDetailsService,
	routeService service.IRouteService,
) IResultUsecase {
	return &ResultUsecase{
		geocodingService:    geocodingService,
		placeDetailsService: placeDetailsService,
		routeService:        routeService,
	}
}

// ComputeOptimizedRoute ルート提案機能のメイン処理
func (u *ResultUsecase) ComputeOptimizedRoute(req *models.ResultRequest) (*models.ResultResponse, error) {
	// 1. 行きたい場所リストから座標を取得
	if len(req.Places) == 0 {
		return nil, fmt.Errorf("場所リストが空です")
	}

	// Place Details APIで各場所の詳細情報を取得
	var places []models.Place
	var waypoints []service.Waypoint

	for i, placeID := range req.Places {
		details, err := u.placeDetailsService.GetPlaceDetails(placeID, "")
		if err != nil {
			// 詳細取得に失敗した場所はエラーメッセージに含める
			return nil, fmt.Errorf("場所[%d] (place_id: %s) の詳細取得に失敗しました: %w", i, placeID, err)
		}

		places = append(places, models.Place{
			PlaceID:  details.PlaceID,
			Name:     details.Name,
			Lat:      details.Lat,
			Lng:      details.Lng,
			Rating:   details.Rating,
			Address:  details.Address,
			PhotoURL: details.PhotoURL,
		})

		waypoints = append(waypoints, service.Waypoint{
			PlaceID: placeID,
			Lat:     details.Lat,
			Lng:     details.Lng,
		})
	}

	if len(places) == 0 {
		return nil, fmt.Errorf("有効な場所が見つかりませんでした")
	}

	// 出発地点とゴール地点を設定
	// 最初の場所を出発地点、最後の場所をゴール地点とする
	originLat := places[0].Lat
	originLng := places[0].Lng
	destinationLat := places[len(places)-1].Lat
	destinationLng := places[len(places)-1].Lng

	// 経由地点は中間の場所（最初と最後を除く）
	var intermediates []service.Waypoint
	if len(waypoints) > 2 {
		intermediates = waypoints[1 : len(waypoints)-1]
	}

	// 2. Routes APIでルートを計算（経由地順最適化を有効にする）
	departureTime := time.Now().Add(1 * time.Hour) // 1時間後をデフォルトとする
	routeResp, err := u.routeService.ComputeRoute(
		originLat, originLng,
		destinationLat, destinationLng,
		intermediates,
		"DRIVE",
		&departureTime,
	)
	if err != nil {
		return nil, fmt.Errorf("ルート計算に失敗しました: %w", err)
	}

	if len(routeResp.Routes) == 0 {
		return nil, fmt.Errorf("ルートが見つかりませんでした")
	}

	routeData := routeResp.Routes[0]

	// 3. 最適化された順序に従って場所を並び替え
	// OptimizedIntermediateWaypointIndexは経由地点（intermediates）の順序のみを表す
	// origin（最初）とdestination（最後）は順序に含まれない
	optimizedPlaces := make([]models.Place, 0, len(places))
	
	// 最初にorigin（最初の場所）を追加
	optimizedPlaces = append(optimizedPlaces, places[0])
	
	if len(routeData.OptimizedIntermediateWaypointIndex) > 0 && len(intermediates) > 0 {
		// 最適化された順序を使用して経由地点を追加
		// OptimizedIntermediateWaypointIndexは経由地点の元のインデックス（リクエスト時の順序）の配列
		// この配列の順序が最適化された順序を表す
		// 例: OptimizedIntermediateWaypointIndex = [2, 0, 1] の場合、元の2番目、0番目、1番目の順で訪問
		for _, originalIntermediateIdx := range routeData.OptimizedIntermediateWaypointIndex {
			// intermediatesのインデックスをplacesのインデックスに変換（+1はoriginの分）
			placeIdx := originalIntermediateIdx + 1
			if placeIdx > 0 && placeIdx < len(places)-1 { // destinationより前
				optimizedPlaces = append(optimizedPlaces, places[placeIdx])
			}
		}
	} else if len(intermediates) > 0 {
		// 最適化されていない場合は元の順序で経由地点を追加
		for i := 1; i < len(places)-1; i++ {
			optimizedPlaces = append(optimizedPlaces, places[i])
		}
	}
	
	// 最後にdestination（最後の場所）を追加
	if len(places) > 1 {
		optimizedPlaces = append(optimizedPlaces, places[len(places)-1])
	}

	// 4. ルート情報を構築
	var routeLegs []models.RouteLeg
	for i, leg := range routeData.Legs {
		// デバッグ: 距離情報を確認
		fmt.Printf("Leg %d: DistanceMeters=%d, Duration=%s\n", i, leg.DistanceMeters, leg.Duration)
		routeLegs = append(routeLegs, models.RouteLeg{
			StartLocation: models.Coordinate{
				Lat: leg.StartLocation.LatLng.Latitude,
				Lng: leg.StartLocation.LatLng.Longitude,
			},
			EndLocation: models.Coordinate{
				Lat: leg.EndLocation.LatLng.Latitude,
				Lng: leg.EndLocation.LatLng.Longitude,
			},
			DistanceMeters: leg.DistanceMeters,
			Duration:       leg.Duration,
		})
	}

	// Routes API v2では、OptimizedIntermediateWaypointIndexから最適化された順序を取得
	optimizedOrder := routeData.OptimizedIntermediateWaypointIndex
	if len(optimizedOrder) == 0 && len(intermediates) > 0 {
		// 最適化されていない場合は経由地点の元の順序（0から始まる連番）
		optimizedOrder = make([]int, len(intermediates))
		for i := range optimizedOrder {
			optimizedOrder[i] = i
		}
	}

	route := models.Route{
		Legs:           routeLegs,
		DistanceMeters: routeData.DistanceMeters,
		Duration:       routeData.Duration,
		OptimizedOrder: optimizedOrder,
	}

	return &models.ResultResponse{
		Places: optimizedPlaces,
		Route:  route,
	}, nil
}

