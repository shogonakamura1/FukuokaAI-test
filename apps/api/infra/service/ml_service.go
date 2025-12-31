package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type IMLService interface {
	Recommend(mustPlaces []string, interestTags []string, freeText string) (map[string]interface{}, error)
	RecomputeRoute(waypoints []map[string]float64) (map[string]interface{}, error)
}

type MLService struct {
	baseURL string
}

func NewMLService(baseURL string) IMLService {
	return &MLService{
		baseURL: baseURL,
	}
}

func (s *MLService) call(endpoint string, data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	url := s.baseURL + endpoint
	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *MLService) Recommend(mustPlaces []string, interestTags []string, freeText string) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"start":         "Hakata Station",
		"must_places":   mustPlaces,
		"interest_tags": interestTags,
		"free_text":     freeText,
	}
	return s.call("/recommend", req)
}

func (s *MLService) RecomputeRoute(waypoints []map[string]float64) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"start":     "Hakata Station",
		"waypoints": waypoints,
	}
	return s.call("/recompute-route", req)
}

