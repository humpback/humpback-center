package types

import "github.com/humpback/common/models"

// EngineContainer is exported
type EngineContainer struct {
	IP        string           `json:"IP"`
	HostName  string           `json:"HostName"`
	Container models.Container `json:"Container"`
}

// GroupContainer is exported
type GroupContainer struct {
	GroupID       string             `json:"GroupId"`
	MetaID        string             `json:"MetaId"`
	IsRemoveDelay bool               `json:"IsRemoveDelay"`
	Instances     int                `json:"Instances"`
	Placement     Placement          `json:"Placement"`
	WebHooks      WebHooks           `json:"WebHooks"`
	Config        models.Container   `json:"Config"`
	Containers    []*EngineContainer `json:"Containers"`
	CreateAt      int64              `json:"CreateAt"`
	LastUpdateAt  int64              `json:"LastUpdateAt"`
}

// GroupContainers is exported
type GroupContainers []*GroupContainer
