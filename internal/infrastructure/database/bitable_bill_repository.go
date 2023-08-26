package database

import (
	"context"
	"fmt"
	"github.com/wangyuheng/richman/internal/domain"

	"github.com/geeklubcn/feishu-bitable-db/db"
	"strconv"
	"sync"
	"time"
)

const (
	Income = "收入"
	Pay    = "支出"
)

type billRepository struct {
	db    db.DB
	cache sync.Map
}

func NewBillRepository(db db.DB) domain.BillRepository {
	b := &billRepository{db: db}
	return b
}

func (b *billRepository) refresh(appToken string) {
	//ctx := context.Background()
	//fm := b.bitable.ListFields(ctx, appToken, billTable)
	//author := BillTableAuthor
	//
	//for _, f := range fm {
	//	if f.Type == 11 {
	//		author = f.FieldName
	//	}
	//	if f.FieldName == BillTableCategory {
	//		b.cache.Store(fmt.Sprintf("bill-categoryFieldType-appToken-%s", appToken), f.Type)
	//	}
	//}
	//
	//b.cache.Store(fmt.Sprintf("bill-authorFieldName-appToken-%s", appToken), author)
}

func (b *billRepository) getCategoryFieldType(appToken, tableToken string) int {
	if v, ok := b.cache.Load(fmt.Sprintf("bill-categoryFieldType-appToken-%s", appToken)); ok {
		if vv, ok := v.(int); ok {
			return vv
		}
	}
	b.refresh(appToken)
	if v, ok := b.cache.Load(fmt.Sprintf("bill-categoryFieldType-appToken-%s", appToken)); ok {
		if vv, ok := v.(int); ok {
			return vv
		}
	}
	return -1
}

func (b *billRepository) getAuthorFieldName(appToken, tableToken string) string {
	if v, ok := b.cache.Load(fmt.Sprintf("bill-authorFieldName-appToken-%s", appToken)); ok {
		if vv, ok := v.(string); ok {
			return vv
		}
	}
	b.refresh(appToken)
	if v, ok := b.cache.Load(fmt.Sprintf("bill-authorFieldName-appToken-%s", appToken)); ok {
		if vv, ok := v.(string); ok {
			return vv
		}
	}
	return ""
}

func (b *billRepository) Search(appToken, tableToken string, ss []db.SearchCmd) []*domain.Bill {
	res := make([]*domain.Bill, 0)

	ctx := context.Background()
	records := b.db.Read(ctx, appToken, tableToken, ss)

	for _, r := range records {
		it := &domain.Bill{
			Remark:   fmt.Sprintf("%s", r[domain.BillTableRemark]),
			Expenses: fmt.Sprintf("%s", r[domain.BillTableExpenses]),
			Month:    fmt.Sprintf("%s", r[domain.BillTableMonth]),
		}
		cs := make([]string, 0)

		fc := r[domain.BillTableCategory]
		if fcs, ok := fc.([]interface{}); ok {
			for _, fc := range fcs {
				if c, ok := fc.(string); ok {
					cs = append(cs, c)
				}
			}
		} else if c, ok := fc.(string); ok {
			cs = append(cs, c)
		}

		it.Categories = cs

		if r[domain.BillTableAmount] != nil {
			it.Amount, _ = strconv.ParseFloat(r[domain.BillTableAmount].(string), 10)
		}

		if r[domain.BillTableDate] != nil {
			it.Date = int64(r[domain.BillTableDate].(float64))
		}

		res = append(res, it)
	}
	if len(res) > 0 {
		b.cache.Store(fmt.Sprintf("remark-search-%+v", ss), res)
	}
	return res
}

func (b *billRepository) Save(appToken, tableToken string, bill *domain.Bill) error {
	ctx := context.Background()

	if bill.Date == 0 {
		bill.Date = time.Now().UnixNano() / 1e6
	}
	if bill.Expenses == "" {
		bill.Expenses = Pay
	}

	var categoryV interface{}
	if b.getCategoryFieldType(appToken, tableToken) == 3 {
		categoryV = bill.Categories[0]
	} else {
		categoryV = bill.Categories
	}

	_, err := b.db.Create(ctx, appToken, tableToken, map[string]interface{}{
		domain.BillTableRemark:   bill.Remark,
		domain.BillTableCategory: categoryV,
		domain.BillTableAmount:   bill.Amount,
		domain.BillTableDate:     bill.Date,
		domain.BillTableExpenses: bill.Expenses,
		domain.BillTableAuthor:   bill.AuthorName,
	})
	if err != nil {
		b.refresh(appToken)
	}
	return err
}
