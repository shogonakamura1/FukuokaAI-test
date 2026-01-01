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
			Duration       string `json:"duration"` // "3600s"形式
		} `json:"legs"`
		DistanceMeters                      int   `json:"distanceMeters"`
		Duration                           string `json:"duration"`
		OptimizedIntermediateWaypointIndex []int `json:"optimizedIntermediateWaypointIndex,omitempty"` // Routes API v2の正しいフィールド名
	} `json:"routes"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
		Details []struct {
			Type  string `json:"@type"`
			Field string `json:"field"`
		} `json:"details,omitempty"`
	} `json:"error,omitempty"`
}

// IntermediateWaypoint 経由地点
type IntermediateWaypoint struct {
	Location struct {
		LatLng struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"latLng"`
	} `json:"location"`
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
	Intermediates        []IntermediateWaypoint `json:"intermediates,omitempty"`
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
		TravelMode: travelMode,
		// RoutingPreferenceは省略可能。TRAFFIC_AWAREを使用する場合は設定
		// ただし、Routes API v2では、RoutingPreferenceの値が正しくないとエラーになる可能性がある
		// RoutingPreference: "TRAFFIC_AWARE",
	}

	// 経由地点を追加（intermediatesが空でない場合のみ）
	if len(intermediates) > 0 {
		for _, waypoint := range intermediates {
			reqBody.Intermediates = append(reqBody.Intermediates, IntermediateWaypoint{
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

		// OptimizeWaypointOrderは、intermediatesが2つ以上ある場合のみ有効化
		// Routes API v2では、intermediatesが1つ以下の場合、OptimizeWaypointOrderをtrueにするとエラーになる
		if len(intermediates) >= 2 {
			reqBody.OptimizeWaypointOrder = true
		}
	}

	// 出発時刻を設定（指定されている場合）
	// RoutingPreferenceをTRAFFIC_AWAREに設定する場合、departureTimeが必須
	if departureTime != nil {
		reqBody.DepartureTime = departureTime.Format(time.RFC3339)
		reqBody.RoutingPreference = "TRAFFIC_AWARE"
	}
	// departureTimeが指定されていない場合は、RoutingPreferenceを設定しない（デフォルトのルーティングを使用）

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// デバッグ用：リクエストボディをログに出力
	fmt.Printf("Routes API Request: %s\n", string(jsonData))

	// APIエンドポイント
	url := fmt.Sprintf("https://routes.googleapis.com/directions/v2:computeRoutes?key=%s", s.apiKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Routes API v2では、optimizeWaypointOrderがtrueの場合、routes.optimized_intermediate_waypoint_indexをフィールドマスクに含める必要がある
	fieldMask := "routes.duration,routes.distanceMeters,routes.legs"
	if reqBody.OptimizeWaypointOrder {
		fieldMask += ",routes.optimized_intermediate_waypoint_index"
	}
	req.Header.Set("X-Goog-FieldMask", fieldMask)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google Routes API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// エラーレスポンスの詳細をログに出力（デバッグ用）
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Routes API Error Response: %s\n", string(body))
	}

	var result RouteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		errorMsg := result.Error.Message
		if errorMsg == "" && result.Error.Status != "" {
			errorMsg = result.Error.Status
		}
		if errorMsg == "" {
			errorMsg = string(body)
		}
		// エラーの詳細情報を追加
		if len(result.Error.Details) > 0 {
			details := ""
			for _, detail := range result.Error.Details {
				details += fmt.Sprintf(" [%s: %s]", detail.Type, detail.Field)
			}
			errorMsg += details
		}
		return nil, fmt.Errorf("Google Routes API error: status %d, message: %s", resp.StatusCode, errorMsg)
	}

	if len(result.Routes) == 0 {
		return nil, fmt.Errorf("no routes found")
	}

	return &result, nil
}

