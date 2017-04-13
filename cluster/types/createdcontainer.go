package types

import "common/models"

// CreateContainerResponse is exported
type CreateContainerResponse struct {
	ID       string   `json:"Id"`
	Name     string   `json:"Name"`
	Warnings []string `json:"Warnings"`
}

// CreatedContainer is exported
type CreatedContainer struct {
	IP       string `json:"IP"`
	HostName string `json:"HostName"`
	models.Container
}

// CreatedContainers is exported
type CreatedContainers []*CreatedContainer

// SetCreatedPair is exported
func (created CreatedContainers) SetCreatedPair(ip string, hostname string, container models.Container) CreatedContainers {

	createdContainer := &CreatedContainer{
		IP:        ip,
		HostName:  hostname,
		Container: container,
	}
	created = append(created, createdContainer)
	return created
}
