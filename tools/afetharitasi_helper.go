package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/network"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	log "github.com/sirupsen/logrus"
	"time"
)

type SingleResponse struct {
	FullText         string `json:"full_text"`
	FormattedAddress string `json:"formatted_address"`
}

func GetAllLocations(ctx context.Context, cache sources.Cache) ([]*locations.Location, error) {
	var d struct {
		Locations []*locations.Location `json:"results"`
	}

	data, exists := cache.Get("locations")
	if exists {
		return data.([]*locations.Location), nil
	}

	res, _, err := network.ProcessGet(ctx, "https://apigo.afetharita.com/feeds/areas?ne_lat=39.91618777305531&ne_lng=47.85149904303703&sw_lat=36.07272886939253&sw_lng=23.872389299415502", map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
	})
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(res, &d); err != nil {
		return nil, err
	}

	cache.SetWithTTL("locations", d.Locations, int64(time.Minute*15), 0)

	return d.Locations, nil
}

func GetSingleLocation(ctx context.Context, locationID int, cache sources.Cache) (*SingleResponse, error) {
	data, exists := cache.Get(fmt.Sprintf("single_location_%d", locationID))
	if exists {
		return data.(*SingleResponse), nil
	}

	resp, _, err := network.ProcessGet(ctx, fmt.Sprintf("https://apigo.afetharita.com/feeds/%d", locationID), map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36",
	})
	if err != nil {
		return nil, err
	}

	singleData := &SingleResponse{}
	if err := json.Unmarshal(resp, singleData); err != nil {
		log.Errorln(string(resp))

		panic(err)
	}

	cache.Set(fmt.Sprintf("single_location_%d", locationID), singleData, 0)

	return singleData, nil
}
