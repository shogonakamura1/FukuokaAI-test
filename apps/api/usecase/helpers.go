package usecase

import (
	"fmt"
	"strings"

	"fukuoka-ai-api/domain/entity"
)

func convertItineraryFromPlaces(places []*entity.TripPlace, startTime string) []entity.TripPlace {
	result := []entity.TripPlace{}
	currentTime := parseTime(startTime)

	for _, place := range places {
		timeRange := formatTimeRange(currentTime, place.StayMinutes)
		// TripPlaceのコピーを作成し、TimeRangeを設定
		placeWithTimeRange := *place
		placeWithTimeRange.TimeRange = timeRange
		placeWithTimeRange.TripID = "" // レスポンスではTripIDを除外
		result = append(result, placeWithTimeRange)
		currentTime = addMinutes(currentTime, place.StayMinutes)
	}

	return result
}

func calculateTimeline(places []*entity.TripPlace, startTime string, mlResp map[string]interface{}) []entity.TripPlace {
	// 簡易実装: 移動時間は30分固定（MVP）
	// 実際にはDirections APIのlegsから取得すべき
	result := []entity.TripPlace{}
	currentTime := parseTime(startTime)

	for i, place := range places {
		if i > 0 {
			// 移動時間（30分固定、MVP）
			currentTime = addMinutes(currentTime, 30)
		}

		timeRange := formatTimeRange(currentTime, place.StayMinutes)
		// TripPlaceのコピーを作成し、TimeRangeを設定
		placeWithTimeRange := *place
		placeWithTimeRange.TimeRange = timeRange
		placeWithTimeRange.TripID = "" // レスポンスではTripIDを除外
		result = append(result, placeWithTimeRange)

		currentTime = addMinutes(currentTime, place.StayMinutes)
	}

	return result
}

// ヘルパー関数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0.0
}

func parseTime(timeStr string) int {
	// "10:00" -> 600 (分)
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 600 // デフォルト10:00
	}
	hours := 0
	minutes := 0
	fmt.Sscanf(parts[0], "%d", &hours)
	fmt.Sscanf(parts[1], "%d", &minutes)
	return hours*60 + minutes
}

func addMinutes(timeMinutes int, minutes int) int {
	return timeMinutes + minutes
}

func formatTimeRange(startMinutes int, durationMinutes int) string {
	startHour := startMinutes / 60
	startMin := startMinutes % 60
	endMinutes := startMinutes + durationMinutes
	endHour := endMinutes / 60
	endMin := endMinutes % 60
	return fmt.Sprintf("%02d:%02d-%02d:%02d", startHour, startMin, endHour, endMin)
}

