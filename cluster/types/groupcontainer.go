package types

import "github.com/humpback/humpback-agent/models"

// EngineContainer is exported
type EngineContainer struct {
	IP        string           `json:"IP"`
	HostName  string           `json:"HostName"`
	Container models.Container `json:"Container"`
}

// GroupContainer is exported
type GroupContainer struct {
	MetaID     string             `json:"MetaId"`
	Instances  int                `json:"Instances"`
	WebHooks   WebHooks           `json:"WebHooks"`
	Config     models.Container   `json:"Config"`
	Containers []*EngineContainer `json:"Containers"`
}

// GroupContainers is exported
type GroupContainers []*GroupContainer
