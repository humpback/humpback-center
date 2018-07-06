package api

import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-center/api/request"
import "github.com/humpback/humpback-center/api/response"
import "github.com/humpback/humpback-center/cluster"

import (
	"net/http"
)

func getGroupAllContainers(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupAllContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve get group all containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve get group all containers request successed. %+v", c.ID, req)
	groupContainers := c.Controller.GetClusterGroupAllContainers(req.GroupID)
	if groupContainers == nil {
		logger.ERROR("[#api#] %s get group all containers %s not found.", c.ID, req.GroupID)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "group not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s getr group all containers %p.", c.ID, groupContainers)
	resp := response.NewGroupAllContainersResponse(req.GroupID, groupContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "group containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getGroupContainers(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve group containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve get group containers request successed. %+v", c.ID, req)
	groupContainer := c.Controller.GetClusterGroupContainers(req.MetaID)
	if groupContainer == nil {
		logger.ERROR("[#api#] %s get containers meta %s not found.", c.ID, req.MetaID)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "meta not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s get group containers %p.", c.ID, groupContainer)
	resp := response.NewGroupContainersResponse(groupContainer)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "group containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getGroupContainersMetaBase(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupContainersMetaBaseRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve group containers meta request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve get group containers meta request successed. %+v", c.ID, req)
	metaBase := c.Controller.GetClusterGroupContainersMetaBase(req.MetaID)
	if metaBase == nil {
		logger.ERROR("[#api#] %s get containers meta %s not found.", c.ID, req.MetaID)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "meta not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s get group meta containers %p.", c.ID, metaBase)
	resp := response.NewGroupContainersMetaBaseResponse(metaBase)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "group containers meta response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getGroupEngines(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupEnginesRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve get group engines request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve get group engines request successed. %+v", c.ID, req)
	engines := c.Controller.GetClusterGroupAllEngines(req.GroupID)
	if engines == nil {
		logger.ERROR("[#api#] %s get group engines group %s not found.", c.ID, req.GroupID)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "group not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s get group engines %p.", c.ID, engines)
	resp := response.NewGroupEnginesResponse(req.GroupID, engines)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "group engines response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func getGroupEngine(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupEngineRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve get engine request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve get engine request successed. %+v", c.ID, req)
	engine := c.Controller.GetClusterEngine(req.Server)
	if engine == nil {
		logger.ERROR("[#api#] %s get  engine %s not found.", c.ID, req.Server)
		result.SetError(request.RequestFailure, request.ErrRequestFailure, "engine not found")
		return c.JSON(http.StatusNotFound, result)
	}

	logger.INFO("[#api#] %s get engine %p.", c.ID, engine)
	resp := response.NewGroupEngineResponse(engine)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "engine response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func postGroupEvent(c *Context) error {

	result := &response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupEventRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve group event request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve group event request successed. %+v", c.ID, req)
	c.Controller.SetClusterGroupEvent(req.GroupID, req.Event)
	resp := response.NewGroupEventResponse("accepted.")
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "group event response")
	result.SetResponse(resp)
	return c.JSON(http.StatusAccepted, result)
}

func postGroupCreateContainers(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupCreateContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve create containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve create containers request successed. %+v", c.ID, req)
	metaid, createdContainers, err := c.Controller.CreateClusterContainers(req.GroupID, req.Instances, req.WebHooks, req.Config, req.Option)
	if err != nil {
		logger.ERROR("[#api#] %s create containers to group %s error: %s", c.ID, req.GroupID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterGroupNotFound {
			return c.JSON(http.StatusNotFound, result)
		} else if err == cluster.ErrClusterCreateContainerNameConflict {
			return c.JSON(http.StatusConflict, result)
		} else if err == cluster.ErrClusterCreateContainerTagAlreadyUsing {
			return c.JSON(http.StatusConflict, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewGroupCreateContainersResponse(req.GroupID, metaid, req.Instances, createdContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "create containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func putGroupUpdateContainers(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupUpdateContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve update containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve update containers request successed. %+v", c.ID, req)
	updatedContainers, err := c.Controller.UpdateClusterContainers(req.MetaID, req.Instances, req.WebHooks, req.Config)
	if err != nil {
		logger.ERROR("[#api#] %s update containers to meta %s error: %s", c.ID, req.MetaID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterMetaDataNotFound {
			return c.JSON(http.StatusNotFound, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewGroupUpdateContainersResponse(req.MetaID, req.Instances, updatedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "update containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func putGroupOperateContainers(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupOperateContainersRequest(c.Request())
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

	resp := response.NewGroupOperateContainersResponse(req.MetaID, req.Action, operatedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "operate containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func putGroupOperateContainer(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupOperateContainerRequest(c.Request())
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

	resp := response.NewGroupOperateContainersResponse(metaID, req.Action, operatedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "operate container response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func putGroupUpgradeContainers(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupUpgradeContainersRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve upgrade containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve upgrade containers request successed. %+v", c.ID, req)
	upgradeContainers, err := c.Controller.UpgradeContainers(req.MetaID, req.ImageTag)
	if err != nil {
		logger.ERROR("[#api#] %s upgrade containers to meta %s error: %s", c.ID, req.MetaID, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterMetaDataNotFound || err == cluster.ErrClusterGroupNotFound {
			return c.JSON(http.StatusNotFound, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewGroupUpgradeContainersResponse(req.MetaID, "upgrade containers", upgradeContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "upgrade containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func deleteGroupRemoveContainersOfMetaName(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupRemoveContainersOfMetaNameRequest(c.Request())
	if err != nil {
		logger.ERROR("[#api#] %s resolve remove containers request faild, %s", c.ID, err.Error())
		result.SetError(request.RequestInvalid, request.ErrRequestInvalid, err.Error())
		return c.JSON(http.StatusBadRequest, result)
	}

	logger.INFO("[#api#] %s resolve remove containers request successed. %+v", c.ID, req)
	metaid, removedContainers, err := c.Controller.RemoveContainersOfMetaName(req.GroupID, req.MetaName)
	if err != nil {
		logger.ERROR("[#api#] %s remove containers to meta %s %s error: %s", c.ID, req.GroupID, req.MetaName, err.Error())
		result.SetError(request.RequestFailure, request.ErrRequestFailure, err.Error())
		if err == cluster.ErrClusterMetaDataNotFound || err == cluster.ErrClusterGroupNotFound {
			return c.JSON(http.StatusNotFound, result)
		}
		return c.JSON(http.StatusInternalServerError, result)
	}

	resp := response.NewGroupRemoveContainersResponse(metaid, removedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "remove containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func deleteGroupRemoveContainers(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupRemoveContainersRequest(c.Request())
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

	resp := response.NewGroupRemoveContainersResponse(req.MetaID, removedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "remove containers response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}

func deleteGroupRemoveContainer(c *Context) error {

	result := response.ResponseResult{ResponseID: c.ID}
	req, err := request.ResolveGroupRemoveContainerRequest(c.Request())
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

	resp := response.NewGroupRemoveContainersResponse(metaID, removedContainers)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "remove container response")
	result.SetResponse(resp)
	return c.JSON(http.StatusOK, result)
}
