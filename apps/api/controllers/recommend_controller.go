package controllers

import (
	"fukuoka-ai-api/models"
	"fukuoka-ai-api/usecase"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RecommendController リコメンド機能のコントローラー
type RecommendController struct {
	recommendUsecase usecase.IRecommendUsecase
}

// NewRecommendController 新しいRecommendControllerを作成
func NewRecommendController(recommendUsecase usecase.IRecommendUsecase) *RecommendController {
	return &RecommendController{
		recommendUsecase: recommendUsecase,
	}
}

// Recommend リコメンド機能のエンドポイント
func (c *RecommendController) Recommend(ctx *gin.Context) {
	var req models.RecommendRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": "リクエストの形式が不正です: " + err.Error(),
		}})
		return
	}

	// バリデーション
	if len(req.MustPlaces) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": "寄りたい場所が指定されていません",
		}})
		return
	}

	if len(req.InterestTags) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": "興味タグが指定されていません",
		}})
		return
	}

	// ユースケースを呼び出し
	response, err := c.recommendUsecase.Recommend(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "INTERNAL_ERROR"
		message := err.Error()

		// エラーメッセージからエラーコードを判定
		errMsg := err.Error()
		if strings.Contains(errMsg, "座標取得") {
			statusCode = http.StatusBadRequest
			errorCode = "GEOCODING_ERROR"
		} else if strings.Contains(errMsg, "Google Places API") {
			errorCode = "PLACES_API_ERROR"
		} else if strings.Contains(errMsg, "APIキー") || strings.Contains(errMsg, "API_KEY") {
			statusCode = http.StatusInternalServerError
			errorCode = "CONFIGURATION_ERROR"
		}

		ctx.JSON(statusCode, gin.H{"error": gin.H{
			"code":    errorCode,
			"message": message,
		}})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

