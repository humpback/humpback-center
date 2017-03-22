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
ResolveClusterGroupContainersRequest
Method:  GET
Route:   /v1/cluster/groups/{groupid}/containers
GroupID: cluster groupid
*/
type ClusterGroupContainersRequest struct {
	GroupID string `json:"GroupId"`
}

func ResolveClusterGroupContainersRequest(r *http.Request) (*ClusterGroupContainersRequest, error) {

	vars := mux.Vars(r)
	groupid := strings.TrimSpace(vars["groupid"])
	if len(strings.TrimSpace(groupid)) == 0 {
		return nil, fmt.Errorf("groupid invalid, can not be empty")
	}

	return &ClusterGroupContainersRequest{
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
	GroupID string `json:"GroupId"`
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
	Server string `json:"Server"`
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
	GroupID string `json:"GroupId"`
	Event   string `json:"Event"`
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
ResolveClusterCreateContainersRequest
Method:  POST
Route:   /v1/cluster/containers
GroupID:   cluster groupid
Instances: create container instance count
WebHook:   web hook site url
Config:    container config
*/
type ClusterCreateContainersRequest struct {
	GroupID   string           `json:"GroupId"`
	Instances int              `json:"Instances"`
	WebHook   string           `json:"WebHook"`
	Config    models.Container `json:"Config"`
}

func ResolveClusterCreateContainersRequest(r *http.Request) (*ClusterCreateContainersRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &ClusterCreateContainersRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.GroupID)) == 0 {
		return nil, fmt.Errorf("create containers cluster groupid invalid, can not be empty")
	}

	if request.Instances <= 0 {
		return nil, fmt.Errorf("create containers instances invalid, should be larger than 0")
	}

	if len(strings.TrimSpace(request.Config.Name)) == 0 {
		return nil, fmt.Errorf("create containers name can not be empty")
	}
	return request, nil
}

/*
ResolveClusterUpdateContainersRequest
Method:  PUT
Route:   /v1/cluster/containers
MetaID:   containers metaid
Instances: update containers instance count
WebHook:   web hook site url
*/
type ClusterUpdateContainersRequest struct {
	MetaID    string `json:"MetaId"`
	Instances int    `json:"Instances"`
	WebHook   string `json:"WebHook"`
}

func ResolveClusterUpdateContainersRequest(r *http.Request) (*ClusterUpdateContainersRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &ClusterUpdateContainersRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.MetaID)) == 0 {
		return nil, fmt.Errorf("set containers metaid invalid, can not be empty")
	}

	if request.Instances <= 0 {
		return nil, fmt.Errorf("set containers instances invalid, should be larger than 0")
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
	MetaID string `json:"MetaId"`
	Action string `json:"Action"`
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
	ContainerID string `json:"ContainerId"`
	Action      string `json:"Action"`
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
ResolveClusterUpgradeContainersRequest
Method:  PUT
Route:   /v1/cluster/containers/upgrade
MetaID:    containers metaid
ImageTag:  upgrade new image tag
*/
type ClusterUpgradeContainersRequest struct {
	MetaID   string `json:"MetaId"`
	ImageTag string `json:"ImageTag"`
}

func ResolveClusterUpgradeContainersRequest(r *http.Request) (*ClusterUpgradeContainersRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &ClusterUpgradeContainersRequest{}
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
	MetaID string `json:"MetaId"`
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
	ContainerID string `json:"ContainerId"`
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
