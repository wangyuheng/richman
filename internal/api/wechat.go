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
	aiCaller   client.OpenAICaller
	facade     business.Facade
	user       biz.User
}

func NewWechat(cfg *config.Config, facade business.Facade, user biz.User, aiCaller client.OpenAICaller) Wechat {
	idempotent, _ := lru.New(256)
	return &wechat{
		token:      cfg.WechatToken,
		idempotent: idempotent,
		facade:     facade,
		aiCaller:   aiCaller,
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
	h := w.facade.BuildHandler(resp.FunctionCall, resp.Content)
	if h.NeedAuth {
		logger.Infof("exec handler %s", h.Name)
		operator, userExist := w.user.Unique(ctx, req.FromUserName)
		if !userExist || operator.Name == "" {
			logger.Info("user not found, input required.")
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.NotFoundUserName)
			return
		}
		ctx.Set("OPERATOR", operator)
	}
	res, err := h.Handle(ctx)
	if err != nil {
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, common.Err(err))
		return
	}
	w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, res)
	return
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
