package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Netflix/go-env"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	usersRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/users"
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
	locationRepository := locationsRepository.NewRepository(mongoClient)
	userRepository := usersRepository.NewRepository(mongoClient)

	admin := NewAdmin(locationRepository, cache)

	app.Use(cors.New())

	entries := app.Group("/entries", func(c *fiber.Ctx) error {
		authKey := c.Get("Auth-Key")

		user, err := userRepository.GetUser(ctx, authKey)
		if err != nil {
			return c.Status(401).SendString("User not found.")
		}

		if user.PermLevel < usersRepository.PermModerator {
			return c.Status(401).SendString("You are not allowed to access here.")
		}

		return c.Next()
	})

	entries.Get("", admin.GetLocationEntries)
	entries.Get("/:entry_id", admin.GetSingleEntry)
	entries.Post("/:entry_id", admin.UpdateEntry)

	app.Get("/monitor", monitor.New())

	app.Get("/get-location", func(c *fiber.Ctx) error {
		locations, err := tools.GetAllLocations(ctx, cache)
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
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

			filteredLocations := make([]*locationsRepository.Location, 0)

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
			Count    int                           `json:"count"`
			Location *locationsRepository.Location `json:"location"`
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

		locations, err := tools.GetAllLocations(ctx, cache)
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

		if err := locationRepository.ResolveLocation(ctx, &locationsRepository.LocationDB{
			EntryID:          body.ID,
			Type:             body.LocationType,
			Location:         location,
			Corrected:        len(body.Reason) > 0,
			OriginalAddress:  originalLocation,
			CorrectedAddress: body.NewAddress,
			Reason:           body.Reason,
			OpenAddress:      body.OpenAddress,
			Apartment:        body.Apartment,
		}); err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		return c.SendString("Successfully added!")
	})

	if err := app.Listen(":80"); err != nil {
		panic(err)
	}
}
