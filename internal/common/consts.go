package common

import (
	"strconv"
	"strings"
)

const (
	Income = "收入"
	Pay    = "支出"
)

func ConfirmExpenses(s string) string {
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
