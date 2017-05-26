package notify

import "github.com/humpback/gounits/logger"
import "gopkg.in/gomail.v1"

// SMTPEndPoint is exported
type SMTPEndPoint struct {
	IEndPoint
	EndPoint
	mailer *gomail.Mailer
}

// NewSMTPEndpoint is exported
func NewSMTPEndpoint(endpoint EndPoint) IEndPoint {

	mailer := gomail.NewMailer(endpoint.Host, endpoint.User, endpoint.Password, endpoint.Port)
	return &SMTPEndPoint{
		EndPoint: endpoint,
		mailer:   mailer,
	}
}

// DoEvent is exported
func (endpoint *SMTPEndPoint) DoEvent(event *Event, data interface{}) {

	if !endpoint.Enabled {
		return
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", endpoint.Sender)
	msg.SetHeader("To", event.ContactInfo)
	msg.SetHeader("Subject", event.makeSubjectText())
	msg.SetBody("text/html", data.(string))
	if err := endpoint.mailer.Send(msg); err != nil {
		logger.ERROR("[#notify#] smtp endpoint post error: %s", err.Error())
	}
}
