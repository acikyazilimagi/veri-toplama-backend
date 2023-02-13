package main

import (
	"context"
	app2 "github.com/YusufOzmen01/veri-kontrol-backend/app"
	"github.com/YusufOzmen01/veri-kontrol-backend/middleware"
	"github.com/YusufOzmen01/veri-kontrol-backend/services"
	"github.com/dgraph-io/ristretto"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/handler"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	"github.com/sirupsen/logrus"
)

func main() {
	rand.Seed(time.Now().UnixMilli())
	app := app2.NewApp()
	middleware.Middlewares(app.App)
	app.Register()

	ctx := context.Background()

	cache, err := ristretto.NewCache(&ristretto.Config{
		MaxCost:     1 << 30,
		NumCounters: 1e7,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UnixMilli())

	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		panic("mongo URL is empty")
	}

	// http client
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	mongoClient := sources.NewMongoClient(ctx, mongoURL, "database")
	locationRepository := locationsRepository.NewRepository(mongoClient)
	//userRepository := usersRepository.NewRepository(mongoClient)

	// services
	feeedServcie := services.NewFeedsServices(locationRepository, client, cache)
	admin := services.NewAdmin(locationRepository, cache)

	logrus.Infoln("Startup complete")

	app.App.Get("/get-location", handler.GetFeeds(feeedServcie))
	app.App.Post("/resolve", handler.ResolveHandler)

	adminG := app.App.Group("/admin", handler.AdminHandler)
	// TODO: soralım
	entriesG := adminG.Group("/entries")
	entriesG.Get("", admin.GetLocationEntries)
	entriesG.Get("/:entry_id", admin.GetSingleEntry)
	entriesG.Post("/:entry_id", admin.UpdateEntry)

	app.Run()
}
