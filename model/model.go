package model

type App struct {
	AppId             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	VerificationToken string `json:"verification_token"`
}

type Bill struct {
	Remark     string   `json:"remark"`
	Categories []string `json:"categories"`
	Amount     float64  `json:"amount"`
	Month      string   `json:"month"`
	Date       int64    `json:"date"`
	Expenses   string   `json:"expenses"`
	AuthorID   string   `json:"author_id"`
}
