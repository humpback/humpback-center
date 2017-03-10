package types

import "github.com/humpback/humpback-agent/models"

// CreatedContainer is exported
type CreatedContainer struct {
	EngineIP string `json:"EngineIP"`
	models.Container
}

// CreatedContainers is exported
type CreatedContainers []*CreatedContainer

// SetContainer is exported
func (created CreatedContainers) SetContainer(engineIP string, container models.Container) CreatedContainers {

	createdContainer := &CreatedContainer{
		EngineIP:  engineIP,
		Container: container,
	}
	created = append(created, createdContainer)
	return created
}
