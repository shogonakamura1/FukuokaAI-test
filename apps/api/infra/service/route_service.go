package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// IRouteService ルートサービスのインターフェース
type IRouteService interface {
	ComputeRoute(originLat, originLng float64, destinationLat, destinationLng float64, intermediates []Waypoint, travelMode string, departureTime *time.Time) (*RouteResponse, error)
}

// RouteService Google Maps Routes APIを使用したルートサービス
type RouteService struct {
	apiKey string
	client *http.Client
}

// Waypoint 経由地点
type Waypoint struct {
	PlaceID string  `json:"place_id,omitempty"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

// RouteResponse ルートAPIのレスポンス
type RouteResponse struct {
	Routes []struct {
		Legs []struct {
			StartLocation struct {
				LatLng struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"latLng"`
			} `json:"startLocation"`
			EndLocation struct {
				LatLng struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"latLng"`
			} `json:"endLocation"`
			DistanceMeters int    `json:"distanceMeters"`
			Duration        string `json:"duration"` // "3600s"形式
		} `json:"legs"`
		DistanceMeters int      `json:"distanceMeters"`
		Duration       string   `json:"duration"`
		OptimizedOrder []int    `json:"optimizedOrder,omitempty"`
	} `json:"routes"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// RouteRequest ルートAPIのリクエスト
type RouteRequest struct {
	Origin struct {
		Location struct {
			LatLng struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"latLng"`
		} `json:"location"`
	} `json:"origin"`
	Destination struct {
		Location struct {
			LatLng struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"latLng"`
		} `json:"location"`
	} `json:"destination"`
	Intermediates []struct {
		Location struct {
			LatLng struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"latLng"`
		} `json:"location"`
	} `json:"intermediates,omitempty"`
	TravelMode          string `json:"travelMode"`
	RoutingPreference   string `json:"routingPreference,omitempty"`
	OptimizeWaypointOrder bool `json:"optimizeWaypointOrder,omitempty"`
	DepartureTime       string `json:"departureTime,omitempty"` // RFC3339形式
}

// NewRouteService 新しいRouteServiceを作成
func NewRouteService() IRouteService {
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	return &RouteService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// ComputeRoute ルートを計算
func (s *RouteService) ComputeRoute(originLat, originLng float64, destinationLat, destinationLng float64, intermediates []Waypoint, travelMode string, departureTime *time.Time) (*RouteResponse, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("GOOGLE_MAPS_API_KEY is not set")
	}

	// リクエストボディを作成
	reqBody := RouteRequest{
		Origin: struct {
			Location struct {
				LatLng struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"latLng"`
			} `json:"location"`
		}{
			Location: struct {
				LatLng struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"latLng"`
			}{
				LatLng: struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					Latitude:  originLat,
					Longitude: originLng,
				},
			},
		},
		Destination: struct {
			Location struct {
				LatLng struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"latLng"`
			} `json:"location"`
		}{
			Location: struct {
				LatLng struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"latLng"`
			}{
				LatLng: struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					Latitude:  destinationLat,
					Longitude: destinationLng,
				},
			},
		},
		TravelMode:          travelMode,
		RoutingPreference:   "TRAFFIC_AWARE",
		OptimizeWaypointOrder: true,
	}

	// 経由地点を追加
	for _, waypoint := range intermediates {
		reqBody.Intermediates = append(reqBody.Intermediates, struct {
			Location struct {
				LatLng struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"latLng"`
			} `json:"location"`
		}{
			Location: struct {
				LatLng struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				} `json:"latLng"`
			}{
				LatLng: struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					Latitude:  waypoint.Lat,
					Longitude: waypoint.Lng,
				},
			},
		})
	}

	// 出発時刻を設定（指定されている場合）
	if departureTime != nil {
		reqBody.DepartureTime = departureTime.Format(time.RFC3339)
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// APIエンドポイント
	url := fmt.Sprintf("https://routes.googleapis.com/directions/v2:computeRoutes?key=%s", s.apiKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-FieldMask", "routes.duration,routes.distanceMeters,routes.legs,routes.optimizedOrder")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google Routes API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result RouteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google Routes API error: status %d, message: %s", result.Error.Code, result.Error.Message)
	}

	if len(result.Routes) == 0 {
		return nil, fmt.Errorf("no routes found")
	}

	return &result, nil
}

