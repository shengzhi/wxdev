// 小程序模板消息

package miniapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type MiniProgramState string

const (
	State_Developer MiniProgramState = "developer"
	State_Trial     MiniProgramState = "trial"
	State_Formal    MiniProgramState = "formal"
)

// TmplData 微信小程序模板消息
type TmplData struct {
	ToUser     string                   `json:"touser"`
	TemplateID string                   `json:"template_id"`
	Page       string                   `json:"page"`
	FormID     string                   `json:"form_id"`          // Prepayid or form id
	Keyword    string                   `json:"emphasis_keyword"` // 模板需要放大的关键词，不填则默认无放大
	Data       map[string]tmplFieldData `json:"data"`
	FontColor  string                   `json:"color"` // 模板内容字体的颜色，不填默认黑色
}

type tmplFieldData struct {
	Value string `json:"value"`
	Color string `json:"color,omitempty"`
}

// NewTmplData 创建模板
func NewTmplData(openid, templateid, formid string) *TmplData {
	return &TmplData{
		ToUser:     openid,
		TemplateID: templateid,
		FormID:     formid,
		Data:       make(map[string]tmplFieldData),
	}
}

// Link 设置跳转页
func (t *TmplData) Link(page string) {
	t.Page = page
}

// Put 追加数据项
func (t *TmplData) Put(key, value, color string) {
	t.Data[key] = tmplFieldData{Value: value, Color: color}
}

const wxapp_tmpl_message_url = "https://api.weixin.qq.com/cgi-bin/message/wxopen/template/send?access_token=%s"

// SendWXAppTemplate 发送微信小程序模板
func (c *WXMiniClient) SendWXAppTemplate(data *TmplData) error {
	token, err := c.getAccessToken()
	if err != nil {
		return err
	}
	uri := fmt.Sprintf(wxapp_tmpl_message_url, token)
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return err
	}
	var res *http.Response
	res, err = http.Post(uri, "application/json", &buf)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("Send template message failed, error code:%d, message:%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

type SubscribeMsgTmpl struct {
	ToUser     string                   `json:"touser"`
	TemplateID string                   `json:"template_id"`
	Page       string                   `json:"page"`
	State      MiniProgramState         `json:"miniprogram_state"`
	Lang       string                   `json:"lang"` // 进入小程序查看”的语言类型，支持zh_CN(简体中文)、en_US(英文)、zh_HK(繁体中文)、zh_TW(繁体中文)，默认为zh_CN
	Data       map[string]tmplFieldData `json:"data"`
}

func NewSubscribeMsgTmpl(openid, templateid, page string) *SubscribeMsgTmpl {
	return &SubscribeMsgTmpl{
		ToUser:     openid,
		TemplateID: templateid,
		Lang:       "zh_CN",
		State:      State_Formal,
		Page:       page,
		Data:       make(map[string]tmplFieldData),
	}
}

// Put 追加数据项
func (t *SubscribeMsgTmpl) Put(key, value string) *SubscribeMsgTmpl {
	t.Data[key] = tmplFieldData{Value: value}
	return t
}

const wxapp_subscribe_message_tmpl = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s"

// SendSubscribeMsg 发送订阅消息.
func (c *WXMiniClient) SendSubscribeMsg(tmpl SubscribeMsgTmpl) error {
	token, err := c.getAccessToken()
	if err != nil {
		return err
	}
	uri := fmt.Sprintf(wxapp_subscribe_message_tmpl, token)
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(tmpl)
	if err != nil {
		return err
	}
	var res *http.Response
	res, err = http.Post(uri, "application/json", &buf)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err = json.NewDecoder(res.Body).Decode(&result); err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("Send subscribe message failed, error code:%d, message:%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}
