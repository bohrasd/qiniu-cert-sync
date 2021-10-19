package qiniu

import (
	"encoding/json"
	"strings"
)

type Https struct {
	CertID      string `json:"certID"`
	ForceHttps  bool   `json:"forceHttps"`
	Http2Enable bool   `json:"http2Enable"`
}

type Domain struct {
	Name  string `json:"name"`
	Https `json:"https"`
}

func (m QiniuManager) ChangeDomainCert(commonName, certID string) (err error) {
	reqBody := Https{
		CertID:      certID,
		ForceHttps:  false,
		Http2Enable: true,
	}

	_, reqErr := doRequest(m.mac, "PUT", strings.Join([]string{"/domain", commonName, "httpsconf"}, "/"), reqBody)
	if reqErr != nil {
		err = reqErr
		return
	}
	return
}

func (m QiniuManager) GetDomain(commonName string) (domain Domain, err error) {
	resData, reqErr := doRequest(m.mac, "GET", strings.Join([]string{"/domain", commonName}, "/"), nil)
	if reqErr != nil {
		err = reqErr
		return
	}

	umErr := json.Unmarshal(resData, &domain)
	if umErr != nil {
		err = umErr
		return
	}
	return
}
