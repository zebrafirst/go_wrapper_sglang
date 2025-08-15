package api_v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	finderHttp "git.iflytek.com/AIaaS/finderhttp-self"
	"git.iflytek.com/AIaaS/storage-gateway-sdk-self/common"
	sgErr "git.iflytek.com/AIaaS/storage-gateway-sdk-self/sg_err"
)

// UploadQuery 上传文件时的可选项
type UploadQuery struct {
	GetLink   bool   `label:"get_link"`   // 是否返回外链
	LinkTtl   int    `label:"link_ttl"`   // 外链的访问超时时间，单位：秒
	SplitHost bool   `label:"split_host"` // 是否只返回文件 url 的 path 部分
	Filename  string `label:"filename"`   // 文件名，不传时使用文件的 md5 值。文件名相同时覆盖文件
	Expose    bool   `label:"expose"`     // 值为 true 时，外链中的文件名为原始文件名，并附加 x_location 参数；否则使用 fileKey 作为文件名
	Type      string `label:"type"`       // 存储引擎选择，枚举类型：[s3, hbase]。不传时根据文件大小自动决定
}

type UploadReq struct {
	Namespace string      // 业务独立命名空间
	File      io.Reader   // 文件内容
	Length    int         // 文件内容长度，单位：字节(Byte)。如果文件大小超过 50 MB，请使用分片上传功能
	XTtl      int         // 设置文件存储的有效期限，单位为秒（second）
	Query     UploadQuery // 默认不提供外链下载，需要提供外链下载时使用此参数
}

type fileData struct {
	FileKey  string `json:"key"`       // 文件唯一标识符
	FileMD5  string `json:"md5"`       // 文件 md5 值
	Location string `json:"location"`  // 文件额外信息。文件名无法唯一确定文件，但关联此信息时就能定位到文件
	FileLink string `json:"link"`      // 文件外链
	LinkPath string `json:"link_path"` // 文件外链路由，不含 host 部分
}

type uploadResult struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Sid     string   `json:"sid"`
	Data    fileData `json:"data"`
}

type UploadRes struct {
	Sid      string
	FileKey  string
	FileMD5  string
	Link     string
	Location string
}

type DownloadReq struct {
	Namespace string
	KeyName   string // 文件唯一标识符或者文件名
	Location  string `label:"x_location"` // Location 为空字符串时 KeyName 视为 FileKey；否则视为 Filename
}

type downloadResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Sid     string `json:"sid"`
}

type DownloadRes struct {
	File io.ReadCloser
}

type UpdateReq struct {
	Namespace string
	KeyName   string    // 文件唯一标识符或者文件名
	Location  string    `label:"x_location"` // Location 为空字符串时 KeyName 视为 FileKey；否则视为 Filename
	File      io.Reader // 文件内容
	Length    int       // 文件内容字节数
}

type md5Data struct {
	MD5 string `json:"md5"`
}

type updateResult struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Sid     string  `json:"sid"`
	Data    md5Data `json:"data"`
}

type UpdateRes struct {
	Sid string
	MD5 string
}

type DeleteReq struct {
	Namespace string
	KeyName   string // 文件唯一标识符或者文件名
	Location  string `label:"x_location"` // Location 为空字符串时 KeyName 视为 FileKey；否则视为 Filename
}

type deleteResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Sid     string `json:"sid"`
}

type DeleteRes struct {
	Sid string `json:"sid"`
}

type GetLinkQuery struct {
	LinkTtl   int  `label:"link_ttl"`
	SplitHost bool `label:"split_host"`
}

type GetLinkReq struct {
	Namespace string
	FileKey   string
	Query     GetLinkQuery
}

type getLinkData struct {
	Link     string `json:"link"`
	LinkPath string `json:"link_path"`
}

type getLinkResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Sid     string      `json:"sid"`
	Data    getLinkData `json:"data"`
}

type GetLinkRes struct {
	Sid  string
	Link string
}

