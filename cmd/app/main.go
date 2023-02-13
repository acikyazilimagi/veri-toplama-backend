package main

import (
	"context"
	"github.com/Netflix/go-env"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/handler"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	usersRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/users"
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
	1:  {36.852702785393014, 36.87286376953126, 36.535570922786015, 35.88409423828126},
	2:  {36.2104851748389, 36.81861877441407, 35.84286468375614, 35.82984924316407},
	3:  {36.495937096205274, 36.649870522206335, 36.064120488812605, 35.4740187605459},
	4:  {36.50903585150776, 36.402143998719424, 36.47976138594277, 36.31474829364722},
	5:  {36.64234742932176, 36.3232450328562, 36.53629731173617, 36.029282092441115},
	6:  {36.116001873480265, 36.06470054394251, 36.0627178139989, 35.91771907373497},
	7:  {38.306102934215616, 38.11981201171876, 37.29478306827433, 35.481778157374286},
	8:  {37.35461473302187, 38.0755896764663, 36.85431769725969, 36.67725839531126},
	9:  {39.065058845523424, 40.013647871307754, 37.86798402826048, 36.687836853946884},
	10: {38.160827052916495, 39.33362355320935, 37.44250898099215, 37.35608449070936},
}

type ResolveBody struct {
	ID            int    `json:"id"`
	LocationType  int    `json:"type"`
	NewAddress    string `json:"new_address"`
	OpenAddress   string `json:"open_address"`
	Apartment     string `json:"apartment"`
	Reason        string `json:"reason"`
	TweetContents string `json:"tweet_contents"`
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

	processedIDs := make([]int, 0)

	logrus.Infoln("Pulling entries")
	locs, err := locationRepository.GetLocations(ctx)
	if err != nil {
		logrus.Errorf("Couldn't get all locations: %s", err)
	}

	for _, loc := range locs {
		processedIDs = append(processedIDs, loc.EntryID)
	}
	logrus.Infoln("Startup complete")

	app.Use(cors.New())

	adminG := app.Group("/admin", func(c *fiber.Ctx) error {
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

	entriesG := adminG.Group("/entries")

	entriesG.Get("", admin.GetLocationEntries)
	entriesG.Get("/:entry_id", admin.GetSingleEntry)
	entriesG.Post("/:entry_id", admin.UpdateEntry)

	app.Get("/monitor", monitor.New())

	app.Get("/get-location", handler.GetLocationHandler(ctx,
		locationRepository,
		cache,
		processedIDs))

	app.Post("/resolve", handler.ResolveValidationHandler(ctx,
		locationRepository,
		userRepository,
		cache,
		processedIDs))

	if err := app.Listen(":3000"); err != nil {
		panic(err)
	}
}
