package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// IPlaceDetailsService 場所詳細サービスのインターフェース
type IPlaceDetailsService interface {
	GetPlaceDetails(placeID string, photoReference string) (*PlaceDetails, error)
}

// PlaceDetails 場所の詳細情報
type PlaceDetails struct {
	PlaceID       string  `json:"place_id"`
	Name          string  `json:"name"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	PhotoURL      string  `json:"photo_url"`
	Rating        float64 `json:"rating"`
	ReviewSummary string  `json:"review_summary"`
	Address       string  `json:"address"`
	Category      string  `json:"category"`
}

// PlaceDetailsService Google Places Place Details APIを使用した詳細取得サービス
type PlaceDetailsService struct {
	apiKey string
	client *http.Client
}

// NewPlaceDetailsService 新しいPlaceDetailsServiceを作成
func NewPlaceDetailsService() IPlaceDetailsService {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	return &PlaceDetailsService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// PlaceDetailsResponse Google Places Place Details APIのレスポンス
type PlaceDetailsResponse struct {
	Status  string `json:"status"`
	Result  struct {
		PlaceID       string   `json:"place_id"`
		Name          string   `json:"name"`
		Rating        float64  `json:"rating,omitempty"`
		FormattedAddress string `json:"formatted_address,omitempty"`
		Geometry      struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		Types []string `json:"types,omitempty"`
		Reviews []struct {
			Text string `json:"text"`
			Rating int `json:"rating"`
		} `json:"reviews,omitempty"`
	} `json:"result"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// GetPlaceDetails 場所の詳細情報を取得
func (s *PlaceDetailsService) GetPlaceDetails(placeID string, photoReference string) (*PlaceDetails, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_MAPS_API_KEY is not set")
	}

	baseURL := "https://maps.googleapis.com/maps/api/place/details/json"
	params := url.Values{}
	params.Add("place_id", placeID)
	params.Add("key", s.apiKey)
	params.Add("language", "ja")
	params.Add("fields", "place_id,name,rating,formatted_address,geometry,types,reviews")

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := s.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google Places API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result PlaceDetailsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "OK" {
		return nil, fmt.Errorf("Google Places API error: %s - %s", result.Status, result.ErrorMessage)
	}

	details := &PlaceDetails{
		PlaceID: result.Result.PlaceID,
		Name:    result.Result.Name,
		Lat:     result.Result.Geometry.Location.Lat,
		Lng:     result.Result.Geometry.Location.Lng,
		Rating:  result.Result.Rating,
		Address: result.Result.FormattedAddress,
	}

	// カテゴリを取得（最初のタイプを使用）
	if len(result.Result.Types) > 0 {
		details.Category = result.Result.Types[0]
	}

	// レビュー要約（最初のレビューの最初の100文字）
	if len(result.Result.Reviews) > 0 {
		reviewText := result.Result.Reviews[0].Text
		if len(reviewText) > 100 {
			reviewText = reviewText[:100] + "..."
		}
		details.ReviewSummary = reviewText
	}

	// 写真URLを生成
	if photoReference != "" {
		photoURL := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/photo?maxwidth=400&photoreference=%s&key=%s",
			photoReference, s.apiKey)
		details.PhotoURL = photoURL
	}

	return details, nil
}

