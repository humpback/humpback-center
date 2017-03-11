package types

// RemovedContainer is exported
type RemovedContainer struct {
	IP          string `json:"IP"`
	ContainerID string `json:"ContainerID"`
	Result      string `json:"Result"`
}

// RemovedContainers is exported
type RemovedContainers []*RemovedContainer

// SetRemovedPair is exported
func (removed RemovedContainers) SetRemovedPair(ip string, containerid string, err error) RemovedContainers {

	result := "remove successed."
	if err != nil {
		result = err.Error()
	}

	removedContainer := &RemovedContainer{
		IP:          ip,
		ContainerID: containerid,
		Result:      result,
	}
	removed = append(removed, removedContainer)
	return removed
}
