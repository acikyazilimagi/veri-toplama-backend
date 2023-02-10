package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/Netflix/go-env"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
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

	locationRepository := locations.NewRepository(mongoClient)

	files, err := os.ReadDir("merge_data")
	if err != nil {
		panic(err)
	}

	locs, err := tools.GetAllLocations(context.Background(), sources.NewCache(1<<30, 1e7, 64))
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		log.Infof("Starting merging of file %s", file.Name())

		f, err := os.Open(fmt.Sprintf("merge_data/%s", file.Name()))
		if err != nil {
			panic(err)
		}

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

			location := make([]float64, 0)

			for _, loc := range locs {
				if loc.EntryID == int(id) {
					location = loc.Loc
				}
			}

			data := &locations.LocationDB{
				EntryID:          int(id),
				Corrected:        len(rec[3]) > 0,
				Location:         location,
				OriginalAddress:  rec[1],
				CorrectedAddress: rec[2],
				Reason:           rec[3],
			}

			if len(rec) > 4 {
				data.OpenAddress = rec[4]
			}

			if len(rec) > 5 {
				data.Apartment = rec[5]
			}

			if err := locationRepository.ResolveLocation(ctx, data); err != nil {
				panic(err)
			}

			log.Infof("%d/%d", i, len(recs)-1)
		}

		if err := f.Close(); err != nil {
			panic(err)
		}
	}
}
