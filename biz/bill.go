package biz

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/geeklubcn/richman/client"
	"github.com/geeklubcn/richman/db"
)

type BillBiz interface {
	Record(authorID, cmd string) error
	GetCategory(remark string) []string
	CurMonthTotal() float64
}

type billBiz struct {
	db db.Bills
}

func NewBill(cli client.Bitable) BillBiz {
	return &billBiz{db: db.NewBills(cli)}
}

func (b *billBiz) Record(authorID, content string) error {
	cmds := strings.Split(strings.TrimSpace(content), " ")
	switch len(cmds) {
	case 2:
		remark := cmds[0]
		amount, expenses, err := b.parseAmount(cmds[1])
		if err != nil {
			return err
		}
		categories := b.GetCategory(remark)
		if len(categories) == 0 {
			return fmt.Errorf("猜不出【%s】是什么分类。先按照完整格式提交一下，下次我就记住了。 \r\n 格式： 备注 分类 金额。比如： 泡面 餐费 100", remark)
		}
		return b.db.Save(&db.Bill{
			Remark:     remark,
			Categories: categories,
			Amount:     amount,
			Expenses:   expenses,
			AuthorID:   authorID,
		})
	case 3:
		remark := cmds[0]
		amount, expenses, err := b.parseAmount(cmds[2])
		if err != nil {
			return err
		}
		categories := []string{cmds[1]}
		return b.db.Save(&db.Bill{
			Remark:     remark,
			Categories: categories,
			Amount:     amount,
			Expenses:   expenses,
			AuthorID:   authorID,
		})
	default:
		return fmt.Errorf("格式错误。记账格式为： 备注 分类 金额。 \r\n 比如： 泡面 餐费 100 \r\n 或者： 加班费 工资收入 +100 \r\n 不是首次输入，可以忽略分类，比如： 泡面 100")
	}
}

func (b *billBiz) parseAmount(cmd string) (amount float64, expenses string, err error) {
	expenses = db.Pay
	if strings.HasPrefix(cmd, "+") {
		expenses = db.Income
	}
	amount, err = strconv.ParseFloat(strings.TrimPrefix(cmd, "+"), 10)
	if err != nil {
		return 0, "", fmt.Errorf("金额格式错误。%s", cmd)
	}
	return amount, expenses, err
}

func (b *billBiz) GetCategory(remark string) []string {
	records := b.db.Search([]db.SearchCmd{
		{
			Key:      db.BillTableRemark,
			Operator: "=",
			Val:      fmt.Sprintf("\"%s\"", remark),
		},
	})

	if len(records) > 0 {
		has := make(map[string]bool)
		res := make([]string, 0)

		for _, c := range records[0].Categories {
			if has[c] {
				continue
			}
			has[c] = true
			res = append(res, c)
		}
		return res
	}
	return nil
}

func (b *billBiz) CurMonthTotal() float64 {
	var total float64
	records := b.db.Search([]db.SearchCmd{
		{
			Key:      db.BillTableMonth,
			Operator: "=",
			Val:      fmt.Sprintf("\"%d 月\"", time.Now().Month()),
		},
		{
			Key:      db.BillTableExpenses,
			Operator: "=",
			Val:      fmt.Sprintf("\"%s\"", db.Pay),
		},
	})

	for _, r := range records {
		total += r.Amount
	}

	return total
}
