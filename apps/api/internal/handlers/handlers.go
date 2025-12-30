package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"fukuoka-ai-api/internal/models"
	"fukuoka-ai-api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	db          *sql.DB
	mlServiceURL string
	repo        *repository.Repository
}

func NewHandler(db *sql.DB, mlServiceURL string) *Handler {
	return &Handler{
		db:          db,
		mlServiceURL: mlServiceURL,
		repo:        repository.NewRepository(db),
	}
}

func (h *Handler) getUserID(c *gin.Context) string {
	// MVP: 簡易的なuser_idヘッダから取得
	// 本番ではIDトークン検証が必要
	return c.GetHeader("X-User-Id")
}

type CreateTripRequest struct {
	MustPlaces   []string `json:"must_places"`
	InterestTags []string `json:"interest_tags"`
	FreeText     string   `json:"free_text,omitempty"`
}

type CreateTripResponse struct {
	TripID     string           `json:"trip_id"`
	ShareID    string           `json:"share_id"`
	Itinerary  []models.Place   `json:"itinerary"`
	Candidates []models.Place   `json:"candidates"`
	Route      *models.Route    `json:"route,omitempty"`
}

func (h *Handler) CreateTrip(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{
			"code":    "UNAUTHORIZED",
			"message": "ユーザーIDが必要です",
		}})
		return
	}

	var req CreateTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		}})
		return
	}

	// ユーザーが存在しない場合は作成
	if err := h.repo.EnsureUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "ユーザー作成に失敗しました",
		}})
		return
	}

	// Python MLサービスに推薦依頼
	mlReq := map[string]interface{}{
		"start":         "Hakata Station",
		"must_places":   req.MustPlaces,
		"interest_tags": req.InterestTags,
		"free_text":     req.FreeText,
	}

	mlResp, err := h.callMLService("/recommend", mlReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "ML_SERVICE_ERROR",
			"message": fmt.Sprintf("推薦サービスエラー: %v", err),
		}})
		return
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

	if err := h.repo.CreateTrip(trip); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "DATABASE_ERROR",
			"message": "旅程の保存に失敗しました",
		}})
		return
	}

	// スポットを保存
	itineraryData, ok := mlResp["initial_itinerary"].([]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "ML_SERVICE_ERROR",
			"message": "初期旅程データの形式が不正です",
		}})
		return
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
		if err := h.repo.CreateTripPlace(place); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "スポットの保存に失敗しました",
			}})
			return
		}
	}

	// 共有情報を保存
	if err := h.repo.CreateShare(shareID, tripID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "DATABASE_ERROR",
			"message": "共有情報の保存に失敗しました",
		}})
		return
	}

	// レスポンス作成
	candidates := []models.Place{}
	if cands, ok := mlResp["candidates"].([]interface{}); ok {
		for _, item := range cands {
			placeData, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			candidates = append(candidates, models.Place{
				PlaceID:      getString(placeData, "place_id"),
				Name:         getString(placeData, "name"),
				Lat:          getFloat64(placeData, "lat"),
				Lng:          getFloat64(placeData, "lng"),
				Category:     getString(placeData, "category"),
				Reason:       getString(placeData, "reason"),
				ReviewSummary: getString(placeData, "review_summary"),
				PhotoURL:     getString(placeData, "photo_url"),
			})
		}
	}

	response := CreateTripResponse{
		TripID:     tripID,
		ShareID:    shareID,
		Itinerary:  h.convertItinerary(itineraryData),
		Candidates: candidates,
	}

	if route, ok := mlResp["route"].(map[string]interface{}); ok {
		if polyline, ok := route["polyline"].(string); ok {
			response.Route = &models.Route{Polyline: polyline}
		}
	}

	c.JSON(http.StatusOK, response)
}

type RecomputeTripRequest struct {
	OrderedPlaceIDs []string          `json:"ordered_place_ids"`
	StayMinutesMap  map[string]int    `json:"stay_minutes_map,omitempty"`
}

