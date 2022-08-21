package service

import (
	"fmt"
	"sync"

	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/geeklubcn/richman/model"
	"github.com/geeklubcn/richman/repo"
	lru "github.com/hashicorp/golang-lru"
)

type AuthorSvc interface {
	Get(openId string, category model.Category) (*model.Author, bool)
	Save(it *model.Author) (string, error)
}

func NewAuthorSvc(appId, appSecret string) AuthorSvc {
	cache, _ := lru.New(1024)
	authors := repo.NewAuthors(appId, appSecret)

	svc := &authorSvc{authors: authors, cache: cache}
	svc.once.Do(func() {
		go svc.Warmup()
	})

	return svc
}

type authorSvc struct {
	authors repo.Authors
	cache   *lru.Cache
	once    sync.Once
}

func (a *authorSvc) Warmup() {
	records := a.authors.Search([]db.SearchCmd{})
	for _, it := range records {
		a.cacheItem(it)
	}
}

func (a *authorSvc) Get(openId string, category model.Category) (*model.Author, bool) {
	if v, ok := a.cache.Get(a.cacheKey(openId)); ok {
		return v.(*model.Author), true
	}

	cmd := make([]db.SearchCmd, 0)
	if category == model.CategoryFeishu {
		cmd = append(cmd, db.SearchCmd{Key: repo.AuthorFeishuOpenId, Operator: "=", Val: openId})
	} else {
		cmd = append(cmd, db.SearchCmd{Key: repo.AuthorWechatOpenId, Operator: "=", Val: openId})
	}
	records := a.authors.Search(cmd)
	for _, it := range records {
		a.cacheItem(it)
	}
	if len(records) > 0 {
		return records[0], true
	}
	return nil, false
}

func (a *authorSvc) Save(it *model.Author) (string, error) {
	res, err := a.authors.Save(it)
	if err != nil {
		a.cacheItem(it)
	}
	return res, err
}

func (a *authorSvc) cacheKey(openId string) string {
	return fmt.Sprintf("author-openId-%s", openId)
}

func (a *authorSvc) cacheItem(it *model.Author) {
	if it.WechatOpenId != "" {
		a.cache.Add(a.cacheKey(it.WechatOpenId), it)
	}
	if it.FeishuOpenId != "" {
		a.cache.Add(a.cacheKey(it.FeishuOpenId), it)
	}
}
