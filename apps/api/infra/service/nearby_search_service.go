package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// INearbySearchService 周辺検索サービスのインターフェース
type INearbySearchService interface {
	SearchNearby(lat, lng, radius float64, interestTags []string) ([]PlaceResult, error)
}

// NearbySearchService Google Places Nearby Search APIを使用した周辺検索サービス
type NearbySearchService struct {
	apiKey string
	client *http.Client
}

// PlaceResult 周辺検索の結果
type PlaceResult struct {
	PlaceID       string  `json:"place_id"`
	Name          string  `json:"name"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	Rating        float64 `json:"rating"`
	PhotoReference string `json:"photo_reference,omitempty"`
	Types         []string `json:"types,omitempty"`
}

// NewNearbySearchService 新しいNearbySearchServiceを作成
func NewNearbySearchService() INearbySearchService {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	return &NearbySearchService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// NearbySearchResponse Google Places Nearby Search APIのレスポンス
type NearbySearchResponse struct {
	Status  string `json:"status"`
	Results []struct {
		PlaceID       string   `json:"place_id"`
		Name          string   `json:"name"`
		Rating        float64  `json:"rating,omitempty"`
		UserRatingsTotal int   `json:"user_ratings_total,omitempty"`
		Geometry      struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		Photos []struct {
			PhotoReference string `json:"photo_reference"`
		} `json:"photos,omitempty"`
		Types []string `json:"types,omitempty"`
	} `json:"results"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// tagToKeyword 興味タグをGoogle Places APIのキーワード/タイプにマッピング
func tagToKeyword(tag string) (keyword string, placeType string) {
	keywordMap := map[string]string{
		"カフェ":     "cafe",
		"レストラン":  "restaurant",
		"神社":      "shrine",
		"寺":       "temple",
		"公園":      "park",
		"自然":      "park",
		"観光":      "tourist_attraction",
		"ショッピング": "shopping_mall",
		"博物館":     "museum",
		"美術館":     "art_gallery",
	}

	// タイプとして使用できるもの
	typeMap := map[string]string{
		"カフェ":     "cafe",
		"レストラン":  "restaurant",
		"神社":      "shrine",
		"寺":       "temple",
		"公園":      "park",
		"自然":      "park",
		"観光":      "tourist_attraction",
		"ショッピング": "shopping_mall",
		"博物館":     "museum",
		"美術館":     "art_gallery",
	}

	if k, ok := keywordMap[tag]; ok {
		if t, ok := typeMap[tag]; ok {
			return "", t // typeとして使用
		}
		return k, ""
	}

	// デフォルトはキーワードとして使用
	return tag, ""
}

// SearchNearby 指定された座標の周辺を検索
func (s *NearbySearchService) SearchNearby(lat, lng, radius float64, interestTags []string) ([]PlaceResult, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_MAPS_API_KEY is not set")
	}

	var allResults []PlaceResult
	seenPlaceIDs := make(map[string]bool)

	// 各興味タグで検索
	for _, tag := range interestTags {
		keyword, placeType := tagToKeyword(tag)

		baseURL := "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
		params := url.Values{}
		params.Add("location", fmt.Sprintf("%.6f,%.6f", lat, lng))
		params.Add("radius", fmt.Sprintf("%.0f", radius))
		params.Add("key", s.apiKey)
		params.Add("language", "ja")

		if placeType != "" {
			params.Add("type", placeType)
		} else if keyword != "" {
			params.Add("keyword", keyword)
		} else {
			params.Add("keyword", tag)
		}

		reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

		resp, err := s.client.Get(reqURL)
		if err != nil {
			continue // エラーが発生しても次のタグで続行
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		var result NearbySearchResponse
		if err := json.Unmarshal(body, &result); err != nil {
			continue
		}

		if result.Status != "OK" {
			continue
		}

		// 結果を追加（重複排除）
		for _, r := range result.Results {
			if !seenPlaceIDs[r.PlaceID] {
				seenPlaceIDs[r.PlaceID] = true
				photoRef := ""
				if len(r.Photos) > 0 {
					photoRef = r.Photos[0].PhotoReference
				}
				allResults = append(allResults, PlaceResult{
					PlaceID:        r.PlaceID,
					Name:           r.Name,
					Lat:            r.Geometry.Location.Lat,
					Lng:            r.Geometry.Location.Lng,
					Rating:         r.Rating,
					PhotoReference: photoRef,
					Types:          r.Types,
				})
			}
		}
	}

	return allResults, nil
}

