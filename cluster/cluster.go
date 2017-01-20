package cluster

import "github.com/humpback/discovery"
import "github.com/humpback/discovery/backends"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-center/models"

import (
	"fmt"
	"sync"
)

// Cluster is exported
type Cluster struct {
	sync.RWMutex
	discovery *discovery.Discovery
	engines   map[string]*Engine
	groups    map[string]*models.Group
	stopCh    chan struct{}
}

// NewCluster is exported
func NewCluster(discovery *discovery.Discovery) (*Cluster, error) {

	return &Cluster{
		discovery: discovery,
		engines:   make(map[string]*Engine),
		groups:    make(map[string]*models.Group),
		stopCh:    make(chan struct{}),
	}, nil
}

func (cluster *Cluster) Start() error {

	logger.INFO("[#cluster#] cluster discovery watching...")
	if cluster.discovery != nil {
		cluster.discovery.Watch(cluster.stopCh, cluster.watchHandleFunc)
		return nil
	}
	return fmt.Errorf("cluster discovery invalid.")
}

func (cluster *Cluster) Stop() {

	close(cluster.stopCh)
}

func (cluster *Cluster) watchHandleFunc(added backends.Entries, removed backends.Entries, err error) {

	if err != nil {
		logger.ERROR("[#cluster#] cluster discovery handlefunc error:%s", err.Error())
		return
	}

	//for _, entry := range removed {
	// set engine state to disconnected.
	//}

	//for _, entry := range added {
	// set engine state to pending.
	//}
}
