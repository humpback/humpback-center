package response

import "github.com/humpback/humpback-center/models"

type Server struct {
	Key    string `json:"key"`
	IPAddr string `json:"ipaddr"`
	Status int    `status`
	//...
}

type Group struct {
	ID      string   `json:"id"`
	Servers []Server `json:"servers"`
}

/*
Method:  GET
Router:  /v1/cluster/groups
*/
type ClusterGroupsResponse struct {
	Groups []Group `json:"groups"`
}

func NewClusterGroupsResponse(groups []*models.Group) *ClusterGroupsResponse {

	response := &ClusterGroupsResponse{
		Groups: make([]Group, 0),
	}
	return response
}
