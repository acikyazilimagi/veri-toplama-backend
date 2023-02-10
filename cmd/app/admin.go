package main

import (
	"encoding/json"
	"fmt"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"strconv"
)

type Admin interface {
	GetLocationEntries(c *fiber.Ctx) error
	GetSingleEntry(c *fiber.Ctx) error
	UpdateEntry(c *fiber.Ctx) error
}

type admin struct {
	locations locations.Repository
	cache     sources.Cache
}

func NewAdmin(locations locations.Repository, cache sources.Cache) Admin {
	return &admin{
		locations: locations,
		cache:     cache,
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
	body := &ResolveBody{}

	if err := json.Unmarshal(c.Body(), body); err != nil {
		return c.SendString(err.Error())
	}

	locs, err := tools.GetAllLocations(c.Context(), a.cache)
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
