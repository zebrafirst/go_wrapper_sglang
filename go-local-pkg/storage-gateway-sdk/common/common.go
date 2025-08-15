package common

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"time"

	finderHttp "git.iflytek.com/AIaaS/finderhttp-self"
	sgErr "git.iflytek.com/AIaaS/storage-gateway-sdk-self/sg_err"
)

// DealConn 不支持处理大文本(response) 如：文件下载
func DealConn(cli finderHttp.Client, ctx context.Context, r *finderHttp.Request) (*finderHttp.Response, error) {
	resp, err := cli.Request(ctx, r)
	if err != nil {
		return nil, sgErr.NewError(sgErr.GET_RESP_ERR_CODE, err)
	}

	return resp, nil
}

// NewAuthHeaders 构建鉴权 headers
// @requestUrl: like http://api.xfyun.cn/api/v1/cdss
// @method: GET. POST. etc....
// @body : 请求body
func NewAuthHeaders(requestUrl, method, apiKey, apiSecret string, body []byte) map[string]string {
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

func Struct2Map(s interface{}, tagName string) url.Values {
	m := make(map[string][]string)

	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	typ := v.Type()

	for i := 0; i < typ.NumField(); i++ {
		fieldType := typ.Field(i)
		vv := v.Field(i)
		tag := fieldType.Tag.Get(tagName)
		if tag == "" {
			tag = fieldType.Name
		}

		if vv.Kind() == reflect.Ptr {
			vv = vv.Elem()
		}

		values := make([]string, 0)
		switch vv.Kind() {
		case reflect.Struct, reflect.Interface:
			continue
		case reflect.Slice:
			for j := 0; j < vv.Len(); j++ {
				if vvv := vv.Index(j); vvv.IsValid() {
					if sv := fmt.Sprintf("%v", vvv); len(sv) != 0 {
						values = append(values, sv)
					}
				}
			}
		case reflect.String:
			if vv.IsValid() && vv.Len() != 0 {
				values = append(values, vv.String())
			}
		default:
			if vv.IsValid() {
				if sv := fmt.Sprintf("%v", vv); len(sv) != 0 {
					values = append(values, sv)
				}
			}
		}

		if len(values) != 0 {
			m[tag] = values
		}
	}

	return m
}

func JudgeKeyOrName(keyName, location string) (string, error) {
	if len(keyName) == 0 {
		return "", sgErr.NewError(sgErr.PARAMETER_ILLEGAL_ERR_CODE, errors.New("keyName length is 0"))
	}
	if len(location) == 0 {
		return keyName, nil
	}
	return keyName + "?x_location=" + location, nil
}

type FSlice struct {
	SliceId int
	File    io.ReadSeeker
	Length  int
}

func SliceFile(file io.Reader, size int) ([]FSlice, error) {
	chunks := make([]FSlice, 0)
	order := 0
	// 开辟缓存区
	buf := make([]byte, size)

	for {
		// io.ReadFull 将 file 写入到指定大小的缓存区
		// 需要考虑最后一次写入时 buf 未满的情况
		n, err := io.ReadFull(file, buf)
		if err == nil || err == io.EOF {
			order += 1
			// 将读取到的数据转换为 io.ReadSeeker 并添加到切片中
			if n > 0 {
				chunks = append(chunks, FSlice{
					SliceId: order,
					File:    bytes.NewReader(buf[:n]),
					Length:  n,
				})
			}

			if err == io.EOF {
				break
			}
		} else if !errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, err
		} else { // 最后一个切片，n < size
			order += 1
			chunks = append(chunks, FSlice{
				SliceId: order,
				File:    bytes.NewReader(buf[:n]),
				Length:  n,
			})
			break
		}
	}
	return chunks, nil
}
