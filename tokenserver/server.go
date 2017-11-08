// Token server

package tokenserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/groupcache/singleflight"
)

type tokenReply struct {
	ErrCode int32  `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Token   string `json:"access_token"`
	Expires int64  `json:"expires_in"`
}

func (r tokenReply) checkErr() error {
	if r.ErrCode != 0 {
		return fmt.Errorf("Code:%d,Message: %s", r.ErrCode, r.ErrMsg)
	}
	return nil
}

type ticketReply struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Ticket  string `json:"ticket"`
	Expires int64  `json:"expires_in"`
}

func (r ticketReply) checkErr() error {
	if r.ErrCode != 0 {
		return fmt.Errorf("Code:%d,Message: %s", r.ErrCode, r.ErrMsg)
	}
	return nil
}

type app struct {
	appid, secret string
	token         tokenReply
	expiredTime   time.Time
}

// isValid 是否有效
func (a app) isValid() bool {
	return time.Now().Before(a.expiredTime)
}

type ticketApp struct {
	appid       string
	ticket      ticketReply
	expiredTime time.Time
}

// isValid 是否有效
func (ta ticketApp) isValid() bool {
	return time.Now().Before(ta.expiredTime)
}

// Server Token 中控服务
type Server struct {
	apps        map[string]app
	ticketApps  map[string]ticketApp
	flightGroup singleflight.Group
}

// NewServer 创建Server程序
func NewServer() *Server {
	return &Server{
		apps:       make(map[string]app, 0),
		ticketApps: make(map[string]ticketApp, 0),
	}
}

// Register 注册微信APP账号
func (s *Server) Register(appid, secret string) {
	s.apps[appid] = app{
		appid: appid, secret: secret, token: tokenReply{},
	}
}

// GetToken 获取Access Token
func (s *Server) GetToken(appid string) (string, string, error) {
	app, has := s.apps[appid]
	if !has {
		return "", "", fmt.Errorf("APPID is not registered")
	}

	if app.isValid() {
		return app.token.Token, app.expiredTime.Format("2006-01-02 15:04:05"), nil
	}
	token, err := s.flightGroup.Do(appid, func() (interface{}, error) {
		var reply tokenReply
		err := s.getAccessToken(app.appid, app.secret, &reply)
		if err != nil {
			return nil, err
		}
		return reply, reply.checkErr()
	})
	if err != nil {
		return "", "", err
	}
	app = s.apps[appid]
	if app.isValid() {
		return app.token.Token, app.expiredTime.Format("2006-01-02 15:04:05"), nil
	}
	app.token = token.(tokenReply)
	app.expiredTime = time.Now().Add(time.Second * time.Duration(app.token.Expires))
	s.apps[app.appid] = app
	return app.token.Token, app.expiredTime.Format("2006-01-02 15:04:05"), nil
}

func (s *Server) getAccessToken(appid, secret string, v interface{}) error {
	uri := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	res, err := http.Get(fmt.Sprintf(uri, appid, secret))
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(v)
}

// GetJSAPITicket 获取微信JSAPI_Ticket
func (s *Server) GetJSAPITicket(appid string) (string, string, error) {
	app, has := s.ticketApps[appid]
	if has && app.isValid() {
		return app.ticket.Ticket, app.expiredTime.Format("2006-01-02 15:04:05"), nil
	}
	key := fmt.Sprintf("ticket_%s", appid)
	resp, err := s.flightGroup.Do(key, func() (interface{}, error) {
		var reply ticketReply
		err := s.doGetTicket(appid, &reply)
		return reply, err
	})
	if err != nil {
		return "", "", err
	}
	app = ticketApp{appid: appid, ticket: resp.(ticketReply)}
	if err = app.ticket.checkErr(); err != nil {
		return "", "", err
	}
	app.expiredTime = time.Now().Add(time.Second * time.Duration(app.ticket.Expires))
	s.ticketApps[appid] = app
	return app.ticket.Ticket, app.expiredTime.Format("2006-01-02 15:04:05"), nil
}

func (s *Server) doGetTicket(appid string, v interface{}) error {
	token, _, err := s.GetToken(appid)
	if err != nil {
		return err
	}
	const uri = "https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=%s&type=jsapi"
	res, err := http.Get(fmt.Sprintf(uri, token))
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	return json.NewDecoder(res.Body).Decode(v)
}

// Run 启动server 并监听HTTP端口
func (s *Server) Run(port int) {
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		appid := r.URL.Query().Get("appid")
		if appid == "" {
			w.WriteHeader(400)
			return
		}
		token, expired, err := s.GetToken(appid)
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, `{"errcode":%d,"errmsg":%v}`, 500, err)
		} else {
			fmt.Fprintf(w, `{"token":"%s","expired":"%s"}`, token, expired)
		}
		w.Header().Set("Content-Type", "text/json")
	})
	http.HandleFunc("/jsapiticket", func(w http.ResponseWriter, r *http.Request) {
		appid := r.URL.Query().Get("appid")
		if appid != "" {
			w.WriteHeader(400)
			return
		}
		ticket, expired, err := s.GetJSAPITicket(appid)
		if err != nil {
			log.Println(err)
			fmt.Fprintf(w, `{"errcode":%d,"errmsg":%v}`, 500, err)
		} else {
			fmt.Fprintf(w, `{"ticket":"%s","expired":"%s"}`, ticket, expired)
		}
		w.Header().Set("Content-Type", "text/json")
	})

	httpServer := http.Server{Addr: fmt.Sprintf(":%d", port)}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatalf("启动HttpServer失败,错误:%v\r\n", err)
			}
		}
	}()
	<-stop
	fmt.Println("server is stopping...")
	httpServer.Shutdown(context.Background())
	fmt.Println("server is gracefully stopped")
}
