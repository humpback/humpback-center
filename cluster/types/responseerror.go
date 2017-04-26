package types

import "github.com/humpback/gounits/http"

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
func ParseHTTPResponseError(response *http.Response) string {

	responseError := &ResponseError{}
	if err := response.JSON(responseError); err != nil {
		return fmt.Sprintf("agent api error, httpcode:%d", response.StatusCode())
	}
	return responseError.Detail
}
