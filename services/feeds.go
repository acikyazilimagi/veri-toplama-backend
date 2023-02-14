package services

import (
	"context"
	"fmt"
	locationsRepository "github.com/acikkaynak/veri-toplama-backend/repository/locations"
	"github.com/acikkaynak/veri-toplama-backend/util"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"math/rand"
)

const afetharitaURL = "https://apigo.afetharita.com/feeds/areas?ne_lat=39.91618777305531&ne_lng=47.85149904303703&sw_lat=36.07272886939253&sw_lng=23.872389299415502"

type repository interface {
	IsDuplicate(ctx context.Context, tweetContents string) (bool, error)
}

type locationServiceClient interface {
	GetLocation(locationID int) (*singleResponse, error)
	GetLocations(c *fiber.Ctx) ([]*locationsRepository.Location, error)
}

type feedService struct {
	repository            repository
	locationServiceClient locationServiceClient
	httpClient            httpClient
}

func NewFeedsServices(repository repository, locationServiceClient locationServiceClient) *feedService {
	return &feedService{
		repository:            repository,
		locationServiceClient: locationServiceClient,
	}
}

func (f *feedService) GeetFeeds(c *fiber.Ctx) (*FeetFeedsResponse, error) {
	logrus.Infoln("Pulling entries")
	locs, err := f.locationServiceClient.GetLocations(c)
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

func (f *feedService) filterLocations(c *fiber.Ctx, locations []*locationsRepository.Location) ([]*locationsRepository.Location, error) {
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

func (f *feedService) filterLocationsByEntryID(locations []*locationsRepository.Location) []*locationsRepository.Location {
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

func (f feedService) filterLocationsByCityID(c *fiber.Ctx, locations []*locationsRepository.Location) ([]*locationsRepository.Location, error) {
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

func (f feedService) filterLocationsByStartingAt(c *fiber.Ctx, locations []*locationsRepository.Location) ([]*locationsRepository.Location, error) {
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

func (f *feedService) getSelectedLocation(locations []*locationsRepository.Location) *locationsRepository.Location {
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

		singleData, err := f.locationServiceClient.GetLocation(s.EntryID)
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
