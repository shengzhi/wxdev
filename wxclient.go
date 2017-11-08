package wxdev

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

// WXClient 公众号客户端
type WXClient struct {
	appid          string
	tokenServerURL *url.URL
	jsapiTicket    struct {
		ticket      string
		expiredTime time.Time
	}
	flightG singleflight.Group
}

// NewWXClient 创建公众号客户端
func NewWXClient(appid string, options ...OptionFunc) *WXClient {
	c := &WXClient{appid: appid}
	for _, fn := range options {
		fn(c)
	}
	return c
}

func (c *WXClient) httpGet(uri string, v interface{}) error {
	res, err := http.Get(uri)
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
	res, err := http.Post(uri, "application/json", bytes.NewBuffer(body))
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

// MediaObject 多媒体对象
type MediaObject struct {
	FileName string
	Type     string
	Data     io.ReadCloser
	Size     int64
}

// DownloadMedia 下载多媒体文件
func (c *WXClient) DownloadMedia(mediaid string) (*MediaObject, error) {
	const uri = "http://file.api.weixin.qq.com/cgi-bin/media/get?access_token=%s&media_id=%s"
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	var mediaObj *MediaObject
	res, err := http.Get(fmt.Sprintf(uri, token, mediaid))
	if err != nil {
		if res != nil {
			res.Body.Close()
		}
		return nil, err
	}
	mediaObj = &MediaObject{}
	mediaObj.Type = res.Header.Get("Content-Type")
	mediaObj.Data = res.Body
	mediaObj.Size, _ = strconv.ParseInt(res.Header.Get("Content-Lengt"), 10, 64)
	disp := res.Header.Get("Content-disposition")
	if len(disp) > 0 {
		slice := strings.Split(disp, "filename=")
		if len(slice) > 1 {
			mediaObj.FileName = strings.Trim(slice[1], `"`)
		}
	}
	return mediaObj, nil
}
