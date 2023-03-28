// 生成小程序码

package miniapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// CodeGenArg 小程序码生成参数
type CodeGenArg struct {
	Sence     string `json:"scene"`
	Path      string `json:"page,omitempty"`
	CheckPath bool   `json:"check_path,omitempty"`
	// Env 要打开的小程序版本。正式版为 "release"，体验版为 "trial"，开发版为 "develop"。默认是正式版.
	Env       string `json:"env_version,omitempty"`
	Width     int    `json:"width,omitempty"`
	AutoColor bool   `json:"auto_color,omitempty"`
	LineColor struct {
		R string `json:"r"`
		G string `json:"g"`
		B string `json:"b"`
	} `json:"line_color,omitempty"`
	// IsHyaline 默认是false，是否需要透明底色，为 true 时，生成透明底色的小程序.
	IsHyaline bool `json:"is_hyaline,omitempty"`
}

// WXACode_A 适用于需要的码数量较少的业务场景
// 通过该接口生成的小程序码，永久有效，数量限制见文末说明，请谨慎使用。
// 用户扫描该码进入小程序后，将直接进入 path 对应的页面
func (c *WXMiniClient) WXACode_A(arg CodeGenArg) (io.Reader, error) {
	const uri = "https://api.weixin.qq.com/wxa/getwxacode?access_token=%s"
	return c.genWXACode(uri, arg)
}

// WXACode_B 适用于需要的码数量极多，或仅临时使用的业务场景
// 通过该接口生成的小程序码，永久有效，数量暂无限制。用户扫描该码进入小程序后，
// 开发者需在对应页面获取的码中 scene 字段的值，再做处理逻辑。
// 使用如下代码可以获取到二维码中的 scene 字段的值。
// 调试阶段可以使用开发工具的条件编译自定义参数 scene=xxxx 进行模拟，
// 开发工具模拟时的 scene 的参数值需要进行 urlencode
func (c *WXMiniClient) WXACode_B(arg CodeGenArg) (io.Reader, error) {
	const uri = "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=%s"
	if arg.Sence == "" {
		return nil, fmt.Errorf("Sence is mandatory")
	}
	arg.Sence = url.QueryEscape(arg.Sence)
	return c.genWXACode(uri, arg)
}

// WXACode_C 适用于需要的码数量较少的业务场景
// 通过该接口生成的小程序二维码，永久有效，数量限制见文末说明，请谨慎使用。
// 用户扫描该码进入小程序后，将直接进入 path 对应的页面
func (c *WXMiniClient) WXACode_C(arg CodeGenArg) (io.Reader, error) {
	const uri = "https://api.weixin.qq.com/cgi-bin/wxaapp/createwxaqrcode?access_token=%s"
	return c.genWXACode(uri, arg)
}

func (c *WXMiniClient) genWXACode(uri string, data interface{}) (io.Reader, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(data)
	res, err := http.Post(fmt.Sprintf(uri, token), "application/json", &buf)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	buf.Reset()
	io.Copy(&buf, res.Body)
	return &buf, nil
}
