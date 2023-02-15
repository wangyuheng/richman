package biz

import (
	"context"
	"fmt"
	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/common"
	"github.com/wangyuheng/richman/internal/model"
	"github.com/wangyuheng/richman/internal/repo"
	"sync"
	"time"
)

type Bill interface {
	CurMonthTotal(appToken string, expenses common.Expenses, amount float64) float64
	ListCategory(appToken string) []string
	GetCategory(appToken, remark string) []string
	Save(ctx context.Context, appToken string, item *model.Bill) error
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

func (b *bill) GetCategory(appToken, remark string) []string {
	if v, ok := b.cache.Load(b.categoryCacheKey(appToken, remark)); ok {
		if vv, ok := v.([]string); ok {
			return vv
		}
	}

	records := b.bills.Search(appToken, []db.SearchCmd{
		{
			Key:      repo.BillTableRemark,
			Operator: "=",
			Val:      remark,
		},
	})
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
				b.cache.Store(b.categoryCacheKey(appToken, remark), res)
				return res
			}
		}
		if len(res) > 0 {
			b.cache.Store(b.categoryCacheKey(appToken, remark), res)
		}
		return res
	}
	return nil
}

func (b *bill) CurMonthTotal(appToken string, expenses common.Expenses, amount float64) float64 {
	var total float64
	records := b.bills.Search(appToken, []db.SearchCmd{
		{
			Key:      repo.BillTableMonth,
			Operator: "=",
			Val:      fmt.Sprintf("%d 月", time.Now().Month()),
		},
		{
			Key:      repo.BillTableExpenses,
			Operator: "=",
			Val:      string(expenses),
		},
	})

	for _, r := range records {
		total += r.Amount
	}

	return total + amount
}

func (b *bill) Save(_ context.Context, appToken string, item *model.Bill) error {
	return b.bills.Save(appToken, item)
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

func (b *bill) categoryCacheKey(appToken, remark string) string {
	return fmt.Sprintf("bill:category:appToken:%s:remark:%s", appToken, remark)
}
