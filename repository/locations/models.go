package locations

import (
	"github.com/acikkaynak/veri-toplama-backend/repository/users"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Location struct {
	EntryID          int       `json:"entry_id"`
	Loc              []float64 `json:"loc"`
	Epoch            int       `json:"epoch"`
	OriginalMessage  string    `json:"original_message"`
	OriginalLocation string    `json:"original_location"`
}

type LocationDB struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	EntryID          int                `json:"entry_id" bson:"entry_id"`
	Sender           *users.User        `json:"sender" bson:"sender"`
	Location         []float64          `json:"location" bson:"location"`
	Corrected        bool               `json:"corrected" bson:"corrected"`
	Verified         bool               `json:"verified" bson:"verified"`
	OriginalAddress  string             `json:"original_address" bson:"original_address"`
	CorrectedAddress string             `json:"corrected_address" bson:"corrected_address"`
	OpenAddress      string             `json:"open_address" bson:"open_address"`
	Apartment        string             `json:"apartment" bson:"apartment"`
	Type             int                `json:"type" bson:"type"`
	Reason           string             `json:"reason" bson:"reason"`
	TweetContents    string             `json:"tweet_contents" bson:"tweet_contents"`
}
