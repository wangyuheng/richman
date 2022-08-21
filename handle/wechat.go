package handle

import (
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/geeklubcn/richman/config"
	"github.com/geeklubcn/richman/model"
	"github.com/geeklubcn/richman/service"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

type Wechat interface {
	CheckSignature(ctx *gin.Context)
	Dispatch(ctx *gin.Context)
}

type wechat struct {
	token      string
	appSvc     service.AppSvc
	authorSvc  service.AuthorSvc
	bookSvc    service.BookSvc
	idempotent *lru.Cache
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

func NewWechat(checkToken string, appSvc service.AppSvc, authorSvc service.AuthorSvc, bookSvc service.BookSvc) Wechat {
	w := &wechat{
		token:     checkToken,
		appSvc:    appSvc,
		authorSvc: authorSvc,
		bookSvc:   bookSvc,
	}
	w.idempotent, _ = lru.New(256)
	return w
}

func (w *wechat) CheckSignature(ctx *gin.Context) {
	signature := ctx.Query("signature")
	timestamp := ctx.Query("timestamp")
	nonce := ctx.Query("nonce")
	echostr := ctx.Query("echostr")
	if w.check(w.token, signature, timestamp, nonce) {
		_, _ = ctx.Writer.WriteString(echostr)
		return
	}
	log.Printf("check sign fail! signature:%s,timestamp:%s,nonce:%s,echostr:%s",
		signature, timestamp, nonce, echostr)
}

func (w *wechat) Dispatch(ctx *gin.Context) {
	var req WxReq
	if err := ctx.BindXML(&req); err != nil {
		logrus.Error("unmarshal xml fail!", err)
		return
	}
	logrus.Infof("receive xml:%+v", req)
	// idempotent
	if idempotent.Contains(req.MsgID) {
		logrus.WithContext(ctx).Infof("ignore repeat wechat msg:%+v", req)
		return
	}
	idempotent.Add(req.MsgID, true)
	// handle panic
	defer func() {
		if p := recover(); p != nil {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, fmt.Sprintf("something is wrong with %s", p))
		}
	}()
	req.Content = strings.TrimSpace(req.Content)
	req.Content = strings.ReplaceAll(req.Content, " ", " ")
	if strings.HasPrefix(req.Content, "wechat_r_") {
		cmds := strings.Split(req.Content, "_r_")
		_, _ = w.authorSvc.Save(&model.Author{
			AppId:        cmds[1],
			FeishuOpenId: cmds[3],
			WechatOpenId: req.FromUserName,
		})
		if _, err := w.bookSvc.Save(cmds[1], cmds[3], cmds[2], string(model.CategoryWechat)); err != nil {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, err.Error())
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, "绑定成功，可以开始记账啦 \r\n记账格式为： 备注 分类 金额。 \r\n 比如： 泡面 餐费 100 \r\n 或者： 加班费 工资收入 +100 \r\n 不是首次输入，可以忽略分类，比如： 泡面 100")
		return
	}

	if strings.Contains(req.Content, "id") &&
		strings.Contains(req.Content, "secret") &&
		strings.Contains(req.Content, "token") {
		var app model.App
		if err := json.Unmarshal([]byte(req.Content), &app); err != nil {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, "json格式不正确，可以按照这个格式修改 {\"id\":\"cli_a257f60e6bbab00c\",\"secret\":\"TVhkohuKkamGFU3cabXuFhdlLoS3EwhL\",\"token\":\"OzCFbkwGSckR6vo1pM4L7c8HU3j0MoeP\"}")
			return
		}
		if err := w.register(app); err != nil {
			w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, err.Error())
			return
		}
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, fmt.Sprintf("绑定成功，开始使用吧。事件订阅地址\r\n%s/feishu/webhook/%s", config.GetConfig().SeverUrl, app.AppId))
		return
	}
	if author, exists := w.authorSvc.Get(req.FromUserName, model.CategoryWechat); exists {
		msg := facades[author.AppId].BillSvc.Record(author.AppId, author.FeishuOpenId, req.Content, model.CategoryWechat)
		w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, msg)
		return
	}
	w.returnTextMsg(ctx, req.ToUserName, req.FromUserName, "没找到绑定的账号信息")
	return
}

func (w *wechat) register(app model.App) error {
	_, err := w.appSvc.Save(&app)
	if err != nil {
		return err

	}
	err = register(app, w.authorSvc, w.bookSvc)
	if err != nil {
		return err

	}
	return nil
}

func (w *wechat) returnTextMsg(ctx *gin.Context, from, to, content string) {
	res, _ := xml.Marshal(WxResp{
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
