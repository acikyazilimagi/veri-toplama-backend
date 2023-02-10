package users

import (
	"context"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type Repository interface {
}

type repository struct {
	mongo sources.MongoClient
}

func NewRepository(mongo sources.MongoClient) Repository {
	return &repository{
		mongo: mongo,
	}
}

const (
	PermSubmit   = 1
	PermValidate = 2
	PermAdmin    = 999
)

type User struct {
	Name        string `json:"name" bson:"name"`
	Discord     string `json:"discord" bson:"discord"`
	AuthKeyHash uint32 `json:"auth_key_hash" bson:"auth_key_hash"`
	PermLevel   int    `json:"perm_level" bson:"perm_level"`
}

func (r *repository) GetUser(ctx context.Context, authKey string) (*User, error) {
	cur := r.mongo.FindOne(ctx, "users", bson.E{
		Key:   "auth_key_hash",
		Value: tools.Hash(authKey),
	})

	user := &User{}
	if err := cur.Decode(user); err != nil {
		logrus.Errorln(err)
		return nil, err
	}

	return user, nil
}

func (r *repository) AddUser(ctx context.Context, name, discord string, permLevel int) (string, error) {
	authKey := tools.RandomString(32)

	if err := r.mongo.InsertOne(ctx, "users", &User{
		Name:        name,
		Discord:     discord,
		AuthKeyHash: tools.Hash(authKey),
		PermLevel:   permLevel,
	}); err != nil {
		logrus.Errorln(err)

		return "", err
	}

	return authKey, nil
}
