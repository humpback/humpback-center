package notify

import "github.com/humpback/gounits/http"
import "github.com/humpback/gounits/logger"

import (
	"time"
)

// APIEndPoint is exported
type APIEndPoint struct {
	IEndPoint
	EndPoint
	client *http.HttpClient
}

// NewAPIEndPoint is exported
func NewAPIEndPoint(endpoint EndPoint) IEndPoint {

	return &APIEndPoint{
		EndPoint: endpoint,
		client:   http.NewWithTimeout(15 * time.Second),
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

	response, err := endpoint.client.PostJSON(endpoint.URL, nil, value, endpoint.Headers)
	if err != nil {
		logger.ERROR("[#notify#] api endpoint error: %s", err.Error())
		return
	}
	defer response.Close()
	if response.StatusCode() != 200 {
		logger.ERROR("[#notify#] api endpoint response code: %d", response.StatusCode())
	}
}
