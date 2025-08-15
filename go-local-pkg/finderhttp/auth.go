package finderhttp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"
)

// HmacAuth
// 构建鉴权headers
// @requestUrl: like http://api.xfyun.cn
// @method: GET. POST. etc....
// @body : 请求body
func NewAuthHeaders(requestUrl, method string, apiKey, apiSecret string, body []byte) map[string]string {
	bodySign := ""
	if body == nil {
		bodySign = sha256Base64([]byte(nil))
	} else {
		bodySign = sha256Base64(body)
	}

	bodySign = "SHA256=" + bodySign

	u, err := url.Parse(requestUrl)
	if err != nil {
		panic("parse url error" + err.Error())
	}
	host := u.Host
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 MST")
	requestLine := method + " " + u.Path + " HTTP/1.1"
	signUrl := fmt.Sprintf("host: %s\ndate: %s\n%s\ndigest: %s", host, date, requestLine, bodySign)
	signature := hmacSha256Base64([]byte(apiSecret), []byte(signUrl))

	authorization := fmt.Sprintf(`api_key="%s", algorithm="hmac-sha256", headers="host date request-line digest", signature="%s"`, apiKey, signature)
	return map[string]string{
		"host":          host,
		"date":          date,
		"authorization": authorization,
		"digest":        bodySign,
	}
}

func sha256Base64(b []byte) string {
	h := sha256.New()
	h.Write(b)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func hmacSha256Base64(secret []byte, data []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
