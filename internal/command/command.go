package command

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"strings"
)

type Command int

const (
	Make Command = iota
	Bind
	Record
	Category
	NotFound
)

const (
	BindSuccess = "绑定成功，可以开始记账啦 \r\n记账格式为： 备注 分类 金额。 \r\n 比如： 泡面 餐费 100 \r\n 或者： 加班费 工资收入 +100 \r\n 不是首次输入，可以忽略分类，比如： 泡面 100"
	NotBind     = "请先绑定菜单。可以把记账文档发给我. 如: https://richman.feishu.cn/base/bascnzqgwKBqIQxp272MoZh1fhd"
)

func Err(err error) string {
	return fmt.Sprintf("发生了一个错误！ %s", err.Error())
}

func Parse(s string) Command {
	l := len(strings.Split(s, " "))
	switch {
	case strings.Contains(s, "整"),
		strings.Contains(s, "搞"),
		strings.Contains(s, "整一个"),
		strings.Contains(s, "搞一个"):
		return Make
	case govalidator.IsURL(s) && strings.Contains(s, "feishu.cn/base/"):
		return Bind
	case s == "分类":
		return Category
	case l == 2, l == 3:
		return Record
	}
	return NotFound
}

func Trim(content string) string {
	res := strings.ReplaceAll(content, " ", " ")

	res = strings.TrimSpace(content)
	res = strings.Trim(content, "\r")
	res = strings.Trim(content, "\n")
	res = strings.TrimSpace(content)
	return res
}
