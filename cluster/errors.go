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
	//cluster container not found
	ErrClusterContainerNotFound = errors.New("cluster container not found")
	//cluster group no docker engine available
	ErrClusterNoEngineAvailable = errors.New("cluster no docker-engine available")
	//cluster containers instances invalid.
	ErrClusterContainersInstancesInvalid = errors.New("cluster containers instances invalid")
	//cluster create containers name conflict
	ErrClusterCreateContainerNameConflict = errors.New("cluster create containers name conflict, this cluster already exists")
	//cluster create containers all failure
	ErrClusterCreateContainerFailure = errors.New("cluster create containers failure")
	//cluster containers is upgrading
	ErrClusterContainersUpgrading = errors.New("cluster containers state is upgrading")
	//cluster containers is migrating
	ErrClusterContainersMigrating = errors.New("cluster containers state is migrating")
	//cluster containers is setting
	ErrClusterContainersSetting = errors.New("cluster containers state is setting")
	//cluster containers instances no change
	ErrClusterContainersInstancesNoChange = errors.New("cluster containers instances no change")
)
