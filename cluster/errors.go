package cluster

import "errors"

// cluster errors define
var (
	//cluster discovery is nil.
	ErrClusterDiscoveryInvalid = errors.New("cluster discovery invalid")
	//cluster group notfound
	ErrClusterGroupNotFound = errors.New("cluster group not found")
	//create container name conflict
	ErrClusterCreateContainerNameConflict = errors.New("create container name conflict, this name already exists")
)
