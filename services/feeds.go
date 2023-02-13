package services

import (
	"context"
	"encoding/json"
	"fmt"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	"github.com/YusufOzmen01/veri-kontrol-backend/util"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net/http"
	"time"
)

const afetharitaURL = "https://apigo.afetharita.com/feeds/areas?ne_lat=39.91618777305531&ne_lng=47.85149904303703&sw_lat=36.07272886939253&sw_lng=23.872389299415502"

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type cache interface {
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{}, cost int64) bool
	SetWithTTL(key, value interface{}, cost int64, ttl time.Duration) bool
}

type repository interface {
	IsDuplicate(ctx context.Context, tweetContents string) (bool, error)
}

type feedServices struct {
	repository repository
	httpClient httpClient
	cache      cache
}

func NewFeedsServices(repository repository, httpClient httpClient, cache cache) *feedServices {
	return &feedServices{
		repository: repository,
		httpClient: httpClient,
		cache:      cache,
	}
}

func (f *feedServices) GeetFeeds(c *fiber.Ctx) (*FeetFeedsResponse, error) {
	logrus.Infoln("Pulling entries")
	locs, err := f.getAllLocations()
	if err != nil {
		logrus.Errorf("Couldn't get all locations: %s", err)
		return nil, fmt.Errorf("Couldn't get all locations: %s", err)
	}

	locations, err := f.filterLocations(c, locs)
	if err != nil {
		logrus.Errorf("Couldn't filter locations: %s", err)
		return nil, fmt.Errorf("Couldn't filter locations: %s", err)
	}

	if len(locations) == 0 {
		return nil, fmt.Errorf("locations is empty")
	}

	selected := f.getSelectedLocation(locations)

	res := &FeetFeedsResponse{
		Count:    len(locations),
		Location: selected,
	}

	return res, nil
}

func (f *feedServices) getAllLocations() ([]*locationsRepository.Location, error) {
	var d struct {
		Locations []*locationsRepository.Location `json:"results"`
	}

	data, exists := f.cache.Get("locations")
	if exists {
		return data.([]*locationsRepository.Location), nil
	}

	req, err := http.NewRequest("GET", afetharitaURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	resp, err := f.httpClient.Do(req)
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

	f.cache.SetWithTTL("locations", d.Locations, int64(time.Minute*15), 0)

	return d.Locations, nil
}

func (f *feedServices) getLocation(locationID int) (*singleResponse, error) {
	data, exists := f.cache.Get(fmt.Sprintf("single_location_%d", locationID))
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

	resp, err := f.httpClient.Do(req)
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

	f.cache.Set(fmt.Sprintf("single_location_%d", locationID), singleData, 0)

	return &singleData, nil
}

func (f *feedServices) filterLocations(c *fiber.Ctx, locations []*locationsRepository.Location) ([]*locationsRepository.Location, error) {
	locs := f.filterLocationsByEntryID(locations)
	if len(locs) == 0 {
		return nil, fmt.Errorf("locations is empty")
	}

	l, err := f.filterLocationsByCityID(c, locs)
	if err != nil {
		logrus.Errorf("Couldn't filter by city id locations: %s", err)
		return nil, fmt.Errorf("Couldn't filter by city id locations: %s", err)
	}

	filteredLocations, err := f.filterLocationsByStartingAt(c, l)
	if err != nil {
		return nil, fmt.Errorf("Couldn't filter by startingAt locations: %s", err)
	}

	return filteredLocations, nil
}

func (f *feedServices) filterLocationsByEntryID(locations []*locationsRepository.Location) []*locationsRepository.Location {
	processedIDs := make([]int, 0)

	for _, loc := range locations {
		processedIDs = append(processedIDs, loc.EntryID)
	}

	for _, id := range processedIDs {
		for i, loc := range locations {
			if id == loc.EntryID {
				locations = append(locations[:i], locations[i+1:]...)

				continue
			}
		}
	}

	return locations
}

func (f feedServices) filterLocationsByCityID(c *fiber.Ctx, locations []*locationsRepository.Location) ([]*locationsRepository.Location, error) {
	cityID := c.QueryInt("city_id")
	if cityID < 1 {
		return nil, fmt.Errorf("city id is not found")
	}

	filteredLocations := make([]*locationsRepository.Location, 0)

	box := util.GetCity(cityID)
	for _, loc := range locations {
		if box[0] >= loc.Loc[0] && box[1] >= loc.Loc[1] && box[2] <= loc.Loc[0] && box[3] <= loc.Loc[1] {
			filteredLocations = append(filteredLocations, loc)
		}
	}

	return filteredLocations, nil
}

func (f feedServices) filterLocationsByStartingAt(c *fiber.Ctx, locations []*locationsRepository.Location) ([]*locationsRepository.Location, error) {
	startingAt := c.QueryInt("starting_at")
	if startingAt < 1 {
		return nil, fmt.Errorf("startingAt is not found")
	}

	filteredLocations := make([]*locationsRepository.Location, 0)

	for _, loc := range locations {
		if loc.Epoch >= startingAt {
			filteredLocations = append(filteredLocations, loc)
		}
	}

	return filteredLocations, nil
}

func (f *feedServices) getSelectedLocation(locations []*locationsRepository.Location) *locationsRepository.Location {
	var (
		selected  *locationsRepository.Location
		fullText  string
		processed = make([]int, 0)
	)

	for {
		randIndex := rand.Intn(len(locations))

		for _, i := range processed {
			if randIndex == i {
				continue
			}
		}

		processed = append(processed, randIndex)

		s := locations[randIndex]

		singleData, err := f.getLocation(s.EntryID)
		if err != nil {
			logrus.Errorln(err)
			break
		}

		exists, err := f.repository.IsDuplicate(context.TODO(), singleData.FullText)
		if err != nil {
			logrus.Errorln(err)
			break
		}

		if !exists {
			selected = s
			fullText = singleData.FullText

			break
		}
	}

	selected.OriginalMessage = fullText
	selected.OriginalLocation = fmt.Sprintf("https://www.google.com/maps/?q=%f,%f&ll=%f,%f&z=21",
		selected.Loc[0], selected.Loc[1], selected.Loc[0], selected.Loc[1])

	return selected
}

type FeetFeedsResponse struct {
	Count    int                           `json:"count"`
	Location *locationsRepository.Location `json:"location"`
}

type singleResponse struct {
	FullText         string `json:"full_text"`
	FormattedAddress string `json:"formatted_address"`
}
