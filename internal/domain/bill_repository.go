package domain

import (
	"github.com/geeklubcn/feishu-bitable-db/db"
)

const (
	BillTableRemark   = "备注"
	BillTableCategory = "分类"
	BillTableAmount   = "金额"
	BillTableDate     = "日期"
	BillTableMonth    = "月份"
	BillTableExpenses = "收支"
	BillTableAuthor   = "花钱小能手"
)

type BillRepository interface {
	Save(appToken, tableToken string, bill *Bill) error
	Search(appToken, tableToken string, ss []db.SearchCmd) []*Bill
}
