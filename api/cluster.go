package api

import "github.com/humpback/humpback-center/api/request"
import "github.com/humpback/humpback-center/api/response"
import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/gounits/logger"

import (
	"net/http"
)

func getClusterGroups(c *Context) error {

	logger.INFO("[#api#] %s resolve getgroups request successed.", c.ID)
	groups := c.Controller.GetClusterGroups()
	logger.INFO("[#api#] %s getgroups %d.", c.ID, len(groups))
	resp := response.NewClusterGroupsResponse(groups)
	result := &response.ResponseResult{ResponseID: c.ID}
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster groups response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getClusterGroup(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterGroupRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve getgroup request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve getgroup request successed. %+v", c.ID, req)
	group := c.Controller.GetClusterGroup(req.GroupID)
	if group == nil {
		logger.ERROR("[#api#] %s getgroup %s not found.", c.ID, req.GroupID)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "cluster group not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s getgroup %p.", c.ID, group)
	resp := response.NewClusterGroupResponse(group)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster group response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getClusterGroupEngines(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterGroupEnginesRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve getengines request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve getengines request successed. %+v", c.ID, req)
	engines := c.Controller.GetClusterGroupEngines(req.GroupID)
	if engines == nil {
		logger.ERROR("[#api#] %s getengines group %s not found.", c.ID, req.GroupID)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "cluster group not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s getengines %p.", c.ID, engines)
	resp := response.NewClusterGroupEnginesResponse(engines)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster group engines response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getClusterEngine(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterEngineRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve getengine request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve getengine request successed. %+v", c.ID, req)
	engine := c.Controller.GetClusterEngine(req.Server)
	if engine == nil {
		logger.ERROR("[#api#] %s getengine %s not found.", c.ID, req.Server)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "cluster engine not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s getengine %p.", c.ID, engine)
	resp := response.NewClusterEngineResponse(engine)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster engine response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func postClusterGroupEvent(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterGroupEventRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve groupevent request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve groupevent request successed. %+v", c.ID, req)
	c.Controller.SetClusterGroupEvent(req.GroupID, req.Event)
	resp := response.NewClusterGroupEventResponse("accepted.")
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster group event response")
	result.SetResponse(resp)
	return c.JSON(http.StatusAccepted, result)
}

func postClusterCreateContainer(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterCreateContainerRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve createcontainer request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve createcontainer request successed. %+v", c.ID, req)
	createdParis, err := c.Controller.CreateContainer(req.GroupID, req.Instances, req.Config)
	if err != nil {
		logger.ERROR("[#api#] %s createcontainer to group %s error: %s", c.ID, req.GroupID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterGroupNotFound {
			return c.JSON(http.StatusNotFound, result)
		} else if err == cluster.ErrClusterCreateContainerNameConflict {
			return c.JSON(http.StatusConflict, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewClusterCreateContainerResponse(req.GroupID, req.Instances, createdParis)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster createcontainer response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}
