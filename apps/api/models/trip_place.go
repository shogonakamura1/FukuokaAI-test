package models

type TripPlace struct {
	ID            string  `json:"id,omitempty"`
	TripID        string  `json:"trip_id,omitempty"` // レスポンスでは除外
	PlaceID       string  `json:"place_id"`
	Name          string  `json:"name"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
	Kind          string  `json:"kind"` // must, recommended, start
	StayMinutes   int     `json:"stay_minutes"`
	OrderIndex    int     `json:"order_index"`
	TimeRange     string  `json:"time_range,omitempty"` // 計算フィールド（DBには保存しない）
	Reason        string  `json:"reason,omitempty"`
	ReviewSummary string  `json:"review_summary,omitempty"`
	PhotoURL      string  `json:"photo_url,omitempty"`
	Category      string  `json:"category,omitempty"` // レスポンス用（DBには保存しない）
}

