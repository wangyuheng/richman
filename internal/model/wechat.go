package model

import "encoding/xml"

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
