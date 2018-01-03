// 客服模块

package wxdev

import (
	"fmt"
)

// CSMsgReply 客服消息
type CSMsgReply struct {
	fields map[string]interface{}
}

// NewCSMsgReply 创建客服消息
func NewCSMsgReply(to string) CSMsgReply {
	reply := CSMsgReply{
		fields: make(map[string]interface{}),
	}
	reply.fields["touser"] = to
	return reply
}
func (reply CSMsgReply) data() interface{} {
	return reply.fields
}

// WithText 文本消息
func (reply CSMsgReply) WithText(content string) {
	reply.fields["msgtype"] = "text"
	reply.fields["text"] = struct {
		Content string `json:"content"`
	}{content}
}

// WithImage 图片消息
func (reply CSMsgReply) WithImage(mediaid string) {
	reply.fields["msgtype"] = "image"
	reply.fields["image"] = struct {
		MediaID string `json:"media_id"`
	}{mediaid}
}

// WithVoice 语音消息
func (reply CSMsgReply) WithVoice(mediaid string) {
	reply.fields["msgtype"] = "voice"
	reply.fields["voice"] = struct {
		MediaID string `json:"media_id"`
	}{mediaid}
}

// WithVideo 视频消息
func (reply CSMsgReply) WithVideo(mediaid, thumbMediaid, title, desc string) {
	reply.fields["msgtype"] = "video"
	reply.fields["video"] = struct {
		MediaID string `json:"media_id"`
		Thumb   string `json:"thumb_media_id"`
		Title   string `json:"title"`
		Desc    string `json:"description"`
	}{mediaid, thumbMediaid, title, desc}
}

// SendCSMsg 发送客服消息
func (c *WXClient) SendCSMsg(msg CSMsgReply) error {
	const uri = "https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=%s"
	token, _ := c.getAccessToken()
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	err := c.httpPost(fmt.Sprintf(uri, token), msg.data(), &result)
	if err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("%d-%s", result.ErrCode, result.ErrMsg)
	}
	return nil
}
