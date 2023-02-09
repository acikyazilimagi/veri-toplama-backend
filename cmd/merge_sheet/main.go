package main

import (
	"context"
	"encoding/csv"
	"github.com/Netflix/go-env"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/repository"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

type Environment struct {
	MongoUri string `env:"mongo_uri"`
}

func main() {
	var environment Environment
	ctx := context.Background()

	if _, err := env.UnmarshalFromEnviron(&environment); err != nil {
		panic(err)
	}

	mongoClient := sources.NewMongoClient(ctx, environment.MongoUri, "database")

	locationRepository := repository.NewRepository(mongoClient)

	f, err := os.Open("data.csv")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)

	recs, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}

	for i, rec := range recs {
		if i == 0 {
			continue // Header kısmını geçmek için
		}

		id, _ := strconv.ParseInt(rec[0], 10, 32)

		data := &repository.LocationDB{
			EntryID:          int(id),
			Corrected:        true,
			OriginalAddress:  rec[1],
			CorrectedAddress: rec[2],
			Reason:           rec[3],
			OpenAddress:      rec[4],
			Apartment:        rec[5],
		}

		if err := locationRepository.ResolveLocation(ctx, data); err != nil {
			panic(err)
		}

		log.Infof("%d/%d", i, len(recs)-1)
	}
}
