package main

import (
	"context"
	app2 "github.com/acikkaynak/veri-toplama-backend/app"
	"github.com/acikkaynak/veri-toplama-backend/db/sources"
	usersRepository "github.com/acikkaynak/veri-toplama-backend/repository/users"
	"github.com/acikkaynak/veri-toplama-backend/services"
	"github.com/dgraph-io/ristretto"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/acikkaynak/veri-toplama-backend/handler"
	locationsRepository "github.com/acikkaynak/veri-toplama-backend/repository/locations"
)

// TODO: postresql eklenecek
func main() {
	// neden ihtiyacımız var?
	rand.Seed(time.Now().UnixMilli())

	app := app2.NewApp()

	app.SetMiddlewares()
	app.Register()

	cache, err := ristretto.NewCache(&ristretto.Config{
		MaxCost:     1 << 30,
		NumCounters: 1e7,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		panic("mongo URL is empty")
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	mongoClient := sources.NewMongoClient(context.TODO(), mongoURL, "database")

	locationRepository := locationsRepository.NewRepository(mongoClient)
	userRepository := usersRepository.NewRepository(mongoClient)

	locationService := services.NewLocationService(client, cache)
	feedService := services.NewFeedsServices(locationRepository, locationService)
	resolveService := services.NewResolveService(locationService, userRepository, locationRepository)
	admin := services.NewAdmin(locationRepository, locationService)

	app.App.Get("/get-location", handler.GetFeeds(feedService))
	app.App.Post("/resolve", handler.Resolve(resolveService))

	adminG := app.App.Group("/admin", handler.AdminHandler)
	entriesG := adminG.Group("/entries")
	entriesG.Get("", admin.GetLocationEntries)
	entriesG.Get("/:entry_id", admin.GetSingleEntry)
	entriesG.Post("/:entry_id", admin.UpdateEntry)

	app.Run(":80")
}
