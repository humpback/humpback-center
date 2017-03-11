package types

import "github.com/humpback/humpback-agent/models"

// CreatedContainer is exported
type CreatedContainer struct {
	IP string `json:"IP"`
	models.Container
}

// CreatedContainers is exported
type CreatedContainers []*CreatedContainer

// SetContainer is exported
func (created CreatedContainers) SetCreatedPair(ip string, container models.Container) CreatedContainers {

	createdContainer := &CreatedContainer{
		IP:        ip,
		Container: container,
	}
	created = append(created, createdContainer)
	return created
}
