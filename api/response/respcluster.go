package response

import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/cluster/types"
import "github.com/humpback/common/models"

/*
GroupAllContainersResponse is exported
Method:  GET
Route:   /v1/groups/{groupid}/collections
*/
type GroupAllContainersResponse struct {
	GroupID    string                 `json:"GroupId"`
	Containers *types.GroupContainers `json:"Containers"`
}

// NewGroupAllContainersResponse is exported
func NewGroupAllContainersResponse(groupid string, containers *types.GroupContainers) *GroupAllContainersResponse {

	return &GroupAllContainersResponse{
		GroupID:    groupid,
		Containers: containers,
	}
}

/*
GroupContainersResponse is exported
Method:  GET
Route:   /v1/groups/collections/{metaid}
*/
type GroupContainersResponse struct {
	Container *types.GroupContainer `json:"Container"`
}

// NewGroupContainersResponse is exported
func NewGroupContainersResponse(container *types.GroupContainer) *GroupContainersResponse {

	return &GroupContainersResponse{
		Container: container,
	}
}

// ContainersMetaBase is exported
type ContainersMetaBase struct {
	GroupID   string         `json:"GroupId"`
	MetaID    string         `json:"MetaId"`
	Instances int            `json:"Instances"`
	WebHooks  types.WebHooks `json:"WebHooks"`
	ImageTag  string         `json:"ImageTag"`
	models.Container
}

/*
GroupContainersMetaBaseResponse is exported
Method:  GET
Route:   /v1/groups/collections/{metaid}/base
*/
type GroupContainersMetaBaseResponse struct {
	MetaBase *ContainersMetaBase `json:"MetaBase"`
}

// NewGroupContainersMetaBaseResponse is exported
func NewGroupContainersMetaBaseResponse(metaBase *cluster.MetaBase) *GroupContainersMetaBaseResponse {

	containersMetaBase := &ContainersMetaBase{
		GroupID:   metaBase.GroupID,
		MetaID:    metaBase.MetaID,
		Instances: metaBase.Instances,
		WebHooks:  metaBase.WebHooks,
		ImageTag:  metaBase.ImageTag,
		Container: metaBase.Config,
	}

	return &GroupContainersMetaBaseResponse{
		MetaBase: containersMetaBase,
	}
}

/*
GroupEnginesResponse is exported
Method:  GET
Route:   /v1/groups/{groupid}/engines
*/
type GroupEnginesResponse struct {
	GroupID string            `json:"GroupId"`
	Engines []*cluster.Engine `json:"Engines"`
}

// NewGroupEnginesResponse is exported
func NewGroupEnginesResponse(groupid string, engines []*cluster.Engine) *GroupEnginesResponse {

	return &GroupEnginesResponse{
		GroupID: groupid,
		Engines: engines,
	}
}

/*
GroupEngineResponse is exported
Method:  GET
Route:   /v1/groups/engines/{server}
*/
type GroupEngineResponse struct {
	Engine *cluster.Engine `json:"Engine"`
}

// NewGroupEngineResponse is exported
func NewGroupEngineResponse(engine *cluster.Engine) *GroupEngineResponse {

	return &GroupEngineResponse{
		Engine: engine,
	}
}

/*
GroupEventResponse is exported
Method:  POST
Route:   /v1/groups/event
*/
type GroupEventResponse struct {
	Message string `json:"Message"`
}

// NewGroupEventResponse is exported
func NewGroupEventResponse(message string) *GroupEventResponse {

	return &GroupEventResponse{
		Message: message,
	}
}

/*
GroupCreateContainersResponse is exported
Method:  POST
Route:   /v1/groups/collections
*/
type GroupCreateContainersResponse struct {
	GroupID    string                   `json:"GroupId"`
	MetaID     string                   `json:"MetaId"`
	Created    string                   `json:"Created"`
	Containers *types.CreatedContainers `json:"Containers"`
}

// NewGroupCreateContainersResponse is exported
func NewGroupCreateContainersResponse(groupid string, metaid string, instances int, containers *types.CreatedContainers) *GroupCreateContainersResponse {

	created := "created all"
	if instances > len(*containers) {
		created = "created partial"
	}

	return &GroupCreateContainersResponse{
		GroupID:    groupid,
		MetaID:     metaid,
		Created:    created,
		Containers: containers,
	}
}

/*
GroupUpdateContainersResponse is exported
Method:  PUT
Route:   /v1/groups/collections
*/
type GroupUpdateContainersResponse struct {
	MetaID     string                   `json:"MetaId"`
	Updated    string                   `json:"Updated"`
	Containers *types.CreatedContainers `json:"Containers"`
}

// NewGroupUpdateContainersResponse is exported
func NewGroupUpdateContainersResponse(metaid string, instances int, containers *types.CreatedContainers) *GroupUpdateContainersResponse {

	updated := "updated all"
	if instances > len(*containers) {
		updated = "updated partial"
	}

	return &GroupUpdateContainersResponse{
		MetaID:     metaid,
		Updated:    updated,
		Containers: containers,
	}
}

/*
GroupOperateContainersResponse is exported
Method:  PUT
Route1:  /v1/groups/collections/action
Route2:  /v1/groups/container/action
*/
type GroupOperateContainersResponse struct {
	MetaID     string                    `json:"MetaId"`
	Action     string                    `json:"Action"`
	Containers *types.OperatedContainers `json:"Containers"`
}

// NewGroupOperateContainersResponse is exported
func NewGroupOperateContainersResponse(metaid string, action string, containers *types.OperatedContainers) *GroupOperateContainersResponse {

	return &GroupOperateContainersResponse{
		MetaID:     metaid,
		Action:     action,
		Containers: containers,
	}
}

/*
GroupUpgradeContainersResponse is exported
Method:  PUT
Route:   /v1/groups/collections/upgrade
*/
type GroupUpgradeContainersResponse struct {
	MetaID     string                   `json:"MetaId"`
	Upgrade    string                   `json:"Upgrade"`
	Containers *types.UpgradeContainers `json:"Containers"`
}

// NewGroupUpgradeContainersResponse is exported
func NewGroupUpgradeContainersResponse(metaid string, upgrade string, containers *types.UpgradeContainers) *GroupUpgradeContainersResponse {

	return &GroupUpgradeContainersResponse{
		MetaID:     metaid,
		Upgrade:    upgrade,
		Containers: containers,
	}
}

/*
GroupRemoveContainersResponse is exported
Method:  PUT
Route1:  /v1/groups/collections/{metaid}
Route2:  /v1/groups/container/{containerid}
*/
type GroupRemoveContainersResponse struct {
	MetaID     string                   `json:"MetaId"`
	Containers *types.RemovedContainers `json:"Containers"`
}

// NewGroupRemoveContainersResponse is exported
func NewGroupRemoveContainersResponse(metaid string, containers *types.RemovedContainers) *GroupRemoveContainersResponse {

	return &GroupRemoveContainersResponse{
		MetaID:     metaid,
		Containers: containers,
	}
}
