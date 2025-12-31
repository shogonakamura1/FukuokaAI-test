package controllers

import (
	"net/http"
	"strings"

	"fukuoka-ai-api/models"
	"fukuoka-ai-api/usecase"

	"github.com/gin-gonic/gin"
)

type TripController struct {
	tripUsecase usecase.TripUsecase
}

func NewTripController(tripUsecase usecase.TripUsecase) *TripController {
	return &TripController{
		tripUsecase: tripUsecase,
	}
}

func (c *TripController) getUserID(ctx *gin.Context) string {
	// MVP: 簡易的なuser_idヘッダから取得
	// 本番ではIDトークン検証が必要
	return ctx.GetHeader("X-User-Id")
}

type CreateTripRequest struct {
	MustPlaces   []string `json:"must_places"`
	InterestTags []string `json:"interest_tags"`
	FreeText     string   `json:"free_text,omitempty"`
}

type CreateTripResponse struct {
	TripID     string            `json:"trip_id"`
	ShareID    string            `json:"share_id"`
	Itinerary  []models.TripPlace `json:"itinerary"`
	Candidates []models.TripPlace `json:"candidates"`
	Route      *models.Route      `json:"route,omitempty"`
}

func (c *TripController) CreateTrip(ctx *gin.Context) {
	userID := c.getUserID(ctx)
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{
			"code":    "UNAUTHORIZED",
			"message": "ユーザーIDが必要です",
		}})
		return
	}

	var req CreateTripRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		}})
		return
	}

	output, err := c.tripUsecase.CreateTrip(userID, req.MustPlaces, req.InterestTags, req.FreeText)
	if err != nil {
		// エラーメッセージからエラーコードを判定（簡易実装）
		statusCode := http.StatusInternalServerError
		errorCode := "INTERNAL_ERROR"
		message := err.Error()

		if contains(err.Error(), "見つかりません") {
			statusCode = http.StatusNotFound
			errorCode = "NOT_FOUND"
		} else if contains(err.Error(), "権限") {
			statusCode = http.StatusForbidden
			errorCode = "FORBIDDEN"
		} else if contains(err.Error(), "推薦サービス") {
			errorCode = "ML_SERVICE_ERROR"
		} else if contains(err.Error(), "保存") || contains(err.Error(), "取得") {
			errorCode = "DATABASE_ERROR"
		}

		ctx.JSON(statusCode, gin.H{"error": gin.H{
			"code":    errorCode,
			"message": message,
		}})
		return
	}

	response := CreateTripResponse{
		TripID:     output.TripID,
		ShareID:    output.ShareID,
		Itinerary:  output.Itinerary,
		Candidates: output.Candidates,
		Route:      output.Route,
	}

	ctx.JSON(http.StatusOK, response)
}

type RecomputeTripRequest struct {
	OrderedPlaceIDs []string       `json:"ordered_place_ids"`
	StayMinutesMap  map[string]int `json:"stay_minutes_map,omitempty"`
}

func (c *TripController) RecomputeTrip(ctx *gin.Context) {
	tripID := ctx.Param("trip_id")
	userID := c.getUserID(ctx)

	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": gin.H{
			"code":    "UNAUTHORIZED",
			"message": "ユーザーIDが必要です",
		}})
		return
	}

	var req RecomputeTripRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": err.Error(),
		}})
		return
	}

	output, err := c.tripUsecase.RecomputeTrip(userID, tripID, req.OrderedPlaceIDs, req.StayMinutesMap)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "INTERNAL_ERROR"
		message := err.Error()

		if contains(err.Error(), "見つかりません") {
			statusCode = http.StatusNotFound
			errorCode = "NOT_FOUND"
		} else if contains(err.Error(), "権限") {
			statusCode = http.StatusForbidden
			errorCode = "FORBIDDEN"
		} else if contains(err.Error(), "ルート再計算") {
			errorCode = "ML_SERVICE_ERROR"
		} else if contains(err.Error(), "保存") || contains(err.Error(), "取得") || contains(err.Error(), "更新") {
			errorCode = "DATABASE_ERROR"
		}

		ctx.JSON(statusCode, gin.H{"error": gin.H{
			"code":    errorCode,
			"message": message,
		}})
		return
	}

	response := gin.H{
		"itinerary": output.Itinerary,
	}

	if output.Route != nil {
		response["route"] = gin.H{"polyline": output.Route.Polyline}
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *TripController) GetShare(ctx *gin.Context) {
	shareID := ctx.Param("share_id")

	output, err := c.tripUsecase.GetShare(shareID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "INTERNAL_ERROR"
		message := err.Error()

		if contains(err.Error(), "見つかりません") {
			statusCode = http.StatusNotFound
			errorCode = "NOT_FOUND"
		} else if contains(err.Error(), "取得") {
			errorCode = "DATABASE_ERROR"
		}

		ctx.JSON(statusCode, gin.H{"error": gin.H{
			"code":    errorCode,
			"message": message,
		}})
		return
	}

	response := gin.H{
		"trip": gin.H{
			"id":         output.Trip.ID,
			"title":      output.Trip.Title,
			"start_time": output.Trip.StartTime,
		},
		"itinerary": output.Itinerary,
	}

	if output.Route != nil {
		response["route"] = gin.H{"polyline": output.Route.Polyline}
	}

	ctx.JSON(http.StatusOK, response)
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

