package handler

import (
	"fmt"
	"github.com/YusufOzmen01/veri-kontrol-backend/util"

	"context"
	"encoding/json"

	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	usersRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/users"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"regexp"
	"strconv"
	"time"
)

type ResolveBody struct {
	ID            int    `json:"id"`
	LocationType  int    `json:"type"`
	NewAddress    string `json:"new_address"`
	OpenAddress   string `json:"open_address"`
	Apartment     string `json:"apartment"`
	Reason        string `json:"reason"`
	TweetContents string `json:"tweet_contents"`
}

func ResolveValidationHandler(ctx context.Context, locationRepository locations.Repository, userRepository usersRepository.Repository, cache sources.Cache, processedIDs []int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		body := &ResolveBody{}

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
}
