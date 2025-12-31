package usecase

import (
	"fukuoka-ai-api/models"
	"math"
	"sort"
)

// haversineDistance 2点間の距離を計算（Haversine公式、メートル単位）
func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371000 // 地球の半径（メートル）

	// 度をラジアンに変換
	lat1Rad := lat1 * math.Pi / 180
	lng1Rad := lng1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lng2Rad := lng2 * math.Pi / 180

	// 差を計算
	deltaLat := lat2Rad - lat1Rad
	deltaLng := lng2Rad - lng1Rad

	// Haversine公式
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// buildMinimumSpanningTree 最小全域木を構築（Kruskal法の簡易版）
// 全ての点を最小距離で連結する
func buildMinimumSpanningTree(coordinates []models.Coordinate) []models.Edge {
	if len(coordinates) <= 1 {
		return []models.Edge{}
	}

	// 全ての可能なエッジを生成
	type edgeWithDistance struct {
		from     int
		to       int
		distance float64
	}

	var edges []edgeWithDistance
	for i := 0; i < len(coordinates); i++ {
		for j := i + 1; j < len(coordinates); j++ {
			dist := haversineDistance(
				coordinates[i].Lat, coordinates[i].Lng,
				coordinates[j].Lat, coordinates[j].Lng,
			)
			edges = append(edges, edgeWithDistance{
				from:     i,
				to:       j,
				distance: dist,
			})
		}
	}

	// 距離でソート
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].distance < edges[j].distance
	})

	// Union-Find構造でサイクルを回避しながらエッジを追加
	parent := make([]int, len(coordinates))
	for i := range parent {
		parent[i] = i
	}

	var find func(int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}

	union := func(x, y int) bool {
		px, py := find(x), find(y)
		if px == py {
			return false // サイクルが発生
		}
		parent[px] = py
		return true
	}

	var mstEdges []models.Edge
	for _, e := range edges {
		if union(e.from, e.to) {
			mstEdges = append(mstEdges, models.Edge{
				From:     coordinates[e.from],
				To:       coordinates[e.to],
				Distance: e.distance,
			})
			// 全ての点が連結されたら終了
			if len(mstEdges) == len(coordinates)-1 {
				break
			}
		}
	}

	// 理論的には不要: 完全グラフでKruskal法を使用している場合、
	// n-1本のエッジで全ての頂点が連結されるため、独立したコンポーネントは残らない
	// 以下の処理は防御的プログラミングとして残していたが、実際には実行されない
	// components := getComponents(coordinates, mstEdges)
	// if len(components) > 1 {
	// 	// 各コンポーネント間の最短距離を計算して結合
	// 	additionalEdges := connectComponents(components, coordinates)
	// 	mstEdges = append(mstEdges, additionalEdges...)
	// }

	return mstEdges
}

// getComponents グラフの連結成分を取得
func getComponents(coordinates []models.Coordinate, edges []models.Edge) [][]int {
	visited := make([]bool, len(coordinates))
	adjList := make([][]int, len(coordinates))

	// 隣接リストを構築
	for _, edge := range edges {
		fromIdx := findCoordinateIndex(coordinates, edge.From)
		toIdx := findCoordinateIndex(coordinates, edge.To)
		if fromIdx >= 0 && toIdx >= 0 {
			adjList[fromIdx] = append(adjList[fromIdx], toIdx)
			adjList[toIdx] = append(adjList[toIdx], fromIdx)
		}
	}

	var components [][]int
	for i := 0; i < len(coordinates); i++ {
		if !visited[i] {
			var component []int
			dfs(i, adjList, visited, &component)
			components = append(components, component)
		}
	}

	return components
}

// dfs 深さ優先探索
func dfs(node int, adjList [][]int, visited []bool, component *[]int) {
	visited[node] = true
	*component = append(*component, node)
	for _, neighbor := range adjList[node] {
		if !visited[neighbor] {
			dfs(neighbor, adjList, visited, component)
		}
	}
}

// findCoordinateIndex 座標のインデックスを検索
func findCoordinateIndex(coordinates []models.Coordinate, coord models.Coordinate) int {
	for i, c := range coordinates {
		if c.Lat == coord.Lat && c.Lng == coord.Lng {
			return i
		}
	}
	return -1
}

// connectComponents 独立したコンポーネントを最短距離で結合
func connectComponents(components [][]int, coordinates []models.Coordinate) []models.Edge {
	if len(components) <= 1 {
		return []models.Edge{}
	}

	var additionalEdges []models.Edge

	// 各コンポーネントペア間の最短距離を計算
	for i := 0; i < len(components); i++ {
		for j := i + 1; j < len(components); j++ {
			minDist := math.MaxFloat64
			var minFrom, minTo models.Coordinate

			// コンポーネントiとjの間の最短距離を探す
			for _, idx1 := range components[i] {
				for _, idx2 := range components[j] {
					dist := haversineDistance(
						coordinates[idx1].Lat, coordinates[idx1].Lng,
						coordinates[idx2].Lat, coordinates[idx2].Lng,
					)
					if dist < minDist {
						minDist = dist
						minFrom = coordinates[idx1]
						minTo = coordinates[idx2]
					}
				}
			}

			if minDist < math.MaxFloat64 {
				additionalEdges = append(additionalEdges, models.Edge{
					From:     minFrom,
					To:       minTo,
					Distance: minDist,
				})
			}
		}
	}

	// 最短距離のエッジのみを返す（全てのコンポーネントを連結する最小のエッジ数）
	if len(additionalEdges) > 0 {
		sort.Slice(additionalEdges, func(i, j int) bool {
			return additionalEdges[i].Distance < additionalEdges[j].Distance
		})
		// 必要な数だけ返す（コンポーネント数-1）
		needed := len(components) - 1
		if len(additionalEdges) > needed {
			return additionalEdges[:needed]
		}
	}

	return additionalEdges
}

// calculateSearchRadius 枝の長さから検索半径を計算（枝の長さ/√3）
func calculateSearchRadius(edgeLength float64) float64 {
	return edgeLength / math.Sqrt(3)
}

