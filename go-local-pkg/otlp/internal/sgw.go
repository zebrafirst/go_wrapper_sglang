package internal

import (
	"bytes"
	"context"
	apiV1 "git.iflytek.com/AIaaS/storage-gateway-sdk-self/api_v1"
	"git.iflytek.com/AIaaS/storage-gateway-sdk-self/common"
	"os"
	"path/filepath"
	"time"
)

type Sgw struct {
	sgwClient apiV1.APIV1Client
	Namespace string
	LinkTtl   int
}

type SgwConfig struct {
	SgwAccessKey string
	SgwSecretKey string
	SgwAddress   string
	SgwUpTimeOut time.Duration // 上传超时时间 改成ms，小于eventlog的flushtimeout
	SgwLogPath   string
	SgwLogLevel  int
	Namespace    string
	LinkTtl      int
}

func NewSgw(conf *SgwConfig) (*Sgw, error) {
	if len(conf.SgwLogPath) == 0 {
		path, _ := os.Getwd()
		conf.SgwLogPath = filepath.Join(path, "sgwsdk.log")
	}
	sgwClient, err := apiV1.NewV1Client(&common.CliConf{
		UseFinder: false,
		Address:   []string{conf.SgwAddress},
		FinderURL: "",
		//LB:        finderhttp.LBRoundRobin(),
		LogName:  conf.SgwLogPath,
		LogLevel: conf.SgwLogLevel, // error
		UseTLS:   false,
		TimeOut:  int(conf.SgwUpTimeOut.Milliseconds()),
	}, conf.SgwAccessKey, conf.SgwSecretKey)
	if err != nil {
		return nil, err
	}

	return &Sgw{
		sgwClient: sgwClient,
		Namespace: conf.Namespace,
		LinkTtl:   conf.LinkTtl,
	}, nil
}

func (s *Sgw) Upload(ctx context.Context, file []byte) (string, error) {
	res, err := s.sgwClient.FileUpload(ctx, apiV1.UploadReq{
		Namespace: s.Namespace,
		File:      bytes.NewReader(file),
		Length:    len(file),
		XTtl:      s.LinkTtl,
		Query: apiV1.UploadQuery{
			GetLink:   true,
			LinkTtl:   s.LinkTtl,
			SplitHost: false,
			//Filename:  "test.otlplog",
			//Expose:    true,
		},
	})
	if err != nil {
		return "", err
	}
	linkUrl := res.Link
	return linkUrl, nil
}
