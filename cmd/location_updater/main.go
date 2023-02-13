package main

import (
	"context"
	"fmt"
	"github.com/Netflix/go-env"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	"github.com/YusufOzmen01/veri-kontrol-backend/util"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
)

type Environment struct {
	MongoUri string `env:"mongo_uri"`
}

func main() {
	ctx := context.Background()
	var environment Environment

	if _, err := env.UnmarshalFromEnviron(&environment); err != nil {
		panic(err)
	}

	_ = sources.NewCache(1<<30, 1e7, 64)

	mongoClient := sources.NewMongoClient(ctx, environment.MongoUri, "database")
	locationRepository := locationsRepository.NewRepository(mongoClient)

	locs, err := locationRepository.GetNewFormatDocuments(ctx)

	if err != nil {
		panic(err)
	}

	for i, loc := range locs {
		if len(loc.CorrectedAddress) < 1 {
			locs = append(locs[:i], locs[i+1:]...)
		}
	}

	fmt.Println(len(locs))

	chunks := lo.Chunk(locs, 10)

	wg := &sync.WaitGroup{}

	for i, chunk := range chunks {
		//locs[i].EntryID to send request to api and hydrate the db with tweet_contents
		for _, l := range chunk {
			wg.Add(1)

			go func(loc *locationsRepository.LocationDB) {
				defer wg.Done()
				longUrl, err := util.GatherLongUrlFromShortUrl(loc.CorrectedAddress)
				if err != nil {
					logrus.Errorln(err)
					return
				}
				locUrl := util.URLtoLatLng(longUrl)
				latVal, err := strconv.ParseFloat(locUrl["lat"], 64)
				if err != nil {
					logrus.Error(err)
					return
				}
				lngVal, err := strconv.ParseFloat(locUrl["lng"], 64)
				if err != nil {
					logrus.Error(err)
					return

				}
				loc.Location[0] = latVal
				loc.Location[1] = lngVal

				if err := locationRepository.ResolveLocation(ctx, loc); err != nil {
					panic(err)
				}

			}(l)
		}

		wg.Wait()

		logrus.Infof("%d/%d complete.", i, len(chunks))
	}

	/*locs, err := locationRepository.GetDocumentsWithNoTweetContents(ctx)

	if err != nil {
		panic(err)
	}

	chunks := lo.Chunk(locs, 10)

	wg := &sync.WaitGroup{}

	for i, chunk := range chunks {
		//locs[i].EntryID to send request to api and hydrate the db with tweet_contents
		for _, l := range chunk {
			wg.Add(1)

			go func(loc *locationsRepository.LocationDB) {
				defer wg.Done()

				resp, err := tools.GetSingleLocation(ctx, loc.EntryID, cache)
				if err != nil {
					panic(err)
				}

				loc.TweetContents = resp.FullText

				if err := locationRepository.ResolveLocation(ctx, loc); err != nil {
					panic(err)
				}
			}(l)
		}

		wg.Wait()

		logrus.Infof("%d/%d complete.", i, len(chunks))*/

	//wg.Wait()
}
