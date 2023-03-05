package api

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"github.com/gin-gonic/gin"
	lru "github.com/hashicorp/golang-lru"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/biz"
	"github.com/wangyuheng/richman/internal/command"
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
	book       biz.Book
	user       biz.User
}

func NewWechat(cfg *config.Config, bill biz.Bill, book biz.Book, user biz.User) Wechat {
	idempotent, _ := lru.New(256)
	return &wechat{
		token:      cfg.WechatToken,
		idempotent: idempotent,
		bill:       bill,
		book:       book,
		user:       user,
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
	if err := ctx.BindXML(&req); err != nil {
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
	// 查询用户信息
	operator, exist := w.user.Unique(ctx, req.FromUserName)
	if !exist {
		operator = &model.User{
			UID: req.FromUserName,
		}
		if err := w.user.Save(ctx, *operator); err != nil {
			logger.WithError(err).Error("save operator fail")
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, fmt.Sprintf("something is wrong with %s", err.Error()))
			return
		}
	}

	cmd := common.Trim(req.Content)

	switch c := command.Parse(cmd); c.Type {
	case command.Analysis:
		book, exists := w.book.QueryByUID(ctx, operator.UID)
		if !exists {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.NotBind)
			return
		}

		in := w.bill.CurMonthTotal(book.AppToken, common.Income, 0)
		out := w.bill.CurMonthTotal(book.AppToken, common.Pay, 0)

		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.Analysis(in, out))
		return
	case command.Make:
		if book, exists := w.book.QueryByUID(ctx, operator.UID); exists {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.MakeSuccess(book.URL))
			return
		}
		res, err := w.book.Generate(ctx, *operator)
		if err != nil {
			logger.WithError(err).Error("Handle Make Cmd Err")
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.Err(err))
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.MakeSuccess(res.URL))
		return
	case command.Bind:
		err := w.book.Bind(ctx, cmd, *operator)
		if err != nil {
			logger.WithError(err).Error("Handle Bind Cmd Err")
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.Err(err))
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.BindSuccess)
		return
	case command.Category:
		book, exists := w.book.QueryByUID(ctx, operator.UID)
		if !exists {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.NotBind)
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, strings.Join(w.bill.ListCategory(book.AppToken), "\r\n"))
		return
	case command.Bill:
		book, exists := w.book.QueryByUID(ctx, operator.UID)
		if !exists {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.NotBind)
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, book.URL)
		return
	case command.User:
		name := c.Data.(string)
		operator = &model.User{
			UID:  req.FromUserName,
			Name: name,
		}
		if err := w.user.Save(ctx, *operator); err != nil {
			logger.WithError(err).Error("save operator fail")
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, fmt.Sprintf("something is wrong with %s", err.Error()))
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.Welcome(name))
		return
	case command.RecordUsual:
		book, exists := w.book.QueryByUID(ctx, operator.UID)
		if !exists {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.NotBind)
			return
		}
		d := c.Data.(command.RecordUsualData)
		if d.Amount <= 0 {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.AmountIllegal)
			return
		}

		categories := w.bill.GetCategory(book.AppToken, d.Remark)
		if len(categories) == 0 {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.NouFoundCategory(d.Remark))
			return
		}
		total := w.bill.CurMonthTotal(book.AppToken, d.Expenses, d.Amount)
		if err := w.bill.Save(ctx, book.AppToken, &model.Bill{
			Remark:     d.Remark,
			Categories: categories,
			Amount:     d.Amount,
			Expenses:   string(d.Expenses),
			AuthorID:   operator.UID,
			AuthorName: operator.Name,
		}); err != nil {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.Err(err))
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.RecordSuccess(total, d.Expenses))
		return
	case command.Record:
		if operator.Name == "" {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.NotFoundUserName)
			return
		}
		book, exists := w.book.QueryByUID(ctx, operator.UID)
		if !exists {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.NotBind)
			return
		}
		d := c.Data.(command.RecordData)
		if d.Amount <= 0 {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.AmountIllegal)
			return
		}
		total := w.bill.CurMonthTotal(book.AppToken, d.Expenses, d.Amount)
		if err := w.bill.Save(ctx, book.AppToken, &model.Bill{
			Remark:     d.Remark,
			Categories: []string{d.Category},
			Amount:     d.Amount,
			Expenses:   string(d.Expenses),
			AuthorID:   operator.UID,
			AuthorName: operator.Name,
		}); err != nil {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.Err(err))
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.RecordSuccess(total, d.Expenses))
		return
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
