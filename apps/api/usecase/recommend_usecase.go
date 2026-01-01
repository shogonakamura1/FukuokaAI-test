package usecase

import (
	"fmt"
	"fukuoka-ai-api/infra/service"
	"fukuoka-ai-api/models"
	"sort"
)

// IRecommendUsecase リコメンド機能のユースケースインターフェース
type IRecommendUsecase interface {
	Recommend(req *models.RecommendRequest) (*models.RecommendResponse, error)
}

// RecommendUsecase リコメンド機能のユースケース実装
type RecommendUsecase struct {
	geocodingService    service.IGeocodingService
	nearbySearchService service.INearbySearchService
	placeDetailsService service.IPlaceDetailsService
}

// NewRecommendUsecase 新しいRecommendUsecaseを作成
func NewRecommendUsecase(
	geocodingService service.IGeocodingService,
	nearbySearchService service.INearbySearchService,
	placeDetailsService service.IPlaceDetailsService,
) IRecommendUsecase {
	return &RecommendUsecase{
		geocodingService:    geocodingService,
		nearbySearchService: nearbySearchService,
		placeDetailsService: placeDetailsService,
	}
}

// Recommend リコメンド機能のメイン処理
func (u *RecommendUsecase) Recommend(req *models.RecommendRequest) (*models.RecommendResponse, error) {
	// 処理フロー1: 出発地点とゴール地点を指定したのち、それ以外の必ず寄りたい場所の座標を得る
	startPlace := req.StartPlace
	if startPlace == "" {
		startPlace = "Hakata Station"
	}

	// 出発地点の座標を取得
	startLat, startLng, _, err := u.geocodingService.GetCoordinates(startPlace)
	if err != nil {
		return nil, fmt.Errorf("出発地点の座標取得に失敗しました: %w", err)
	}

	// ゴール地点の座標を取得（指定されている場合）
	var goalLat, goalLng float64
	if req.GoalPlace != "" {
		goalLat, goalLng, _, err = u.geocodingService.GetCoordinates(req.GoalPlace)
		if err != nil {
			return nil, fmt.Errorf("ゴール地点の座標取得に失敗しました: %w", err)
		}
	} else {
		// ゴール地点が指定されていない場合は出発地点と同じにする
		goalLat, goalLng = startLat, startLng
	}

	// 寄りたい場所の座標を取得
	var mustPlaceCoords []models.Coordinate
	for _, placeName := range req.MustPlaces {
		lat, lng, _, err := u.geocodingService.GetCoordinates(placeName)
		if err != nil {
			// 見つからない場所はスキップ
			continue
		}
		mustPlaceCoords = append(mustPlaceCoords, models.Coordinate{
			Lat:  lat,
			Lng:  lng,
			Name: placeName,
		})
	}

	// 全ての座標をまとめる（出発地点、ゴール地点、寄りたい場所）
	allCoordinates := []models.Coordinate{
		{Lat: startLat, Lng: startLng, Name: startPlace},
	}
	allCoordinates = append(allCoordinates, mustPlaceCoords...)
	allCoordinates = append(allCoordinates, models.Coordinate{
		Lat: goalLat, Lng: goalLng, Name: req.GoalPlace,
	})

	// 処理フロー2: 出発地点とゴール地点を含む全ての座標で、それぞれ一番距離が近い組み合わせを作り、
	// 全ての点が線でつながるようにする（独立した枝ができた場合、枝同士で一番近い地点同士を結ぶ）
	edges := buildMinimumSpanningTree(allCoordinates)

	// 処理フロー3: 全ての枝で、半径が 枝の長さ/√3 となる円内で、興味タグで検索をnearby search APIで検索する
	var allCandidates []service.PlaceResult
	seenPlaceIDs := make(map[string]bool)

	for _, edge := range edges {
		// エッジの中点を計算
		midLat := (edge.From.Lat + edge.To.Lat) / 2
		midLng := (edge.From.Lng + edge.To.Lng) / 2

		// 検索半径を計算（枝の長さ/√3）
		searchRadius := calculateSearchRadius(edge.Distance)

		// 周辺検索を実行
		results, err := u.nearbySearchService.SearchNearby(midLat, midLng, searchRadius, req.InterestTags)
		if err != nil {
			// エラーが発生しても次のエッジで続行
			continue
		}

		// 結果を追加（重複排除）
		for _, result := range results {
			if !seenPlaceIDs[result.PlaceID] {
				seenPlaceIDs[result.PlaceID] = true
				allCandidates = append(allCandidates, result)
			}
		}
	}

	// 処理フロー4: その結果を一枚の写真と共にレビューの高い順に合計10件表示する
	// まず、関連性が低い結果をフィルタリング
	var filteredCandidates []service.PlaceResult
	for _, candidate := range allCandidates {
		score := calculateRelevanceScore(candidate, req.InterestTags)
		// 関連性スコアが5.0未満の結果は除外（タグと全く関係ない可能性が高い）
		if score >= 5.0 {
			filteredCandidates = append(filteredCandidates, candidate)
		}
	}
	
	// タグとの関連性を考慮してソート
	sort.Slice(filteredCandidates, func(i, j int) bool {
		scoreI := calculateRelevanceScore(filteredCandidates[i], req.InterestTags)
		scoreJ := calculateRelevanceScore(filteredCandidates[j], req.InterestTags)
		
		// 関連性スコアが同じ場合は評価で比較
		if scoreI == scoreJ {
			return filteredCandidates[i].Rating > filteredCandidates[j].Rating
		}
		
		// 関連性スコアが高いものを優先
		return scoreI > scoreJ
	})
	
	// フィルタリング後の候補を使用
	allCandidates = filteredCandidates

	// 最大4件に制限
	maxResults := 4
	if len(allCandidates) > maxResults {
		allCandidates = allCandidates[:maxResults]
	}

	// Place Details APIで詳細情報を取得
	var places []models.Place
	for _, candidate := range allCandidates {
		details, err := u.placeDetailsService.GetPlaceDetails(candidate.PlaceID, candidate.PhotoReference)
		if err != nil {
			// 詳細取得に失敗した場合は基本情報のみを使用
			places = append(places, models.Place{
				PlaceID:  candidate.PlaceID,
				Name:     candidate.Name,
				Lat:      candidate.Lat,
				Lng:      candidate.Lng,
				Rating:   candidate.Rating,
				PhotoURL: "", // 写真URLは取得できない
			})
			continue
		}

		places = append(places, models.Place{
			PlaceID:       details.PlaceID,
			Name:          details.Name,
			Lat:           details.Lat,
			Lng:           details.Lng,
			PhotoURL:      details.PhotoURL,
			Rating:        details.Rating,
			ReviewSummary: details.ReviewSummary,
			Category:      details.Category,
			Address:       details.Address,
		})
	}

	return &models.RecommendResponse{
		Places: places,
	}, nil
}

