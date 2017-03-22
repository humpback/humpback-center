package api

import "github.com/humpback/humpback-center/api/request"
import "github.com/humpback/humpback-center/api/response"
import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/gounits/logger"

import (
	"net/http"
)

func getClusterGroupContainers(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterGroupContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve get cluster group containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve get cluster group containers request successed. %+v", c.ID, req)
	groupContainers := c.Controller.GetClusterGroupContainers(req.GroupID)
	if groupContainers == nil {
		logger.ERROR("[#api#] %s get cluster group containers %s not found.", c.ID, req.GroupID)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "cluster group not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s get cluster group containers %p.", c.ID, groupContainers)
	resp := response.NewClusterGroupContainersResponse(req.GroupID, groupContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster group containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getClusterGroupEngines(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterGroupEnginesRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve get cluster group engines request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve get cluster group engines request successed. %+v", c.ID, req)
	engines := c.Controller.GetClusterGroupEngines(req.GroupID)
	if engines == nil {
		logger.ERROR("[#api#] %s get cluster group engines group %s not found.", c.ID, req.GroupID)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "cluster group not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s get cluster group engines %p.", c.ID, engines)
	resp := response.NewClusterGroupEnginesResponse(req.GroupID, engines)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster group engines response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getClusterEngine(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterEngineRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve get engine request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve get engine request successed. %+v", c.ID, req)
	engine := c.Controller.GetClusterEngine(req.Server)
	if engine == nil {
		logger.ERROR("[#api#] %s get  engine %s not found.", c.ID, req.Server)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "cluster engine not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s get engine %p.", c.ID, engine)
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

func postClusterCreateContainers(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterCreateContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve create containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve create containers request successed. %+v", c.ID, req)
	metaid, createdContainers, err := c.Controller.CreateClusterContainers(req.GroupID, req.Instances, req.WebHook, req.Config)
	if err != nil {
		logger.ERROR("[#api#] %s create containers to group %s error: %s", c.ID, req.GroupID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterGroupNotFound {
			return c.JSON(http.StatusNotFound, result)
		} else if err == cluster.ErrClusterCreateContainerNameConflict {
			return c.JSON(http.StatusConflict, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewClusterCreateContainersResponse(req.GroupID, metaid, req.Instances, createdContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster create containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func putClusterUpdateContainers(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterUpdateContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve update containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve update containers request successed. %+v", c.ID, req)
	updatedContainers, err := c.Controller.UpdateClusterContainers(req.MetaID, req.Instances, req.WebHook)
	if err != nil {
		logger.ERROR("[#api#] %s update containers to meta %s error: %s", c.ID, req.MetaID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterMetaDataNotFound {
			return c.JSON(http.StatusNotFound, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewClusterUpdateContainersResponse(req.MetaID, req.Instances, updatedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster update containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func putClusterOperateContainers(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterOperateContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve operate containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve operate containers request successed. %+v", c.ID, req)
	operatedContainers, err := c.Controller.OperateContainers(req.MetaID, req.Action)
	if err != nil {
		logger.ERROR("[#api#] %s operate %s containers to meta %s error: %s", c.ID, req.Action, req.MetaID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterMetaDataNotFound || err == cluster.ErrClusterGroupNotFound {
			return c.JSON(http.StatusNotFound, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewClusterOperateContainersResponse(req.MetaID, req.Action, operatedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster operate containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func putClusterOperateContainer(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterOperateContainerRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve operate container request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve operate container request successed. %+v", c.ID, req)
	metaID, operatedContainers, err := c.Controller.OperateContainer(req.ContainerID, req.Action)
	if err != nil {
		logger.ERROR("[#api#] %s operate %s container to %s error: %s", c.ID, req.Action, req.ContainerID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterMetaDataNotFound || err == cluster.ErrClusterGroupNotFound {
			return c.JSON(http.StatusNotFound, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewClusterOperateContainersResponse(metaID, req.Action, operatedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster operate container response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func putClusterUpgradeContainers(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterUpgradeContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve upgrade containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve upgrade containers request successed. %+v", c.ID, req)
	if err := c.Controller.UpgradeContainers(req.MetaID, req.ImageTag); err != nil {
		logger.ERROR("[#api#] %s upgrade containers to meta %s error: %s", c.ID, req.MetaID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterMetaDataNotFound || err == cluster.ErrClusterGroupNotFound {
			return c.JSON(http.StatusNotFound, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewClusterUpgradeContainersResponse(req.MetaID, "upgrade containers accepted")
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster upgrade containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusAccepted, result)
}

func deleteClusterRemoveContainers(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterRemoveContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve remove containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve remove containers request successed. %+v", c.ID, req)
	removedContainers, err := c.Controller.RemoveContainers(req.MetaID)
	if err != nil {
		logger.ERROR("[#api#] %s remove containers to meta %s error: %s", c.ID, req.MetaID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterMetaDataNotFound || err == cluster.ErrClusterGroupNotFound {
			return c.JSON(http.StatusNotFound, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewClusterRemoveContainersResponse(req.MetaID, removedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster remove containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func deleteClusterRemoveContainer(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveClusterRemoveContainerRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve remove container request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve remove container request successed. %+v", c.ID, req)
	metaID, removedContainers, err := c.Controller.RemoveContainer(req.ContainerID)
	if err != nil {
		logger.ERROR("[#api#] %s remove container to %s error: %s", c.ID, req.ContainerID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterMetaDataNotFound || err == cluster.ErrClusterGroupNotFound || err == cluster.ErrClusterContainerNotFound {
			return c.JSON(http.StatusNotFound, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewClusterRemoveContainersResponse(metaID, removedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster remove container response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}
