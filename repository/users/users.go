package users

import (
	"context"
	"fmt"
	"github.com/YusufOzmen01/veri-kontrol-backend/core/sources"
	"github.com/YusufOzmen01/veri-kontrol-backend/tools"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Repository interface {
	GetUser(ctx context.Context, authKey string) (*User, error)
	AddUser(ctx context.Context, name, discord string, permLevel int) (string, error)
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
	PermSubmit    = 1
	PermModerator = 2
)

type User struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Discord     string             `json:"discord" bson:"discord"`
	AuthKeyHash uint32             `json:"auth_key_hash" bson:"auth_key_hash"`
	PermLevel   int                `json:"perm_level" bson:"perm_level"`
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
		if u.AuthKeyHash == tools.Hash(authKey) {
			return u, nil
		}
	}

	return nil, fmt.Errorf("user not found")
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
