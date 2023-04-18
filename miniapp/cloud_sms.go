// 云开发-短信.
package miniapp

type (
	SMSSendReq struct {
		// Env 云开发环境ID.
		Env  string  `json:"env"`
		Link URLLink `json:"url_link"`
		// TmplId 短信模版 ID。(844110: 营销类短信模版 ID).
		TmplId string `json:"template_id"`
		// Params 短信模版变量数组.
		Params []string `json:"template_param_list"`
		// Phones 手机号列表，单次请求最多支持 1000 个境内手机号，手机号必须以+86开头.
		Phones []string `json:"phone_number_list"`
		// UseShortName 是否使用小程序简称.
		UseShortName bool `json:"use_short_name,omitempty"`
		// ResourceAppID 资源方appid，第三方代开发时可填第三方appid或小程序appid，应为所填环境所属的账号APPID.
		ResourceAppID string `json:"resource_appid"`
	}
	SMSSendResp struct {
		Results []struct {
			SerialNo    string `json:"serial_no"`
			PhoneNumber string `json:"phone_number"`
			Code        string `json:"code"`
			Message     string `json:"message"`
			// IsoCode 国家码或地区码.
			IsoCode string `json:"iso_code"`
		} `json:"send_status_list"`
	}
)

// SendSMS 发送携带 URL Link 的短信.
func (c *WXMiniClient) SendSMS(req SMSSendReq) (SMSSendResp, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return SMSSendResp{}, err
	}
	uri := url_sms_send.Format(token)
	var resp struct {
		reply
		SMSSendResp
	}
	err = c.httpPost(uri, req, &resp)
	if err != nil {
		return resp.SMSSendResp, err
	}
	if err = resp.Error(); err != nil {
		return resp.SMSSendResp, err
	}
	return resp.SMSSendResp, nil
}
