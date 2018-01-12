// 微信消息管理模块

package wxdev

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// WXEventType 事件类型
type WXEventType string

// 事件类型定义
const (
	EventTypeSubscribe   WXEventType = "subscribe"
	EventTypeUnsubscribe WXEventType = "unsubscribe"
	EventTypeScan        WXEventType = "SCAN"
	EventTypeLocation    WXEventType = "LOCATION"
	EventTypeClick       WXEventType = "CLICK"
	EventTypeView        WXEventType = "VIEW"
	// 自定义菜单事件
	EventTypeScanPush       WXEventType = "scancode_push"
	EventTypeScanWait       WXEventType = "scancode_waitmsg"
	EventTypePhoto          WXEventType = "pic_sysphoto"
	EventTypePhotoOrAlbum   WXEventType = "pic_photo_or_album"
	EventTypePicWeiXin      WXEventType = "pic_weixin"
	EventTypeLocationSelect WXEventType = "location_select"
)

// WXMessageType 微信消息类型
type WXMessageType string

// 微信消息类型定义
const (
	WXMsgTypeEvent      WXMessageType = "event"
	WXMsgTypeText       WXMessageType = "text"
	WXMsgTypeImage      WXMessageType = "image"
	WXMsgTypeVoice      WXMessageType = "voice"
	WXMsgTypeVideo      WXMessageType = "video"
	WXMsgTypeShortVideo WXMessageType = "shortvideo"
	WXMsgTypeLink       WXMessageType = "link"
	WXMsgTypeLocation   WXMessageType = "location"
	WXMsgTypeMusic      WXMessageType = "music"
	WXMsgTypeNews       WXMessageType = "news"
	WXMsgTypeTransferCS WXMessageType = "transfer_customer_service"
)

// WXMessageHandler 消息处理器
type WXMessageHandler func(WXMessageRequest) WXMessageResponse

// WXMessageResponse 微信回复消息
type WXMessageResponse interface{}

// WXMessageOKResponse 返回success
type WXMessageOKResponse struct{}

func (c *WXClient) validateSign(nonce, timestamp, expectSign string) bool {
	params := []string{c.validationToken, nonce, timestamp}
	sort.Strings(params)
	s1 := sha1.New()
	io.WriteString(s1, strings.Join(params, ""))
	actualSign := fmt.Sprintf("%x", s1.Sum(nil))
	return actualSign == expectSign
}

// CheckSignature 微信接入验证
func (c *WXClient) CheckSignature(qryArgs url.Values) (string, error) {
	nonce := qryArgs.Get("nonce")
	timestamp := qryArgs.Get("timestamp")
	expectSign := qryArgs.Get("signature")
	echostr := qryArgs.Get("echostr")
	if c.validateSign(nonce, timestamp, expectSign) {
		return echostr, nil
	}
	return "", fmt.Errorf("微信接入验证失败")
}

// MessageHandleFunc 消息处理函数
func (c *WXClient) MessageHandleFunc(handler func(WXMessageRequest) WXMessageResponse) {
	c.msgHandler = WXMessageHandler(handler)
}

// ServerGin 支持Gin web Framework
func (c *WXClient) ServerGin(ctx *gin.Context) {
	c.ServeHTTP(ctx.Writer, ctx.Request)
}

// ServeHTTP 自定义消息处理器
func (c *WXClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method == "GET" {
		if echo, err := c.CheckSignature(r.URL.Query()); err != nil {
			fmt.Fprintf(w, "failed")
			w.WriteHeader(200)
		} else {
			fmt.Fprintf(w, "%s", echo)
			w.WriteHeader(200)
		}
		return
	}
	d := xml.NewDecoder(r.Body)
	var msgReq WXMessageRequest
	if err := d.Decode(&msgReq); err != nil {
		log.Println("Decode wechat message request failed,error:", err)
		fmt.Fprintf(w, "success")
		w.WriteHeader(200)
		return
	}
	if c.msgHandler == nil {
		fmt.Fprintf(w, "success")
		w.WriteHeader(200)
		return
	}
	resp := c.msgHandler(msgReq)
	if _, ok := resp.(WXMessageOKResponse); ok {
		fmt.Fprintf(w, "success")
		w.WriteHeader(200)
		return
	}
	coder := xml.NewEncoder(w)
	coder.Encode(resp)
	coder.Flush()
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(200)
}