func (c *apiV1Cli) FileUpload(ctx context.Context, req UploadReq) (UploadRes, error) {
	res := UploadRes{}

	if req.Query.GetLink == true && req.Query.LinkTtl <= 0 {
		return res, sgErr.NewError(sgErr.PARAMETER_ILLEGAL_ERR_CODE, errors.New("LinkTtl must be a positive integer"))
	}

	if req.Length > LenLimit {
		return res, sgErr.NewError(sgErr.FILE_LENGTH_OVER_CODE, nil)
	}

	r := &finderHttp.Request{
		Path: fmt.Sprintf("%s/%s?%s", V1PATH, req.Namespace,
			common.Struct2Map(req.Query, "label").Encode()),
		Method:        http.MethodPost,
		Body:          req.File,
		ContentLength: req.Length,
		Handlers: []finderHttp.Handler{
			func(r *http.Request, url string) error {
				r.Header.Set("X-TTL", strconv.Itoa(req.XTtl))
				headers := common.NewAuthHeaders(url, http.MethodPost, c.auth.ApiKey, c.auth.ApiSecret, nil)
				for k, v := range headers {
					r.Header.Set(k, v)
				}
				return nil
			},
		},
	}

	resp, err := common.DealConn(c.cli, ctx, r)
	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, sgErr.NewError(sgErr.IO_READ_ERR_CODE, err)
	}

	midRes := uploadResult{}
	err = json.Unmarshal(msg, &midRes)
	if err != nil {
		return res, sgErr.NewError(sgErr.JSON_UNMARSHAL_ERR_CODE, err)
	}

	if resp.StatusCode != 200 {
		return res, sgErr.WarpHttpErr(resp.StatusCode, midRes.Code, midRes.Sid, midRes.Message)
	}

	if midRes.Code != 0 {
		return res, sgErr.WarpMsgErr(midRes.Code, midRes.Sid, midRes.Message)
	}

	res = UploadRes{
		Sid:      midRes.Sid,
		FileKey:  midRes.Data.FileKey,
		FileMD5:  midRes.Data.FileMD5,
		Location: midRes.Data.Location,
	}

	if req.Query.GetLink {
		if req.Query.SplitHost {
			res.Link = midRes.Data.LinkPath
		} else {
			res.Link = midRes.Data.FileLink
		}
	}

	return res, nil
}

// FileDownload FileKey 与 (Filename, Location) 只需要其中之一即可
func (c *apiV1Cli) FileDownload(ctx context.Context, req DownloadReq) (DownloadRes, error) {
	res := DownloadRes{}
	subPath, err := common.JudgeKeyOrName(req.KeyName, req.Location)
	if err != nil {
		return res, err
	}

	r := &finderHttp.Request{
		Path:   fmt.Sprintf("%s/%s/%s", V1PATH, req.Namespace, subPath),
		Method: http.MethodGet,
		Handlers: []finderHttp.Handler{
			func(r *http.Request, url string) error {
				headers := common.NewAuthHeaders(url, http.MethodGet, c.auth.ApiKey, c.auth.ApiSecret, nil)
				for k, v := range headers {
					r.Header.Set(k, v)
				}
				return nil
			},
		},
	}

	resp, err := c.cli.Request(ctx, r)
	if err != nil {
		return res, sgErr.NewError(sgErr.GET_RESP_ERR_CODE, err)
	}

	// 只有下载需要考虑大文本传输问题
	// http 状态码不为 200 时才反序列化json；
	// 否则不处理，直接返回response.Body
	if resp.StatusCode != 200 {
		defer resp.Body.Close()

		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return res, sgErr.NewError(sgErr.IO_READ_ERR_CODE, err)
		}

		midRes := &downloadResult{}
		err = json.Unmarshal(msg, &midRes)
		if err != nil {
			return res, sgErr.NewError(sgErr.JSON_UNMARSHAL_ERR_CODE, err)
		}

		return res, sgErr.WarpHttpErr(resp.StatusCode, midRes.Code, midRes.Sid, midRes.Message)
	}

	res.File = resp.Body

	return res, nil
}

