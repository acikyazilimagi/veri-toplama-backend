package handler

import (
	"context"
	"fmt"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
	"github.com/YusufOzmen01/veri-kontrol-backend/util"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"time"
)

func GetFeeds(c *fiber.Ctx) error {
	ctx := context.Background()
	cache := sources.NewCache(1<<30, 1e7, 64)
	rand.Seed(time.Now().UnixMilli())

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		panic("mongo URL is empty")
	}

	mongoClient := sources.NewMongoClient(ctx, mongoURL, "database")
	locationRepository := locationsRepository.NewRepository(mongoClient)

	// TODO:
	processedIDs := make([]int, 0)

	logrus.Infoln("Pulling entries")
	locs, err := locationRepository.GetLocations(ctx)
	if err != nil {
		logrus.Errorf("Couldn't get all locations: %s", err)
	}

	for _, loc := range locs {
		processedIDs = append(processedIDs, loc.EntryID)
	}

	locations, err := tools.GetAllLocations(ctx, cache)
	if err != nil {
		logrus.Errorln(err)

		return c.SendString(err.Error())
	}

	for _, id := range processedIDs {
		for i, loc := range locations {
			if id == loc.EntryID {
				locations = append(locations[:i], locations[i+1:]...)

				continue
			}
		}
	}

	cityID := c.QueryInt("city_id")
	if cityID > 0 {
		// box := cities[cityID]
		box := util.GetCity(cityID)
		if box == nil {
			// TODO: akışı kesmeliyiz
		}

		filteredLocations := make([]*locationsRepository.Location, 0)

		for _, loc := range locations {
			if box[0] >= loc.Loc[0] && box[1] >= loc.Loc[1] && box[2] <= loc.Loc[0] && box[3] <= loc.Loc[1] {
				filteredLocations = append(filteredLocations, loc)
			}
		}

		locations = filteredLocations
	}

	startingAt := c.QueryInt("starting_at")
	if startingAt > 0 {
		filteredLocations := make([]*locationsRepository.Location, 0)

		for _, loc := range locations {
			if loc.Epoch >= startingAt {
				filteredLocations = append(filteredLocations, loc)
			}
		}

		locations = filteredLocations
	}

	if len(locations) == 0 {
		return c.JSON(struct {
			Count    int                           `json:"count"`
			Location *locationsRepository.Location `json:"location"`
		}{
			Count:    0,
			Location: nil,
		})
	}

	var selected *locationsRepository.Location
	fullText := ""
	processed := make([]int, 0)

	for {
		randIndex := rand.Intn(len(locations))

		if len(processed) == len(locations) {
			return c.JSON(struct {
				Count    int                           `json:"count"`
				Location *locationsRepository.Location `json:"location"`
			}{
				Count:    0,
				Location: nil,
			})
		}

		for _, i := range processed {
			if randIndex == i {
				continue
			}
		}

		processed = append(processed, randIndex)

		s := locations[randIndex]

		singleData, err := tools.GetSingleLocation(ctx, s.EntryID, cache)
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		exists, err := locationRepository.IsDuplicate(c.Context(), singleData.FullText)
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		if !exists {
			selected = s
			fullText = singleData.FullText

			break
		}
	}

	selected.OriginalMessage = fullText
	selected.OriginalLocation = fmt.Sprintf("https://www.google.com/maps/?q=%f,%f&ll=%f,%f&z=21", selected.Loc[0], selected.Loc[1], selected.Loc[0], selected.Loc[1])

	return c.JSON(struct {
		Count    int                           `json:"count"`
		Location *locationsRepository.Location `json:"location"`
	}{
		Count:    len(locations),
		Location: selected,
	})
}
