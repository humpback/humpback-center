package request

import "github.com/gorilla/mux"
import "github.com/humpback/humpback-agent/models"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

/*
ResolveClusterGroupRequest
Method:  GET
Route:   /v1/cluster/groups/{groupid}
GroupID: cluster groupid
*/
type ClusterGroupRequest struct {
	GroupID string `json:"groupid"`
}

func ResolveClusterGroupRequest(r *http.Request) (*ClusterGroupRequest, error) {

	vars := mux.Vars(r)
	groupid := strings.TrimSpace(vars["groupid"])
	if len(strings.TrimSpace(groupid)) == 0 {
		return nil, fmt.Errorf("groupid invalid, can not be empty")
	}

	return &ClusterGroupRequest{
		GroupID: groupid,
	}, nil
}

/*
ResolveClusterGroupEnginesRequest
Method:  GET
Route:   /v1/cluster/groups/{groupid}/engines
GroupID: cluster groupid
*/
type ClusterGroupEnginesRequest struct {
	GroupID string `json:"groupid"`
}

func ResolveClusterGroupEnginesRequest(r *http.Request) (*ClusterGroupEnginesRequest, error) {

	vars := mux.Vars(r)
	groupid := strings.TrimSpace(vars["groupid"])
	if len(strings.TrimSpace(groupid)) == 0 {
		return nil, fmt.Errorf("groupid invalid, can not be empty")
	}

	return &ClusterGroupEnginesRequest{
		GroupID: groupid,
	}, nil
}

/*
ResolveClusterEngineRequest
Method:  GET
Route:   /v1/cluster/engines/{server}
Server: cluster engine host address
*/
type ClusterEngineRequest struct {
	Server string `json:"server"`
}

func ResolveClusterEngineRequest(r *http.Request) (*ClusterEngineRequest, error) {

	vars := mux.Vars(r)
	server := strings.TrimSpace(vars["server"])
	if len(strings.TrimSpace(server)) == 0 {
		return nil, fmt.Errorf("engine server invalid, can not be empty")
	}

	return &ClusterEngineRequest{
		Server: server,
	}, nil
}

const (
	GROUP_CREATE_EVENT = "create"
	GROUP_REMOVE_EVENT = "remove"
	GROUP_CHANGE_EVENT = "change"
)

type ClusterGroupEventRequest struct {
	GroupID string `json:"groupid"`
	Event   string `json:"event"`
}

/*
ResolveClusterGroupEventRequest
Method:  POST
Route:   /v1/cluster/groups/event
GroupID: cluster groupid
Event:   event type
*/
func ResolveClusterGroupEventRequest(r *http.Request) (*ClusterGroupEventRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &ClusterGroupEventRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	request.Event = strings.ToLower(request.Event)
	request.Event = strings.TrimSpace(request.Event)
	if request.Event != GROUP_CREATE_EVENT &&
		request.Event != GROUP_REMOVE_EVENT &&
		request.Event != GROUP_CHANGE_EVENT {
		return nil, fmt.Errorf("event type invalid.")
	}
	return request, nil
}

/*
ResolveClusterCreateContainerRequest
Method:  POST
Route:   /v1/cluster/containers
GroupID:   cluster groupid
Instances: create container instance count
Config:    container config
*/
type ClusterCreateContainerRequest struct {
	GroupID   string           `json:"groupid"`
	Instances int              `json:"instances"`
	Config    models.Container `json:"config"`
}

func ResolveClusterCreateContainerRequest(r *http.Request) (*ClusterCreateContainerRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &ClusterCreateContainerRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.GroupID)) == 0 {
		return nil, fmt.Errorf("create containers cluster groupid invalid, can not be empty")
	}

	if request.Instances < 0 {
		return nil, fmt.Errorf("create containers instances invalid, should be larger than 0")
	}

	if len(strings.TrimSpace(request.Config.Name)) == 0 {
		return nil, fmt.Errorf("create containers name can not be empty")
	}
	return request, nil
}

/*
ResolveClusterOperateContainersRequest
Method:  PUT
Route:   /v1/cluster/collections/action
MetaID:  containers metaid
Action:  operate action (start|stop|restart|kill|pause|unpause)
*/
type ClusterOperateContainersRequest struct {
	MetaID string `json:"metaid"`
	Action string `json:"action"`
}

func ResolveClusterOperateContainersRequest(r *http.Request) (*ClusterOperateContainersRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &ClusterOperateContainersRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.MetaID)) == 0 {
		return nil, fmt.Errorf("operate containers metaid invalid, can not be empty")
	}
	return request, nil
}

/*
ResolveClusterOperateContainerRequest
Method:  PUT
Route:   /v1/cluster/containers/action
ContainerID: operate container id
Action:  operate action (start|stop|restart|kill|pause|unpause)
*/
type ClusterOperateContainerRequest struct {
	ContainerID string `json:"containerid"`
	Action      string `json:"action"`
}

func ResolveClusterOperateContainerRequest(r *http.Request) (*ClusterOperateContainerRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &ClusterOperateContainerRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.ContainerID)) == 0 {
		return nil, fmt.Errorf("operate containers metaid invalid, can not be empty")
	}
	return request, nil
}

/*
ResolveClusterUpgradeContainerRequest
Method:  PUT
Route:   /v1/cluster/containers/upgrade
MetaID:    containers metaid
ImageTag:  upgrade new image tag
*/
type ClusterUpgradeContainerRequest struct {
	MetaID   string `json:"metaid"`
	ImageTag string `json:"imagetag"`
}

func ResolveClusterUpgradeContainerRequest(r *http.Request) (*ClusterUpgradeContainerRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &ClusterUpgradeContainerRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.MetaID)) == 0 {
		return nil, fmt.Errorf("upgrade containers metaid invalid, can not be empty")
	}
	return request, nil
}

/*
ResolveClusterRemoveContainersRequest
Method:  DELETE
Route:   /v1/cluster/collections/{metaid}
MetaID:  remove containers metaid
*/
type ClusterRemoveContainersRequest struct {
	MetaID string `json:"metaid"`
}

func ResolveClusterRemoveContainersRequest(r *http.Request) (*ClusterRemoveContainersRequest, error) {

	vars := mux.Vars(r)
	metaid := strings.TrimSpace(vars["metaid"])
	if len(strings.TrimSpace(metaid)) == 0 {
		return nil, fmt.Errorf("remove containers metaid invalid, can not be empty")
	}
	request := &ClusterRemoveContainersRequest{
		MetaID: metaid,
	}
	return request, nil
}

/*
ResolveClusterRemoveContainerRequest
Method:  DELETE
Route:   /v1/cluster/containers/{containerid}
ContainerID:  remove container id
*/
type ClusterRemoveContainerRequest struct {
	ContainerID string `json:"containerid"`
}

func ResolveClusterRemoveContainerRequest(r *http.Request) (*ClusterRemoveContainerRequest, error) {

	vars := mux.Vars(r)
	containerid := strings.TrimSpace(vars["containerid"])
	if len(strings.TrimSpace(containerid)) == 0 {
		return nil, fmt.Errorf("remove containers containerid invalid, can not be empty")
	}
	request := &ClusterRemoveContainerRequest{
		ContainerID: containerid,
	}
	return request, nil
}
