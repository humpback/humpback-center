package response

import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/cluster/types"

/*
ClusterGroupContainersResponse
Method:  GET
Route:   /v1/cluster/groups/{groupid}/containers
*/
type ClusterGroupContainersResponse struct {
	GroupID    string                 `json:"GroupId"`
	Containers *types.GroupContainers `json:"Containers"`
}

func NewClusterGroupContainersResponse(groupid string, containers *types.GroupContainers) *ClusterGroupContainersResponse {

	return &ClusterGroupContainersResponse{
		GroupID:    groupid,
		Containers: containers,
	}
}

/*
ClusterGroupEnginesResponse
Method:  GET
Route:   /v1/cluster/groups/{groupid}/engines
*/
type ClusterGroupEnginesResponse struct {
	GroupID string            `json:"GroupId"`
	Engines []*cluster.Engine `json:"Engines"`
}

func NewClusterGroupEnginesResponse(groupid string, engines []*cluster.Engine) *ClusterGroupEnginesResponse {

	return &ClusterGroupEnginesResponse{
		GroupID: groupid,
		Engines: engines,
	}
}

/*
ClusterEngineResponse
Method:  GET
Route:   /v1/cluster/engines/{engineid}
*/
type ClusterEngineResponse struct {
	Engine *cluster.Engine `json:"Engine"`
}

func NewClusterEngineResponse(engine *cluster.Engine) *ClusterEngineResponse {

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
	Message string `json:"Message"`
}

func NewClusterGroupEventResponse(message string) *ClusterGroupEventResponse {

	return &ClusterGroupEventResponse{
		Message: message,
	}
}

/*
ClusterCreateContainersResponse
Method:  POST
Route:   /v1/cluster/containers
GroupID: cluster groupid
MetaID:  cluster containers metaid
Created: create result message.
Containers: created containers pairs.
*/
type ClusterCreateContainersResponse struct {
	GroupID    string                   `json:"GroupId"`
	MetaID     string                   `json:"MetaId"`
	Created    string                   `json:"Created"`
	Containers *types.CreatedContainers `json:"Containers"`
}

func NewClusterCreateContainersResponse(groupid string, metaid string, instances int, containers *types.CreatedContainers) *ClusterCreateContainersResponse {

	created := "created all"
	if instances > len(*containers) {
		created = "created partial"
	}

	return &ClusterCreateContainersResponse{
		GroupID:    groupid,
		MetaID:     metaid,
		Created:    created,
		Containers: containers,
	}
}

/*
ClusterSetContainersResponse
Method:  PUT
Route:   /v1/cluster/containers
MetaID:  cluster containers metaid
Seted:   set result message.
Containers: set containers pairs.
*/
type ClusterSetContainersResponse struct {
	MetaID     string                   `json:"MetaId"`
	Seted      string                   `json:"Seted"`
	Containers *types.CreatedContainers `json:"Containers"`
}

func NewClusterSetContainersResponse(metaid string, instances int, containers *types.CreatedContainers) *ClusterSetContainersResponse {

	seted := "seted all"
	if instances > len(*containers) {
		seted = "seted partial"
	}

	return &ClusterSetContainersResponse{
		MetaID:     metaid,
		Seted:      seted,
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
	MetaID     string                    `json:"MetaId"`
	Action     string                    `json:"Action"`
	Containers *types.OperatedContainers `json:"Containers"`
}

func NewClusterOperateContainersResponse(metaid string, action string, containers *types.OperatedContainers) *ClusterOperateContainersResponse {

	return &ClusterOperateContainersResponse{
		MetaID:     metaid,
		Action:     action,
		Containers: containers,
	}
}

/*
ClusterUpgradeContainersResponse
Method:  PUT
Route:   /v1/cluster/containers/upgrade
MetaID:  containers metaid
*/
type ClusterUpgradeContainersResponse struct {
	MetaID  string `json:"MetaId"`
	Upgrade string `json:"Upgrade"`
}

func NewClusterUpgradeContainersResponse(metaid string, upgrade string) *ClusterUpgradeContainersResponse {

	return &ClusterUpgradeContainersResponse{
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
	MetaID     string                   `json:"MetaId"`
	Containers *types.RemovedContainers `json:"Containers"`
}

func NewClusterRemoveContainersResponse(metaid string, containers *types.RemovedContainers) *ClusterRemoveContainersResponse {

	return &ClusterRemoveContainersResponse{
		MetaID:     metaid,
		Containers: containers,
	}
}
