package wxdev

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/golang/groupcache/singleflight"
	"github.com/shengzhi/util/helper"
)

var ErrNoTokenServer = errors.New("No specify token server")

// OptionFunc 配置函数
type OptionFunc func(*WXClient)

// WithTokenServer 设置Token server
func WithTokenServer(uri string) OptionFunc {
	return func(c *WXClient) {
		var err error
		c.tokenServerURL, err = url.Parse(uri)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func WithValidationToken(token string) OptionFunc {
	return func(c *WXClient) {
		c.validationToken = token
	}
}

// WXClient 公众号客户端
type WXClient struct {
	appid          string
	tokenServerURL *url.URL
	jsapiTicket    struct {
		ticket      string
		expiredTime time.Time
	}
	validationToken string
	msgHandler      WXMessageHandler
	flightG         singleflight.Group
	isdebug         bool
}

// NewWXClient 创建公众号客户端
func NewWXClient(appid string, options ...OptionFunc) *WXClient {
	c := &WXClient{appid: appid}
	for _, fn := range options {
		fn(c)
	}
	return c
}

// EnableDebug 启用debug模式
func (c *WXClient) EnableDebug() {
	c.isdebug = true
}
func (c *WXClient) dumpRequest(req *http.Request) {
	if c.isdebug {
		data, _ := httputil.DumpRequest(req, true)
		fmt.Println(string(data))
	}
}
func (c *WXClient) dumpResponse(resp *http.Response) {
	if c.isdebug {
		data, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(data))
	}
}
func (c *WXClient) httpDo(req *http.Request) (*http.Response, error) {
	c.dumpRequest(req)
	client := http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	resp, err := client.Do(req)
	c.dumpResponse(resp)
	return resp, err
}

func (c *WXClient) httpGet(uri string, v interface{}) error {
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

func (c *WXClient) httpPost(uri string, data, v interface{}) error {
	body, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(body))
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

// JSSDKSignature JSSDK 签名对象
type JSSDKSignature struct {
	AppID, Noncestr, Sign string
	Timestamp             int64
}

var randomeGenerator = helper.CreateRandomGenerator(helper.DigitsAndLetters)

// GenJSAPISign 生成JSSDK签名
func (c *WXClient) GenJSAPISign(u *url.URL) (JSSDKSignature, error) {
	ticket, err := c.jsapitkt()
	if err != nil {
		return JSSDKSignature{}, err
	}
	uri := fmt.Sprintf("http://%s%s", u.Host, u.RequestURI())
	noncestr := randomeGenerator(16)
	timestamp := time.Now().Unix()
	plainTxt := []byte(fmt.Sprintf("jsapi_ticket=%s&noncestr=%s&timestamp=%d&url=%s",
		ticket, noncestr, timestamp, uri))
	h := sha1.New()
	h.Write(plainTxt)
	b := h.Sum(nil)
	sign := hex.EncodeToString(b)
	return JSSDKSignature{AppID: c.appid, Noncestr: noncestr, Timestamp: timestamp, Sign: sign}, nil
}

type wxreply struct {
	Ticket, Expired string
}

func (c *WXClient) jsapitkt() (string, error) {
	if time.Now().Before(c.jsapiTicket.expiredTime) {
		return c.jsapiTicket.ticket, nil
	}
	resp, err := c.flightG.Do("getjsapiticket", func() (interface{}, error) {
		var result wxreply
		err := c.getJSAPITicket(&result)
		return result, err
	})
	if err != nil {
		return "", err
	}
	reply := resp.(wxreply)
	c.jsapiTicket.ticket = reply.Ticket
	c.jsapiTicket.expiredTime, _ = time.ParseInLocation("2006-01-02 15:04:05", reply.Expired, time.Local)
	return c.jsapiTicket.ticket, nil
}

func (c *WXClient) getJSAPITicket(v interface{}) error {
	if c.tokenServerURL == nil {
		return ErrNoTokenServer
	}
	u, _ := url.Parse(fmt.Sprintf("jsapiticket?appid=%s", c.appid))
	return c.httpGet(c.tokenServerURL.ResolveReference(u).String(), v)
}

func (c *WXClient) getAccessToken() (string, error) {
	if c.tokenServerURL == nil {
		return "", ErrNoTokenServer
	}
	var err error
	var token string
	for i := 0; i < 5; i++ {
		resp, err := c.flightG.Do("getaccesstoken", func() (interface{}, error) {
			u, _ := url.Parse(fmt.Sprintf("token?appid=%s", c.appid))
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
