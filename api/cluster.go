package api

import "github.com/humpback/humpback-center/api/request"
import "github.com/humpback/humpback-center/api/response"
import "github.com/humpback/gounits/logger"

import (
	"net/http"
)

func getClusterGroups(c *Context) error {

	logger.INFO("[#api#] %s resolve get clustergroups request successed.", c.ID)
	groups := c.Controller.GetClusterGroups()
	logger.INFO("[#api#] %s get clustergroups %d.", c.ID, len(groups))
	resp := response.NewClusterGroupsResponse(groups)
	result := &response.ResponseResult{ResponseID: c.ID}
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "response cluster groups")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getClusterGroup(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterGroupRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve get clustergroup request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve get clustergroup request successed. %+v", c.ID, req)
	group := c.Controller.GetClusterGroup(req.GroupID)
	if group == nil {
		logger.ERROR("[#api#] %s get clustergroup %s not found.", c.ID, req.GroupID)
		result.SetError(request.RequestNotFound, request.ErrRequestNotFound, req.GroupID+" not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s get clustergroup %p.", c.ID, group)
	resp := response.NewClusterGroupResponse(group)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "response cluster group")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getClusterEngine(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterEngineRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve get clusterengine request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve get clusterengine request successed. %+v", c.ID, req)
	engine := c.Controller.GetClusterEngine(req.EngineID)
	if engine == nil {
		logger.ERROR("[#api#] %s get clusterengine %s not found.", c.ID, req.EngineID)
		result.SetError(request.RequestNotFound, request.ErrRequestNotFound, req.EngineID+" not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s get clusterengine %p.", c.ID, engine)
	resp := response.NewClusterEngineResponse(engine)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "response cluster engine")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func postClusterGroupEvent(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterGroupEventRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve post clustergroupevent request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve post clustergroupevent request successed. %+v", c.ID, req)
	c.Controller.SetClusterGroupEvent(req.GroupID, req.Event)
	resp := response.NewClusterGroupEventResponse("accepted.")
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "response cluster group event")
	result.SetResponse(resp)
	return c.JSON(http.StatusAccepted, result)
}
