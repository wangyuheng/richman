package service

import (
	"context"

	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/geeklubcn/richman/model"
)

const (
	authorDatabase     = "Richman"
	authorTable        = "authors"
	authorAppId        = "app_id"
	authorFeishuOpenId = "feishu_open_id"
	authorWechatOpenId = "wechat_open_id"
)

type AuthorSvc interface {
	Get(openId string, category model.Category) (*model.Author, bool)
	Save(it *model.Author, category model.Category) (string, error)
}

func NewAuthorSvc(appId, appSecret string) AuthorSvc {
	ctx := context.Background()
	it, _ := db.NewDB(appId, appSecret)
	_, _ = it.SaveTable(ctx, authorDatabase, db.Table{
		Name: authorTable,
		Fields: []db.Field{
			{Name: authorFeishuOpenId, Type: db.String},
			{Name: authorWechatOpenId, Type: db.String},
			{Name: authorAppId, Type: db.String},
		},
	})
	return authorSvc{it}
}

type authorSvc struct {
	db db.DB
}

func (a authorSvc) Get(openId string, category model.Category) (*model.Author, bool) {
	ctx := context.Background()
	cmd := make([]db.SearchCmd, 0)
	if category == model.CategoryFeishu {
		cmd = append(cmd, db.SearchCmd{Key: authorFeishuOpenId, Operator: "=", Val: openId})
	} else {
		cmd = append(cmd, db.SearchCmd{Key: authorWechatOpenId, Operator: "=", Val: openId})
	}
	records := a.db.Read(ctx, authorDatabase, authorTable, cmd)
	for _, r := range records {
		return &model.Author{
			AppId:        db.GetString(r, authorAppId),
			FeishuOpenId: db.GetString(r, authorFeishuOpenId),
			WechatOpenId: db.GetString(r, authorWechatOpenId),
		}, true
	}
	return nil, false
}

func (a authorSvc) Save(it *model.Author, category model.Category) (string, error) {
	ctx := context.Background()

	cmd := make([]db.SearchCmd, 0)
	if category == model.CategoryFeishu {
		cmd = append(cmd, db.SearchCmd{Key: authorFeishuOpenId, Operator: "=", Val: it.FeishuOpenId})
	} else {
		cmd = append(cmd, db.SearchCmd{Key: authorWechatOpenId, Operator: "=", Val: it.WechatOpenId})
	}

	for _, r := range a.db.Read(ctx, authorDatabase, authorTable, cmd) {
		_ = a.db.Delete(ctx, authorDatabase, authorTable, db.GetID(r))
	}

	return a.db.Create(ctx, authorDatabase, authorTable, map[string]interface{}{
		appsAppId:          it.AppId,
		authorFeishuOpenId: it.FeishuOpenId,
		authorWechatOpenId: it.WechatOpenId,
	})
}