// calculateRelevanceScore 検索結果と検索タグの関連性スコアを計算
// スコアが高いほど検索タグと関連性が高い
func calculateRelevanceScore(candidate service.PlaceResult, interestTags []string) float64 {
	score := 0.0
	
	// 1. MatchedTagsの数を考慮（複数のタグで見つかった場合は高スコア）
	matchedCount := float64(len(candidate.MatchedTags))
	score += matchedCount * 10.0 // マッチしたタグ数 × 10
	
	// 2. Typesフィールドと検索タグのマッピングを比較
	// タグからGoogle Places APIのタイプへのマッピング
	tagToTypeMap := map[string][]string{
		"カフェ":     {"cafe", "cafe", "food", "point_of_interest", "establishment"},
		"レストラン":  {"restaurant", "food", "point_of_interest", "establishment"},
		"神社":      {"shrine", "place_of_worship", "point_of_interest", "establishment"},
		"寺":       {"temple", "place_of_worship", "point_of_interest", "establishment"},
		"公園":      {"park", "point_of_interest", "establishment"},
		"自然":      {"park", "natural_feature", "point_of_interest", "establishment"},
		"観光":      {"tourist_attraction", "tourist_attraction", "point_of_interest", "establishment"},
		"ショッピング": {"shopping_mall", "store", "point_of_interest", "establishment"},
		"博物館":     {"museum", "point_of_interest", "establishment"},
		"美術館":     {"art_gallery", "museum", "point_of_interest", "establishment"},
	}
	
	// 各検索タグについて、Typesとの一致を確認
	for _, tag := range interestTags {
		expectedTypes, exists := tagToTypeMap[tag]
		if !exists {
			continue
		}
		
		// Typesフィールドに期待されるタイプが含まれているか確認
		for _, placeType := range candidate.Types {
			for _, expectedType := range expectedTypes {
				if placeType == expectedType {
					// 主要タイプ（最初の2つ）に一致する場合は高スコア
					if placeType == expectedTypes[0] {
						score += 20.0 // 完全一致
					} else if len(expectedTypes) > 1 && placeType == expectedTypes[1] {
						score += 15.0 // 準一致
					} else {
						score += 5.0 // 部分一致
					}
					break
				}
			}
		}
	}
	
	// 3. 評価も少し考慮（関連性が同じ場合の補助的な要素）
	score += candidate.Rating * 0.5
	
	return score
}

