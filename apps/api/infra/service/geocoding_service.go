package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// IGeocodingService ジオコーディングサービスのインターフェース
type IGeocodingService interface {
	GetCoordinates(placeName string) (lat, lng float64, placeID string, err error)
}

// GeocodingService Google Places Text Search APIを使用したジオコーディングサービス
type GeocodingService struct {
	apiKey string
	client *http.Client
}

// NewGeocodingService 新しいGeocodingServiceを作成
func NewGeocodingService() IGeocodingService {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		// エラーを返すか、デフォルト値を設定するかは要件による
		// ここでは空文字列のまま進める（呼び出し時にエラーになる）
	}

	return &GeocodingService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// TextSearchResponse Google Places Text Search APIのレスポンス
type TextSearchResponse struct {
	Status  string `json:"status"`
	Results []struct {
		PlaceID  string `json:"place_id"`
		Name     string `json:"name"`
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"results"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// GetCoordinates 場所名から座標を取得
func (s *GeocodingService) GetCoordinates(placeName string) (lat, lng float64, placeID string, err error) {
	if s.apiKey == "" {
		return 0, 0, "", fmt.Errorf("GOOGLE_MAPS_API_KEY is not set")
	}

	// 博多駅のデフォルト値
	if placeName == "" || placeName == "Hakata Station" || placeName == "博多駅" {
		return 33.5904, 130.4208, "", nil
	}

	baseURL := "https://maps.googleapis.com/maps/api/place/textsearch/json"
	params := url.Values{}
	params.Add("query", fmt.Sprintf("%s 福岡", placeName))
	params.Add("key", s.apiKey)
	params.Add("language", "ja")

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := s.client.Get(reqURL)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to call Google Places API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to read response: %w", err)
	}

	var result TextSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, 0, "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "OK" {
		return 0, 0, "", fmt.Errorf("Google Places API error: %s - %s", result.Status, result.ErrorMessage)
	}

	if len(result.Results) == 0 {
		return 0, 0, "", fmt.Errorf("place not found: %s", placeName)
	}

	firstResult := result.Results[0]
	return firstResult.Geometry.Location.Lat, firstResult.Geometry.Location.Lng, firstResult.PlaceID, nil
}

