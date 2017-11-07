// 微信小程序开发工具包

package miniapp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shengzhi/wxdev/crypt"
)

// OptionFunc 配置函数
type OptionFunc func(*WXMiniClient)

// option 配置
type option struct {
	appid, appsecret string
}

// WXMiniClient 微信小程序客户端
type WXMiniClient struct {
	opt option
}

// NewClient 创建客户端
func NewClient(appid, secret string) *WXMiniClient {
	return &WXMiniClient{
		opt: option{
			appid: appid, appsecret: secret,
		},
	}
}

// WXAppSession 微信小程序会话
type WXAppSession struct {
	ErrCode    int
	ErrMsg     string
	OpenID     string
	SessionKey string `json:"session_key"`
}

const sessionkey_url = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"

func (c *WXMiniClient) httpGet(uri string, v interface{}) error {
	res, err := http.Get(uri)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	d := json.NewDecoder(res.Body)
	err = d.Decode(v)
	if err != nil {
		return fmt.Errorf("Decode JSON error:%v", err)
	}
	return nil
}

func (c *WXMiniClient) httpPost(uri string, data, v interface{}) error {
	var buf bytes.Buffer
	w := json.NewEncoder(&buf)
	w.Encode(data)
	res, err := http.Post(uri, "application/json", &buf)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	d := json.NewDecoder(res.Body)
	err = d.Decode(v)
	if err != nil {
		return fmt.Errorf("Decode JSON error:%v", err)
	}
	return nil
}

// GetSessionKey 获取小程序session key
func (c *WXMiniClient) GetSessionKey(code string) (WXAppSession, error) {
	uri := fmt.Sprintf(sessionkey_url, c.opt.appid, c.opt.appsecret, code)
	var s WXAppSession
	err := c.httpGet(uri, &s)
	return s, err
}

// WXAppSign 小程序签名验证
func (c *WXMiniClient) WXAppSign(rawdata, sessionkey string) string {
	var cipherTxt bytes.Buffer
	cipherTxt.WriteString(rawdata)
	cipherTxt.WriteString(sessionkey)
	return crypt.SHA1(cipherTxt.Bytes())
}

// WXAppDecript 小程序解密
func (c *WXMiniClient) WXAppDecript(crypted, sessionkey, iv string) ([]byte, error) {
	cryptedByte, _ := base64.StdEncoding.DecodeString(crypted)
	key, _ := base64.StdEncoding.DecodeString(sessionkey)
	ivbyte, _ := base64.StdEncoding.DecodeString(iv)
	return crypt.AESDecrypt(cryptedByte, key, ivbyte)
}
