package handler

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
	"github.com/wangyuheng/richman/internal/common"
	"github.com/wangyuheng/richman/internal/domain"
	"github.com/wangyuheng/richman/internal/usecase"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"
)

type WechatHandler interface {
	CheckSignature(ctx *gin.Context)
	Dispatch(ctx *gin.Context)
}

type wechatHandler struct {
	token         string
	resCache      *lru.Cache
	aiService     domain.AIService
	billUseCase   usecase.BillUseCase
	userUseCase   usecase.UserUseCase
	ledgerUseCase usecase.LedgerUseCase
	running       *running
}

type running struct {
	toggle *lru.Cache
}

func (r *running) Start(msgID string) {
	r.toggle.Add(fmt.Sprintf("%s:runing", msgID), true)
}

func (r *running) End(msgID string) {
	r.toggle.Remove(fmt.Sprintf("%s:runing", msgID))
}

func (r *running) IsRunning(msgID string) bool {
	return r.toggle.Contains(fmt.Sprintf("%s:runing", msgID))
}

// func NewWechatHandler(cfg *config.Config, user biz.User, aiService client.OpenaiService) WechatHandler {
func NewWechatHandler(cfg *config.Config, billUseCase usecase.BillUseCase, aiService domain.AIService, ledgerUseCase usecase.LedgerUseCase, userUseCase usecase.UserUseCase) WechatHandler {
	resCache, _ := lru.New(256)
	runningCache, _ := lru.New(256)
	return &wechatHandler{
		token:         cfg.WechatToken,
		resCache:      resCache,
		aiService:     aiService,
		billUseCase:   billUseCase,
		ledgerUseCase: ledgerUseCase,
		userUseCase:   userUseCase,
		running: &running{
			toggle: runningCache,
		},
	}
}

func (w *wechatHandler) CheckSignature(ctx *gin.Context) {
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

func (w *wechatHandler) Dispatch(ctx *gin.Context) {
	var req WxReq
	var err error
	if err = ctx.BindXML(&req); err != nil {
		logrus.Error("unmarshal req xml fail!", err)
		_ = ctx.AbortWithError(400, fmt.Errorf("unmarshal req xml fail"))
		return
	}
	ctx.Set(common.CurrentUserID, req.FromUserName)

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
	if w.running.IsRunning(req.MsgID) {
		for i := 0; i < 5; i++ {
			logger.Info("wait for running finish")
			time.Sleep(1 * time.Second)
			if !w.running.IsRunning(req.MsgID) {
				break
			}
		}
	}
	if r, ok := w.resCache.Get(req.MsgID); ok {
		logger.Info("get res from cache")
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, r.(string))
		return
	}
	w.running.Start(req.MsgID)
	defer func() {
		w.running.End(req.MsgID)
	}()
	res, err := w.handleWechatTextMessage(ctx, req.Content, req.FromUserName)
	if err != nil {
		w.resCache.Add(req.MsgID, common.Err(err))
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.Err(err))
		return
	}
	w.resCache.Add(req.MsgID, res)
	w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, res)
	return
}

func (w *wechatHandler) handleWechatTextMessage(ctx context.Context, content, UID string) (string, error) {
	cmd := common.Trim(content)

	var js map[string]string
	if json.Unmarshal([]byte(cmd), &js) == nil {
		return common.NotSupport, nil
	}
	if cmd == "搞一个" {
		return common.NotSupport, nil
	}

	operator := &domain.User{
		UID: UID,
	}

	resp, err := w.aiService.CallFunctions(ctx, cmd, buildAIFunctions())
	if err != nil {
		return "", err
	}
	h := w.buildHandler(resp.FunctionCall, resp.Content)
	if h.NeedAuth {
		logrus.WithContext(ctx).Infof("exec handler %s", h.Name)
		userExist := false
		operator, userExist = w.userUseCase.GetByID(UID)
		if !userExist || operator.Name == "" {
			logrus.WithContext(ctx).Info("user not found, input required.")
			return common.NotFoundUserName, nil
		}
	}
	return h.Handle(operator)
}

