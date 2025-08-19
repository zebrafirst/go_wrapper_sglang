package logging

import (
	"bytes"
	"context"
	apiV1 "git.iflytek.com/AIaaS/storage-gateway-sdk-self/api_v1"
	"git.iflytek.com/AIaaS/storage-gateway-sdk-self/common"
	"os"
	"path/filepath"
	"time"
)

const (
	DEFAULT_TIMEOUT = 30
)

var (
	sgwClient apiV1.APIV1Client
)

type SgwConfig struct {
	SgwAccessKey string
	SgwSecretKey string
	SgwAddress   string
	SgwUpTimeOut time.Duration
	SgwLogPath   string
	SgwLogLevel  int
}

func initSgw(conf *SgwConfig) (err error) {
	if len(conf.SgwLogPath) == 0 {
		path, _ := os.Getwd()
		conf.SgwLogPath = filepath.Join(path, "sgw_sdk.log")
	}
	if sgwClient, err = apiV1.NewV1Client(&common.CliConf{
		UseFinder: false,
		Address:   []string{conf.SgwAddress},
		FinderURL: "",
		//LB:        finderhttp.LBRoundRobin(),
		LogName:  conf.SgwLogPath,
		LogLevel: conf.SgwLogLevel, // error
		UseTLS:   false,
		TimeOut:  int(conf.SgwUpTimeOut.Seconds()),
	}, conf.SgwAccessKey, conf.SgwSecretKey); err != nil {
		return err
	}
	return nil
}

func uploadSgw(ctx context.Context, file []byte, ns string, linkTTL int) (string, error) {
	res, err := sgwClient.FileUpload(ctx, apiV1.UploadReq{
		Namespace: ns,
		File:      bytes.NewReader(file),
		Length:    len(file),
		XTtl:      linkTTL,
		Query: apiV1.UploadQuery{
			GetLink:   true,
			LinkTtl:   linkTTL,
			SplitHost: false,
			//Filename:  "test.logging",
			//Expose:    true,
		},
	})
	if err != nil {
		return "", err
	}
	linkUrl := res.Link
	return linkUrl, nil
}
