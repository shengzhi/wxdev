// 公众号自定义菜单

package wxdev

import (
	"fmt"
)

// Button 菜单
type Button map[string]interface{}

// AddSubMenu 增加子菜单
func (b Button) AddSubMenu(menu Button) Button {
	if v, ok := b["sub_button"]; ok {
		b["sub_button"] = append(v.([]Button), menu)
	} else {
		b["sub_button"] = []Button{menu}
	}
	return b
}

// NewParentMenu creates parent menu
func NewParentMenu(name string) Button {
	return map[string]interface{}{"name": name}
}

// NewButtonClick creates click button
func NewButtonClick(name, key string) Button {
	return map[string]interface{}{
		"type": "click",
		"name": name,
		"key":  key,
	}
}

// NewButtonView creates view button
func NewButtonView(name, url string) Button {
	return map[string]interface{}{
		"type": "view",
		"name": name,
		"url":  url,
	}
}

// NewButtonScanPush 扫码推事件
func NewButtonScanPush(name, key string) Button {
	return map[string]interface{}{
		"type": "scancode_push",
		"name": name,
		"key":  key,
	}
}

// NewButtonScanWait 扫码带提示事件
func NewButtonScanWait(name, key string) Button {
	return map[string]interface{}{
		"type": "scancode_waitmsg",
		"name": name,
		"key":  key,
	}
}

// NewButtonPhoto 拍照按钮
func NewButtonPhoto(name, key string) Button {
	return map[string]interface{}{
		"type": "pic_sysphoto",
		"name": name,
		"key":  key,
	}
}

// NewButtonPhotoOrAlbum 拍照或相册按钮
func NewButtonPhotoOrAlbum(name, key string) Button {
	return map[string]interface{}{
		"type": "pic_photo_or_album",
		"name": name,
		"key":  key,
	}
}

// NewButtonWXPic 微信相册发图
func NewButtonWXPic(name, key string) Button {
	return map[string]interface{}{
		"type": "pic_weixin",
		"name": name,
		"key":  key,
	}
}

// NewButtonLocation 发送位置
func NewButtonLocation(name, key string) Button {
	return map[string]interface{}{
		"type": "location_select",
		"name": name,
		"key":  key,
	}
}

// NewButtonMedia 图片/音频/视频素材
func NewButtonMedia(name, mediaid string) Button {
	return map[string]interface{}{
		"type":     "media_id",
		"name":     name,
		"media_id": mediaid,
	}
}

// NewButtonArticle 图文消息
func NewButtonArticle(name, mediaid string) Button {
	return map[string]interface{}{
		"type":     "view_limited",
		"name":     name,
		"media_id": mediaid,
	}
}

// NewButtonMiniApp 打开小程序按钮
// appid: 小程序的appid（仅认证公众号可配置）
// pagepath: 小程序的页面路径
// 参数uri: 网页 链接，用户点击菜单可打开链接，不超过1024字节。不支持小程序的老版本客户端将打开本url。
func NewButtonMiniApp(name, appid, pagepath, uri string) Button {
	return map[string]interface{}{
		"type": "miniprogram",
		"name": name, "url": uri,
		"appid": appid, "pagepath": pagepath,
	}
}

// CreateMenu 创建菜单
func (c *WXClient) CreateMenu(items ...Button) error {
	const uri = "https://api.weixin.qq.com/cgi-bin/menu/create?access_token=%s"
	token, err := c.getAccessToken()
	if err != nil {
		return err
	}
	var data = struct {
		Buttons []Button `json:"button"`
	}{items}
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	err = c.httpPost(fmt.Sprintf(uri, token), data, &result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("%d-%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// ClearMenu 清空菜单
func (c *WXClient) ClearMenu() error {
	const uri = "https://api.weixin.qq.com/cgi-bin/menu/delete?access_token=%s"
	token, err := c.getAccessToken()
	if err != nil {
		return err
	}
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	err = c.httpGet(fmt.Sprintf(uri, token), &result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("%d-%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// MatchRule 个性化菜单匹配规则
type MatchRule map[string]string

// MatchTag 标签匹配
func (m MatchRule) MatchTag(tagid string) MatchRule {
	m["tag_id"] = tagid
	return m
}

// MatchSex 性别匹配
func (m MatchRule) MatchSex(sex WXSexType) MatchRule {
	m["sex"] = fmt.Sprintf("%d", sex)
	return m
}

// MatchCountry 国家匹配
func (m MatchRule) MatchCountry(country string) MatchRule {
	m["country"] = country
	return m
}

// MatchProvince 匹配省份
func (m MatchRule) MatchProvince(province string) MatchRule {
	m["province"] = province
	return m
}

// MatchCity 匹配城市
func (m MatchRule) MatchCity(city string) MatchRule {
	m["city"] = city
	return m
}

// MatchLanguage 匹配语言
func (m MatchRule) MatchLanguage(language string) MatchRule {
	m["language"] = language
	return m
}

func (m MatchRule) validate() error {
	if len(m) <= 0 {
		return fmt.Errorf("Match rule cannot be empty")
	}
	if _, hasCity := m["city"]; hasCity {
		if _, hasProvince := m["province"]; !hasProvince {
			return fmt.Errorf("Must specify province")
		}
	}
	if _, has := m["province"]; has {
		if _, hasCountry := m["country"]; !hasCountry {
			return fmt.Errorf("Must specify country")
		}
	}
	return nil
}

// CreatePersonalMenu 创建个性化菜单
func (c *WXClient) CreatePersonalMenu(rule MatchRule, items ...Button) (string, error) {
	if err := rule.validate(); err != nil {
		return "", err
	}
	token, _ := c.getAccessToken()
	const uri = "https://api.weixin.qq.com/cgi-bin/menu/addconditional?access_token=%s"
	var result struct {
		MenuID string `json:"menuid"`
	}
	var data = struct {
		Buttons []Button  `json:"button"`
		Rule    MatchRule `json:"matchrule"`
	}{items, rule}
	err := c.httpPost(fmt.Sprintf(uri, token), data, &result)
	return result.MenuID, err
}

// DeletePersonalMenu 删除个性化菜单
func (c *WXClient) DeletePersonalMenu(menuid string) error {
	const uri = "https://api.weixin.qq.com/cgi-bin/menu/delconditional?access_token=%s"
	token, _ := c.getAccessToken()
	var data = struct {
		MenuID string `json:"menuid"`
	}{menuid}
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	err := c.httpPost(fmt.Sprintf(uri, token), data, &result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("%d-%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// TestPersonalMenu 测试个性化菜单
// wxuserid 可以是粉丝的OpenID，也可以是粉丝的微信号
func (c *WXClient) TestPersonalMenu(wxuserid string) ([]Button, error) {
	const uri = "https://api.weixin.qq.com/cgi-bin/menu/trymatch?access_token=%s"
	token, _ := c.getAccessToken()
	var data = struct {
		UserID string `json:"user_id"`
	}{wxuserid}
	var result struct {
		ErrCode int      `json:"errcode"`
		ErrMsg  string   `json:"errmsg"`
		Buttons []Button `json:"button"`
	}
	err := c.httpPost(fmt.Sprintf(uri, token), data, &result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("%d-%s", result.ErrCode, result.ErrMsg)
	}
	return result.Buttons, nil
}
