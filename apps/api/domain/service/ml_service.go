package service

type MLService interface {
	Recommend(mustPlaces []string, interestTags []string, freeText string) (map[string]interface{}, error)
	RecomputeRoute(waypoints []map[string]float64) (map[string]interface{}, error)
}

