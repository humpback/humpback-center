package request

import "github.com/gorilla/mux"
import "github.com/humpback/humpback-agent/models"
import "github.com/humpback/humpback-center/cluster/types"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

/*
GroupAllContainersRequest is exported
Method:  GET
Route:   /v1/groups/{groupid}/collections
*/
type GroupAllContainersRequest struct {
	GroupID string `json:"GroupId"`
}

// ResolveGroupAllContainersRequest is exported
func ResolveGroupAllContainersRequest(r *http.Request) (*GroupAllContainersRequest, error) {

	vars := mux.Vars(r)
	groupid := strings.TrimSpace(vars["groupid"])
	if len(strings.TrimSpace(groupid)) == 0 {
		return nil, fmt.Errorf("groupid invalid, can not be empty")
	}

	return &GroupAllContainersRequest{
		GroupID: groupid,
	}, nil
}

/*
GroupContainersRequest is exported
Method:  GET
Route:   /v1/groups/collections/{metaid}
*/
type GroupContainersRequest struct {
	MetaID string `json:"MetaId"`
}

// ResolveGroupContainersRequest is exported
func ResolveGroupContainersRequest(r *http.Request) (*GroupContainersRequest, error) {

	vars := mux.Vars(r)
	metaid := strings.TrimSpace(vars["metaid"])
	if len(strings.TrimSpace(metaid)) == 0 {
		return nil, fmt.Errorf("metaid invalid, can not be empty")
	}

	request := &GroupContainersRequest{
		MetaID: metaid,
	}
	return request, nil
}

/*
GroupContainersMetaBaseRequest is exported
Method:  GET
Route:   /v1/groups/collections/{metaid}/base
*/
type GroupContainersMetaBaseRequest struct {
	MetaID string `json:"MetaId"`
}

// ResolveGroupContainersMetaBaseRequest is exported
func ResolveGroupContainersMetaBaseRequest(r *http.Request) (*GroupContainersMetaBaseRequest, error) {

	vars := mux.Vars(r)
	metaid := strings.TrimSpace(vars["metaid"])
	if len(strings.TrimSpace(metaid)) == 0 {
		return nil, fmt.Errorf("metaid invalid, can not be empty")
	}

	request := &GroupContainersMetaBaseRequest{
		MetaID: metaid,
	}
	return request, nil
}

/*
GroupEnginesRequest is exported
Method:  GET
Route:   /v1/groups/{groupid}/engines
*/
type GroupEnginesRequest struct {
	GroupID string `json:"GroupId"`
}

// ResolveGroupEnginesRequest is exported
func ResolveGroupEnginesRequest(r *http.Request) (*GroupEnginesRequest, error) {

	vars := mux.Vars(r)
	groupid := strings.TrimSpace(vars["groupid"])
	if len(strings.TrimSpace(groupid)) == 0 {
		return nil, fmt.Errorf("groupid invalid, can not be empty")
	}

	return &GroupEnginesRequest{
		GroupID: groupid,
	}, nil
}

/*
GroupEngineRequest is exported
Method:  GET
Route:   /v1/groups/engines/{server}
*/
type GroupEngineRequest struct {
	Server string `json:"Server"`
}

// ResolveGroupEngineRequest is exported
func ResolveGroupEngineRequest(r *http.Request) (*GroupEngineRequest, error) {

	vars := mux.Vars(r)
	server := strings.TrimSpace(vars["server"])
	if len(strings.TrimSpace(server)) == 0 {
		return nil, fmt.Errorf("engine server invalid, can not be empty")
	}

	return &GroupEngineRequest{
		Server: server,
	}, nil
}

const (
	GROUP_CREATE_EVENT = "create"
	GROUP_REMOVE_EVENT = "remove"
	GROUP_CHANGE_EVENT = "change"
)

/*
GroupEventRequest is exported
Method:  POST
Route:   /v1/groups/event
*/
type GroupEventRequest struct {
	GroupID string `json:"GroupId"`
	Event   string `json:"Event"`
}

