package handlers

import (
	"fmt"
	"net/http"

	"fukuoka-ai-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateTripRequest struct {
	MustPlaces   []string `json:"must_places"`
	InterestTags []string `json:"interest_tags"`
	FreeText     string   `json:"free_text,omitempty"`
}

type CreateTripResponse struct {
	TripID     string             `json:"trip_id"`
	ShareID    string             `json:"share_id"`
	Itinerary  []models.TripPlace `json:"itinerary"`
	Candidates []models.TripPlace `json:"candidates"`
	Route      *models.Route      `json:"route,omitempty"`
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
	// DBから保存したplacesを取得してTimeRangeを計算
	savedPlaces, err := h.repo.GetTripPlaces(tripID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "DATABASE_ERROR",
			"message": "スポットの取得に失敗しました",
		}})
		return
	}
	itinerary := h.convertItineraryFromPlaces(savedPlaces, trip.StartTime)

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
				// TripIDは空（candidatesはDBに保存しない）
			})
		}
	}

	response := CreateTripResponse{
		TripID:     tripID,
		ShareID:    shareID,
		Itinerary:  itinerary,
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
	OrderedPlaceIDs []string       `json:"ordered_place_ids"`
	StayMinutesMap  map[string]int `json:"stay_minutes_map,omitempty"`
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
		"start":     "Hakata Station",
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

