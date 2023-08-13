package api

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/biz"
	"github.com/wangyuheng/richman/internal/business"
	"github.com/wangyuheng/richman/internal/client"
	"github.com/wangyuheng/richman/internal/common"
	"github.com/wangyuheng/richman/internal/model"
	"runtime/debug"
	"sort"
	"strings"
	"time"
)

type Wechat interface {
	CheckSignature(ctx *gin.Context)
	Dispatch(ctx *gin.Context)
}

type wechat struct {
	token      string
	idempotent *lru.Cache
	bill       biz.Bill
	user       biz.User
	ledgerSvr  business.LedgerSvr
	aiCaller   client.OpenAICaller
}

func NewWechat(cfg *config.Config, bill biz.Bill, user biz.User, ledgerSvr business.LedgerSvr, aiCaller client.OpenAICaller) Wechat {
	idempotent, _ := lru.New(256)
	return &wechat{
		token:      cfg.WechatToken,
		idempotent: idempotent,
		bill:       bill,
		user:       user,
		ledgerSvr:  ledgerSvr,
		aiCaller:   aiCaller,
	}
}

func (w *wechat) CheckSignature(ctx *gin.Context) {
	signature := ctx.Query("signature")
	timestamp := ctx.Query("timestamp")
	nonce := ctx.Query("nonce")
	echostr := ctx.Query("echostr")
	logger := logrus.WithContext(ctx).
		WithField("token", w.token).
		WithField("signature", signature).
		WithField("timestamp", timestamp).
		WithField("nonce", nonce).
		WithField("echostr", echostr)

	logger.Info("check wechat sign")
	if w.check(w.token, signature, timestamp, nonce) {
		_, _ = ctx.Writer.WriteString(echostr)
		return
	}
	logger.Error("check sign fail")
	_ = ctx.AbortWithError(400, fmt.Errorf("heck sign fail"))
}

func (w *wechat) Dispatch(ctx *gin.Context) {
	var req model.WxReq
	var err error
	if err = ctx.BindXML(&req); err != nil {
		logrus.Error("unmarshal req xml fail!", err)
		_ = ctx.AbortWithError(400, fmt.Errorf("unmarshal req xml fail"))
		return
	}
	logger := logrus.WithContext(ctx).WithField("req", fmt.Sprintf("%+v", req))
	logger.Info("receive req xml")
	// handle panic
	defer func() {
		if p := recover(); p != nil {
			logger.Errorf("handle dispatch req panic! err:%+v, stack:%s", p, debug.Stack())
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, fmt.Sprintf("something is wrong with %s", p))
			return
		}
	}()
	// idempotent
	if w.idempotent.Contains(req.MsgID) {
		logger.Info("ignore repeat req")
		return
	}
	w.idempotent.Add(req.MsgID, true)

	ctx.Set("OPERATOR", &model.User{UID: req.FromUserName})
	cmd := common.Trim(req.Content)

	resp, err := w.aiCaller.CallFunctions(ctx, cmd, business.BuildFunctions())
	if err != nil {
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.Err(err))
		return
	}
	h := w.buildHandler(resp.FunctionCall, resp.Content)
	if h.needAuth {
		operator, userExist := w.user.Unique(ctx, req.FromUserName)
		if !userExist || operator.Name == "" {
			logger.Info("user not found, input required.")
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.NotFoundUserName)
			return
		}
		ctx.Set("OPERATOR", operator)
	}
	res, err := h.handle(ctx)
	if err != nil {
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.Err(err))
		return
	}
	w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, res)
	return
}

type handler struct {
	name     string
	needAuth bool
	handle   func(ctx context.Context) (string, error)
}