// ResolveGroupEventRequest is exported
func ResolveGroupEventRequest(r *http.Request) (*GroupEventRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &GroupEventRequest{}
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
GroupCreateContainersRequest is exported
Method:  POST
Route:   /v1/groups/collections
*/
type GroupCreateContainersRequest struct {
	GroupID   string           `json:"GroupId"`
	Instances int              `json:"Instances"`
	WebHooks  types.WebHooks   `json:"WebHooks"`
	Config    models.Container `json:"Config"`
}

// ResolveGroupCreateContainersRequest is exported
func ResolveGroupCreateContainersRequest(r *http.Request) (*GroupCreateContainersRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &GroupCreateContainersRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.GroupID)) == 0 {
		return nil, fmt.Errorf("create containers groupid invalid, can not be empty")
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
GroupUpdateContainersRequest is exported
Method:  PUT
Route:   /v1/groups/collections
*/
type GroupUpdateContainersRequest struct {
	MetaID    string         `json:"MetaId"`
	Instances int            `json:"Instances"`
	WebHooks  types.WebHooks `json:"WebHooks"`
}

// ResolveGroupUpdateContainersRequest is exported
func ResolveGroupUpdateContainersRequest(r *http.Request) (*GroupUpdateContainersRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &GroupUpdateContainersRequest{}
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
GroupOperateContainersRequest is exported
Method:  PUT
Route:   /v1/groups/collections/action
*/
type GroupOperateContainersRequest struct {
	MetaID string `json:"MetaId"`
	Action string `json:"Action"`
}

// ResolveGroupOperateContainersRequest is exported
func ResolveGroupOperateContainersRequest(r *http.Request) (*GroupOperateContainersRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &GroupOperateContainersRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.MetaID)) == 0 {
		return nil, fmt.Errorf("operate containers metaid invalid, can not be empty")
	}
	return request, nil
}

/*
GroupOperateContainerRequest is exported
Method:  PUT
Route:   /v1/groups/container/action
*/
type GroupOperateContainerRequest struct {
	ContainerID string `json:"ContainerId"`
	Action      string `json:"Action"`
}

// ResolveGroupOperateContainerRequest is exported
func ResolveGroupOperateContainerRequest(r *http.Request) (*GroupOperateContainerRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &GroupOperateContainerRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.ContainerID)) == 0 {
		return nil, fmt.Errorf("operate containers metaid invalid, can not be empty")
	}
	return request, nil
}

/*
GroupUpgradeContainersRequest is exported
Method:  PUT
Route:   /v1/groups/collections/upgrade
*/
type GroupUpgradeContainersRequest struct {
	MetaID   string `json:"MetaId"`
	ImageTag string `json:"ImageTag"`
}

// ResolveGroupUpgradeContainersRequest is exported
func ResolveGroupUpgradeContainersRequest(r *http.Request) (*GroupUpgradeContainersRequest, error) {

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	request := &GroupUpgradeContainersRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil, err
	}

	if len(strings.TrimSpace(request.MetaID)) == 0 {
		return nil, fmt.Errorf("upgrade containers metaid invalid, can not be empty")
	}
	return request, nil
}

/*
GroupRemoveContainersRequest is exported
Method:  DELETE
Route:   /v1/groups/collections/{metaid}
*/
type GroupRemoveContainersRequest struct {
	MetaID string `json:"MetaId"`
}

// ResolveGroupRemoveContainersRequest is exported
func ResolveGroupRemoveContainersRequest(r *http.Request) (*GroupRemoveContainersRequest, error) {

	vars := mux.Vars(r)
	metaid := strings.TrimSpace(vars["metaid"])
	if len(strings.TrimSpace(metaid)) == 0 {
		return nil, fmt.Errorf("remove containers metaid invalid, can not be empty")
	}
	request := &GroupRemoveContainersRequest{
		MetaID: metaid,
	}
	return request, nil
}

/*
GroupRemoveContainerRequest is exported
Method:  DELETE
Route:   /v1/groups/container/{containerid}
*/
type GroupRemoveContainerRequest struct {
	ContainerID string `json:"ContainerId"`
}

// ResolveGroupRemoveContainerRequest is exported
func ResolveGroupRemoveContainerRequest(r *http.Request) (*GroupRemoveContainerRequest, error) {

	vars := mux.Vars(r)
	containerid := strings.TrimSpace(vars["containerid"])
	if len(strings.TrimSpace(containerid)) == 0 {
		return nil, fmt.Errorf("remove containers containerid invalid, can not be empty")
	}

	request := &GroupRemoveContainerRequest{
		ContainerID: containerid,
	}
	return request, nil
}
