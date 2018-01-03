// 多媒体素材

package wxdev

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// StuffType 素材类型
type StuffType string

// 素材类型定义
const (
	StuffTypeImage StuffType = "image"
	StuffTypeVoice StuffType = "voice"
	StuffTypeVideo StuffType = "video"
	StuffTypeThumb StuffType = "thumb"
)

const url_uploadMedia = "https://api.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s"

// UploadTempStuffFile 上传文件至临时素材库
func (c *WXClient) UploadTempStuffFile(t StuffType, filename string) (string, error) {
	_, name := filepath.Split(filename)
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return c.UploadTempStuff(t, name, file)
}

// UploadTempStuff 上传临时素材
func (c *WXClient) UploadTempStuff(t StuffType, filename string, file io.Reader) (string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	defer w.Close()
	fw, err := w.CreateFormFile("media", filename)
	if err != nil {
		return "", err
	}
	_, err = io.CopyBuffer(fw, file, nil)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(&buf, "\r\n--%s--\r\n", w.Boundary())
	token, _ := c.getAccessToken()
	uri := fmt.Sprintf(url_uploadMedia, token, t)
	req, err := http.NewRequest("POST", uri, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	// reqData, _ := httputil.DumpRequest(req, true)
	// fmt.Println(string(reqData))
	client := http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}
	// respData, _ := httputil.DumpResponse(resp, true)
	// fmt.Println(string(respData))
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		MediaID string `json:"media_id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("%d-%s", result.ErrCode, result.ErrMsg)
	}
	return result.MediaID, nil
}

// MediaObject 多媒体对象
type MediaObject struct {
	FileName string
	Type     string
	Data     io.ReadCloser
	Size     int64
}

func (c *WXClient) downloadStuff(uri, mediaid string) (MediaObject, error) {
	var mediaObj MediaObject
	token, err := c.getAccessToken()
	if err != nil {
		return mediaObj, err
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf(uri, token, mediaid), nil)
	res, err := c.httpDo(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return mediaObj, err
	}
	var buf bytes.Buffer
	_, err = io.CopyBuffer(&buf, res.Body, nil)
	if err != nil {
		return mediaObj, err
	}
	mediaObj.Type = res.Header.Get("Content-Type")
	mediaObj.Data = ioutil.NopCloser(&buf)
	mediaObj.Size, _ = strconv.ParseInt(res.Header.Get("Content-Lengt"), 10, 64)
	disp := res.Header.Get("Content-disposition")
	if len(disp) > 0 {
		slice := strings.Split(disp, "filename=")
		if len(slice) > 1 {
			mediaObj.FileName = strings.Trim(slice[1], `"`)
		}
	}
	return mediaObj, nil
}

// DownloadMedia 下载多媒体文件
func (c *WXClient) DownloadMedia(mediaid string) (MediaObject, error) {
	const uri = "http://file.api.weixin.qq.com/cgi-bin/media/get?access_token=%s&media_id=%s"
	return c.downloadStuff(uri, mediaid)
}

// DownloadHQVoice 下载高清语音文件 格式为speex，16K采样率
func (c *WXClient) DownloadHQVoice(mediaid string) (MediaObject, error) {
	const uri = "https://api.weixin.qq.com/cgi-bin/media/get/jssdk?access_token=%s&media_id=%s"
	return c.downloadStuff(uri, mediaid)
}

// DownloadVideo 下载视频文件,返回视频链接
func (c *WXClient) DownloadVideo(mediaid string) (string, error) {
	const uri = "http://file.api.weixin.qq.com/cgi-bin/media/get?access_token=%s&media_id=%s"
	token, err := c.getAccessToken()
	if err != nil {
		return "", err
	}
	var result struct {
		ErrCode  int    `json:"errcode"`
		ErrMsg   string `json:"errmsg"`
		VideoURL string `json:"video_url"`
	}
	err = c.httpGet(fmt.Sprintf(uri, token, mediaid), &result)
	if err != nil {
		return result.VideoURL, err
	}
	if result.ErrCode != 0 {
		return result.VideoURL, fmt.Errorf("%d-%s", result.ErrCode, result.ErrMsg)
	}
	return result.VideoURL, nil
}
