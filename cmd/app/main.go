package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Netflix/go-env"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/repository"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type Environment struct {
	MongoUri string `env:"mongo_uri"`
}

var cities = map[int][]float64{
	1: {36.852702785393014, 36.87286376953126, 36.535570922786015, 35.88409423828126},
	2: {36.2104851748389, 36.81861877441407, 35.84286468375614, 35.82984924316407},
	3: {36.495937096205274, 36.649870522206335, 36.064120488812605, 35.4740187605459},
	4: {36.50903585150776, 36.402143998719424, 36.47976138594277, 36.31474829364722},
	5: {36.64234742932176, 36.3232450328562, 36.53629731173617, 36.029282092441115},
	6: {36.116001873480265, 36.06470054394251, 36.0627178139989, 35.91771907373497},
}

type ResolveBody struct {
	ID           int    `json:"id"`
	LocationType int    `json:"type"`
	NewAddress   string `json:"new_address"`
	OpenAddress  string `json:"open_address"`
	Apartment    string `json:"apartment"`
	Reason       string `json:"reason"`
}

func main() {
	app := fiber.New()
	ctx := context.Background()
	cache := sources.NewCache(1<<30, 1e7, 64)

	rand.Seed(time.Now().UnixMilli())

	var environment Environment
	if _, err := env.UnmarshalFromEnviron(&environment); err != nil {
		panic(err)
	}

	mongoClient := sources.NewMongoClient(ctx, environment.MongoUri, "database")
	locationRepository := repository.NewRepository(mongoClient)

	app.Use(cors.New())
	app.Get("/monitor", monitor.New())

	app.Get("/get-location", func(c *fiber.Ctx) error {
		locations := make([]*repository.Location, 0)

		data, exists := cache.Get("locations")
		if !exists {
			lcs, err := tools.GetAllLocations(ctx)
			if err != nil {
				logrus.Errorln(err)

				return c.SendString(err.Error())
			}

			locations = lcs

			cache.SetWithTTL("locations", locations, int64(time.Minute*15), 0)
		} else {
			locations = data.([]*repository.Location)
		}

		locs, err := locationRepository.GetLocations(ctx)
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		for _, l := range locs {
			for i, loc := range locations {
				if l.EntryID == loc.EntryID {
					locations = append(locations[:i], locations[i+1:]...)

					continue
				}
			}
		}

		cityID := c.QueryInt("city_id")
		if cityID > 0 {
			box := cities[cityID]

			filteredLocations := make([]*repository.Location, 0)

			for _, loc := range locations {
				if box[0] >= loc.Loc[0] && box[1] >= loc.Loc[1] && box[2] <= loc.Loc[0] && box[3] <= loc.Loc[1] {
					filteredLocations = append(filteredLocations, loc)
				}
			}

			locations = filteredLocations
		}

		randIndex := rand.Intn(len(locations))
		selected := locations[randIndex]

		singleData, err := tools.GetSingleLocation(ctx, selected.EntryID)
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		selected.OriginalMessage = singleData.FullText
		selected.OriginalLocation = fmt.Sprintf("https://www.google.com/maps/?q=%f,%f&ll=%f,%f&z=21", selected.Loc[0], selected.Loc[1], selected.Loc[0], selected.Loc[1])

		return c.JSON(struct {
			Count    int                  `json:"count"`
			Location *repository.Location `json:"location"`
		}{
			Count:    len(locations),
			Location: selected,
		})
	})

	app.Post("/resolve", func(c *fiber.Ctx) error {
		body := &ResolveBody{}

		if err := json.Unmarshal(c.Body(), body); err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		exists, err := locationRepository.IsResolved(ctx, body.ID)
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		if exists {
			return c.SendString("this location is already checked")
		}

		locations, err := tools.GetAllLocations(ctx)
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		singleData, err := tools.GetSingleLocation(ctx, body.ID)
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		originalLocation := ""
		location := make([]float64, 0)

		for _, loc := range locations {
			if loc.EntryID == body.ID {
				originalLocation = fmt.Sprintf("https://www.google.com/maps/?q=%f,%f&ll=%f,%f&z=21", loc.Loc[0], loc.Loc[1], loc.Loc[0], loc.Loc[1])
				location = loc.Loc
			}
		}

		if err := locationRepository.ResolveLocation(ctx, &repository.LocationDB{
			EntryID:          body.ID,
			Type:             body.LocationType,
			Location:         location,
			Corrected:        body.NewAddress == singleData.FullText,
			OriginalAddress:  originalLocation,
			CorrectedAddress: body.NewAddress,
			Reason:           body.Reason,
			OpenAddress:      body.OpenAddress,
			Apartment:        body.Apartment,
		}); err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		return c.SendString("Added!")
	})

	if err := app.Listen(":80"); err != nil {
		panic(err)
	}
}
