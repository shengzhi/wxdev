package miniapp

import "fmt"

type APIURL string

const (
	url_getPhoneNumber           APIURL = "https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s"
	wxapp_subscribe_message_tmpl APIURL = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s"
	url_soter_verify             APIURL = "https://api.weixin.qq.com/cgi-bin/soter/verify_signature?access_token=%s"
	url_schema_generate          APIURL = "https://api.weixin.qq.com/wxa/generatescheme?access_token=%s"
	url_schema_query             APIURL = "https://api.weixin.qq.com/wxa/queryscheme?access_token=%s"
	url_link_generate            APIURL = "https://api.weixin.qq.com/wxa/generate_urllink?access_token=%s"
	url_link_query               APIURL = "https://api.weixin.qq.com/wxa/query_urllink?access_token=%s"
	url_link_short               APIURL = "https://api.weixin.qq.com/wxa/genwxashortlink?access_token=%s"
	url_sms_send                 APIURL = "https://api.weixin.qq.com/tcb/sendsmsv2?access_token=%s"
	url_activity_create          APIURL = "https://api.weixin.qq.com/cgi-bin/message/wxopen/activityid/create?access_token=%s"
)

func (uri APIURL) Format(args ...interface{}) string {
	return fmt.Sprintf(string(uri), args...)
}
