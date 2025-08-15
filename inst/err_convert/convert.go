package errconvert

import (
	"encoding/json"
	"strings"
)

var (
	UnknowErrCode           = Th3apiErr{Code: 1100, Message: "Unknow err"}                  // 调用三方插件未知错误
	CallTh3ApiUnknowErrCode = Th3apiErr{Code: 1101, Message: "Call th3api unknow err"}      // 调三方api未知错误
	UnmarshalExtraBodyErr   = Th3apiErr{Code: 1102, Message: "Unmarshal extra body failed"} // 反序列化extra body 失败
)

type Th3apiErr struct {
	Code    int
	Message string
}

func (th3apiErr *Th3apiErr) Error() string {
	bs, _ := json.Marshal(th3apiErr)
	return string(bs)
}

// 对调用三方api返回的err进行封装处理，移除敏感信息
func WrapperErr(err error, code int) error {
	if err != nil {
		errStr := err.Error()
		// errStr示例  status: 401 Unauthorized, message: The API key format is incorrect. Request id: 0217545711362391e36c6b25bd05e5381899274b67bacae2be21e
		// 移除掉 Request id: 后面部分
		if i := strings.Index(errStr, "Request id:"); i != -1 {
			errStr = errStr[:i]
		}
		th3apiErr := &Th3apiErr{
			Code:    code,
			Message: errStr,
		}
		return th3apiErr
	}
	return nil
}
