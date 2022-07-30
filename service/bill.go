package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/geeklubcn/richman/client"

	"github.com/geeklubcn/richman/model"

	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/geeklubcn/richman/repo"
)

type BillSvc interface {
	Record(appId, authorId, cmd string, category model.Category) string
	GetCategory(appToken, remark string) []string
}

type billSvc struct {
	repo  repo.Bills
	books BookSvc
}

func NewBillSvc(appId, appSecret string, bookSvc BookSvc, bitable client.Bitable) BillSvc {
	return &billSvc{
		repo:  repo.NewBills(appId, appSecret, bitable),
		books: bookSvc,
	}
}

func (b *billSvc) Record(appId, authorId, content string, category model.Category) string {
	cmds := strings.Split(strings.TrimSpace(content), " ")
	book, exists := b.books.GetByOpenId(authorId)
	if len(cmds) != 1 && !exists {
		return fmt.Sprintf("请先绑定菜单。可以把记账文档发给我. 如%s", "https://richman.feishu.cn/base/bascnzqgwKBqIQxp272MoZh1fhd")
	}
	switch len(cmds) {
	case 1:
		cmd := cmds[0]
		switch cmd {
		case "账单":
			return fmt.Sprintf("https://richman.feishu.cn/base/%s", book.AppToken)
		case "微信", "wechat", "wx", "weixin":
			return "wechat_r_" + appId + "_r_" + book.AppToken + "_r_" + authorId
		default:
			ss := strings.Split(cmd, "feishu.cn/base/")
			if len(ss) < 2 {
				return fmt.Sprintf("url[%s] format illegal", cmd)
			}
			s := ss[1]
			l := strings.Index(s, "?")
			if l2 := strings.Index(s, "/"); l2 > 0 && l2 < l {
				l = l2
			}
			_, err := b.books.Save(appId, authorId, s[0:l], string(category))
			if err != nil {
				return err.Error()
			}
			return "绑定成功，可以开始记账啦 \r\n记账格式为： 备注 分类 金额。 \r\n 比如： 泡面 餐费 100 \r\n 或者： 加班费 工资收入 +100 \r\n 不是首次输入，可以忽略分类，比如： 泡面 100"
		}
	case 2:
		remark := cmds[0]
		amount, expenses, err := b.parseAmount(cmds[1])
		if err != nil {
			return err.Error()
		}
		categories := b.GetCategory(book.AppToken, remark)
		if len(categories) == 0 {
			return fmt.Sprintf("猜不出【%s】是什么分类。先按照完整格式提交一下，下次我就记住了。 \r\n 格式： 备注 分类 金额。比如： 泡面 餐费 100", remark)
		}
		err = b.repo.Save(book.AppToken, &model.Bill{
			Remark:     remark,
			Categories: categories,
			Amount:     amount,
			Expenses:   expenses,
			AuthorID:   authorId,
		})
		if err == nil {
			return fmt.Sprintf("记账成功。本月已支出 %.2f", b.curMonthTotal(book.AppToken))
		}
		return err.Error()
	case 3:
		remark := cmds[0]
		amount, expenses, err := b.parseAmount(cmds[2])
		if err != nil {
			return err.Error()
		}
		categories := []string{cmds[1]}
		err = b.repo.Save(book.AppToken, &model.Bill{
			Remark:     remark,
			Categories: categories,
			Amount:     amount,
			Expenses:   expenses,
			AuthorID:   authorId,
		})
		if err == nil {
			return fmt.Sprintf("记账成功。本月已支出 %.2f", b.curMonthTotal(book.AppToken))
		}
		return err.Error()
	default:
		return fmt.Sprintf("格式错误。记账格式为： 备注 分类 金额。 \r\n 比如： 泡面 餐费 100 \r\n 或者： 加班费 工资收入 +100 \r\n 不是首次输入，可以忽略分类，比如： 泡面 100")
	}
}

func (b *billSvc) parseAmount(cmd string) (amount float64, expenses string, err error) {
	expenses = repo.Pay
	if strings.HasPrefix(cmd, "+") {
		expenses = repo.Income
	}
	amount, err = strconv.ParseFloat(strings.TrimPrefix(cmd, "+"), 10)
	if err != nil {
		return 0, "", fmt.Errorf("金额格式错误。%s", cmd)
	}
	return amount, expenses, err
}

func (b *billSvc) GetCategory(appToken, remark string) []string {
	records := b.repo.Search(appToken, []db.SearchCmd{
		{
			Key:      repo.BillTableRemark,
			Operator: "=",
			Val:      remark,
		},
	})
	// distinct
	if len(records) > 0 {
		has := make(map[string]bool)
		res := make([]string, 0)

		for _, r := range records {
			if len(r.Categories) > 0 {
				for _, c := range r.Categories {
					if has[c] {
						continue
					}
					has[c] = true
					res = append(res, c)
				}
				return res
			}
		}

		return res
	}
	return nil
}

func (b *billSvc) curMonthTotal(appToken string) float64 {
	var total float64
	records := b.repo.Search(appToken, []db.SearchCmd{
		{
			Key:      repo.BillTableMonth,
			Operator: "=",
			Val:      fmt.Sprintf("%d 月", time.Now().Month()),
		},
		{
			Key:      repo.BillTableExpenses,
			Operator: "=",
			Val:      repo.Pay,
		},
	})

	for _, r := range records {
		total += r.Amount
	}

	return total
}
