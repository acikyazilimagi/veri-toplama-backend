package services

import (
	"encoding/json"
	"fmt"
	locationsRepository "github.com/acikkaynak/veri-toplama-backend/repository/locations"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"time"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type cache interface {
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{}, cost int64) bool
	SetWithTTL(key, value interface{}, cost int64, ttl time.Duration) bool
}

type locationService struct {
	httpClient httpClient
	cache      cache
}

func NewLocationService(httpClient httpClient, cache cache) *locationService {
	return &locationService{
		httpClient: httpClient,
		cache:      cache,
	}
}

func (s *locationService) GetLocation(locationID int) (*singleResponse, error) {
	data, exists := s.cache.Get(fmt.Sprintf("single_location_%d", locationID))
	if exists {
		return data.(*singleResponse), nil
	}

	url := fmt.Sprintf("https://apigo.afetharita.com/feeds/%d", locationID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	singleData := singleResponse{}
	if err := json.Unmarshal(body, &singleData); err != nil {
		return nil, err
	}

	s.cache.Set(fmt.Sprintf("single_location_%d", locationID), singleData, 0)

	return &singleData, nil
}

func (s *locationService) GetLocations(c *fiber.Ctx) ([]*locationsRepository.Location, error) {
	var d struct {
		Locations []*locationsRepository.Location `json:"results"`
	}

	data, exists := s.cache.Get("locations")
	if exists {
		return data.([]*locationsRepository.Location), nil
	}

	req, err := http.NewRequest("GET", afetharitaURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &d); err != nil {
		return nil, err
	}

	s.cache.SetWithTTL("locations", d.Locations, int64(time.Minute*15), 0)

	return d.Locations, nil
}

type singleResponse struct {
	FullText         string `json:"full_text"`
	FormattedAddress string `json:"formatted_address"`
}