// WXMessageRequest 微信消息
type WXMessageRequest struct {
	XMLName                            xml.Name      `xml:"xml"`
	ToUserName                         string        `xml:",omitempty"`
	FromUserName                       string        `xml:",omitempty"`
	CreateTime                         int64         `xml:",omitempty"`
	MsgType                            WXMessageType `xml:",omitempty"`
	Content                            string        `xml:",omitempty"`
	MsgId                              int64         `xml:",omitempty"`
	PicUrl                             string        `xml:",omitempty"`
	MediaId                            string        `xml:",omitempty"`
	ThumbMediaId                       string        `xml:",omitempty"` //视频消息缩略图的媒体id，可以调用多媒体文件下载接口拉取数据。
	Format                             string        `xml:",omitempty"`
	Recognition                        string        `xml:",omitempty"` //语音识别结果，UTF8编码
	Location_X                         float64       `xml:",omitempty"`
	Location_Y                         float64       `xml:",omitempty"`
	Scale                              int           `xml:",omitempty"` //地图缩放大小
	Poiname, Label                     string        `xml:",omitempty"` // 地理位置信息
	Title                              string        `xml:",omitempty"`
	Description                        string        `xml:",omitempty"`
	Url                                string        `xml:",omitempty"`
	Event                              WXEventType   `xml:",omitempty"`
	EventKey                           string        `xml:",omitempty"`
	Ticket                             string        `xml:",omitempty"`
	Latitude, Longitude, Precision     float64       `xml:",omitempty"`
	ScanCodeInfo, ScanType, ScanResult string        `xml:",omitempty"`
	SendPicsInfo                       PicInfo       `xml:",omitempty"`
}

// PicInfo 发送的图片信息
type PicInfo struct {
	Count       int      `xml:",omitempty"`
	MD5SumItems []string `xml:"PicList>item>PicMd5Sum,omitempty"`
}

// IsEvent 下发消息是否来源于事件
func (r WXMessageRequest) IsEvent() bool {
	return r.MsgType == WXMsgTypeEvent
}

// IsScanEvent 消息是否为扫码事件
func (r WXMessageRequest) IsScanEvent() bool {
	if r.MsgType != WXMsgTypeEvent {
		return false
	}
	if r.Event == EventTypeScan {
		return true
	}
	if r.Event == EventTypeSubscribe && strings.HasPrefix(r.EventKey, "qrscene_") {
		return true
	}
	return false
}

type WXMsgResponse struct {
	XMLName      xml.Name      `xml:"xml"`
	ToUserName   string        `xml:",omitempty"`
	FromUserName string        `xml:",omitempty"`
	CreateTime   int64         `xml:",omitempty"`
	MsgType      WXMessageType `xml:",omitempty"`
}

// WXTextMsgResponse 回复微信消息
type WXTextMsgResponse struct {
	WXMsgResponse
	Content CDataContent `xml:",omitempty"`
}

type CDataContent struct {
	Text string `xml:",cdata"`
}

// CDataWrap 包装为CDATA 内容
func CDataWrap(content string) CDataContent {
	return CDataContent{Text: content}
}

// WXMediaMsgResponse 图片/语音体消息
type WXMediaMsgResponse struct {
	WXMsgResponse
	MediaID     CDataContent `xml:"MediaId,omitempty"`
	Title       CDataContent `xml:"Title,omitempty"`
	Description CDataContent `xml:"Description,omitempty"`
}

