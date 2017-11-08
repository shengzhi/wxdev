// 公众号模板消息
package wxdev

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// TmplData 模板数据
type TmplData struct {
	ToUser     string                   `json:"touser"`
	TemplateID string                   `json:"template_id"`
	URL        string                   `json:"url"`
	FontColor  string                   `json:"color"`
	Data       map[string]tmplFieldData `json:"data"`
	MiniApp    struct {
		AppID    string `json:"appid"`
		PagePath string `json:"pagepath"`
	} `json:"miniprogram"`
}

type tmplFieldData struct {
	Value string `json:"value"`
	Color string `json:"color"`
}

// NewTmplData 创建模板对象
func NewTmplData(openid, tmplid string) *TmplData {
	return &TmplData{
		TemplateID: tmplid,
		ToUser:     openid,
		Data:       make(map[string]tmplFieldData, 0),
	}
}

// LinkMiniApp 设置模板消息跳转至的小程序
func (t *TmplData) LinkMiniApp(appid, page string) {
	t.MiniApp.AppID, t.MiniApp.PagePath = appid, page
}

// Put 追加数据项
func (t *TmplData) Put(key, value, color string) {
	t.Data[key] = tmplFieldData{Value: value, Color: color}
}

// TmplMessageSendReply 模板消息发送结果
type TmplMessageSendReply struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	MsgID   int64  `json:"msgid"`
}

const tmpl_message_url = "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s"

// SendTmplMessage 发送模板消息
func (c *WXClient) SendTmplMessage(data *TmplData) (int64, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return 0, err
	}
	uri := fmt.Sprintf(tmpl_message_url, token)
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return 0, err
	}
	var res *http.Response
	res, err = http.Post(uri, "application/json", &buf)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return 0, err
	}
	var result TmplMessageSendReply
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return 0, err
	}
	if result.ErrCode != 0 {
		return 0, fmt.Errorf("WXDev: Send template message failed, error code:%d, message:%s", result.ErrCode, result.ErrMsg)
	}
	return result.MsgID, nil
}
