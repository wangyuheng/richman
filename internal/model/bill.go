package model

type Bill struct {
	Remark     string   `json:"remark"`
	Categories []string `json:"categories"`
	Amount     float64  `json:"amount"`
	Month      string   `json:"month"`
	Date       int64    `json:"date"`
	Expenses   string   `json:"expenses"`
	AuthorID   string   `json:"author_id"`
	AuthorName string   `json:"author_name"`
}
