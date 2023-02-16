package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/acikkaynak/veri-toplama-backend/models"
	locationsRepository "github.com/acikkaynak/veri-toplama-backend/repository/locations"
	usersRepository "github.com/acikkaynak/veri-toplama-backend/repository/users"
	"github.com/acikkaynak/veri-toplama-backend/util"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"regexp"
	"strconv"
	"time"
)

type userRepositoryClient interface {
	GetUser(ctx context.Context, authKey string) (*usersRepository.User, error)
}

type locationsRepositoryClient interface {
	ResolveLocation(ctx context.Context, location *locationsRepository.LocationDB) error
}

type resolveService struct {
	locationServiceClient     locationServiceClient
	userRepositoryClient      userRepositoryClient
	locationsRepositoryClient locationsRepositoryClient
}

func NewResolveService(locationServiceClient locationServiceClient, userRepositoryClient userRepositoryClient, locationsRepositoryClient locationsRepositoryClient) *resolveService {
	return &resolveService{
		locationServiceClient:     locationServiceClient,
		userRepositoryClient:      userRepositoryClient,
		locationsRepositoryClient: locationsRepositoryClient,
	}
}

func (s *resolveService) Resolve(c *fiber.Ctx) error {
	logrus.Infoln("Pulling entries")
	locs, err := s.locationServiceClient.GetLocations(c)
	if err != nil {
		logrus.Errorf("Couldn't get all locations: %s", err)
	}

	body, err := s.filterLocations(c, locs)
	if err != nil {
		return err
	}

	originalLocation, location, err := s.getOriginalLocations(body, locs)
	if err != nil {
		return err
	}

	sender, err := s.userRepositoryClient.GetUser(c.Context(), c.Get("Auth-Key"))
	if err != nil {
		return err
	}

	rs := &locationsRepository.LocationDB{
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
	}

	if err := s.locationsRepositoryClient.ResolveLocation(c.Context(), rs); err != nil {
		logrus.Errorln(err)
		return err
	}

	return nil
}

func (s *resolveService) filterLocations(c *fiber.Ctx, locs []*locationsRepository.Location) (*models.ResolveBody, error) {
	processedIDs := make([]int, 0)
	for _, loc := range locs {
		processedIDs = append(processedIDs, loc.EntryID)
	}

	body := &models.ResolveBody{}
	if err := json.Unmarshal(c.Body(), body); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	if err := matchNewAddress(body.NewAddress); err != nil {
		return nil, err
	}

	for _, id := range processedIDs {
		if body.ID == id {
			return nil, fmt.Errorf("this location is already checked")
		}
	}

	return body, nil
}

func (s *resolveService) getOriginalLocations(body *models.ResolveBody, locs []*locationsRepository.Location) (string, []float64, error) {
	originalLocation := ""
	location := make([]float64, 2)

	for _, loc := range locs {
		if loc.EntryID == body.ID {
			originalLocation = getOriginalLocation(loc.Loc)
			location = loc.Loc
		}
	}

	input := []string{"HATA YOK", "Hata Yok", "hata yok"}

	if body.Reason != input[0] && body.Reason != input[1] && body.Reason != input[2] {
		if body.NewAddress != "" {
			longUrl, err := util.GatherLongUrlFromShortUrl(body.NewAddress)
			if err != nil {
				logrus.Errorln(err)
				return "", nil, err
			}

			locUrl := util.URLtoLatLng(longUrl)
			latVal, err := strconv.ParseFloat(locUrl["lat"], 64)
			if err != nil {
				logrus.Error(err)
				return "", nil, err
			}

			lngVal, err := strconv.ParseFloat(locUrl["lng"], 64)
			if err != nil {
				logrus.Error(err)
				return "", nil, err
			}

			location[0] = latVal
			location[1] = lngVal
		}
	}

	return originalLocation, location, nil
}

func matchNewAddress(neAddress string) error {
	match, err := regexp.MatchString("^https://goo.gl/maps/.*", neAddress)
	if err != nil {
		logrus.Error(err)
		return fmt.Errorf("invalid Google Maps URL!")
	}
	if !match {
		return fmt.Errorf("Invalid Google Maps URL!")
	}

	return nil
}

func getOriginalLocation(loc []float64) string {
	return fmt.Sprintf("https://www.google.com/maps/?q=%f,%f&ll=%f,%f&z=21", loc[0], loc[1], loc[0], loc[1])
}
