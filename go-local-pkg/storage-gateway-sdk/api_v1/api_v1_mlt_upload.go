package api_v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"

	finderHttp "git.iflytek.com/AIaaS/finderhttp-self"
	"git.iflytek.com/AIaaS/storage-gateway-sdk-self/common"
	sgErr "git.iflytek.com/AIaaS/storage-gateway-sdk-self/sg_err"
)

type MltUploadInitReq struct {
	Namespace string
	Filename  string // 必须要传文件名
}

type uploadIdData struct {
	UploadId string `json:"upload_id"`
}

type mltUploadInitResult struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Sid     string       `json:"sid"`
	Data    uploadIdData `json:"data"`
}

type MltUploadInitRes struct {
	Sid      string
	UploadId string
}

type MltUploadReq struct {
	Namespace string
	UploadId  string
	Filename  string        // 需要与初始化时的 Filename 一致
	SliceId   int           // 从 1 开始
	File      io.ReadSeeker // 文件内容
	Length    int           // 文件内容长度，单位：字节(Byte)
}

type mltUploadData struct {
	Info string `json:"info"`
}

type mltUploadResult struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Sid     string        `json:"sid"`
	Data    mltUploadData `json:"data"`
}

type MltUploadRes struct {
	Sid  string
	Info string
}

type MltUploadFinQuery struct {
	GetLink  bool   `label:"get_link"`
	LinkTtl  int    `label:"link_ttl"`
	Filename string `label:"filename"`
}

type MltFinInfo struct {
	SliceId int    `json:"slice_id"`
	Info    string `json:"info"`
}

type MltFinInfos struct {
	Info []MltFinInfo `json:"info"`
}

type MltUploadFinReq struct {
	Namespace string
	UploadId  string
	Query     MltUploadFinQuery
	Infos     MltFinInfos
}

type mltUploadFinData struct {
	FileKey  string `json:"key"`
	Link     string `json:"link"`
	Location string `json:"location"`
}

type mltUploadFinResult struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Sid     string           `json:"sid"`
	Data    mltUploadFinData `json:"data"`
}

type MltUploadFinRes struct {
	Sid      string
	FileKey  string
	Link     string
	Location string
}

type MltUploadAllReq struct {
	Namespace string
	Filename  string // 必须要传文件名
	File      io.Reader
	Size      int // 分片大小，单位为 Byte，默认为 10 MB，最大不超过 50 MB
	Worker    int // 并发数，默认为 1
	Query     MltUploadFinQuery
}

type MltUploadCLReq struct {
	Namespace string
	UploadId  string
	Filename  string
}

type mltUploadCLResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Sid     string `json:"sid"`
}

type MltUploadCLRes struct {
	Sid string
}

func withName(path, filename string) string {
	if len(filename) != 0 {
		path += "?filename=" + filename
	}
	return path
}

