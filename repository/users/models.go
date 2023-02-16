package users

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Discord     string             `json:"discord" bson:"discord"`
	AuthKeyHash uint32             `json:"auth_key_hash" bson:"auth_key_hash"`
	PermLevel   int                `json:"perm_level" bson:"perm_level"`
}
