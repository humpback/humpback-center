package request

import (
	"errors"
)

const (
	RequestSuccessed int = 0
	RequestInvalid   int = -1000
	RequestNotFound  int = -1001
	RequestConflict  int = -1003
	RequestException int = -1004
)

var (
	ErrRequestSuccessed = errors.New("request successed.")
	ErrRequestInvalid   = errors.New("request resolve invalid.")
	ErrRequestNotFound  = errors.New("request resource not found.")
	ErrRequestConflict  = errors.New("request resource conflict.")
	ErrRequestException = errors.New("request server exception.")
)
