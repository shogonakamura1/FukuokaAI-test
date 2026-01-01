package models

// GeocodingRequest ジオコーディング機能のリクエスト
type GeocodingRequest struct {
	PlaceName string `json:"place_name" binding:"required"` // 場所名
}

// GeocodingResponse ジオコーディング機能のレスポンス
type GeocodingResponse struct {
	PlaceID string  `json:"place_id"` // Google Place ID
	Lat     float64 `json:"lat"`      // 緯度
	Lng     float64 `json:"lng"`      // 経度
	Name    string  `json:"name"`     // 場所名
}

