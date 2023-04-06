package wxdev

import (
	"fmt"

	"github.com/shengzhi/util/dtime"
)

type WXSexType byte

const (
	WXSexTypeMale   WXSexType = 1
	WXSexTypeFemale WXSexType = 2
)

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

type WXBoolType byte

// ToBool 转换为bool类型
func (t WXBoolType) ToBool() bool {
	return t != 0
}

// WXUserInfo 微信公众号用户基本信息
type WXUserInfo struct {
	Subscribe     WXBoolType     `json:"subscribe"`
	OpenID        string         `json:"openid"`
	NickName      string         `json:"nickname"`
	Sex           WXSexType      `json:"sex"`
	Language      string         `json:"language"`
	City          string         `json:"city"`
	Province      string         `json:"province"`
	Country       string         `json:"country"`
	HeadImgUrl    string         `json:"headimgurl"`
	SubscribeTime dtime.JSONTime `json:"subscribe_time"`
	UnionID       string         `json:"unionid"`
	Remark        string         `json:"remark"`
	GroupID       int            `json:"groupid"`
	TagIDList     []int          `json:"tagid_list"`
	ErrCode       int            `json:"errcode"`
	ErrMsg        string         `json:"errmsg"`
}

// GetUserInfo 获取用户基本信息
func (c *WXClient) GetUserInfo(openid string) (user WXUserInfo, err error) {
	const uri = "https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN"
	token, err := c.getAccessToken()
	if err != nil {
		return user, err
	}
	err = c.httpGet(fmt.Sprintf(uri, token, openid), &user)
	if err != nil {
		return
	}
	if user.ErrCode != 0 {
		err = fmt.Errorf("%d-%s", user.ErrCode, user.ErrMsg)
	}
	return
}

// BatchGetUserInfo 批量获取用户信息
func (c *WXClient) BatchGetUserInfo(openids ...string) ([]WXUserInfo, error) {
	if len(openids) <= 0 || len(openids) > 100 {
		return nil, fmt.Errorf("Cannot be more than 100 records one time")
	}
	const uri = "https://api.weixin.qq.com/cgi-bin/user/info/batchget?access_token=%s"
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}
	type openidInfo struct {
		OpenID string `json:"openid"`
		Lang   string `json:"lang,omitempty"`
	}
	var data struct {
		Users []openidInfo `json:"user_list"`
	}
	for _, openid := range openids {
		data.Users = append(data.Users, openidInfo{OpenID: openid})
	}
	var result struct {
		ErrCode int          `json:"errcode"`
		ErrMsg  string       `json:"errmsg"`
		Users   []WXUserInfo `json:"user_info_list"`
	}
	err = c.httpPost(fmt.Sprintf(uri, token), data, &result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("%d-%s", result.ErrCode, result.ErrMsg)
	}
	return result.Users, nil
}

type LoginAccessToken struct {
	ErrCode        int    `json:"errcode"`
	ErrMsg         string `json:"errmsg"`
	AccessToken    string `json:"access_token"`
	Expires        int    `json:"expires_in"`
	RefreshToken   string `json:"refresh_token"`
	OpenID         string `json:"openid"`
	Scope          string `json:"scope"`
	UnionID        string `json:"unionid"`
	IsSnapshotUser int    `json:"is_snapshotuser"`
}

func (c *WXClient) GetLoginAccessToken(code string) (LoginAccessToken, error) {
	const uri = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	var token LoginAccessToken
	err := c.httpGet(fmt.Sprintf(uri, c.appid, c.appsecret, code), &token)
	if err != nil {
		return token, err
	}
	if token.ErrCode != 0 {
		return token, fmt.Errorf("code:%d,message:%s", token.ErrCode, token.ErrMsg)
	}
	return token, nil
}
