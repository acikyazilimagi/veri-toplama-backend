package models

type ResolveBody struct {
	ID            int    `json:"id"`
	LocationType  int    `json:"type"`
	NewAddress    string `json:"new_address"`
	OpenAddress   string `json:"open_address"`
	Apartment     string `json:"apartment"`
	Reason        string `json:"reason"`
	TweetContents string `json:"tweet_contents"`
}
