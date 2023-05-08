// 小程序链接.
package miniapp

type URLSchemaGenReq struct {
	JumpWxa struct {
		// Path 通过 scheme 码进入的小程序页面路径，必须是已经发布的小程序存在的页面，不可携带 query。path 为空时会跳转小程序主页.
		Path string `json:"path,omitempty"`
		// Query 通过 scheme 码进入小程序时的 query，最大1024个字符，只支持数字，大小写英文以及部分特殊字符：`!#$&'()*+,/:;=?@-._~%`.
		Query string `json:"query,omitempty"`
		// 默认值"release"。要打开的小程序版本。正式版为"release"，体验版为"trial"，开发版为"develop"，仅在微信外打开时生效.
		Env string `json:"env_version,omitempty"`
	} `json:"jump_wxa"`
}

type OpenLink string

// GenerateURLSchema 该接口用于获取小程序 scheme 码，适用于短信、邮件、外部网页、微信内等拉起小程序的业务场景。
// 通过该接口，可以选择生成到期失效和永久有效的小程序码，有数量限制，目前仅针对国内非个人主体的小程序开放,
// 详情参考 https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/url-scheme.html.
func (c *WXMiniClient) GenerateURLSchema(req URLSchemaGenReq) (OpenLink, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return OpenLink(""), err
	}
	uri := url_schema_generate.Format(token)
	var resp struct {
		reply
		OpenLink OpenLink `json:"openlink"`
	}
	err = c.httpPost(uri, req, &resp)
	if err != nil {
		return resp.OpenLink, err
	}
	if err = resp.Error(); err != nil {
		return resp.OpenLink, err
	}
	return resp.OpenLink, nil
}

type URLSchemaDetail struct {
	SchemaInfo struct {
		AppID      string `json:"appid"`
		Path       string `json:"path"`
		Query      string `json:"query"`
		CreateTime int64  `json:"create_time"`
		ExpireTime int64  `json:"expire_time"`
		Env        string `json:"env_version"`
	} `json:"scheme_info"`
	// VisiOpenID 访问scheme的用户openid，为空表示未被访问过.
	VisitOpenID string `json:"visit_openid"`
}

func (c *WXMiniClient) QueryURLSchema(schema OpenLink) (URLSchemaDetail, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return URLSchemaDetail{}, err
	}
	uri := url_schema_query.Format(token)
	var resp struct {
		reply
		URLSchemaDetail
	}
	req := map[string]string{"scheme": string(schema)}
	err = c.httpPost(uri, req, &resp)
	if err != nil {
		return resp.URLSchemaDetail, err
	}
	if err = resp.Error(); err != nil {
		return resp.URLSchemaDetail, err
	}
	return resp.URLSchemaDetail, nil
}

type (
	URLLinkGenerateReq struct {
		// Path 通过 scheme 码进入的小程序页面路径，必须是已经发布的小程序存在的页面，不可携带 query。path 为空时会跳转小程序主页.
		Path string `json:"path,omitempty"`
		// Query 通过 scheme 码进入小程序时的 query，最大1024个字符，只支持数字，大小写英文以及部分特殊字符：`!#$&'()*+,/:;=?@-._~%`.
		Query string `json:"query,omitempty"`
		// 默认值"release"。要打开的小程序版本。正式版为"release"，体验版为"trial"，开发版为"develop"，仅在微信外打开时生效.
		Env string `json:"env_version,omitempty"`

		CloudBase struct {
			// Path 通过 scheme 码进入的小程序页面路径，必须是已经发布的小程序存在的页面，不可携带 query。path 为空时会跳转小程序主页.
			Path string `json:"path,omitempty"`
			// Query 通过 scheme 码进入小程序时的 query，最大1024个字符，只支持数字，大小写英文以及部分特殊字符：`!#$&'()*+,/:;=?@-._~%`.
			Query string `json:"query,omitempty"`
			// 默认值"release"。要打开的小程序版本。正式版为"release"，体验版为"trial"，开发版为"develop"，仅在微信外打开时生效.
			Env string `json:"env_version,omitempty"`
			// Domain 静态网站自定义域名，不填则使用默认域名.
			Domain string `json:"domain,omitempty"`
			// ResourceAppID 第三方批量代云开发时必填，表示创建该 env 的 appid （小程序/第三方平台）.
			ResourceAppID string `json:"resource_appid,omitempty"`
		} `json:"cloud_base,omitempty"`
	}

	URLLink string
)

