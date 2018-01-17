package notify

import "github.com/humpback/gounits/httpx"
import "github.com/humpback/gounits/logger"

import (
	"context"
	"time"
	"net"
	"net/http"
)

// APIEndPoint is exported
type APIEndPoint struct {
	IEndPoint
	EndPoint
	client *httpx.HttpClient
}

// NewAPIEndPoint is exported
func NewAPIEndPoint(endpoint EndPoint) IEndPoint {

	client := httpx.NewClient().
		SetTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   45 * time.Second,
				KeepAlive: 90 * time.Second,
			}).DialContext,
			DisableKeepAlives:     false,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout,
			ExpectContinueTimeout: http.DefaultTransport.(*http.Transport).ExpectContinueTimeout,
		})

	return &APIEndPoint{
		EndPoint: endpoint,
		client:   client,
	}
}

// DoEvent is exported
func (endpoint *APIEndPoint) DoEvent(event *Event, data interface{}) {

	if !endpoint.Enabled {
		return
	}

	value := map[string]interface{}{
		"From":        endpoint.Sender,
		"To":          event.ContactInfo,
		"Subject":     event.makeSubjectText(),
		"Body":        data,
		"ContentType": "HTML",
		"MailType":    "Smtp",
		"SmtpSetting": map[string]interface{}{},
	}

	response, err := endpoint.client.PostJSON(context.Background(), endpoint.URL, nil, value, endpoint.Headers)
	if err != nil {
		logger.ERROR("[#notify#] api endpoint error: %s", err.Error())
		return
	}
	defer response.Close()
	if response.StatusCode() >= http.StatusBadRequest {
		logger.ERROR("[#notify#] api endpoint response code: %d", response.StatusCode())
	}
}
