package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	usersRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/users"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
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

func ResolveValidationHandler(ctx context.Context, locationRepository locations.Repository, userRepository usersRepository.Repository, cache sources.Cache) fiber.Handler {
	return func(c *fiber.Ctx) error {
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

		var sender *usersRepository.User

		authKey := c.Get("Auth-Key")
		userData, err := userRepository.GetUser(c.Context(), authKey)
		if err == nil {
			sender = userData
		}

		if err := locationRepository.ResolveLocation(ctx, &locationsRepository.LocationDB{
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

		return c.SendString("Successfully added!")
	}
}
