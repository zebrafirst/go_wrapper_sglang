package common

import (
	"net/http"
	"time"
)

func DefaultClient(timeout int) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxConnsPerHost = 1024
	transport.MaxIdleConnsPerHost = 128

	c := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}
	return c
}