func buildAIFunctions() domain.AI {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	currentDate := time.Now().Format("2006/01/02")
	expenses := []string{"收入", "支出"}
	bookkeepingRequired := []string{"remark", "amount", "expenses", "category"}
	return domain.AI{
		Introduction: fmt.Sprintf("你叫Richman 是一个基于飞书表格的记账软件。当前时间是 %s 如果不明确用户的意图，可以指导用户使用这个如见。比如：可以通过输入包子花了15或者工资收入100 用来记账", currentTime),
		Functions: []domain.AIFunction{
			{
				Name:        "bookkeeping",
				Description: "记账工具，支持记录收入支出",
				Parameters: domain.AIParameter{
					Type: "object",
					Properties: map[string]domain.AIProperty{
						"remark": {
							Type:        "string",
							Description: "名称或描述",
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
							Description: "账单分类",
						},
					},
					Required: &bookkeepingRequired,
				},
			},
			{
				Name:        "query_bill",
				Description: "查询账单信息",
				Parameters: domain.AIParameter{
					Type: "object",
					Properties: map[string]domain.AIProperty{
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
				Parameters: domain.AIParameter{
					Type:       "object",
					Properties: map[string]domain.AIProperty{},
				},
			},
			{
				Name:        "get_category",
				Description: "获取分类",
				Parameters: domain.AIParameter{
					Type:       "object",
					Properties: map[string]domain.AIProperty{},
				},
			},
			{
				Name:        "get_user_identity",
				Description: "获取用户的称呼",
				Parameters: domain.AIParameter{
					Type: "object",
					Properties: map[string]domain.AIProperty{
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
				Parameters: domain.AIParameter{
					Type:       "object",
					Properties: map[string]domain.AIProperty{},
				},
			},
		},
	}

}
func (w *wechatHandler) buildHandler(call *domain.AIFunctionCall, content string) Handler {
	if call != nil {
		switch call.Name {
		case "get_source_code":
			return Handler{
				Name:     call.Name,
				NeedAuth: false,
				Handle: func(operator *domain.User) (string, error) {
					return "https://github.com/wangyuheng/richman", nil
				},
			}
		case "get_ledger":
			return Handler{
				Name:     call.Name,
				NeedAuth: true,
				Handle: func(operator *domain.User) (string, error) {
					var ledger *domain.Ledger
					var err error
					ledger, exists := w.ledgerUseCase.QueryByUID(operator.UID)
					if !exists {
						ledger, err = w.ledgerUseCase.Allocated(*operator)
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
				Handle: func(operator *domain.User) (string, error) {
					ledger, exists := w.ledgerUseCase.QueryByUID(operator.UID)
					if !exists {
						ledger, _ = w.ledgerUseCase.Allocated(*operator)
					}
					return strings.Join(w.billUseCase.ListCategory(ledger.AppToken, ledger.TableToken), "\r\n"), nil
				},
			}
		case "query_bill":
			return Handler{
				Name:     call.Name,
				NeedAuth: true,
				Handle: func(operator *domain.User) (string, error) {
					ledger, exists := w.ledgerUseCase.QueryByUID(operator.UID)
					if !exists {
						ledger, _ = w.ledgerUseCase.Allocated(*operator)
					}
					in := w.billUseCase.CurMonthTotal(ledger.AppToken, ledger.TableToken, common.Income, 0)
					out := w.billUseCase.CurMonthTotal(ledger.AppToken, ledger.TableToken, common.Pay, 0)

					return common.Analysis(in, out), nil
				},
			}
		case "bookkeeping":
			return Handler{
				Name:     call.Name,
				NeedAuth: true,
				Handle: func(operator *domain.User) (string, error) {
					var args BookkeepingArgs
					_ = json.Unmarshal([]byte(call.Arguments), &args)
					amount, err := strconv.ParseFloat(args.Amount, 64)
					if err != nil {
						return common.AmountIllegal, nil
					}
					ledger, exists := w.ledgerUseCase.QueryByUID(operator.UID)
					if !exists {
						ledger, _ = w.ledgerUseCase.Allocated(*operator)
					}
					total := w.billUseCase.CurMonthTotal(ledger.AppToken, ledger.TableToken, common.Expenses(args.Expenses), amount)
					if err := w.billUseCase.Save(ledger.AppToken, ledger.TableToken, &domain.Bill{
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
				Handle: func(operator *domain.User) (string, error) {
					var args GetUserIdentityArgs
					_ = json.Unmarshal([]byte(call.Arguments), &args)
					operator.Name = args.Name

					if err := w.userUseCase.Save(*operator); err != nil {
						logrus.WithError(err).Error("save operator fail")
						return "", err
					}
					go func() {
						if _, exist := w.ledgerUseCase.QueryByUID(operator.UID); !exist {
							_, _ = w.ledgerUseCase.Allocated(*operator)
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
			Handle: func(operator *domain.User) (string, error) {
				return content, nil
			},
		}
	}
	return Handler{
		Name:     "NoThing",
		NeedAuth: false,
		Handle: func(operator *domain.User) (string, error) {
			return "拜个早年吧", nil
		},
	}
}

type Handler struct {
	Name     string
	NeedAuth bool
	Handle   func(operator *domain.User) (string, error)
}

func (w *wechatHandler) returnTextMsg(ctx *gin.Context, from, to, content string) {
	res, _ := xml.Marshal(WxResp{
		ToUserName:   to,
		FromUserName: from,
		CreateTime:   time.Now().UnixNano() / 1e9,
		MsgType:      "text",
		Content:      content,
	})
	_, _ = ctx.Writer.Write(res)
}

func (w *wechatHandler) check(token, signature, timestamp, nonce string) bool {
	l := sort.StringSlice{token, timestamp, nonce}
	sort.Strings(l)
	str := strings.Join(l, "")
	if signature == fmt.Sprintf("%x", sha1.Sum([]byte(str))) {
		return true
	}
	return false
}

type WxReq struct {
	XMLName xml.Name `xml:"xml"`
	// ToUserName 开发者微信号
	ToUserName string `xml:"ToUserName"`
	// FromUserName 发送方帐号（一个OpenID）
	FromUserName string `xml:"FromUserName"`
	// CreateTime 消息创建时间 （整型）
	CreateTime int64 `xml:"CreateTime"`
	// MsgType 消息类型（文本消息为 text ）
	MsgType string `xml:"MsgType"`
	// Content 文本消息内容
	Content string `xml:"Content"`
	// MsgID 消息类型（消息id，64位整型）
	MsgID string `xml:"MsgId"`
}
type WxResp struct {
	XMLName xml.Name `xml:"xml"`
	// ToUserName 接收方帐号（收到的OpenID）
	ToUserName string `xml:"ToUserName"`
	// FromUserName 开发者微信号
	FromUserName string `xml:"FromUserName"`
	// CreateTime 消息创建时间 （整型）
	CreateTime int64 `xml:"CreateTime"`
	// MsgType 消息类型（文本消息为 text ）
	MsgType string `xml:"MsgType"`
	// Content 文本消息内容
	Content string `xml:"Content"`
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
