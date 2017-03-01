package request

import (
	"errors"
)

const (
	RequestSuccessed int = 0
	RequestInvalid   int = -1001
	RequestFailure   int = -1002
)

var (
	ErrRequestSuccessed = errors.New("request successed")
	ErrRequestInvalid   = errors.New("request resolve error")
	ErrRequestFailure   = errors.New("request failure error")
)
