package types

// OperatedContainer is exported
type OperatedContainer struct {
	IP          string `json:"IP"`
	ContainerID string `json:"ContainerID"`
	Result      string `json:"Result"`
}

// OperatedContainers is exported
type OperatedContainers []*OperatedContainer

// SetOperatedPair is exported
func (operated OperatedContainers) SetOperatedPair(ip string, containerid string, err error) OperatedContainers {

	result := "operate successed."
	if err != nil {
		result = err.Error()
	}

	operatedContainer := &OperatedContainer{
		IP:          ip,
		ContainerID: containerid,
		Result:      result,
	}
	operated = append(operated, operatedContainer)
	return operated
}
