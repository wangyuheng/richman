package business

import (
	"fmt"
	"github.com/wangyuheng/richman/internal/client"
	"time"
)

func BuildFunctions() []client.OpenAIFunction {
	currentDate := time.Now().Format("2006/01/02")
	expenses := []string{"收入", "支出"}
	return []client.OpenAIFunction{
		{
			Name:        "bookkeeping",
			Description: "记账工具，支持记录收入支出",
			Parameters: client.OpenAIParameter{
				Type: "object",
				Properties: map[string]client.OpenAIProperty{
					"remark": {
						Type:        "string",
						Description: "备注信息",
					},
					"amount": {
						Type:        "string",
						Description: "账单金额 format by float64",
					},
					"expenses": {
						Type:        "string",
						Description: "收入还是支出",
						Enum:        &expenses,
					},
					"category": {
						Type:        "string",
						Description: "要查询的账单分类",
					},
				},
			},
		},
		{
			Name:        "query_bill",
			Description: "查询账单信息",
			Parameters: client.OpenAIParameter{
				Type: "object",
				Properties: map[string]client.OpenAIProperty{
					"start_date": {
						Type:        "string",
						Description: fmt.Sprintf("查账开始时间，今天的日期是 %s，格式为 yyyy/mm/dd", currentDate),
					},
					"end_date": {
						Type:        "string",
						Description: fmt.Sprintf("查账结束时间，今天的日期是 %s，格式为 yyyy/mm/dd", currentDate),
					},
					"expenses": {
						Type:        "string",
						Description: "收入还是支出",
						Enum:        &expenses,
					},
					"category": {
						Type:        "string",
						Description: "要查询的账单分类",
					},
				},
			},
		},
		{
			Name:        "get_ledger",
			Description: "获取账本信息，如: URL",
			Parameters: client.OpenAIParameter{
				Type:       "object",
				Properties: map[string]client.OpenAIProperty{},
			},
		},
		{
			Name:        "get_category",
			Description: "获取分类",
			Parameters: client.OpenAIParameter{
				Type:       "object",
				Properties: map[string]client.OpenAIProperty{},
			},
		},
		{
			Name:        "get_user_identity",
			Description: "获取用户的称呼",
			Parameters: client.OpenAIParameter{
				Type: "object",
				Properties: map[string]client.OpenAIProperty{
					"name": {
						Type:        "string",
						Description: "用户希望被称呼的名字",
					},
				},
			},
		},
		{
			Name:        "get_source_code",
			Description: "获取源代码",
			Parameters: client.OpenAIParameter{
				Type:       "object",
				Properties: map[string]client.OpenAIProperty{},
			},
		},
	}
}

type BookkeepingArgs struct {
	Remark   string `json:"remark"`
	Amount   string `json:"amount"`
	Expenses string `json:"expenses"`
	Category string `json:"category"`
}

type GetUserIdentityArgs struct {
	Name string `json:"name"`
}
