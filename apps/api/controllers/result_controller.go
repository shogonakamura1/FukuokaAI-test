package controllers

import (
	"fukuoka-ai-api/models"
	"fukuoka-ai-api/usecase"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ResultController ルート提案機能のコントローラー
type ResultController struct {
	resultUsecase usecase.IResultUsecase
}

// NewResultController 新しいResultControllerを作成
func NewResultController(resultUsecase usecase.IResultUsecase) *ResultController {
	return &ResultController{
		resultUsecase: resultUsecase,
	}
}

// Result リコメンド追加してルートを提案するエンドポイント
func (c *ResultController) Result(ctx *gin.Context) {
	var req models.ResultRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": "リクエストの形式が不正です: " + err.Error(),
		}})
		return
	}

	// バリデーション
	if len(req.Places) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": "場所リストが指定されていません",
		}})
		return
	}

	// ユースケースを呼び出し
	response, err := c.resultUsecase.ComputeOptimizedRoute(&req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "INTERNAL_ERROR"
		message := err.Error()

		// エラーメッセージからエラーコードを判定
		if strings.Contains(message, "場所リストが空") {
			statusCode = http.StatusBadRequest
			errorCode = "INVALID_REQUEST"
		} else if strings.Contains(message, "有効な場所") {
			statusCode = http.StatusBadRequest
			errorCode = "INVALID_REQUEST"
		} else if strings.Contains(message, "ルート計算") || strings.Contains(message, "Routes API") {
			errorCode = "ROUTES_API_ERROR"
		}

		ctx.JSON(statusCode, gin.H{"error": gin.H{
			"code":    errorCode,
			"message": message,
		}})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

