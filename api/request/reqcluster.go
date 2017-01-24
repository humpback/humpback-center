package request

import "github.com/gorilla/mux"

import (
	"fmt"
	"net/http"
	"strings"
)

/*
ClusterGroupRequest
Method:  GET
Route:   /v1/cluster/groups/{groupid}
Image:   remove image name.
*/
type ClusterGroupRequest struct {
	GroupID string `json:"groupid"`
}

func ResolveClusterGroupRequest(request *http.Request) (*ClusterGroupRequest, error) {

	vars := mux.Vars(request)
	groupid := strings.TrimSpace(vars["groupid"])
	if groupid == "" {
		return nil, fmt.Errorf("cluster groupid invalid")
	}

	return &ClusterGroupRequest{
		GroupID: groupid,
	}, nil
}
