package biz

import (
	"context"
	"fmt"
	"strings"

	"github.com/geeklubcn/feishu-bitable-db/db"
)

const (
	bookDatabase = "Richman"
	bookTable    = "books"
	bookAppToken = "app_token"
	bookOpenId   = "open_id"
)

type Book interface {
	GetAppTokenByOpenId(openId string) (string, bool)
	Save(openId, url string) (string, error)
}

type book struct {
	db db.DB
}

func NewBook(appId, appSecret string) Book {
	ctx := context.Background()
	it, _ := db.NewDB(appId, appSecret)
	_, _ = it.SaveTable(ctx, bookDatabase, db.Table{
		Name: "books",
		Fields: []db.Field{
			{Name: bookAppToken, Type: db.String},
			{Name: bookOpenId, Type: db.String},
		},
	})
	return book{it}
}

func (b book) Save(openId, url string) (string, error) {
	ctx := context.Background()

	ss := strings.Split(url, "feishu.cn/base/")
	if len(ss) < 2 {
		return "", fmt.Errorf("url[%s] format illegal", url)
	}
	s := ss[1]
	l := strings.Index(s, "?")
	if l2 := strings.Index(s, "/"); l2 > 0 && l2 < l {
		l = l2
	}

	appToken := s[0:l]

	for _, r := range b.db.Read(ctx, bookDatabase, bookTable, []db.SearchCmd{
		{bookOpenId, "=", openId},
	}) {
		_ = b.db.Delete(ctx, bookDatabase, bookTable, r["id"].(string))
	}

	return b.db.Create(ctx, bookDatabase, bookTable, map[string]interface{}{
		bookAppToken: appToken,
		bookOpenId:   openId,
	})
}

func (b book) GetAppTokenByOpenId(openId string) (string, bool) {
	ctx := context.Background()
	rs := b.db.Read(ctx, bookDatabase, bookTable, []db.SearchCmd{
		{bookOpenId, "=", openId},
	})
	for _, r := range rs {
		if appToken, exists := r[bookAppToken]; exists {
			return appToken.(string), true
		}
	}
	return "", false
}
