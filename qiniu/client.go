package qiniu

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
)

type QiniuManager struct {
	mac *auth.Credentials
}

// 七牛域名管理服务域名
var (
	QiniuHost = "http://api.qiniu.com"
)

type ErrorInfo struct {
	Err  string `json:"error,omitempty"`
	Code int    `json:"code,omitempty"`
}

func NewQiniuManager(accessKey, secretKey string) QiniuManager {
	return QiniuManager{qbox.NewMac(accessKey, secretKey)}
}

func doRequest(mac *auth.Credentials, method, path string, body interface{}) (resData []byte,
	err error) {
	urlStr := fmt.Sprintf("%s%s", QiniuHost, path)
	reqData, _ := json.Marshal(body)
	req, reqErr := http.NewRequest(method, urlStr, bytes.NewReader(reqData))
	if reqErr != nil {
		err = reqErr
		return
	}

	accessToken, signErr := mac.SignRequest(req)
	if signErr != nil {
		err = signErr
		return
	}

	req.Header.Add("Authorization", "QBox "+accessToken)
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, respErr := http.DefaultClient.Do(req)
	if respErr != nil {
		err = respErr
		return
	}
	defer resp.Body.Close()

	resData, ioErr := ioutil.ReadAll(resp.Body)
	if ioErr != nil {
		err = ioErr
		return
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		var errInfo ErrorInfo
		json.Unmarshal(resData, &errInfo)
		err = errors.New(errInfo.Err)
		return
	}

	return
}
