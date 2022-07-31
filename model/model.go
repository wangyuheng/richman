package model

type Category string

const (
	CategoryWechat Category = "wechat"
	CategoryFeishu Category = "feishu"
)

type App struct {
	AppId             string `json:"id"`
	AppSecret         string `json:"secret"`
	VerificationToken string `json:"token"`
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

type Author struct {
	AppId        string `json:"app_id"`
	FeishuOpenId string `json:"feishu_open_id"`
	WechatOpenId string `json:"wechat_open_id"`
}

type Book struct {
	AppId    string   `json:"app_id"`
	AppToken string   `json:"app_token"`
	OpenId   string   `json:"open_id"`
	Category Category `json:"category"`
}

type Dream struct {
	Id       string  `json:"id"`
	Keyword  string  `json:"keyword"`
	Target   float64 `json:"target"`
	CurVal   float64 `json:"cur_val"`
	Progress string  `json:"progress"`
}

type DreamRecord struct {
	Id      string  `json:"id"`
	Keyword string  `json:"keyword"`
	Amount  float64 `json:"amount"`
	Date    int64   `json:"date"`
	Maker   string  `json:"maker"`
}