func (w *wechat) buildHandler(call *client.OpenAIFunctionCall, content string) handler {
	if call != nil {
		switch call.Name {
		case "get_source_code":
			return handler{
				name:     call.Name,
				needAuth: false,
				handle: func(ctx context.Context) (string, error) {
					return "https://github.com/wangyuheng/richman", nil
				},
			}
		case "get_ledger":
			return handler{
				name:     call.Name,
				needAuth: true,
				handle: func(ctx context.Context) (string, error) {
					operator := ctx.Value("OPERATOR").(*model.User)
					var ledger *business.Ledger
					var err error
					ledger, exists := w.ledgerSvr.QueryByUID(ctx, operator.UID)
					if !exists {
						ledger, err = w.ledgerSvr.Generate(ctx, *operator)
						if err != nil {
							return "", err
						}
					}
					return ledger.URL, nil
				},
			}
		case "get_category":
			return handler{
				name:     call.Name,
				needAuth: true,
				handle: func(ctx context.Context) (string, error) {
					operator := ctx.Value("OPERATOR").(*model.User)
					ledger, exists := w.ledgerSvr.QueryByUID(ctx, operator.UID)
					if !exists {
						return common.NotBind, nil
					}
					return strings.Join(w.bill.ListCategory(ledger.AppToken), "\r\n"), nil
				},
			}
		case "query_bill":
			return handler{
				name:     call.Name,
				needAuth: true,
				handle: func(ctx context.Context) (string, error) {
					operator := ctx.Value("OPERATOR").(*model.User)
					ledger, exists := w.ledgerSvr.QueryByUID(ctx, operator.UID)
					if !exists {
						return common.NotBind, nil
					}
					in := w.bill.CurMonthTotal(ledger.AppToken, common.Income, 0)
					out := w.bill.CurMonthTotal(ledger.AppToken, common.Pay, 0)

					return common.Analysis(in, out), nil
				},
			}
		case "bookkeeping":
			return handler{
				name:     call.Name,
				needAuth: true,
				handle: func(ctx context.Context) (string, error) {
					operator := ctx.Value("OPERATOR").(*model.User)

					var args business.BookkeepingArgs
					_ = json.Unmarshal([]byte(call.Arguments), &args)
					ledger, exists := w.ledgerSvr.QueryByUID(ctx, operator.UID)
					if !exists {
						return common.NotBind, nil
					}
					total := w.bill.CurMonthTotal(ledger.AppToken, common.Expenses(args.Expenses), args.Amount)
					if err := w.bill.Save(ctx, ledger.AppToken, &model.Bill{
						Remark:     args.Remark,
						Categories: []string{args.Category},
						Amount:     args.Amount,
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
			return handler{
				name:     call.Name,
				needAuth: false,
				handle: func(ctx context.Context) (string, error) {
					operator := ctx.Value("OPERATOR").(*model.User)

					var args business.GetUserIdentityArgs
					_ = json.Unmarshal([]byte(call.Arguments), &args)
					operator.Name = args.Name

					if err := w.user.Save(ctx, *operator); err != nil {
						logrus.WithContext(ctx).WithError(err).Error("save operator fail")
						return "", err
					}
					go func() {
						if _, exist := w.ledgerSvr.QueryByUID(ctx, operator.UID); !exist {
							_, _ = w.ledgerSvr.Generate(ctx, *operator)
						}
					}()

					return common.Welcome(operator.Name), nil
				},
			}
		}
	}
	if content != "" {
		return handler{
			name:     "ai answer",
			needAuth: true,
			handle: func(ctx context.Context) (string, error) {
				return content, nil
			},
		}
	}
	return handler{
		name:     "NoThing",
		needAuth: false,
		handle: func(ctx context.Context) (string, error) {
			return "拜个早年吧", nil
		},
	}
}

func (w *wechat) returnTextMsg(ctx *gin.Context, from, to, content string) {
	res, _ := xml.Marshal(model.WxResp{
		ToUserName:   to,
		FromUserName: from,
		CreateTime:   time.Now().UnixNano() / 1e9,
		MsgType:      "text",
		Content:      content,
	})
	_, _ = ctx.Writer.Write(res)
}

func (w *wechat) check(token, signature, timestamp, nonce string) bool {
	l := sort.StringSlice{token, timestamp, nonce}
	sort.Strings(l)
	str := strings.Join(l, "")
	if signature == fmt.Sprintf("%x", sha1.Sum([]byte(str))) {
		return true
	}
	return false
}
