package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/acikkaynak/veri-toplama-backend/models"
	"github.com/acikkaynak/veri-toplama-backend/repository/locations"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"strconv"
)

type adminLocationRepository interface {
	GetLocations(ctx context.Context) ([]*locations.LocationDB, error)
	ResolveLocation(ctx context.Context, location *locations.LocationDB) error
	IsResolved(ctx context.Context, locationID int) (bool, error)
	IsDuplicate(ctx context.Context, tweetContents string) (bool, error)
	GetDocumentsWithNoTweetContents(ctx context.Context) ([]*locations.LocationDB, error)
}

type adminFeedService interface {
	GeetFeeds(c *fiber.Ctx) (*FeetFeedsResponse, error)
}

type Admin interface {
	GetLocationEntries(c *fiber.Ctx) error
	GetSingleEntry(c *fiber.Ctx) error
	UpdateEntry(c *fiber.Ctx) error
}

type admin struct {
	locations             adminLocationRepository
	locationServiceClient locationServiceClient
}

func NewAdmin(locations adminLocationRepository, locationServiceClient locationServiceClient) Admin {
	return &admin{
		locations:             locations,
		locationServiceClient: locationServiceClient,
	}
}

func (a *admin) GetLocationEntries(c *fiber.Ctx) error {
	entries, err := a.locations.GetLocations(c.Context())
	if err != nil {
		return c.SendString(err.Error())
	}

	return c.JSON(entries)
}

func (a *admin) GetSingleEntry(c *fiber.Ctx) error {
	entryID, _ := strconv.ParseInt(c.Params("entry_id"), 10, 32)

	entries, err := a.locations.GetLocations(c.Context())
	if err != nil {
		return c.SendString(err.Error())
	}

	for _, entry := range entries {
		if entry.EntryID == int(entryID) {
			return c.JSON(entry)
		}
	}

	return c.SendString("Entry not found.")
}

func (a *admin) UpdateEntry(c *fiber.Ctx) error {
	body := &models.ResolveBody{}

	if err := json.Unmarshal(c.Body(), body); err != nil {
		return c.SendString(err.Error())
	}

	locs, err := a.locationServiceClient.GetLocations(c)
	if err != nil {
		logrus.Errorln(err)
		return c.SendString(err.Error())
	}

	originalLocation := ""
	location := make([]float64, 0)

	for _, loc := range locs {
		if loc.EntryID == body.ID {
			originalLocation = fmt.Sprintf("https://www.google.com/maps/?q=%f,%f&ll=%f,%f&z=21", loc.Loc[0], loc.Loc[1], loc.Loc[0], loc.Loc[1])
			location = loc.Loc
		}
	}

	if err := a.locations.ResolveLocation(c.Context(), &locations.LocationDB{
		EntryID:          body.ID,
		Type:             body.LocationType,
		Location:         location,
		Corrected:        body.Reason == "Hata Yok",
		Verified:         true,
		OriginalAddress:  originalLocation,
		CorrectedAddress: body.NewAddress,
		Reason:           body.Reason,
		OpenAddress:      body.OpenAddress,
		Apartment:        body.Apartment,
	}); err != nil {
		logrus.Errorln(err)
		return c.SendString(err.Error())
	}

	return c.SendString("")
}
