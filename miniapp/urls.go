package miniapp

import "fmt"

type APIURL string

const (
	url_getPhoneNumber           APIURL = "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s"
	wxapp_subscribe_message_tmpl APIURL = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s"
)

func (uri APIURL) Format(args ...interface{}) string {
	return fmt.Sprintf(string(uri), args...)
}
