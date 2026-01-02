package models

// RecommendRequest リコメンド機能のリクエスト
type RecommendRequest struct {
	MustPlaces   []string `json:"must_places" binding:"required"`   // 寄りたい場所（リスト）
	InterestTags []string `json:"interest_tags" binding:"required"` // 興味タグ（リスト）
	StartPlace   string   `json:"start_place,omitempty"`            // 出発地点（オプション、デフォルトは博多駅）
	GoalPlace    string   `json:"goal_place,omitempty"`             // ゴール地点（オプション）
}

// Place 場所情報
type Place struct {
	PlaceID        string  `json:"place_id"`
	Name           string  `json:"name"`
	Lat            float64 `json:"lat"`
	Lng            float64 `json:"lng"`
	PhotoURL       string  `json:"photo_url,omitempty"`
	Rating         float64 `json:"rating,omitempty"`
	ReviewSummary  string  `json:"review_summary,omitempty"`
	Category       string  `json:"category,omitempty"`
	Address        string  `json:"address,omitempty"`
	RelevanceScore float64 `json:"relevance_score,omitempty"` // 関連性スコア
}

// RecommendResponse リコメンド機能のレスポンス
type RecommendResponse struct {
	Places           []Place `json:"places"`             // 推薦場所（最大10件、レビュー順）
	MaxPossibleScore float64 `json:"max_possible_score"` // 理論的最大スコア
}

// Coordinate 座標情報
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Name string `json:"name,omitempty"`
}

// Edge グラフのエッジ（枝）
type Edge struct {
	From     Coordinate `json:"from"`
	To       Coordinate `json:"to"`
	Distance float64    `json:"distance"` // メートル単位
}

