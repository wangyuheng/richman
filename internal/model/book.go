package model

type Book struct {
	AppToken    string `json:"app_token"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	CreatorID   string `json:"creator_id"`
	CreatorName string `json:"creator_name"`
}
