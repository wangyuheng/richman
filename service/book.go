package service

import (
	"fmt"
	"sync"

	"github.com/geeklubcn/feishu-bitable-db/db"
	lru "github.com/hashicorp/golang-lru"
	"github.com/wangyuheng/richman/model"
	"github.com/wangyuheng/richman/repo"
)

type BookSvc interface {
	GetByOpenId(openId string) (*model.Book, bool)
	Save(appId, openId, appToken, category string) (string, error)
}

type bookSvc struct {
	books repo.Books
	cache *lru.Cache
	once  sync.Once
}

func NewBookSvc(appId, appSecret string) BookSvc {
	books := repo.NewBooks(appId, appSecret)
	cache, _ := lru.New(1024)

	svc := &bookSvc{books: books, cache: cache}
	svc.once.Do(func() {
		go svc.Warmup()
	})

	return svc
}

func (b *bookSvc) cacheKey(openId string) string {
	return fmt.Sprintf("book-db-openid-%s", openId)
}

func (b *bookSvc) Warmup() {
	rs := b.books.Search([]db.SearchCmd{})
	for _, r := range rs {
		b.cache.Add(b.cacheKey(r.OpenId), r)
	}
}

func (b *bookSvc) GetByOpenId(openId string) (*model.Book, bool) {
	if v, ok := b.cache.Get(b.cacheKey(openId)); ok {
		return v.(*model.Book), true
	}
	rs := b.books.Search([]db.SearchCmd{
		{repo.BookOpenId, "=", openId},
	})
	for _, r := range rs {
		b.cache.Add(b.cacheKey(openId), r)
	}
	if len(rs) > 0 {
		return rs[0], true
	}
	return nil, false
}

func (b *bookSvc) Save(appId, openId, appToken, category string) (string, error) {
	if _, ok := b.cache.Get(b.cacheKey(openId)); ok {
		b.cache.Remove(b.cacheKey(openId))
	}

	it := &model.Book{
		AppId:    appId,
		AppToken: appToken,
		OpenId:   openId,
		Category: model.Category(category),
	}
	res, err := b.books.Save(it)

	if err != nil {
		b.cache.Add(b.cacheKey(openId), it)
	}

	return res, err
}
