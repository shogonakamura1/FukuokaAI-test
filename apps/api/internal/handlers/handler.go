package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"fukuoka-ai-api/internal/repository"

	"github.com/gin-gonic/gin"
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

