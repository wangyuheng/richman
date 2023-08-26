package domain

type Ledger struct {
	ID          string `json:"id"`
	AppToken    string `json:"app_token"`
	TableToken  string `json:"table_token"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	CreatorID   string `json:"creator_id"`
	CreatorName string `json:"creator_name"`
}
