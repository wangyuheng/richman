package biz

import (
	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/repo"
	"sync"
)

type Bill interface {
	ListCategory(appToken string) []string
}

type bill struct {
	bills repo.Bills
	cache sync.Map
}

func NewBill(_ *config.Config, bills repo.Bills) Bill {
	return &bill{
		bills: bills,
	}
}

func (b *bill) ListCategory(appToken string) []string {
	records := b.bills.Search(appToken, []db.SearchCmd{})
	// distinct
	if len(records) > 0 {
		has := make(map[string]bool)
		res := make([]string, 0)

		for _, r := range records {
			if len(r.Categories) > 0 {
				for _, c := range r.Categories {
					if has[c] {
						continue
					}
					has[c] = true
					res = append(res, c)
				}
			}
		}
		return res
	}
	return nil
}
