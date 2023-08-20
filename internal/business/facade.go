package business

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/internal/biz"
	"github.com/wangyuheng/richman/internal/client"
	"github.com/wangyuheng/richman/internal/common"
	"github.com/wangyuheng/richman/internal/model"
	"strconv"
	"strings"
)

type Facade interface {
	BuildHandler(call *client.OpenAIFunctionCall, content string) Handler
}

type Handler struct {
	Name     string
	NeedAuth bool
	Handle   func(ctx context.Context) (string, error)
}

type facade struct {
	bill      biz.Bill
	user      biz.User
	ledgerSvr LedgerSvr
}

func NewFacade(bill biz.Bill, user biz.User, ledgerSvr LedgerSvr) Facade {
	return &facade{
		bill:      bill,
		user:      user,
		ledgerSvr: ledgerSvr,
	}
}

func (f *facade) BuildHandler(call *client.OpenAIFunctionCall, content string) Handler {
	if call != nil {
		switch call.Name {
		case "get_source_code":
			return Handler{
				Name:     call.Name,
				NeedAuth: false,
				Handle: func(ctx context.Context) (string, error) {
					return "https://github.com/wangyuheng/richman", nil
				},
			}
		case "get_ledger":
			return Handler{
				Name:     call.Name,
				NeedAuth: true,
				Handle: func(ctx context.Context) (string, error) {
					operator := ctx.Value("OPERATOR").(*model.User)
					var ledger *Ledger
					var err error
					ledger, exists := f.ledgerSvr.QueryByUID(ctx, operator.UID)
					if !exists {
						ledger, err = f.ledgerSvr.Generate(ctx, *operator)
						if err != nil {
							return "", err
						}
					}
					return ledger.URL, nil
				},
			}
		case "get_category":
			return Handler{
				Name:     call.Name,
				NeedAuth: true,
				Handle: func(ctx context.Context) (string, error) {
					operator := ctx.Value("OPERATOR").(*model.User)
					ledger, exists := f.ledgerSvr.QueryByUID(ctx, operator.UID)
					if !exists {
						return common.NotBind, nil
					}
					return strings.Join(f.bill.ListCategory(ledger.AppToken, ledger.TableToken), "\r\n"), nil
				},
			}
		case "query_bill":
			return Handler{
				Name:     call.Name,
				NeedAuth: true,
				Handle: func(ctx context.Context) (string, error) {
					operator := ctx.Value("OPERATOR").(*model.User)
					ledger, exists := f.ledgerSvr.QueryByUID(ctx, operator.UID)
					if !exists {
						return common.NotBind, nil
					}
					in := f.bill.CurMonthTotal(ledger.AppToken, ledger.TableToken, common.Income, 0)
					out := f.bill.CurMonthTotal(ledger.AppToken, ledger.TableToken, common.Pay, 0)

					return common.Analysis(in, out), nil
				},
			}
		case "bookkeeping":
			return Handler{
				Name:     call.Name,
				NeedAuth: true,
				Handle: func(ctx context.Context) (string, error) {
					operator := ctx.Value("OPERATOR").(*model.User)

					var args BookkeepingArgs
					_ = json.Unmarshal([]byte(call.Arguments), &args)
					amount, err := strconv.ParseFloat(args.Amount, 64)
					if err != nil {
						return common.AmountIllegal, nil
					}
					ledger, exists := f.ledgerSvr.QueryByUID(ctx, operator.UID)
					if !exists {
						return common.NotBind, nil
					}
					total := f.bill.CurMonthTotal(ledger.AppToken, ledger.TableToken, common.Expenses(args.Expenses), amount)
					if err := f.bill.Save(ctx, ledger.AppToken, ledger.TableToken, &model.Bill{
						Remark:     args.Remark,
						Categories: []string{args.Category},
						Amount:     amount,
						Expenses:   args.Expenses,
						AuthorID:   operator.UID,
						AuthorName: operator.Name,
					}); err != nil {
						return "", err
					}
					return common.RecordSuccess(total, common.Expenses(args.Expenses)), nil
				},
			}
		case "get_user_identity":
			return Handler{
				Name:     call.Name,
				NeedAuth: false,
				Handle: func(ctx context.Context) (string, error) {
					operator := ctx.Value("OPERATOR").(*model.User)

					var args GetUserIdentityArgs
					_ = json.Unmarshal([]byte(call.Arguments), &args)
					operator.Name = args.Name

					if err := f.user.Save(ctx, *operator); err != nil {
						logrus.WithContext(ctx).WithError(err).Error("save operator fail")
						return "", err
					}
					go func() {
						if _, exist := f.ledgerSvr.QueryByUID(ctx, operator.UID); !exist {
							_, _ = f.ledgerSvr.Generate(ctx, *operator)
						}
					}()

					return common.Welcome(operator.Name), nil
				},
			}
		}
	}
	if content != "" {
		return Handler{
			Name:     "ai answer",
			NeedAuth: true,
			Handle: func(ctx context.Context) (string, error) {
				return content, nil
			},
		}
	}
	return Handler{
		Name:     "NoThing",
		NeedAuth: false,
		Handle: func(ctx context.Context) (string, error) {
			return "拜个早年吧", nil
		},
	}
}
