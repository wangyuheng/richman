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
	"github.com/wangyuheng/richman/internal/model"
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
	book       biz.Book
	user       biz.User
	//appSvc     service.AppSvc
	//authorSvc  service.AuthorSvc
	//bookSvc    service.BookSvc

}

func NewWechat(cfg *config.Config, book biz.Book, user biz.User) Wechat {
	idempotent, _ := lru.New(256)
	return &wechat{
		token:      cfg.WechatToken,
		idempotent: idempotent,
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
	logger := logrus.WithContext(ctx).WithField("req", req)
	logger.Info("receive req xml")
	// handle panic
	defer func() {
		if p := recover(); p != nil {
			logger.Errorf("handle dispatch req panic! err:%+v", p)
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
		//TODO wechat
		if err := w.user.Save(ctx, model.User{
			UID:  "",
			Name: "",
		}); err != nil {
			logger.WithError(err).Error("save operator fail")
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, fmt.Sprintf("something is wrong with %s", err.Error()))
			return
		}
	}

	cmd := command.Trim(req.Content)

	switch command.Parse(cmd) {
	case command.Make:
		res, err := w.book.Generate(ctx, *operator)
		if err != nil {
			logger.WithError(err).Error("Handle Make Cmd Err")
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, command.Err(err))
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, res.URL)
		return
	case command.Bind:
		err := w.book.Bind(ctx, cmd, *operator)
		if err != nil {
			logger.WithError(err).Error("Handle Bind Cmd Err")
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, command.Err(err))
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, command.BindSuccess)
		return
	}

	//	if strings.HasPrefix(req.Content, "wechat_r_") {
	//		cmds := strings.Split(req.Content, "_r_")
	//		_, _ = w.authorSvc.Save(&model.Author{
	//			AppId:        cmds[1],
	//			FeishuOpenId: cmds[3],
	//			WechatOpenId: req.FromUserName,
	//		})
	//		if _, err := w.bookSvc.Save(cmds[1], cmds[3], cmds[2], string(model.CategoryWechat)); err != nil {
	//			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, err.Error())
	//			return
	//		}
	//		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, "绑定成功，可以开始记账啦 \r\n记账格式为： 备注 分类 金额。 \r\n 比如： 泡面 餐费 100 \r\n 或者： 加班费 工资收入 +100 \r\n 不是首次输入，可以忽略分类，比如： 泡面 100")
	//		return
	//	}
	//
	//	if strings.Contains(req.Content, "id") &&
	//		strings.Contains(req.Content, "secret") &&
	//		strings.Contains(req.Content, "token") {
	//		var app model.App
	//		if err := json.Unmarshal([]byte(req.Content), &app); err != nil {
	//			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, "json格式不正确，可以按照这个格式修改 {\"id\":\"cli_a257f60e6bbab00c\",\"secret\":\"TVhkohuKkamGFU3cabXuFhdlLoS3EwhL\",\"token\":\"OzCFbkwGSckR6vo1pM4L7c8HU3j0MoeP\"}")
	//			return
	//		}
	//		if err := w.register(app); err != nil {
	//			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, err.Error())
	//			return
	//		}
	//		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, fmt.Sprintf("绑定成功，开始使用吧。事件订阅地址\r\n%s/feishu/webhook/%s", config.GetConfig().SeverUrl, app.AppId))
	//		return
	//	}
	//	if author, exists := w.authorSvc.Get(req.FromUserName, model.CategoryWechat); exists {
	//		msg := facades[author.AppId].BillSvc.Record(author.AppId, author.FeishuOpenId, req.Content, model.CategoryWechat)
	//		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, msg)
	//		return
	//	}
	//	w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, "没找到绑定的账号信息")
	//	return
}

//
//func (w *wechat) register(app model.App) error {
//	_, err := w.appSvc.Save(&app)
//	if err != nil {
//		return err
//
//	}
//	err = register(app, w.authorSvc, w.bookSvc)
//	if err != nil {
//		return err
//
//	}
//	return nil
//}

func (w *wechat) trim(content string) string {
	res := strings.ReplaceAll(content, " ", " ")

	res = strings.TrimSpace(content)
	res = strings.Trim(content, "\r")
	res = strings.Trim(content, "\n")
	res = strings.TrimSpace(content)
	return res
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
