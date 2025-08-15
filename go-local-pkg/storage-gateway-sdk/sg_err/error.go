package sgErr

import (
	"errors"
	"fmt"
)

type SGError struct {
	code int
	err  error
}

const (
	GET_RESP_ERR_CODE = 94000 + iota
	RESP_CODE_ERR_CODE
	IO_READ_ERR_CODE
	SLICE_FILE_ERR_CODE
	JSON_UNMARSHAL_ERR_CODE
	PARAMETER_MISS_ERR_CODE
	PARAMETER_ILLEGAL_ERR_CODE
	RES_NOT_MEET_EXPECT_CODE
	FILE_LENGTH_OVER_CODE
)

var errMsg = map[int]string{
	GET_RESP_ERR_CODE:          "cannot get response",
	RESP_CODE_ERR_CODE:         "response code is not 0",
	IO_READ_ERR_CODE:           "io.ReadAll err",
	SLICE_FILE_ERR_CODE:        "slice file err",
	JSON_UNMARSHAL_ERR_CODE:    "json.Marshal err",
	PARAMETER_MISS_ERR_CODE:    "parameter miss",
	PARAMETER_ILLEGAL_ERR_CODE: "parameter illegal",
	FILE_LENGTH_OVER_CODE:      "file length is over 50MB, try to use MltUpload",
}

func NewError(code int, err error) *SGError {
	if err == nil {
		err = errors.New(errMsg[code])
	}

	return &SGError{
		code: code,
		err:  err,
	}
}

func WarpHttpErr(statusCode, code int, sid, msg string) *SGError {
	return NewError(statusCode, errors.New(fmt.Sprintf("resp code => %d | sid => %s | msg => %s", code, sid, msg)))
}

func WarpMsgErr(code int, sid, msg string) *SGError {
	return NewError(code, errors.New(fmt.Sprintf("sid => %s | message => %s", sid, msg)))
}

func (s *SGError) GetCode() int {
	return s.code
}

func (s *SGError) GetErr() error {
	return s.err
}

func (s *SGError) Error() string {
	return fmt.Sprintf("code: %d | error: %v", s.code, s.err)
}
