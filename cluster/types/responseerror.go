package types

import "github.com/humpback/gounits/httpx"

import (
	"fmt"
)

/*
  Humpback api response exception struct
*/

// ResponseError is exported
type ResponseError struct {
	Code    int    `json:"Code"`
	Detail  string `json:"Detail"`
	Message string `json:"Message"`
}

// ParseHTTPResponseError is exported
func ParseHTTPResponseError(response *httpx.HttpResponse) string {

	responseError := &ResponseError{}
	if err := response.JSON(responseError); err != nil {
		return fmt.Sprintf("engine client error, httpcode: %d", response.StatusCode())
	}
	return responseError.Detail
}
