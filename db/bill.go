package db

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/geeklubcn/richman/client"
	larkCore "github.com/larksuite/oapi-sdk-go/core"
	larkBitable "github.com/larksuite/oapi-sdk-go/service/bitable/v1"
	"github.com/sirupsen/logrus"
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
	Save(bill *Bill) error
	Search(ss []SearchCmd) []Bill
}

type bills struct {
	cli   client.Bitable
	db    Tables
	cache sync.Map
}

func NewBills(cli client.Bitable) Bills {
	return &bills{
		cli: cli,
		db:  newDB(cli),
	}
}

func (b *bills) Search(ss []SearchCmd) []Bill {
	res := make([]Bill, 0)
	if t, ok := b.db.Get(billTable); ok {
		ctx := larkCore.WrapContext(context.Background())

		filters := make([]string, 0)
		for _, s := range ss {
			filters = append(filters, fmt.Sprintf("CurrentValue.[%s]%s%s", s.Key, s.Operator, s.Val))
		}
		filter := "AND(" + strings.Join(filters, ",") + ")"

		records, err := b.cli.ListRecords(ctx, t.TableId, filter)
		if err != nil {
			logrus.Error("list records fail!.", err)
			return nil
		}

		for _, r := range records {
			it := Bill{
				Remark:   fmt.Sprintf("%s", r.Fields[BillTableRemark]),
				Expenses: fmt.Sprintf("%s", r.Fields[BillTableExpenses]),
				Month:    fmt.Sprintf("%s", r.Fields[BillTableMonth]),
			}
			cs := make([]string, 0)

			fc := r.Fields[BillTableCategory]
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

			if r.Fields[BillTableAmount] != nil {
				it.Amount, _ = strconv.ParseFloat(r.Fields[BillTableAmount].(string), 10)
			}

			if r.Fields[BillTableDate] != nil {
				it.Date = int64(r.Fields[BillTableDate].(float64))
			}

			res = append(res, it)
		}
		if len(res) > 0 {
			b.cache.Store(fmt.Sprintf("remark-search-%+v", ss), res)
		}
	}

	return res
}
func (b *bills) Save(bill *Bill) error {
	if t, ok := b.db.Get(billTable); ok {
		if bill.Date == 0 {
			bill.Date = time.Now().UnixNano() / 1e6
		}
		if bill.Expenses == "" {
			bill.Expenses = Pay
		}
		_, err := b.cli.BatchCreateRecord(larkCore.WrapContext(context.Background()), t.TableId, &larkBitable.AppTableRecordBatchCreateReqBody{
			Records: []*larkBitable.AppTableRecord{
				{
					Fields: map[string]interface{}{
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
					},
				},
			},
		})
		return err
	}
	return fmt.Errorf("table:%s must be exists", billTable)
}

type Bill struct {
	Remark     string
	Categories []string
	Amount     float64
	Month      string
	Date       int64
	Expenses   string
	AuthorID   string
}

func (b *Bill) GetCategory() interface{} {
	if len(b.Categories) == 1 {
		return b.Categories[0]
	} else {
		return b.Categories
	}
}
