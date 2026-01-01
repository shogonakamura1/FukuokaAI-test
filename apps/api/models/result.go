package models

// ResultRequest ルート提案機能のリクエスト
type ResultRequest struct {
	Places []string `json:"places" binding:"required"` // 場所IDのリスト
}

// RouteLeg ルートの区間情報
type RouteLeg struct {
	StartLocation  Coordinate `json:"start_location"`
	EndLocation    Coordinate `json:"end_location"`
	DistanceMeters int        `json:"distance_meters"` // メートル単位
	Duration       string     `json:"duration"`        // 所要時間（例: "3600s"）
}

// Route ルート情報
type Route struct {
	Legs           []RouteLeg `json:"legs"`
	DistanceMeters int        `json:"distance_meters"` // 総距離（メートル単位）
	Duration       string     `json:"duration"`        // 総所要時間（例: "3600s"）
	OptimizedOrder []int      `json:"optimized_order"` // 最適化された順序
}

// ResultResponse ルート提案機能のレスポンス
type ResultResponse struct {
	Places []Place `json:"places"` // 最適化された順序の場所リスト
	Route  Route   `json:"route"`  // ルート情報
}

