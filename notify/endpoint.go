package notify

// IEndPoint is exported
// sender endPoint interface
type IEndPoint interface {
	DoEvent(event *Event, data interface{})
}
