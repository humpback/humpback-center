package types

// OperatedContainer is exported
type OperatedContainer struct {
	IP          string `json:"IP"`
	HostName    string `json:"HostName"`
	ContainerID string `json:"ContainerId"`
	Result      string `json:"Result"`
}

// OperatedContainers is exported
type OperatedContainers []*OperatedContainer

// SetOperatedPair is exported
func (operated OperatedContainers) SetOperatedPair(ip string, hostname string, containerid string, action string, err error) OperatedContainers {

	result := action + " successed."
	if err != nil {
		result = action + " failure, " + err.Error()
	}

	operatedContainer := &OperatedContainer{
		IP:          ip,
		HostName:    hostname,
		ContainerID: containerid,
		Result:      result,
	}
	operated = append(operated, operatedContainer)
	return operated
}
