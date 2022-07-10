package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/geeklubcn/feishu-bitable-db/db"
)

const (
	bookDatabase = "Richman"
	bookTable    = "books"
	bookAppToken = "app_token"
	bookOpenId   = "open_id"
)

type BookSvc interface {
	GetAppTokenByOpenId(openId string) (string, bool)
	Save(openId, url string) (string, error)
}

type bookSvc struct {
	db db.DB
}

func NewBookSvc(appId, appSecret string) BookSvc {
	ctx := context.Background()
	it, err := db.NewDB(appId, appSecret)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("init DB err! appId:%s", appId)
		return nil
	}
	_, _ = it.SaveTable(ctx, bookDatabase, db.Table{
		Name: "books",
		Fields: []db.Field{
			{Name: bookAppToken, Type: db.String},
			{Name: bookOpenId, Type: db.String},
		},
	})
	return bookSvc{it}
}

func (b bookSvc) Save(openId, url string) (string, error) {
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
		_ = b.db.Delete(ctx, bookDatabase, bookTable, db.GetID(r))
	}

	return b.db.Create(ctx, bookDatabase, bookTable, map[string]interface{}{
		bookAppToken: appToken,
		bookOpenId:   openId,
	})
}

func (b bookSvc) GetAppTokenByOpenId(openId string) (string, bool) {
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