func (c *apiV1Cli) FileUpdate(ctx context.Context, req UpdateReq) (UpdateRes, error) {
	res := UpdateRes{}

	subPath, err := common.JudgeKeyOrName(req.KeyName, req.Location)
	if err != nil {
		return res, err
	}

	r := &finderHttp.Request{
		Path:          fmt.Sprintf("%s/%s/%s", V1PATH, req.Namespace, subPath),
		Method:        http.MethodPatch,
		Body:          req.File,
		ContentLength: req.Length,
		Handlers: []finderHttp.Handler{
			func(r *http.Request, url string) error {
				headers := common.NewAuthHeaders(url, http.MethodPatch, c.auth.ApiKey, c.auth.ApiSecret, nil)
				for k, v := range headers {
					r.Header.Set(k, v)
				}
				return nil
			},
		},
	}

	resp, err := common.DealConn(c.cli, ctx, r)
	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, sgErr.NewError(sgErr.IO_READ_ERR_CODE, err)
	}

	midRes := updateResult{}
	err = json.Unmarshal(msg, &midRes)
	if err != nil {
		return res, sgErr.NewError(sgErr.JSON_UNMARSHAL_ERR_CODE, err)
	}

	if resp.StatusCode != 200 {
		return res, sgErr.WarpHttpErr(resp.StatusCode, midRes.Code, midRes.Sid, midRes.Message)
	}

	if midRes.Code != 0 {
		return res, sgErr.WarpMsgErr(midRes.Code, midRes.Sid, midRes.Message)
	}

	res = UpdateRes{
		Sid: midRes.Sid,
		MD5: midRes.Data.MD5,
	}

	return res, nil
}

func (c *apiV1Cli) FileDelete(ctx context.Context, req DeleteReq) (DeleteRes, error) {
	res := DeleteRes{}

	subPath, err := common.JudgeKeyOrName(req.KeyName, req.Location)
	if err != nil {
		return res, err
	}

	r := &finderHttp.Request{
		Path:   fmt.Sprintf("%s/%s/%s", V1PATH, req.Namespace, subPath),
		Method: http.MethodDelete,
		Handlers: []finderHttp.Handler{
			func(r *http.Request, url string) error {
				headers := common.NewAuthHeaders(url, http.MethodDelete, c.auth.ApiKey, c.auth.ApiSecret, nil)
				for k, v := range headers {
					r.Header.Set(k, v)
				}
				return nil
			},
		},
	}

	resp, err := common.DealConn(c.cli, ctx, r)
	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, sgErr.NewError(sgErr.IO_READ_ERR_CODE, err)
	}

	midRes := deleteResult{}
	err = json.Unmarshal(msg, &midRes)
	if err != nil {
		return res, sgErr.NewError(sgErr.JSON_UNMARSHAL_ERR_CODE, err)
	}

	if resp.StatusCode != 200 {
		return res, sgErr.WarpHttpErr(resp.StatusCode, midRes.Code, midRes.Sid, midRes.Message)
	}

	if midRes.Code != 0 {
		return res, sgErr.WarpMsgErr(midRes.Code, midRes.Sid, midRes.Message)
	}

	res = DeleteRes{
		Sid: midRes.Sid,
	}
	return res, nil
}

func (c *apiV1Cli) GetLink(ctx context.Context, req GetLinkReq) (GetLinkRes, error) {
	res := GetLinkRes{}

	r := &finderHttp.Request{
		Path: fmt.Sprintf("%s/%s/%s/get_link?%s",
			V1PATH, req.Namespace, req.FileKey, common.Struct2Map(req.Query, "label").Encode()),
		Method: http.MethodGet,
		Handlers: []finderHttp.Handler{
			func(r *http.Request, url string) error {
				headers := common.NewAuthHeaders(url, http.MethodGet, c.auth.ApiKey, c.auth.ApiSecret, nil)
				for k, v := range headers {
					r.Header.Set(k, v)
				}
				return nil
			},
		},
	}

	resp, err := common.DealConn(c.cli, ctx, r)
	if err != nil {
		return res, err
	}

	defer resp.Body.Close()

	msg, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, sgErr.NewError(sgErr.IO_READ_ERR_CODE, err)
	}

	midRes := getLinkResult{}
	err = json.Unmarshal(msg, &midRes)
	if err != nil {
		return res, sgErr.NewError(sgErr.JSON_UNMARSHAL_ERR_CODE, err)
	}

	if resp.StatusCode != 200 {
		return res, sgErr.WarpHttpErr(resp.StatusCode, midRes.Code, midRes.Sid, midRes.Message)
	}

	if midRes.Code != 0 {
		return res, sgErr.WarpMsgErr(midRes.Code, midRes.Sid, midRes.Message)
	}

	res = GetLinkRes{
		Sid: midRes.Sid,
	}

	if req.Query.SplitHost {
		res.Link = midRes.Data.LinkPath
	} else {
		res.Link = midRes.Data.Link
	}

	return res, nil
}