// GenerateURLLInk 获取小程序 URL Link，适用于短信、邮件、网页、微信内等拉起小程序的业务场景。
// 通过该接口，可以选择生成到期失效和永久有效的小程序链接，有数量限制，目前仅针对国内非个人主体的小程序开放,
// 详情参考 https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/url-link.html.
func (c *WXMiniClient) GenerateURLLink(req URLLinkGenerateReq) (URLLink, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return URLLink(""), err
	}
	uri := url_link_generate.Format(token)
	var resp struct {
		reply
		URLLink URLLink `json:"url_link"`
	}
	err = c.httpPost(uri, req, &resp)
	if err != nil {
		return resp.URLLink, err
	}
	if err = resp.Error(); err != nil {
		return resp.URLLink, err
	}
	return resp.URLLink, nil
}

type URLLinkDetail struct {
	LinkInfo struct {
		AppID      string `json:"appid"`
		Path       string `json:"path"`
		Query      string `json:"query"`
		CreateTime int64  `json:"create_time"`
		ExpireTime int64  `json:"expire_time"`
		Env        string `json:"env_version"`
	} `json:"url_link_info"`
	LinkQuota struct {
		// Used 长期有效 url_link 已生成次数.
		Used int64 `json:"long_time_used"`
		// Limit 长期有效 url_link 生成次数上限.
		Limit int64 `json:"long_time_limit"`
	} `json:"url_link_quota"`
	// VisiOpenID 访问scheme的用户openid，为空表示未被访问过.
	VisitOpenID string `json:"visit_openid"`
}

func (c *WXMiniClient) QueryURLLink(link URLLink) (URLLinkDetail, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return URLLinkDetail{}, err
	}
	uri := url_link_query.Format(token)
	var resp struct {
		reply
		URLLinkDetail
	}
	req := map[string]string{"url_link": string(link)}
	err = c.httpPost(uri, req, &resp)
	if err != nil {
		return resp.URLLinkDetail, err
	}
	if err = resp.Error(); err != nil {
		return resp.URLLinkDetail, err
	}
	return resp.URLLinkDetail, nil
}

type ShortURLLinkGenerateReq struct {
	// PageUrl 通过 Short Link 进入的小程序页面路径，必须是已经发布的小程序存在的页面，可携带 query，最大1024个字符.
	PageUrl string `json:"page_url"`
	// PageTitle 页面标题，不能包含违法信息，超过20字符会用... 截断代替.
	PageTitle string `json:"page_title"`
	// Permanent 默认值false。生成的 Short Link 类型，短期有效：false，永久有效：true.
	Permanent bool `json:"is_permanent"`
}

// GenerateShortURLLink 获取小程序 Short Link，适用于微信内拉起小程序的业务场景。
// 目前只开放给电商类目(具体包含以下一级类目：电商平台、商家自营、跨境电商)。
// 通过该接口，可以选择生成到期失效和永久有效的小程序短链,
// 详情参考：https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/shortlink.html,
// ***调用上限***
// Link 将根据是否为到期有效与失效时间参数，分为**短期有效ShortLink ** 与 **永久有效ShortLink **：
// 单个小程序每日生成 ShortLink 上限为50万个（包含短期有效 ShortLink 与长期有效 ShortLink ）
// 单个小程序总共可生成永久有效 ShortLink 上限为10万个，请谨慎调用。
// 短期有效ShortLink 有效时间为30天，单个小程序生成短期有效ShortLink 不设上限.
func (c *WXMiniClient) GenerateShortURLLink(req ShortURLLinkGenerateReq) (URLLink, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return URLLink(""), err
	}
	uri := url_link_short.Format(token)
	var resp struct {
		reply
		URLLink URLLink `json:"url_link"`
	}
	err = c.httpPost(uri, req, &resp)
	if err != nil {
		return resp.URLLink, err
	}
	if err = resp.Error(); err != nil {
		return resp.URLLink, err
	}
	return resp.URLLink, nil
}

type NFCSchemaGenReq struct {
	URLSchemaGenReq
	ModelID string `json:"model_id"` // scheme对应的设备model_id
	SN      string `json:"sn"`       // scheme对应的设备sn，仅一机一码时填写
}

// GenerateNFCSchema 用于获取用于 NFC 的小程序 scheme 码，适用于 NFC 拉起小程序的业务场景。
// 目前仅针对国内非个人主体的小程序开放，详见 NFC 标签打开小程序。
func (c *WXMiniClient) GenerateNFCSchema(req NFCSchemaGenReq) (OpenLink, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return OpenLink(""), err
	}
	uri := url_schema_nfc.Format(token)
	var resp struct {
		reply
		OpenLink OpenLink `json:"openlink"`
	}
	err = c.httpPost(uri, req, &resp)
	if err != nil {
		return resp.OpenLink, err
	}
	if err = resp.Error(); err != nil {
		return resp.OpenLink, err
	}
	return resp.OpenLink, nil
}
