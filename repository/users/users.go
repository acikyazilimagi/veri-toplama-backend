package users

import (
	"context"
	"fmt"
	"github.com/acikkaynak/veri-toplama-backend/db/sources"
	"github.com/acikkaynak/veri-toplama-backend/util"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	PermSubmit    = 1
	PermModerator = 2
)

type repository struct {
	mongo sources.MongoClient
}

func NewRepository(mongo sources.MongoClient) *repository {
	return &repository{
		mongo: mongo,
	}
}

func (r *repository) GetUser(ctx context.Context, authKey string) (*User, error) {
	cur, err := r.mongo.Find(ctx, "users", bson.D{})
	if err != nil {
		return nil, err
	}

	// Saçma bir yöntem farkındayım ama filtreler çalışmadığından böyle yapmak zorunda kaldım

	user := make([]*User, 0)
	if err := cur.All(ctx, &user); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	for _, u := range user {
		if u.AuthKeyHash == util.Hash(authKey) {
			return u, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (r *repository) AddUser(ctx context.Context, name, discord string, permLevel int) (string, error) {
	authKey := util.RandomString(32)

	if err := r.mongo.InsertOne(ctx, "users", &User{
		Name:        name,
		Discord:     discord,
		AuthKeyHash: util.Hash(authKey),
		PermLevel:   permLevel,
	}); err != nil {
		logrus.Errorln(err)

		return "", err
	}

	return authKey, nil
}
