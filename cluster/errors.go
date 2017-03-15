package cluster

import "errors"

// cluster errors define
var (
	//cluster discovery is nil.
	ErrClusterDiscoveryInvalid = errors.New("cluster discovery invalid")
	//cluster meta not found
	ErrClusterMetaDataNotFound = errors.New("cluster metadata not found")
	//cluster group not found
	ErrClusterGroupNotFound = errors.New("cluster group not found")
	//cluster group no docker engine available
	ErrClusterNoEngineAvailable = errors.New("cluster no docker-engine available")
	//create container name conflict
	ErrClusterCreateContainerNameConflict = errors.New("cluster create container name conflict, this cluster already exists")
	//create container all failure
	ErrClusterCreateContainerFailure = errors.New("cluster create container failure")
	//cluster containers is upgrading
	ErrClusterContainersUpgrading = errors.New("cluster containers is upgrading")
)
