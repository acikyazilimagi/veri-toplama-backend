package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/models"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	usersRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/users"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
	"github.com/YusufOzmen01/veri-kontrol-backend/util"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"
)

func ResolveHandler(c *fiber.Ctx) error {
	ctx := context.Background()
	cache := sources.NewCache(1<<30, 1e7, 64)
	rand.Seed(time.Now().UnixMilli())

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		panic("mongo URL is empty")
	}

	mongoClient := sources.NewMongoClient(ctx, mongoURL, "database")
	locationRepository := locationsRepository.NewRepository(mongoClient)
	userRepository := usersRepository.NewRepository(mongoClient)

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

	body := &models.ResolveBody{}

	if err := json.Unmarshal(c.Body(), body); err != nil {
		logrus.Errorln(err)

		return c.SendString(err.Error())
	}

	match, err := regexp.MatchString("^https://goo.gl/maps/.*", body.NewAddress)
	if err != nil {
		logrus.Error(err)
		return err
	}
	if !match {
		return c.SendString("Invalid Google Maps URL!")
	}

	for _, id := range processedIDs {
		if body.ID == id {
			return c.SendString("this location is already checked")
		}
	}

	locations, err := tools.GetAllLocations(ctx, cache)
	if err != nil {
		logrus.Errorln(err)

		return c.SendString(err.Error())
	}

	originalLocation := ""
	location := make([]float64, 2)

	for _, loc := range locations {
		if loc.EntryID == body.ID {
			originalLocation = fmt.Sprintf("https://www.google.com/maps/?q=%f,%f&ll=%f,%f&z=21", loc.Loc[0], loc.Loc[1], loc.Loc[0], loc.Loc[1])
			location = loc.Loc
		}
	}

	var sender *usersRepository.User

	authKey := c.Get("Auth-Key")
	userData, err := userRepository.GetUser(c.Context(), authKey)
	if err == nil {
		sender = userData
	}
	input := []string{"HATA YOK", "Hata Yok", "hata yok"}

	if body.Reason != input[0] && body.Reason != input[1] && body.Reason != input[2] {
		if body.NewAddress != "" {

			longUrl, err := util.GatherLongUrlFromShortUrl(body.NewAddress)
			if err != nil {
				logrus.Errorln(err)
				return c.SendString(err.Error())
			}
			locUrl := util.URLtoLatLng(longUrl)
			latVal, err := strconv.ParseFloat(locUrl["lat"], 64)
			if err != nil {
				logrus.Error(err)
				c.SendString(err.Error())

			}
			lngVal, err := strconv.ParseFloat(locUrl["lng"], 64)
			if err != nil {
				logrus.Error(err)
				c.SendString(err.Error())
			}
			location[0] = latVal
			location[1] = lngVal

		}

	}

	if err := locationRepository.ResolveLocation(ctx, &locationsRepository.LocationDB{
		ID:               primitive.NewObjectIDFromTimestamp(time.Now()),
		EntryID:          body.ID,
		Type:             body.LocationType,
		Location:         location,
		Corrected:        body.Reason == "Hata Yok",
		OriginalAddress:  originalLocation,
		CorrectedAddress: body.NewAddress,
		Reason:           body.Reason,
		Sender:           sender,
		OpenAddress:      body.OpenAddress,
		Apartment:        body.Apartment,
		TweetContents:    body.TweetContents,
	}); err != nil {
		logrus.Errorln(err)

		return c.SendString(err.Error())
	}

	processedIDs = append(processedIDs, body.ID)

	return c.SendString("Successfully added!")
}
