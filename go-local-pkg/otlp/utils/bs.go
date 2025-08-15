package utils

import (
	"fmt"
	"reflect"
	"unsafe"
)

func AnyToString(val interface{}) string {
	var value = ""
	switch val.(type) {
	case bool:
		if val == true {
			value = "true"
		} else {
			value = "false"
		}
	case int, int64, int32, int16, int8, uint8, uint16, uint32, uint64, uint:
		value = fmt.Sprintf("%d", val)
	case float32, float64:
		value = fmt.Sprintf("%f", val)
	case string:
		value = fmt.Sprintf("%s", val)
	default:
		return ""
	}
	return value
}

func B2S(b []byte) string {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	strHeader := reflect.StringHeader{
		Data: bytesHeader.Data,
		Len:  bytesHeader.Len,
	}
	return *(*string)(unsafe.Pointer(&strHeader))
}

func S2B(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func ToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func ToBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	shh := &sliceHeader{
		data: unsafe.Pointer(sh.Data),
		len:  sh.Len,
		cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(shh))
}

type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}
