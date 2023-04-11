// 生物识别.
package miniapp

import (
	"encoding/json"
	"fmt"
)

type SoterResult struct {
	// Raw 调用者传入的challenge.
	Raw string `json:"raw"`
	// FId （仅Android支持）本次生物识别认证的生物信息编号（如指纹识别则是指纹信息在本设备内部编号）.
	FId string `json:"fid"`
	// Counter 防重放特征参数.
	Counter int64 `json:"counter"`
	// TeeName TEE名称（如高通或者trustonic等）.
	TeeName string `json:"tee_n"`
	// TeeVersion TEE版本号.
	TeeVersion string `json:"tee_v"`
	// FpN 指纹以及相关逻辑模块提供商（如FPC等）.
	FPN string `json:"fp_n"`
	// FPV 指纹以及相关模块版本号.
	FPV string `json:"fp_v"`
	// CPUId 机器唯一识别ID.
	CPUId string `json:"cpu_id"`
	// UId 概念同Android系统定义uid，即应用程序编号.
	UId string `json:"uid"`
}

// VerifySignature 用于SOTER 生物认证秘钥签名验证.
func (c *WXMiniClient) VerifySignature(openid, jsonstr, signature string) (SoterResult, error) {
	var result SoterResult
	token, err := c.getAccessToken()
	if err != nil {
		return result, err
	}
	uri := url_soter_verify.Format(token)
	req := map[string]string{
		"openid":         openid,
		"json_string":    jsonstr,
		"json_signature": signature,
	}
	var resp struct {
		reply
		OK bool `json:"is_ok"`
	}
	err = c.httpPost(uri, req, &resp)
	if err != nil {
		return result, err
	}
	if !resp.OK {
		return result, fmt.Errorf("SOTER verify signature failed, code:%d, message:%s", resp.ErrCode, resp.ErrMsg)
	}
	err = json.Unmarshal([]byte(jsonstr), &result)
	return result, err
}
