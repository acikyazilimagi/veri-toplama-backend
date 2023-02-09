package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Netflix/go-env"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"time"
	"veri-kontrol-backend/core/network"
	"veri-kontrol-backend/core/sources"
)

type Environment struct {
	MongoUri string `env:"mongo_uri"`
}

type Location struct {
	EntryID          int       `json:"entry_id"`
	Loc              []float64 `json:"loc"`
	OriginalMessage  string    `json:"original_message"`
	OriginalLocation string    `json:"original_location"`
	epoch            int64     `json:"epoch"`
}

type LocationDB struct {
	EntryID          int    `json:"entry_id" bson:"entry_id"`
	Corrected        bool   `json:"corrected" bson:"corrected"`
	OriginalAddress  string `json:"original_address" bson:"original_address"`
	CorrectedAddress string `json:"corrected_address" bson:"corrected_address"`
	Reason           string `json:"reason" bson:"reason"`
}

type SingleResponse struct {
	FullText         string `json:"full_text"`
	FormattedAddress string `json:"formatted_address"`
}

var cities = map[int][]float64{
	1: {36.852702785393014, 36.87286376953126, 36.535570922786015, 35.88409423828126},
	2: {36.2104851748389, 36.81861877441407, 35.84286468375614, 35.82984924316407},
	3: {36.495937096205274, 36.649870522206335, 36.064120488812605, 35.4740187605459},
	4: {36.50903585150776, 36.402143998719424, 36.47976138594277, 36.31474829364722},
	5: {36.64234742932176, 36.3232450328562, 36.53629731173617, 36.029282092441115},
	6: {36.116001873480265, 36.06470054394251, 36.0627178139989, 35.91771907373497},
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

	app.Get("/get-location", func(c *fiber.Ctx) error {
		locations := make([]*Location, 0)

		data, exists := cache.Get("locations")
		if !exists {
			var d struct {
				Locations []*Location `json:"results"`
			}

			res, _, err := network.ProcessGet(ctx, "https://apigo.afetharita.com/feeds/areas?ne_lat=39.91618777305531&ne_lng=47.85149904303703&sw_lat=36.07272886939253&sw_lng=23.872389299415502", nil)
			if err != nil {
				logrus.Errorln(err)

				return c.SendString(err.Error())
			}

			if err := json.Unmarshal(res, &d); err != nil {
				logrus.Errorln(err)

				return c.SendString(err.Error())
			}

			locations = d.Locations

			cache.SetWithTTL("locations", locations, int64(time.Minute*15), 0)
		} else {
			locations = data.([]*Location)
		}

		cur, err := mongoClient.Find(ctx, "locations", bson.E{})
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		locs := make([]*LocationDB, 0)
		if err := cur.All(ctx, &locs); err != nil {
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

			filteredLocations := make([]*Location, 0)

			for _, loc := range locations {
				if box[0] >= loc.Loc[0] && box[1] >= loc.Loc[1] && box[2] <= loc.Loc[0] && box[3] <= loc.Loc[1] {
					filteredLocations = append(filteredLocations, loc)
				}
			}

			locations = filteredLocations
		}

		randIndex := rand.Intn(len(locations))
		selected := locations[randIndex]

		resp, _, err := network.ProcessGet(ctx, fmt.Sprintf("https://apigo.afetharita.com/feeds/%d", selected.EntryID), nil)
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		singleData := &SingleResponse{}
		if err := json.Unmarshal(resp, singleData); err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		selected.OriginalMessage = singleData.FullText
		selected.OriginalLocation = singleData.FormattedAddress

		return c.JSON(struct {
			Count    int
			Location *Location
		}{
			Count:    len(locations),
			Location: selected,
		})
	})

	app.Get("/resolve", func(c *fiber.Ctx) error {
		id := c.QueryInt("id")
		newAddress := c.Query("new_address")
		reason := c.Query("reason")

		resp, _, err := network.ProcessGet(ctx, fmt.Sprintf("https://apigo.afetharita.com/feeds/%d", id), nil)
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		singleData := &SingleResponse{}
		if err := json.Unmarshal(resp, singleData); err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		exists, err := mongoClient.DoesExist(ctx, "locations", bson.D{bson.E{
			Key:   "entry_id",
			Value: id,
		}})
		if err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		if exists {
			return c.SendString("this location is already checked")
		}

		data := &LocationDB{
			EntryID:          id,
			Corrected:        newAddress == singleData.FullText,
			OriginalAddress:  singleData.FormattedAddress,
			CorrectedAddress: newAddress,
			Reason:           reason,
		}

		if err := mongoClient.InsertOne(ctx, "locations", data); err != nil {
			logrus.Errorln(err)

			return c.SendString(err.Error())
		}

		return c.SendString("Added!")
	})

	app.Listen(":3000")
}
