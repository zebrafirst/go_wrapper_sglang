package finderhttp

import "fmt"

//go:generate stringer -type ErrType  -output error_string.go
type ErrType int

const (
	ErrKnown                  = 0
	ErrNoStorageFound ErrType = iota + 8000
	ErrRespNotJSON
	ErrNoPeerAddr
)

type Error struct {
	Type    ErrType
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%v|%s", e.Type, e.Message)
}

func NewError(t ErrType, msg string) error {
	return &Error{
		Type:    t,
		Message: msg,
	}
}

var (
	ErrNoPeerAddrFound = NewError(ErrNoPeerAddr, "no available peer addr found")
)
