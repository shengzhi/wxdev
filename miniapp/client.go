// 微信小程序开发工具包

package miniapp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/golang/groupcache/singleflight"
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
	opt            option
	tokenServerURL *url.URL
	flightG        singleflight.Group
}

// NewClient 创建客户端
func NewClient(appid, secret string, options ...OptionFunc) *WXMiniClient {
	c := &WXMiniClient{
		opt: option{
			appid: appid, appsecret: secret,
		},
	}
	for _, fn := range options {
		fn(c)
	}
	return c
}

// AppID 返回当前小程序APP ID
func (c *WXMiniClient) AppID() string { return c.opt.appid }

// WXAppSession 微信小程序会话
type WXAppSession struct {
	ErrCode    int
	ErrMsg     string
	OpenID     string
	UnionID    string
	SessionKey string `json:"session_key"`
}

const sessionkey_url = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"

type reply struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (rep reply) Error() error {
	if rep.ErrCode == 0 {
		return nil
	}
	return fmt.Errorf("code:%d,errmsg:%s", rep.ErrCode, rep.ErrMsg)
}

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
	if s.ErrCode != 0 {
		err = fmt.Errorf("%d-%s", s.ErrCode, s.ErrMsg)
	}
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

func (c *WXMiniClient) getAccessToken() (string, error) {
	var err error
	var token string
	for i := 0; i < 5; i++ {
		resp, err := c.flightG.Do("getaccesstoken", func() (interface{}, error) {
			u, _ := url.Parse(fmt.Sprintf("token?appid=%s", c.opt.appid))
			var reply struct{ Token string }
			err := c.httpGet(c.tokenServerURL.ResolveReference(u).String(), &reply)
			return reply.Token, err
		})
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		token = resp.(string)
		break
	}
	return token, err
}

type WXSexType byte

func (t WXSexType) String() string {
	switch t {
	case 1:
		return "M"
	case 2:
		return "F"
	default:
		return "U"
	}
}

// WXUserInfo 微信小程序用户信息
type WXUserInfo struct {
	OpenID     string    `json:"openid"`
	NickName   string    `json:"nickname"`
	Gender     WXSexType `json:"gender"`
	Language   string    `json:"language"`
	City       string    `json:"city"`
	Province   string    `json:"province"`
	Country    string    `json:"country"`
	HeadImgUrl string    `json:"avatarUrl"`
	UnionID    string    `json:"unionId"`
	WaterMark  struct {
		AppID     string `json:"appid"`
		Timestamp int64  `json:"timestamp"`
	} `json:"watermark"`
}

// GetUserInfo 获取微信用户信息
func (c *WXMiniClient) GetUserInfo(iv, cipherTxt, sessionKey string) (WXUserInfo, error) {
	data, err := c.WXAppDecript(cipherTxt, sessionKey, iv)
	if err != nil {
		return WXUserInfo{}, err
	}
	var user WXUserInfo
	err = json.Unmarshal(data, &user)
	return user, err
}

// WXPhoneInfo 微信账号绑定电话信息
type WXPhoneInfo struct {
	Phone     string `json:"phoneNumber"`
	PurePhone string `json:"purePhoneNumber"`
	Country   string `json:"countryCode"`
	WaterMark struct {
		AppID     string `json:"appid"`
		Timestamp int64  `json:"timestamp"`
	} `json:"watermark"`
}

// GetPhoneNumber 获取微信绑定电话号码
//
//	func (c *WXMiniClient) GetPhoneNumber(iv, cipherTxt, sessionKey string) (WXPhoneInfo, error) {
//		data, err := c.WXAppDecript(cipherTxt, sessionKey, iv)
//		if err != nil {
//			return WXPhoneInfo{}, err
//		}
//		var phone WXPhoneInfo
//		err = json.Unmarshal(data, &phone)
//		return phone, err
//	}
func (c *WXMiniClient) GetPhoneNumber(code string) (WXPhoneInfo, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return WXPhoneInfo{}, err
	}
	var resp struct {
		reply
		PhoneInfo WXPhoneInfo `json:"phone_info"`
	}
	req := map[string]string{"code": code}
	err = c.httpPost(url_getPhoneNumber.Format(token), req, &resp)
	if err != nil {
		return WXPhoneInfo{}, err
	}
	if err = resp.Error(); err != nil {
		return WXPhoneInfo{}, err
	}
	return resp.PhoneInfo, nil
}
