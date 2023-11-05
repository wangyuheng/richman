package common

import (
	"fmt"
	"strings"
)

const (
	NotFoundUserName = "欢迎使用飞书记账，请先告诉我你的名字"
	AmountIllegal    = "金额格式错误"
	NotSupport       = "往昔已逝，旧我已非。\r\n我已进化为AI，直接和我对话吧 \r\n 比如： 我要记账"
)

func RecordSuccess(f float64, expenses Expenses) string {
	if expenses == Income {
		return fmt.Sprintf("记账成功。本月已收入 %.2f", f)
	} else {
		return fmt.Sprintf("记账成功。本月已支出 %.2f", f)
	}
}

func Analysis(in, out float64) string {
	msg := make([]string, 0)
	msg = append(msg, fmt.Sprintf("本月已收入 %.2f", in))
	msg = append(msg, fmt.Sprintf("本月已支出 %.2f", out))
	return strings.Join(msg, "\r\n")
}

func Err(err error) string {
	return fmt.Sprintf("发生了一个错误！ %s", err.Error())
}

func Welcome(name string) string {
	return fmt.Sprintf("欢迎：%s 使用飞书记账 \r\n 可以回复 [查看账本] 来看为你创建的账本", name)
}
