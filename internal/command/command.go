package command

import (
	"github.com/asaskevich/govalidator"
	"github.com/wangyuheng/richman/internal/common"
	"strconv"
	"strings"
)

type Commander struct {
	Type Command
	Data interface{}
}

type RecordData struct {
	Remark   string
	Category string
	Amount   float64
	Expenses string
}

type Command int

const (
	Make Command = iota
	Bill
	Bind
	Record
	RecordUsual
	Category
	NotFound
)

func Parse(s string) *Commander {
	switch {
	case strings.Contains(s, "整"),
		strings.Contains(s, "搞"),
		strings.Contains(s, "整一个"),
		strings.Contains(s, "搞一个"):
		return &Commander{Make, s}
	case govalidator.IsURL(s) && strings.Contains(s, "feishu.cn/base/"):
		return &Commander{Bind, s}
	case s == "账单":
		return &Commander{Bill, s}
	case s == "分类":
		return &Commander{Category, s}
	case len(strings.Split(s, " ")) == 3:
		ss := strings.Split(strings.TrimSpace(s), " ")
		expenses := common.Pay
		if strings.HasPrefix(ss[2], "+") {
			expenses = common.Income
		}
		amount, _ := strconv.ParseFloat(strings.TrimPrefix(ss[2], "+"), 10)
		return &Commander{Record, RecordData{
			Remark:   ss[0],
			Category: ss[1],
			Amount:   amount,
			Expenses: expenses,
		}}
	case len(strings.Split(s, " ")) == 2:
		return &Commander{RecordUsual, s}
	}
	return &Commander{NotFound, s}
}
