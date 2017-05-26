package notify

//INotifyEndPointFactory is exported
type INotifyEndPointFactory interface {
	CreateAPIEndPoint(endpoint EndPoint) IEndPoint
	CreateSMTPEndPoint(endpoint EndPoint) IEndPoint
}

//NotifyEndPointFactory is exported
type NotifyEndPointFactory struct {
	INotifyEndPointFactory
}

//CreateAPIEndPoint is exported
func (factory *NotifyEndPointFactory) CreateAPIEndPoint(endpoint EndPoint) IEndPoint {
	return NewAPIEndPoint(endpoint)
}

//CreateSMTPEndPoint is exported
func (factory *NotifyEndPointFactory) CreateSMTPEndPoint(endpoint EndPoint) IEndPoint {
	return NewSMTPEndpoint(endpoint)
}
