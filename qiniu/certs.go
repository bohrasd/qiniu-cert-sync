package qiniu

import (
	"encoding/json"
	"strings"
)

type CertReq struct {
	Name       string
	CommonName string
	Pri        string
	Ca         string
}

type CertUpResp struct {
	CertID string `json:"certID"`
}

type Cert struct {
	CertID     string   `json:"certid"`
	Name       string   `json:"name"`
	CommonName string   `json:"common_name"`
	DNSNames   []string `json:"dnsnames"`
	NotBefore  int      `json:"not_before"`
	NotAfter   int      `json:"not_after"`
	CreateTime int      `json:"create_time"`
}

type CertListResp struct {
	Marker string `json:"certID"`
	Certs  []Cert
}

func (m QiniuManager) DelCert(certID string) (err error) {
	_, reqErr := doRequest(m.mac, "DELETE", strings.Join([]string{"/sslcert", certID}, "/"), nil)
	if reqErr != nil {
		err = reqErr
		return
	}
	return
}

func (m QiniuManager) GetCert(certID string) (certResp Cert, err error) {
	resData, reqErr := doRequest(m.mac, "GET", strings.Join([]string{"/sslcert", certID}, "/"), nil)
	if reqErr != nil {
		err = reqErr
		return
	}
	umErr := json.Unmarshal(resData, &certResp)
	if umErr != nil {
		err = umErr
		return
	}
	return
}

func (m QiniuManager) CertList() (certListResp CertListResp, err error) {
	resData, reqErr := doRequest(m.mac, "GET", "/sslcert", nil)
	if reqErr != nil {
		err = reqErr
		return
	}
	umErr := json.Unmarshal(resData, &certListResp)
	if umErr != nil {
		err = umErr
		return
	}
	return
}

func (m QiniuManager) UploadCert(name, common_name, ca, pri string) (certUpResp CertUpResp,
	err error) {
	reqBody := CertReq{
		Name:       name,
		CommonName: common_name,
		Pri:        pri,
		Ca:         ca,
	}

	resData, reqErr := doRequest(m.mac, "POST", "/sslcert", reqBody)
	if reqErr != nil {
		err = reqErr
		return
	}
	umErr := json.Unmarshal(resData, &certUpResp)
	if umErr != nil {
		err = umErr
		return
	}
	return
}
