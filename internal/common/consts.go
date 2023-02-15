package common

import (
	"strconv"
	"strings"
)

type Expenses string

const (
	Income Expenses = "收入"
	Pay    Expenses = "支出"
)

func ConfirmExpenses(s string) Expenses {
	expenses := Pay
	if strings.HasPrefix(s, "+") {
		expenses = Income
	}
	return expenses
}

func ParseAmount(s string) float64 {
	amount, _ := strconv.ParseFloat(strings.TrimPrefix(s, "+"), 10)
	return amount
}
