package response

import "github.com/humpback/humpback-center/cluster/types"
import "github.com/humpback/humpback-center/models"

/*
ClusterGroupsResponse
Method:  GET
Route:   /v1/cluster/groups
*/
type ClusterGroupsResponse struct {
	Groups []*models.Group `json:"groups"`
}

func NewClusterGroupsResponse(groups []*models.Group) *ClusterGroupsResponse {

	return &ClusterGroupsResponse{
		Groups: groups,
	}
}

/*
ClusterGroupResponse
Method:  GET
Route:   /v1/cluster/groups/{groupid}
*/
type ClusterGroupResponse struct {
	Group *models.Group `json:"group"`
}

func NewClusterGroupResponse(group *models.Group) *ClusterGroupResponse {

	return &ClusterGroupResponse{
		Group: group,
	}
}

/*
ClusterGroupEnginesResponse
Method:  GET
Route:   /v1/cluster/groups/{groupid}/engines
*/
type ClusterGroupEnginesResponse struct {
	Engines []*models.Engine `json:"engines"`
}

func NewClusterGroupEnginesResponse(engines []*models.Engine) *ClusterGroupEnginesResponse {

	return &ClusterGroupEnginesResponse{
		Engines: engines,
	}
}

/*
ClusterEngineResponse
Method:  GET
Route:   /v1/cluster/engines/{engineid}
*/
type ClusterEngineResponse struct {
	Engine *models.Engine `json:"engine"`
}

func NewClusterEngineResponse(engine *models.Engine) *ClusterEngineResponse {

	return &ClusterEngineResponse{
		Engine: engine,
	}
}

/*
ClusterGroupEventResponse
Method:  POST
Route:   /v1/cluster/groups/event
Message: response message
*/
type ClusterGroupEventResponse struct {
	Message string `json:"message"`
}

func NewClusterGroupEventResponse(message string) *ClusterGroupEventResponse {

	return &ClusterGroupEventResponse{
		Message: message,
	}
}

/*
ClusterCreateContainerResponse
Method:  POST
Route:   /v1/cluster/containers
GroupID: cluster groupid
MetaID:  cluster containers metaid
Created: create result message.
Containers: created containers pairs.
*/
type ClusterCreateContainerResponse struct {
	GroupID    string                   `json:"groupid"`
	MetaID     string                   `json:"metaid"`
	Created    string                   `json:"created"`
	Containers *types.CreatedContainers `json:"containers"`
}

func NewClusterCreateContainerResponse(groupid string, metaid string, instances int, containers *types.CreatedContainers) *ClusterCreateContainerResponse {

	created := "created all"
	if instances > len(*containers) {
		created = "created partial"
	}

	return &ClusterCreateContainerResponse{
		GroupID:    groupid,
		MetaID:     metaid,
		Created:    created,
		Containers: containers,
	}
}

/*
ClusterOperateContainerResponse
Method:  PUT
Route1:  /v1/cluster/collections/action
Route2:  /v1/cluster/containers/action
MetaID:    containers metaid
Action:    operate action (start|stop|restart|kill|pause|unpause)
Containers: operated containers pairs.
*/
type ClusterOperateContainersResponse struct {
	MetaID     string                    `json:"metaid"`
	Action     string                    `json:"action"`
	Containers *types.OperatedContainers `json:"containers"`
}

func NewClusterOperateContainersResponse(metaid string, action string, containers *types.OperatedContainers) *ClusterOperateContainersResponse {

	return &ClusterOperateContainersResponse{
		MetaID:     metaid,
		Action:     action,
		Containers: containers,
	}
}

/*
ClusterUpgradeContainerResponse
Method:  PUT
Route:   /v1/cluster/containers/upgrade
MetaID:  containers metaid
*/
type ClusterUpgradeContainerResponse struct {
	MetaID  string `json:"metaid"`
	Upgrade string `json:"upgrade"`
}

func NewClusterUpgradeContainerResponse(metaid string, upgrade string) *ClusterUpgradeContainerResponse {

	return &ClusterUpgradeContainerResponse{
		MetaID:  metaid,
		Upgrade: upgrade,
	}
}

/*
ClusterRemoveContainersResponse
Method:  PUT
Route1:  /v1/cluster/collections/{metaid}
Route2:  /v1/cluster/containers/{containerid}
MetaID:  containers metaid
Containers: removed containers pairs.
*/
type ClusterRemoveContainersResponse struct {
	MetaID     string                   `json:"metaid"`
	Containers *types.RemovedContainers `json:"containers"`
}

func NewClusterRemoveContainersResponse(metaid string, containers *types.RemovedContainers) *ClusterRemoveContainersResponse {

	return &ClusterRemoveContainersResponse{
		MetaID:     metaid,
		Containers: containers,
	}
}
