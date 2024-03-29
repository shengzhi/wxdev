// 微信小程序开发工具包

package miniapp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
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
	httpcli        *http.Client
	isdebug        bool
}

// NewClient 创建客户端
func NewClient(appid, secret string, options ...OptionFunc) *WXMiniClient {
	c := &WXMiniClient{
		httpcli: &http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		},
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

func (c *WXMiniClient) dumpRequest(req *http.Request) {
	if c.isdebug {
		data, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(data))
	}
}
func (c *WXMiniClient) dumpResponse(resp *http.Response) {
	if c.isdebug {
		data, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(data))
	}
}

func (c *WXMiniClient) httpDo(req *http.Request) (*http.Response, error) {
	c.dumpRequest(req)
	resp, err := c.httpcli.Do(req)
	if resp != nil {
		c.dumpResponse(resp)
	}
	return resp, err
}

func (c *WXMiniClient) httpGet(uri string, v interface{}) error {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}
	res, err := c.httpDo(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(v)
}

func (c *WXMiniClient) httpPost(uri string, data, v interface{}) error {
	var buf bytes.Buffer
	coder := json.NewEncoder(&buf)
	coder.SetEscapeHTML(false)
	if err := coder.Encode(data); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", uri, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpDo(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(v)
}
func (c *WXMiniClient) EnableDebug() { c.isdebug = true }

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
	u, err := url.Parse(fmt.Sprintf("token?appid=%s", c.opt.appid))
	if err != nil {
		return "", err
	}
	tokenUri := c.tokenServerURL.ResolveReference(u).String()
	for i := 0; i < 5; i++ {
		resp, err := c.flightG.Do("getaccesstoken", func() (interface{}, error) {
			var reply struct{ Token string }
			err := c.httpGet(tokenUri, &reply)
			return reply.Token, err
		})
		if err != nil {
			if c.isdebug {
				fmt.Printf("get access_token error:%v,url:%s", err, tokenUri)
			}
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

// GetPhoneNumber 获取微信绑定电话号码.
func (c *WXMiniClient) DecryptPhoneNumber(iv, cipherTxt, sessionKey string) (WXPhoneInfo, error) {
	data, err := c.WXAppDecript(cipherTxt, sessionKey, iv)
	if err != nil {
		return WXPhoneInfo{}, err
	}
	var phone WXPhoneInfo
	err = json.Unmarshal(data, &phone)
	return phone, err
}

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
