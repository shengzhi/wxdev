// 账号管理模块

package wxdev

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	// PermanentQRCode 永久二维码
	PermanentQRCode = 0
)

// CreateQRCode 创建二维码
func (c *WXClient) CreateQRCode(sceneValue interface{}, seconds int) (io.Reader, error) {
	const uri = "https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=%s"
	token, _ := c.getAccessToken()
	data := make(map[string]interface{})
	isNumber := func() bool {
		switch sceneValue.(type) {
		case int, int32, int64, int16, int8, uint8, uint16, uint32, uint:
			return true
		default:
			return false
		}
	}
	// 永久二维码
	if seconds <= 0 || seconds > 2592000 {
		if ok := isNumber(); ok {
			data["action_name"] = "QR_LIMIT_SCENE"
			data["action_info"] = map[string]interface{}{
				"scene": map[string]interface{}{"scene_id": sceneValue},
			}
		} else {
			data["action_name"] = "QR_LIMIT_STR_SCENE"
			data["action_info"] = map[string]interface{}{
				"scene": map[string]string{"scene_str": fmt.Sprintf("%s", sceneValue)},
			}
		}
	} else {
		data["expire_seconds"] = seconds
		if ok := isNumber(); ok {
			data["action_name"] = "QR_SCENE"
			data["action_info"] = map[string]interface{}{
				"scene": map[string]interface{}{"scene_id": sceneValue},
			}
		} else {
			data["action_name"] = "QR_STR_SCENE"
			data["action_info"] = map[string]interface{}{
				"scene": map[string]string{"scene_str": fmt.Sprintf("%s", sceneValue)},
			}
		}
	}
	var result struct {
		Ticket string `json:"ticket"`
		URL    string `json:"url"`
		Expire int    `json:"expire_seconds"`
	}
	err := c.httpPost(fmt.Sprintf(uri, token), data, &result)
	if err != nil {
		return nil, err
	}
	const qrcode_uri = "https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=%s"
	req, _ := http.NewRequest("GET", fmt.Sprintf(qrcode_uri, url.QueryEscape(result.Ticket)), nil)
	resp, err := c.httpDo(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		errmsg, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s", string(errmsg))
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	return &buf, err
}