// WXMusicMsgResponse 音乐消息
type WXMusicMsgResponse struct {
	WXMsgResponse
	Title        CDataContent `xml:",omitempty"`
	Description  CDataContent `xml:",omitempty"`
	MusicURL     CDataContent `xml:",omitempty"`
	HQMusicUrl   CDataContent `xml:",omitempty"`
	ThumbMediaId CDataContent `xml:",omitempty"`
}

// WXArticleMsgResponse 图文消息
type WXArticleMsgResponse struct {
	WXMsgResponse
	ArticleCount int
	Articles     []ArticleItem `xml:"Articles>item"`
}

// ArticleItem 图文消息文章列表
type ArticleItem struct {
	Title       CDataContent
	Description CDataContent
	PicUrl      CDataContent
	Url         CDataContent
}

// WXTransferCSResponse 转发消息至微信客服系统
type WXTransferCSResponse struct {
	WXMsgResponse
	Account CDataContent `xml:"TransInfo>KfAccount,omitempty"`
}

// NewTextMsg 创建文本消息
func NewTextMsg(from, to, content string) WXTextMsgResponse {
	return WXTextMsgResponse{
		WXMsgResponse: WXMsgResponse{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   time.Now().Unix(),
			MsgType:      WXMsgTypeText,
		},
		Content: CDataWrap(content),
	}
}

// NewImgMsg 图片消息
func NewImgMsg(from, to, mediaid string) WXMediaMsgResponse {
	return WXMediaMsgResponse{
		WXMsgResponse: WXMsgResponse{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   time.Now().Unix(),
			MsgType:      WXMsgTypeImage,
		},
		MediaID: CDataWrap(mediaid),
	}
}

// NewVoiceMsg 语音消息
func NewVoiceMsg(from, to, mediaid string) WXMediaMsgResponse {
	return WXMediaMsgResponse{
		WXMsgResponse: WXMsgResponse{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   time.Now().Unix(),
			MsgType:      WXMsgTypeVoice,
		},
		MediaID: CDataWrap(mediaid),
	}
}

// NewVideoMsg 视频消息
func NewVideoMsg(from, to, mediaid, title, desc string) WXMediaMsgResponse {
	return WXMediaMsgResponse{
		WXMsgResponse: WXMsgResponse{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   time.Now().Unix(),
			MsgType:      WXMsgTypeVideo,
		},
		MediaID:     CDataWrap(mediaid),
		Title:       CDataWrap(title),
		Description: CDataWrap(desc),
	}
}

// NewMusicMsg 音乐消息
func NewMusicMsg(from, to, mediaid, title, desc, uri, hquri string) WXMusicMsgResponse {
	return WXMusicMsgResponse{
		WXMsgResponse: WXMsgResponse{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   time.Now().Unix(),
			MsgType:      WXMsgTypeMusic,
		},
		ThumbMediaId: CDataWrap(mediaid),
		Title:        CDataWrap(title),
		Description:  CDataWrap(desc),
		MusicURL:     CDataWrap(uri),
		HQMusicUrl:   CDataWrap(hquri),
	}
}

// NewArticleMsg 图文消息
func NewArticleMsg(from, to string, items ...ArticleItem) WXArticleMsgResponse {
	msg := WXArticleMsgResponse{
		WXMsgResponse: WXMsgResponse{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   time.Now().Unix(),
			MsgType:      WXMsgTypeNews,
		},
		ArticleCount: len(items),
	}
	msg.Articles = append(msg.Articles, items...)
	return msg
}

// TransferCustomerService 创建转发客服系统响应
func TransferCustomerService(from, to, account string) WXTransferCSResponse {
	msg := WXTransferCSResponse{
		WXMsgResponse: WXMsgResponse{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   time.Now().Unix(),
			MsgType:      WXMsgTypeTransferCS,
		},
	}
	if account != "" {
		msg.Account = CDataWrap(account)
	}
	return msg
}
