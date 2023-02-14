package locations

import (
	"context"
	"github.com/acikkaynak/veri-toplama-backend/db/sources"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	TypeWreckage   = 1
	TypeSupplyHelp = 2
)

type repository struct {
	mongo sources.MongoClient
}

func NewRepository(mongo sources.MongoClient) *repository {
	return &repository{
		mongo: mongo,
	}
}

func (r *repository) GetLocations(ctx context.Context) ([]*LocationDB, error) {
	cur, err := r.mongo.Find(ctx, "locations", bson.D{})
	if err != nil {
		return nil, err
	}

	locs := make([]*LocationDB, 0)
	if err := cur.All(ctx, &locs); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	return locs, nil
}

func (r *repository) ResolveLocation(ctx context.Context, location *LocationDB) error {
	if err := r.mongo.DeleteOne(ctx, "locations", bson.D{{
		Key:   "entry_id",
		Value: location.EntryID,
	}}); err != nil {
		logrus.Errorln(err)

		return err
	}

	if err := r.mongo.InsertOne(ctx, "locations", location); err != nil {
		logrus.Errorln(err)

		return err
	}

	return nil
}

func (r *repository) IsResolved(ctx context.Context, locationID int) (bool, error) {
	exists, err := r.mongo.DoesExist(ctx, "locations", bson.D{{
		Key:   "entry_id",
		Value: locationID,
	}})
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *repository) IsDuplicate(ctx context.Context, tweetContents string) (bool, error) {
	exists, err := r.mongo.DoesExist(ctx, "locations", bson.D{{
		Key:   "tweet_contents",
		Value: tweetContents,
	}})
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *repository) GetDocumentsWithNoTweetContents(ctx context.Context) ([]*LocationDB, error) {
	// FOR OLD DB COLLECTIONS ONLY, update the tweet_contents for old tweet data where it does not exist, or is empty
	// Do not use in app
	cur, err := r.mongo.Find(ctx, "locations", bson.D{{
		Key:   "tweet_contents",
		Value: nil,
	}})
	if err != nil {
		return nil, err
	}

	locs := make([]*LocationDB, 0)
	if err := cur.All(ctx, &locs); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	return locs, nil
}
