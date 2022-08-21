package repo

import (
	"context"

	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/geeklubcn/richman/model"
)

const (
	authorDatabase     = "Richman"
	authorTable        = "authors"
	AuthorAppId        = "app_id"
	AuthorFeishuOpenId = "feishu_open_id"
	AuthorWechatOpenId = "wechat_open_id"
)

type Authors interface {
	Search(cmd []db.SearchCmd) []*model.Author
	Save(it *model.Author) (string, error)
}

type authors struct {
	db db.DB
}

func NewAuthors(appId, appSecret string) Authors {
	ctx := context.Background()
	it, _ := db.NewDB(appId, appSecret)
	_, _ = it.SaveTable(ctx, authorDatabase, db.Table{
		Name: authorTable,
		Fields: []db.Field{
			{Name: AuthorFeishuOpenId, Type: db.String},
			{Name: AuthorWechatOpenId, Type: db.String},
			{Name: AuthorAppId, Type: db.String},
		},
	})
	return &authors{it}
}

func (a *authors) Search(cmd []db.SearchCmd) []*model.Author {
	ctx := context.Background()
	res := make([]*model.Author, 0)
	records := a.db.Read(ctx, authorDatabase, authorTable, cmd)
	for _, r := range records {
		res = append(res, &model.Author{
			AppId:        db.GetString(r, AuthorAppId),
			FeishuOpenId: db.GetString(r, AuthorFeishuOpenId),
			WechatOpenId: db.GetString(r, AuthorWechatOpenId),
		})
	}
	return res
}

func (a *authors) Save(it *model.Author) (string, error) {
	ctx := context.Background()

	cmd := make([]db.SearchCmd, 0)
	if it.FeishuOpenId != "" {
		cmd = append(cmd, db.SearchCmd{Key: AuthorFeishuOpenId, Operator: "=", Val: it.FeishuOpenId})
	}
	if it.WechatOpenId != "" {
		cmd = append(cmd, db.SearchCmd{Key: AuthorWechatOpenId, Operator: "=", Val: it.WechatOpenId})
	}

	for _, r := range a.db.Read(ctx, authorDatabase, authorTable, cmd) {
		_ = a.db.Delete(ctx, authorDatabase, authorTable, db.GetID(r))
	}

	res, err := a.db.Create(ctx, authorDatabase, authorTable, map[string]interface{}{
		AuthorAppId:        it.AppId,
		AuthorFeishuOpenId: it.FeishuOpenId,
		AuthorWechatOpenId: it.WechatOpenId,
	})
	return res, err
}