func (c *apiV1Cli) MltUploadInit(ctx context.Context, req MltUploadInitReq) (MltUploadInitRes, error) {
	res := MltUploadInitRes{}

	path := fmt.Sprintf("%s/%s/multipart/init", V1PATH, req.Namespace)
	path = withName(path, req.Filename)

	r := &finderHttp.Request{
		Path:   path,
		Method: http.MethodPost,
		Handlers: []finderHttp.Handler{
			func(r *http.Request, url string) error {
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

	midRes := mltUploadInitResult{}
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

	res = MltUploadInitRes{
		Sid:      midRes.Sid,
		UploadId: midRes.Data.UploadId,
	}
	return res, nil
}

func (c *apiV1Cli) MltUpload(ctx context.Context, req MltUploadReq) (MltUploadRes, error) {
	res := MltUploadRes{}

	if req.SliceId <= 0 {
		return res, sgErr.NewError(sgErr.PARAMETER_ILLEGAL_ERR_CODE, errors.New("slice_id must be positive integer"))
	}

	path := fmt.Sprintf("%s/%s/multipart/upload/%s/%d", V1PATH, req.Namespace, req.UploadId, req.SliceId)
	path = withName(path, req.Filename)

	r := &finderHttp.Request{
		Path:          path,
		Method:        http.MethodPost,
		Body:          req.File,
		ContentLength: req.Length,
		Handlers: []finderHttp.Handler{
			func(r *http.Request, url string) error {
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

	midRes := mltUploadResult{}
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

	res = MltUploadRes{
		Sid:  midRes.Sid,
		Info: midRes.Data.Info,
	}
	return res, nil
}

func (c *apiV1Cli) MltUploadFin(ctx context.Context, req MltUploadFinReq) (MltUploadFinRes, error) {
	res := MltUploadFinRes{}

	sort.Slice(req.Infos.Info, func(i, j int) bool {
		return req.Infos.Info[i].SliceId < req.Infos.Info[j].SliceId
	})

	infos, err := json.Marshal(req.Infos)
	if err != nil {
		return res, sgErr.NewError(sgErr.JSON_UNMARSHAL_ERR_CODE, err)
	}

	r := &finderHttp.Request{
		Path: fmt.Sprintf("%s/%s/multipart/finish/%s?%s",
			V1PATH, req.Namespace, req.UploadId, common.Struct2Map(req.Query, "label").Encode()),
		Method:        http.MethodPost,
		Body:          bytes.NewReader(infos),
		ContentLength: len(infos),
		Handlers: []finderHttp.Handler{
			func(r *http.Request, url string) error {
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

	midRes := mltUploadFinResult{}
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

	res = MltUploadFinRes{
		Sid:      midRes.Sid,
		FileKey:  midRes.Data.FileKey,
		Link:     midRes.Data.Link,
		Location: midRes.Data.Location,
	}
	return res, nil
}

func (c *apiV1Cli) MltUploadAll(ctx context.Context, req MltUploadAllReq) (MltUploadFinRes, error) {
	res := MltUploadFinRes{}

	if req.Size <= 0 || req.Size > LenLimit {
		req.Size = SmallLen
	}

	if req.Worker <= 0 {
		req.Worker = DefaultWorker
	}

	// 分割大文件
	files, err := common.SliceFile(req.File, req.Size)
	if err != nil {
		return res, err
	}

	// 并发使用 根据 subCtx 通知并发结束
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 更新初始化请求
	initRes, err := c.MltUploadInit(subCtx, MltUploadInitReq{
		Namespace: req.Namespace,
		Filename:  req.Filename,
	})

	if err != nil {
		return res, err
	}

	type uploadRet struct {
		res     MltUploadRes
		sliceId int
		err     error
	}

	// 更新请求队列
	reqC := make(chan *MltUploadReq, len(files))
	// 收集首个错误结果
	uploadErrC := make(chan uploadRet, 1)
	// 收集所有的正确结果
	uploadResC := make(chan uploadRet, len(files))

	for _, f := range files {
		uploadReq := &MltUploadReq{
			Namespace: req.Namespace,
			UploadId:  initRes.UploadId,
			Filename:  req.Filename,
			SliceId:   f.SliceId,
			File:      f.File,
			Length:    f.Length,
		}
		reqC <- uploadReq
	}

	wg := sync.WaitGroup{}

	// 并发上传请求
	for i := 0; i < req.Worker; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case r, ok := <-reqC:
					if !ok {
						return
					}
					// fmt.Println(r.SliceId)
					uRes, uErr := c.MltUpload(subCtx, *r)
					// 如果出现 err，结束 goroutine
					if uErr != nil {
						uploadErrC <- uploadRet{uRes, r.SliceId, uErr}
						cancel()
						return
					}
					uploadResC <- uploadRet{uRes, r.SliceId, nil}
				case <-ctx.Done():
					return
				default:
					return
				}
			}
		}()
	}

	// 阻塞，等待并发结束并关闭 channel
	wg.Wait()
	close(uploadErrC)
	close(uploadResC)
	close(reqC)

	// 处理收集到错误结果
	for uRet := range uploadErrC {
		if uRet.err != nil {
			return res, uRet.err
		}
	}

	// 处理收集到的正确结果
	if len(uploadResC) != len(files) {
		return res, sgErr.NewError(sgErr.RES_NOT_MEET_EXPECT_CODE, errors.New("res don't meet exception"))
	}

	infos := make([]MltFinInfo, len(uploadResC))
	i := 0
	for ret := range uploadResC {
		infos[i].SliceId = ret.sliceId
		infos[i].Info = ret.res.Info
		i += 1
	}

	// 更新结束请求
	return c.MltUploadFin(ctx, MltUploadFinReq{
		Namespace: req.Namespace,
		UploadId:  initRes.UploadId,
		Query:     req.Query,
		Infos: MltFinInfos{
			Info: infos,
		},
	})
}

func (c *apiV1Cli) MltUploadCL(ctx context.Context, req MltUploadCLReq) (MltUploadCLRes, error) {
	res := MltUploadCLRes{}

	if len(req.Filename) == 0 {
		return res, sgErr.NewError(sgErr.PARAMETER_MISS_ERR_CODE, errors.New("filename length is 0"))
	}

	path := fmt.Sprintf("%s/%s/multipart/cancel/%s", V1PATH, req.Namespace, req.UploadId)
	path = withName(path, req.Filename)

	r := &finderHttp.Request{
		Path:   path,
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

	midRes := mltUploadCLResult{}
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

	res = MltUploadCLRes{
		Sid: midRes.Sid,
	}
	return res, nil
}
