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

type MltUpdateInitReq struct {
	Namespace string
	FileKey   string
}

type mltUpdateInitResult struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Sid     string       `json:"sid"`
	Data    uploadIdData `json:"data"`
}

type MltUpdateInitRes struct {
	Sid      string
	UploadId string
}

type MltUpdateReq struct {
	Namespace string
	UploadId  string
	SliceId   int
	File      io.ReadSeeker
	Length    int
}

type mltUpdateData struct {
	Info string `json:"info"`
}

type mltUpdateResult struct {
	Code    int           `json:"code"`
	Message string        `json:"message"`
	Sid     string        `json:"sid"`
	Data    mltUpdateData `json:"data"`
}

type MltUpdateRes struct {
	Sid  string
	Info string
}

type MltUpdateFinQuery struct {
	GetLink bool `label:"get_link"`
	LinkTtl int  `label:"link_ttl"`
}

type MltUpdateFinReq struct {
	Namespace string
	UploadId  string
	Query     MltUpdateFinQuery
	Infos     MltFinInfos `json:"info"`
}

type mltUpdateFinData struct {
	FileKey  string `json:"key"`
	Link     string `json:"link"`
	Location string `json:"location"`
}

type mltUpdateFinResult struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Sid     string           `json:"sid"`
	Data    mltUpdateFinData `json:"data"`
}

type MltUpdateFinRes struct {
	Sid      string
	FileKey  string
	Link     string
	Location string
}

type MltUpdateAllReq struct {
	Namespace string
	FileKey   string    // 文件唯一标识符，不能为空
	File      io.Reader // 文件内容
	Size      int       // 分片大小，单位为 Byte，默认为 10 MB，最大为 50 MB
	Worker    int       // 并发数，默认为 1
	Query     MltUpdateFinQuery
}

type MltUpdateCLReq struct {
	Namespace string
	UploadId  string
}

type mltUpdateCLResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Sid     string `json:"sid"`
}

type MltUpdateCLRes struct {
	Sid string
}

func (c *apiV1Cli) MltUpdateInit(ctx context.Context, req MltUpdateInitReq) (MltUpdateInitRes, error) {
	res := MltUpdateInitRes{}

	r := &finderHttp.Request{
		Path:   fmt.Sprintf("%s/%s/multipart_update/init/%s", V1PATH, req.Namespace, req.FileKey),
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

	midRes := mltUpdateInitResult{}
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

	res = MltUpdateInitRes{
		Sid:      midRes.Sid,
		UploadId: midRes.Data.UploadId,
	}
	return res, nil
}

func (c *apiV1Cli) MltUpdate(ctx context.Context, req MltUpdateReq) (MltUpdateRes, error) {
	res := MltUpdateRes{}

	r := &finderHttp.Request{
		Path:          fmt.Sprintf("%s/%s/multipart_update/upload/%s/%d", V1PATH, req.Namespace, req.UploadId, req.SliceId),
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

	midRes := mltUpdateResult{}
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

	res = MltUpdateRes{
		Sid:  midRes.Sid,
		Info: midRes.Data.Info,
	}
	return res, nil
}

func (c *apiV1Cli) MltUpdateFin(ctx context.Context, req MltUpdateFinReq) (MltUpdateFinRes, error) {
	res := MltUpdateFinRes{}

	sort.Slice(req.Infos.Info, func(i, j int) bool {
		return req.Infos.Info[i].SliceId < req.Infos.Info[j].SliceId
	})

	infos, err := json.Marshal(req.Infos)
	if err != nil {
		return res, sgErr.NewError(sgErr.JSON_UNMARSHAL_ERR_CODE, err)
	}

	r := &finderHttp.Request{
		Path: fmt.Sprintf("%s/%s/multipart_update/finish/%s?%s",
			V1PATH, req.Namespace, req.UploadId, common.Struct2Map(req.Query, "label").Encode()),
		Method:        http.MethodPost,
		Body:          bytes.NewReader(infos),
		ContentLength: len(infos),
		Handlers: []finderHttp.Handler{
			func(r *http.Request, url string) error {
				headers := common.NewAuthHeaders(url, http.MethodPost, c.auth.ApiKey, c.auth.ApiSecret, infos)
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

	midRes := mltUpdateFinResult{}
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

	res = MltUpdateFinRes{
		Sid:      midRes.Sid,
		FileKey:  midRes.Data.FileKey,
		Link:     midRes.Data.Link,
		Location: midRes.Data.Location,
	}
	return res, nil
}

func (c *apiV1Cli) MltUpdateAll(ctx context.Context, req MltUpdateAllReq) (MltUpdateFinRes, error) {
	res := MltUpdateFinRes{}

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
	initRes, err := c.MltUpdateInit(subCtx, MltUpdateInitReq{
		Namespace: req.Namespace,
		FileKey:   req.FileKey,
	})

	if err != nil {
		return res, err
	}

	type updateRet struct {
		res     MltUpdateRes
		sliceId int
		err     error
	}

	// 更新请求队列
	reqC := make(chan *MltUpdateReq, len(files))
	// 收集首个错误结果
	updateErrC := make(chan updateRet, 1)
	// 收集所有的正确结果
	updateResC := make(chan updateRet, len(files))

	for _, f := range files {
		updateReq := &MltUpdateReq{
			Namespace: req.Namespace,
			UploadId:  initRes.UploadId,
			SliceId:   f.SliceId,
			File:      f.File,
			Length:    f.Length,
		}
		reqC <- updateReq
	}

	wg := sync.WaitGroup{}

	// 并发更新请求
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
					uRes, uErr := c.MltUpdate(subCtx, *r)
					// 如果出现 err 不为空，结束 goroutine
					if uErr != nil {
						updateErrC <- updateRet{uRes, r.SliceId, uErr}
						cancel()
						return
					}
					updateResC <- updateRet{uRes, r.SliceId, nil}
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
	close(updateErrC)
	close(updateResC)
	close(reqC)

	// 处理收集到错误结果
	for uRet := range updateErrC {
		if uRet.err != nil {
			return res, uRet.err
		}
	}

	// 处理收集到的正确结果
	if len(updateResC) != len(files) {
		return res, sgErr.NewError(sgErr.RES_NOT_MEET_EXPECT_CODE, errors.New("res don't meet exception"))
	}

	infos := make([]MltFinInfo, len(updateResC))
	i := 0
	for ret := range updateResC {
		infos[i].SliceId = ret.sliceId
		infos[i].Info = ret.res.Info
		i += 1
	}

	// 更新结束请求
	return c.MltUpdateFin(ctx, MltUpdateFinReq{
		Namespace: req.Namespace,
		UploadId:  initRes.UploadId,
		Query:     req.Query,
		Infos: MltFinInfos{
			Info: infos,
		},
	})
}

func (c *apiV1Cli) MltUpdateCL(ctx context.Context, req MltUpdateCLReq) (MltUpdateCLRes, error) {
	res := MltUpdateCLRes{}

	r := &finderHttp.Request{
		Path:   fmt.Sprintf("%s/%s/multipart_update/cancel/%s", V1PATH, req.Namespace, req.UploadId),
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

	midRes := mltUpdateCLResult{}
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

	res = MltUpdateCLRes{
		Sid: midRes.Sid,
	}
	return res, nil
}
