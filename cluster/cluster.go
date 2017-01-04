package cluster

import "github.com/humpback/discovery"
import "github.com/humpback/discovery/backends"
import "github.com/humpback/gounits/logger"

import "sync"
import "fmt"

// Cluster is exported
type Cluster struct {
	sync.RWMutex
	discovery *discovery.Discovery
	groups    map[string]*Group
}

func NewCluster(d *discovery.Discovery) (*Cluster, error) {

	cluster := &Cluster{
		discovery: d,
		groups:    make(map[string]*Group),
	}
	cluster.watch()
	return nil, nil
}

func (c *Cluster) AddGroup(id string, name string) (*Group, error) {

	return nil, nil
}

func (c *Cluster) RemoveGroup(id string) bool {

	return false
}

func (c *Cluster) watch() error {

	logger.INFO("[#cluster#] cluster discovery watching...")
	if c.discovery != nil {
		c.discovery.Watch(nil, c.watchHandleFunc)
		return nil
	}
	return fmt.Errorf("cluster discovery invalid.")
}

func (c *Cluster) watchHandleFunc(added backends.Entries, removed backends.Entries, err error) {

	if err != nil {
		logger.INFO("[#cluster#] cluster discovery handler error:%s", err.Error())
		return
	}

	//for _, entry := range removed {
	// set engine state to disconnected.
	//}

	//for _, entry := range added {
	// set engine state to pending.
	//}
}
