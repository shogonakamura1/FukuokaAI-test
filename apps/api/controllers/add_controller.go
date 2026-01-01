package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddController 場所追加機能のコントローラー
type AddController struct {
}

// NewAddController 新しいAddControllerを作成
func NewAddController() *AddController {
	return &AddController{}
}

// AddPlace リコメンド場所をリストに追加するエンドポイント
// 注: APIドキュメントでは「入力なし、出力なし」となっているため、単純に200 OKを返す
func (c *AddController) AddPlace(ctx *gin.Context) {
	placeID := ctx.Param("place_id")
	if placeID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": gin.H{
			"code":    "INVALID_REQUEST",
			"message": "place_idが指定されていません",
		}})
		return
	}

	// ここでは単純に200 OKを返す
	// 実際のリスト管理はクライアント側で行う想定
	ctx.JSON(http.StatusOK, gin.H{
		"message": "場所が追加されました",
		"place_id": placeID,
	})
}

