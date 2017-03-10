package cluster

import "errors"

// cluster errors define
var (
	//cluster discovery is nil.
	ErrClusterDiscoveryInvalid = errors.New("cluster discovery invalid")
	//cluster group notfound
	ErrClusterGroupNotFound = errors.New("cluster group not found")
	//cluster group no docker engine available
	ErrClusterNoEngineAvailable = errors.New("cluster no docker-engine available")
	//create container name conflict
	ErrClusterCreateContainerNameConflict = errors.New("cluster create container name conflict, this cluster already exists")
	//create container all failure
	ErrClusterCreateContainerFailure = errors.New("cluster create container failure")
)
