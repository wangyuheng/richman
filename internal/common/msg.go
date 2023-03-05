package common

import (
	"fmt"
	"strings"
)

const (
	BindSuccess      = "绑定成功，可以开始记账啦 \r\n记账格式为： 备注 分类 金额。 \r\n 比如： 泡面 餐费 100 \r\n 或者： 加班费 工资收入 +100 \r\n 不是首次输入，可以忽略分类，比如： 泡面 100"
	NotBind          = "请先绑定菜单。可以把记账文档发给我. 如: https://richman.feishu.cn/base/bascnzqgwKBqIQxp272MoZh1fhd \r\n 或者说 搞一个"
	NotFoundUserName = "请告诉我你的名字\n 如: 用户 张三"
	AmountIllegal    = "金额格式错误"
)

func MakeSuccess(url string) string {
	return fmt.Sprintf("%s\r\n%s", url, BindSuccess)
}

func RecordSuccess(f float64, expenses Expenses) string {
	if expenses == Income {
		return fmt.Sprintf("记账成功。本月已收入 %.2f", f)
	} else {
		return fmt.Sprintf("记账成功。本月已支出 %.2f", f)
	}
}

func NouFoundCategory(remark string) string {
	return fmt.Sprintf("猜不出【%s】是什么分类。先按照完整格式提交一下，下次我就记住了。 \r\n 格式： 备注 分类 金额。比如： 泡面 餐费 100", remark)
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
	return fmt.Sprintf("欢迎：%s \r\n %s", name, BindSuccess)
}
