package controllers

import (
	"fukuoka-ai-api/infra/service"
	"fukuoka-ai-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GeocodingController ジオコーディング機能のコントローラー
type GeocodingController struct {
	geocodingService service.IGeocodingService
}

// NewGeocodingController 新しいGeocodingControllerを作成
func NewGeocodingController(geocodingService service.IGeocodingService) *GeocodingController {
	return &GeocodingController{
		geocodingService: geocodingService,
	}
}

// GetPlaceID 場所名からplace_idを取得するエンドポイント
func (c *GeocodingController) GetPlaceID(ctx *gin.Context) {
	var req models.GeocodingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": "リクエストの形式が不正です: " + err.Error(),
		}})
		return
	}

	// バリデーション
	if req.PlaceName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": "場所名が指定されていません",
		}})
		return
	}

	// ジオコーディングサービスを呼び出し
	lat, lng, placeID, err := c.geocodingService.GetCoordinates(req.PlaceName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "GEOCODING_ERROR",
			"message": "場所の座標取得に失敗しました: " + err.Error(),
		}})
		return
	}

	// place_idが空の場合はエラーを返す
	if placeID == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "GEOCODING_ERROR",
			"message": "place_idを取得できませんでした: " + req.PlaceName,
		}})
		return
	}

	response := models.GeocodingResponse{
		PlaceID: placeID,
		Lat:     lat,
		Lng:     lng,
		Name:    req.PlaceName,
	}

	ctx.JSON(http.StatusOK, response)
}

