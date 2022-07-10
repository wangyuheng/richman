package repo

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/geeklubcn/richman/model"

	"github.com/geeklubcn/feishu-bitable-db/db"
)

const (
	billTable         = "个人账单记录"
	BillTableRemark   = "备注"
	BillTableCategory = "分类"
	BillTableAmount   = "金额"
	BillTableDate     = "日期"
	BillTableMonth    = "月份"
	BillTableExpenses = "收支"
	BillTableAuthor   = "怨种"
)

const (
	Income = "收入"
	Pay    = "支出"
)

type Bills interface {
	Save(appToken string, bill *model.Bill) error
	Search(appToken string, ss []db.SearchCmd) []*model.Bill
}

type bills struct {
	db    db.DB
	cache sync.Map
}

func NewBills(appId, appSecret string) Bills {
	d, _ := db.NewDB(appId, appSecret)
	return &bills{db: d}
}

func (b *bills) Search(appToken string, ss []db.SearchCmd) []*model.Bill {
	res := make([]*model.Bill, 0)

	ctx := context.Background()
	records := b.db.Read(ctx, appToken, billTable, ss)

	for _, r := range records {
		it := &model.Bill{
			Remark:   fmt.Sprintf("%s", r[BillTableRemark]),
			Expenses: fmt.Sprintf("%s", r[BillTableExpenses]),
			Month:    fmt.Sprintf("%s", r[BillTableMonth]),
		}
		cs := make([]string, 0)

		fc := r[BillTableCategory]
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

		if r[BillTableAmount] != nil {
			it.Amount, _ = strconv.ParseFloat(r[BillTableAmount].(string), 10)
		}

		if r[BillTableDate] != nil {
			it.Date = int64(r[BillTableDate].(float64))
		}

		res = append(res, it)
	}
	if len(res) > 0 {
		b.cache.Store(fmt.Sprintf("remark-search-%+v", ss), res)
	}
	return res
}

func (b *bills) Save(appToken string, bill *model.Bill) error {
	ctx := context.Background()

	if bill.Date == 0 {
		bill.Date = time.Now().UnixNano() / 1e6
	}
	if bill.Expenses == "" {
		bill.Expenses = Pay
	}
	_, err := b.db.Create(ctx, appToken, billTable, map[string]interface{}{
		BillTableRemark:   bill.Remark,
		BillTableCategory: bill.Categories,
		BillTableAmount:   bill.Amount,
		BillTableDate:     bill.Date,
		BillTableExpenses: bill.Expenses,
		BillTableAuthor: []map[string]string{
			{
				"id": bill.AuthorID,
			},
		},
	})
	return err
}
