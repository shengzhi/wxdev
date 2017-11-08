package wxdev

import "fmt"

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

type WXBoolType byte

// ToBool 转换为bool类型
func (t WXBoolType) ToBool() bool {
	return t != 0
}

// WXUserInfo 微信公众号用户基本信息
type WXUserInfo struct {
	Subscribe     WXBoolType `json:"subscribe"`
	OpenID        string     `json:"openid"`
	NickName      string     `json:"nickname"`
	Sex           WXSexType  `json:"sex"`
	Language      string     `json:"language"`
	City          string     `json:"city"`
	Province      string     `json:"province"`
	Country       string     `json:"country"`
	HeadImgUrl    string     `json:"headimgurl"`
	SubscribeTime int64      `json:"subscribe_time"`
	UnionID       string     `json:"unionid"`
	Remark        string     `json:"remark"`
	GroupID       int        `json:"groupid"`
	TagIDList     []int      `json:"tagid_list"`
	ErrCode       int        `json:"errcode"`
	ErrMsg        string     `json:"errmsg"`
}

const user_baseinfo_url = "https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN"

// GetUserInfo 获取用户基本信息
func (c *WXClient) GetUserInfo(openid string) (user WXUserInfo, err error) {
	token, err := c.getAccessToken()
	if err != nil {
		return user, err
	}
	uri := fmt.Sprintf(user_baseinfo_url, token, openid)
	err = c.httpGet(uri, &user)
	if err != nil {
		return
	}
	if user.ErrCode != 0 {
		err = fmt.Errorf("接口返回错误: %s", user.ErrMsg)
	}
	return
}
