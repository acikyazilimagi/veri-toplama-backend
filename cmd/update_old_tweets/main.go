package main

import (
	"context"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	locationsRepository "github.com/YusufOzmen01/veri-kontrol-backend/repository/locations"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type Environment struct {
	MongoUri string `env:"mongo_uri"`
}

type repository struct {
	mongo sources.MongoClient
}

type Location struct {
	EntryID          int       `json:"entry_id"`
	Loc              []float64 `json:"loc"`
	OriginalMessage  string    `json:"original_message"`
	OriginalLocation string    `json:"original_location"`
}

type LocationDB struct {
	EntryID          int       `json:"entry_id" bson:"entry_id"`
	Location         []float64 `json:"location" bson:"location"`
	Corrected        bool      `json:"corrected" bson:"corrected"`
	Verified         bool      `json:"verified" bson:"verified"`
	OriginalAddress  string    `json:"original_address" bson:"original_address"`
	CorrectedAddress string    `json:"corrected_address" bson:"corrected_address"`
	OpenAddress      string    `json:"open_address" bson:"open_address"`
	Apartment        string    `json:"apartment" bson:"apartment"`
	Type             int       `json:"type" bson:"type"`
	Reason           string    `json:"reason" bson:"reason"`
	TweetContents    string    `json:"tweet_contents" bson:"tweet_contents"`
}

func main() {
	ctx := context.Background()
	var environment Environment

	mongoClient := sources.NewMongoClient(ctx, environment.MongoUri, "database")

	locationRepository := locationsRepository.NewRepository(mongoClient)

	locs, err := locationRepository.UpdateTweetContents(ctx)

	if err != nil {
		logrus.Errorln(err)
		return
	}

	for i := 0; i < len(locs); i++ {
		//locs[i].EntryID to send request to api and hydrate the db with tweet_contents
		resp, err := tools.GetSingleLocation(ctx, locs[i].EntryID)

		if err != nil {
			logrus.Errorln(err)
			return
		}
		// add resp.fulltext to db
		mongoClient.UpdateOne(ctx, "locations", bson.D{{
			Key:   "entry_id",
			Value: locs[i].EntryID,
		}}, bson.D{{
			Key:   "tweet_contents",
			Value: resp.FullText,
		}})
	}
}