func (h *Handler) RecomputeTrip(c *gin.Context) {
	tripID := c.Param("trip_id")
	userID := h.getUserID(c)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{
			"code":    "UNAUTHORIZED",
			"message": "ユーザーIDが必要です",
		}})
		return
	}

	var req RecomputeTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		}})
		return
	}

	// 旅程を取得
	trip, err := h.repo.GetTrip(tripID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code":    "NOT_FOUND",
			"message": "旅程が見つかりません",
		}})
		return
	}

	if trip.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": gin.H{
			"code":    "FORBIDDEN",
			"message": "この旅程へのアクセス権限がありません",
		}})
		return
	}

	// スポットを順序通りに取得
	places, err := h.repo.GetTripPlacesByIDs(tripID, req.OrderedPlaceIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "DATABASE_ERROR",
			"message": "スポットの取得に失敗しました",
		}})
		return
	}

	// 滞在時間を更新
	for i, place := range places {
		place.OrderIndex = i
		if minutes, ok := req.StayMinutesMap[place.ID]; ok {
			place.StayMinutes = minutes
		}
		if err := h.repo.UpdateTripPlace(place); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "スポットの更新に失敗しました",
			}})
			return
		}
	}

	// Python MLサービスで再計算
	waypoints := []map[string]float64{}
	for _, place := range places {
		waypoints = append(waypoints, map[string]float64{
			"lat": place.Lat,
			"lng": place.Lng,
		})
	}

	mlReq := map[string]interface{}{
		"start":    "Hakata Station",
		"waypoints": waypoints,
	}

	mlResp, err := h.callMLService("/recompute-route", mlReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "ML_SERVICE_ERROR",
			"message": fmt.Sprintf("ルート再計算エラー: %v", err),
		}})
		return
	}

	// タイムライン計算
	itinerary := h.calculateTimeline(places, trip.StartTime, mlResp)

	response := gin.H{
		"itinerary": itinerary,
	}

	if route, ok := mlResp["route"].(map[string]interface{}); ok {
		if polyline, ok := route["polyline"].(string); ok {
			response["route"] = gin.H{"polyline": polyline}
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetShare(c *gin.Context) {
	shareID := c.Param("share_id")

	trip, err := h.repo.GetTripByShareID(shareID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code":    "NOT_FOUND",
			"message": "共有された旅程が見つかりません",
		}})
		return
	}

	places, err := h.repo.GetTripPlaces(trip.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "DATABASE_ERROR",
			"message": "スポットの取得に失敗しました",
		}})
		return
	}

	itinerary := h.convertItineraryFromPlaces(places, trip.StartTime)

	// ルートを再計算
	waypoints := []map[string]float64{}
	for _, place := range places {
		waypoints = append(waypoints, map[string]float64{
			"lat": place.Lat,
			"lng": place.Lng,
		})
	}

	var routePolyline string
	if len(waypoints) > 0 {
		mlReq := map[string]interface{}{
			"start":    "Hakata Station",
			"waypoints": waypoints,
		}
		mlResp, err := h.callMLService("/recompute-route", mlReq)
		if err == nil {
			if route, ok := mlResp["route"].(map[string]interface{}); ok {
				if polyline, ok := route["polyline"].(string); ok {
					routePolyline = polyline
				}
			}
		}
	}

	response := gin.H{
		"trip": gin.H{
			"id":         trip.ID,
			"title":      trip.Title,
			"start_time": trip.StartTime,
		},
		"itinerary": itinerary,
	}

	if routePolyline != "" {
		response["route"] = gin.H{"polyline": routePolyline}
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) callMLService(endpoint string, data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	url := h.mlServiceURL + endpoint
	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (h *Handler) convertItinerary(data []interface{}) []models.Place {
	result := []models.Place{}
	for _, item := range data {
		placeData, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		result = append(result, models.Place{
			ID:           getString(placeData, "id"),
			PlaceID:      getString(placeData, "place_id"),
			Name:         getString(placeData, "name"),
			Lat:          getFloat64(placeData, "lat"),
			Lng:          getFloat64(placeData, "lng"),
			Kind:         getString(placeData, "kind"),
			StayMinutes:  getInt(placeData, "stay_minutes"),
			OrderIndex:   getInt(placeData, "order_index"),
			TimeRange:    getString(placeData, "time_range"),
			Reason:       getString(placeData, "reason"),
			ReviewSummary: getString(placeData, "review_summary"),
			PhotoURL:     getString(placeData, "photo_url"),
		})
	}
	return result
}

func (h *Handler) convertItineraryFromPlaces(places []*models.TripPlace, startTime string) []models.Place {
	result := []models.Place{}
	currentTime := parseTime(startTime)

	for _, place := range places {
		timeRange := formatTimeRange(currentTime, place.StayMinutes)
		result = append(result, models.Place{
			ID:           place.ID,
			PlaceID:      place.PlaceID,
			Name:         place.Name,
			Lat:          place.Lat,
			Lng:          place.Lng,
			Kind:         place.Kind,
			StayMinutes:  place.StayMinutes,
			OrderIndex:   place.OrderIndex,
			TimeRange:    timeRange,
			Reason:       place.Reason,
			ReviewSummary: place.ReviewSummary,
			PhotoURL:     place.PhotoURL,
		})
		currentTime = addMinutes(currentTime, place.StayMinutes)
	}

	return result
}

func (h *Handler) calculateTimeline(places []*models.TripPlace, startTime string, mlResp map[string]interface{}) []models.Place {
	// 簡易実装: 移動時間は30分固定（MVP）
	// 実際にはDirections APIのlegsから取得すべき
	result := []models.Place{}
	currentTime := parseTime(startTime)

	for i, place := range places {
		if i > 0 {
			// 移動時間（30分固定、MVP）
			currentTime = addMinutes(currentTime, 30)
		}

		timeRange := formatTimeRange(currentTime, place.StayMinutes)
		result = append(result, models.Place{
			ID:           place.ID,
			PlaceID:      place.PlaceID,
			Name:         place.Name,
			Lat:          place.Lat,
			Lng:          place.Lng,
			Kind:         place.Kind,
			StayMinutes:  place.StayMinutes,
			OrderIndex:   place.OrderIndex,
			TimeRange:    timeRange,
			Reason:       place.Reason,
			ReviewSummary: place.ReviewSummary,
			PhotoURL:     place.PhotoURL,
		})

		currentTime = addMinutes(currentTime, place.StayMinutes)
	}

	return result
}

// ヘルパー関数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0.0
}

func parseTime(timeStr string) int {
	// "10:00" -> 600 (分)
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 600 // デフォルト10:00
	}
	hours := 0
	minutes := 0
	fmt.Sscanf(parts[0], "%d", &hours)
	fmt.Sscanf(parts[1], "%d", &minutes)
	return hours*60 + minutes
}

func addMinutes(timeMinutes int, minutes int) int {
	return timeMinutes + minutes
}

func formatTimeRange(startMinutes int, durationMinutes int) string {
	startHour := startMinutes / 60
	startMin := startMinutes % 60
	endMinutes := startMinutes + durationMinutes
	endHour := endMinutes / 60
	endMin := endMinutes % 60
	return fmt.Sprintf("%02d:%02d-%02d:%02d", startHour, startMin, endHour, endMin)
}

