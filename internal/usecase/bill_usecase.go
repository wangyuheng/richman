package usecase

import (
	"fmt"
	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/wangyuheng/richman/internal/common"
	"github.com/wangyuheng/richman/internal/domain"
	"sync"
	"time"
)

type BillUseCase interface {
	Save(appToken, tableToken string, bill *domain.Bill) error
	GetCategory(appToken, tableToken, remark string) []string
	CurMonthTotal(appToken, tableToken string, expenses common.Expenses, amount float64) float64
	ListCategory(appToken, tableToken string) []string
}

type billUseCase struct {
	billRepository domain.BillRepository
	cache          sync.Map
}

func NewBillUseCase(billRepository domain.BillRepository) BillUseCase {
	return &billUseCase{
		billRepository: billRepository,
	}
}

func (b *billUseCase) Save(appToken, tableToken string, bill *domain.Bill) error {
	return b.billRepository.Save(appToken, tableToken, bill)
}

func (b *billUseCase) GetCategory(appToken, tableToken, remark string) []string {
	if v, ok := b.cache.Load(b.categoryCacheKey(appToken, remark)); ok {
		if vv, ok := v.([]string); ok {
			return vv
		}
	}

	records := b.billRepository.Search(appToken, tableToken, []db.SearchCmd{
		{
			Key:      domain.BillTableRemark,
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

func (b *billUseCase) CurMonthTotal(appToken, tableToken string, expenses common.Expenses, amount float64) float64 {
	var total float64
	records := b.billRepository.Search(appToken, tableToken, []db.SearchCmd{
		{
			Key:      domain.BillTableMonth,
			Operator: "=",
			Val:      fmt.Sprintf("%d 月", time.Now().Month()),
		},
		{
			Key:      domain.BillTableExpenses,
			Operator: "=",
			Val:      string(expenses),
		},
	})

	for _, r := range records {
		total += r.Amount
	}

	return total + amount
}

func (b *billUseCase) ListCategory(appToken, tableToken string) []string {
	records := b.billRepository.Search(appToken, tableToken, []db.SearchCmd{})
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

func (b *billUseCase) categoryCacheKey(appToken, remark string) string {
	return fmt.Sprintf("bill:category:appToken:%s:remark:%s", appToken, remark)
}
