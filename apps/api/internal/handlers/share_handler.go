package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetShare(c *gin.Context) {
	shareID := c.Param("share_id")

	trip, err := h.repo.GetTripByShareID(shareID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{
			"code":    "NOT_FOUND",
			"message": "共有された旅程が見つかりません",
		}})
		return
	}

	places, err := h.repo.GetTripPlaces(trip.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{
			"code":    "DATABASE_ERROR",
			"message": "スポットの取得に失敗しました",
		}})
		return
	}

	itinerary := h.convertItineraryFromPlaces(places, trip.StartTime)

	// ルートを再計算
	waypoints := []map[string]float64{}
	for _, place := range places {
		waypoints = append(waypoints, map[string]float64{
			"lat": place.Lat,
			"lng": place.Lng,
		})
	}

	var routePolyline string
	if len(waypoints) > 0 {
		mlReq := map[string]interface{}{
			"start":     "Hakata Station",
			"waypoints": waypoints,
		}
		mlResp, err := h.callMLService("/recompute-route", mlReq)
		if err == nil {
			if route, ok := mlResp["route"].(map[string]interface{}); ok {
				if polyline, ok := route["polyline"].(string); ok {
					routePolyline = polyline
				}
			}
		}
	}

	response := gin.H{
		"trip": gin.H{
			"id":         trip.ID,
			"title":      trip.Title,
			"start_time": trip.StartTime,
		},
		"itinerary": itinerary,
	}

	if routePolyline != "" {
		response["route"] = gin.H{"polyline": routePolyline}
	}

	c.JSON(http.StatusOK, response)
}

