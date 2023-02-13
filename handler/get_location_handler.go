package handler

import (
	"context"
	"fmt"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"math/rand"
)

var cities = map[int][]float64{
	1:  {36.852702785393014, 36.87286376953126, 36.535570922786015, 35.88409423828126},
	2:  {36.2104851748389, 36.81861877441407, 35.84286468375614, 35.82984924316407},
	3:  {36.495937096205274, 36.649870522206335, 36.064120488812605, 35.4740187605459},
	4:  {36.50903585150776, 36.402143998719424, 36.47976138594277, 36.31474829364722},
	5:  {36.64234742932176, 36.3232450328562, 36.53629731173617, 36.029282092441115},
	6:  {36.116001873480265, 36.06470054394251, 36.0627178139989, 35.91771907373497},
	7:  {38.53348725642158, 38.78062516773912, 37.32756763881127, 35.45481415037825},
	8:  {37.35461473302187, 38.0755896764663, 36.85431769725969, 36.67725839531126},
	9:  {39.065058845523424, 40.013647871307754, 37.86798402826048, 36.687836853946884},
	10: {38.160827052916495, 39.33362355320935, 37.44250898099215, 37.35608449070936},
}

func GetLocationHandler(ctx context.Context, locationRepository locations.Repository, cache sources.Cache, processedIDs []int) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
			box := cities[cityID]

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
}
