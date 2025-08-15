package api_v1

import (
	"context"
	"log"

	finderHttp "git.iflytek.com/AIaaS/finderhttp-self"
	"git.iflytek.com/AIaaS/storage-gateway-sdk-self/common"
)

const (
	V1PATH        = "/api/v1"
	LenLimit      = 50 * 1024 * 1024
	SmallLen      = 10 * 1024 * 1024
	DefaultWorker = 1
)

type apiV1Cli struct {
	cli  finderHttp.Client
	auth common.Auth
}

type APIV1Client interface {
	FileUpload(ctx context.Context, req UploadReq) (UploadRes, error)
	FileDownload(ctx context.Context, req DownloadReq) (DownloadRes, error)
	FileUpdate(ctx context.Context, req UpdateReq) (UpdateRes, error)
	FileDelete(ctx context.Context, req DeleteReq) (DeleteRes, error)
	GetLink(ctx context.Context, req GetLinkReq) (GetLinkRes, error)

	MltUploadInit(ctx context.Context, req MltUploadInitReq) (MltUploadInitRes, error)
	MltUpload(ctx context.Context, req MltUploadReq) (MltUploadRes, error)
	MltUploadFin(ctx context.Context, req MltUploadFinReq) (MltUploadFinRes, error)
	MltUploadAll(ctx context.Context, req MltUploadAllReq) (MltUploadFinRes, error)
	MltUploadCL(ctx context.Context, req MltUploadCLReq) (MltUploadCLRes, error)

	MltUpdateInit(ctx context.Context, req MltUpdateInitReq) (MltUpdateInitRes, error)
	MltUpdate(ctx context.Context, req MltUpdateReq) (MltUpdateRes, error)
	MltUpdateFin(ctx context.Context, req MltUpdateFinReq) (MltUpdateFinRes, error)
	MltUpdateAll(ctx context.Context, req MltUpdateAllReq) (MltUpdateFinRes, error)
	MltUpdateCL(ctx context.Context, req MltUpdateCLReq) (MltUpdateCLRes, error)

	Close()
}

func NewV1Client(conf *common.CliConf, apiKey, apiSecret string) (APIV1Client, error) {
	common.CheckConf(conf)
	logger, err := finderHttp.NewLogger(conf.LogName, conf.LogLevel)
	if err != nil {
		log.Printf("create logger err: %v, use default logger\n", err)
	}

	if conf.UseFinder {
		conf.Address = nil
	} else {
		conf.FinderURL = ""
	}

	cli, err := finderHttp.NewClient(&finderHttp.ClientOptions{
		FinderUrl:  conf.FinderURL,
		Address:    conf.Address,
		LB:         conf.LB,
		Ping:       conf.Ping,
		Logger:     logger,
		UseTLS:     conf.UseTLS,
		HttpClient: common.DefaultClient(conf.TimeOut),
	})

	if err != nil {
		return nil, err
	}

	a := &apiV1Cli{
		auth: common.Auth{
			ApiKey:    apiKey,
			ApiSecret: apiSecret,
		},
		cli: cli,
	}

	return a, nil
}

func (c *apiV1Cli) Close() {
	c.cli.Destory()
}

// func strgType(st string, length int) string {
// 	if len(st) == 0 {
// 		if length < SmallLen {
// 			st = "hbase"
// 		} else {
// 			st = "s3"
// 		}
// 	}
// 	return st
// }
