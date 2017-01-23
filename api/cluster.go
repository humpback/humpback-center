package api

import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-center/api/request"
import "github.com/humpback/humpback-center/api/response"

import (
	"net/http"
)

func getClusterGroups(c *Context) {

	logger.INFO("[#api#] getClusterGroups id:%s request resolve successed.", c.ID)
	result := &response.ResponseResult{ResponseID: c.ID}
	groups := c.Cluster.GetGroups()
	response := response.NewClusterGroupsResponse(groups)
	result.SetError(request.RequestSuccessed, request.ErrRequestSuccessed, "cluster groups")
	result.SetResponse(response)
	c.JSON(http.StatusOK, result)
}
