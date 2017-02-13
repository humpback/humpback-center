package cluster

import "github.com/humpback/discovery"
import "github.com/humpback/discovery/backends"
import "github.com/humpback/gounits/json"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-center/cluster/types"

import (
	"errors"
	"sync"
)

// Cluster errors define
var (
	//ErrClusterDiscoveryInvalid, discovery is nil.
	ErrClusterDiscoveryInvalid = errors.New("cluster discovery invalid.")
)

// Group is exported
// Servers: cluster server ips, correspond engines's key.
type Group struct {
	ID      string
	Servers []string
}

// Cluster is exported
// engines: map[ip]*Engine
// groups:  map[groupid]*Group
type Cluster struct {
	sync.RWMutex
	Discovery *discovery.Discovery
	engines   map[string]*Engine
	groups    map[string]*Group
	stopCh    chan struct{}
}

// NewCluster is exported
func NewCluster(discovery *discovery.Discovery) (*Cluster, error) {

	if discovery == nil {
		return nil, ErrClusterDiscoveryInvalid
	}

	return &Cluster{
		Discovery: discovery,
		engines:   make(map[string]*Engine),
		groups:    make(map[string]*Group),
		stopCh:    make(chan struct{}),
	}, nil
}

func (cluster *Cluster) Start() error {

	logger.INFO("[#cluster#] cluster discovery watching...")
	if cluster.Discovery != nil {
		cluster.Discovery.Watch(cluster.stopCh, cluster.watchHandleFunc)
		return nil
	}
	return ErrClusterDiscoveryInvalid
}

func (cluster *Cluster) Stop() {

	close(cluster.stopCh)
	logger.INFO("[#cluster#] cluster discovery closed.")
}

func (cluster *Cluster) GetEngine(ip string) *Engine {

	cluster.RLock()
	defer cluster.RUnlock()
	if engine, ret := cluster.engines[ip]; ret {
		return engine
	}
	return nil
}

func (cluster *Cluster) GetGroups() []*Group {

	groups := []*Group{}
	cluster.RLock()
	for _, group := range cluster.groups {
		groups = append(groups, group)
	}
	cluster.RUnlock()
	return groups
}

func (cluster *Cluster) GetGroup(groupid string) *Group {

	cluster.RLock()
	defer cluster.RUnlock()
	if group, ret := cluster.groups[groupid]; ret {
		return group
	}
	return nil
}

func (cluster *Cluster) SetGroup(groupid string, servers []string) {

	cluster.Lock()
	group, ret := cluster.groups[groupid]
	if !ret {
		group = &Group{ID: groupid, Servers: servers}
		cluster.groups[groupid] = group
		logger.INFO("[#cluster#] cluster create group %s(%d)", groupid, len(servers))
	} else {
		group.Servers = servers
		logger.INFO("[#cluster#] cluster set group %s(%d)", groupid, len(servers))
	}

	for _, server := range group.Servers {
		if _, ret := cluster.engines[server]; !ret {
			engine, err := NewEngine(server)
			if err != nil {
				logger.ERROR("[#cluster#] cluster add engine %s error:%s", server, err.Error())
				continue
			}
			cluster.engines[server] = engine
			logger.INFO("[#cluster#] cluster add engine %p:%s", engine, engine.IP)
		}
	}
	cluster.Unlock()
}

func (cluster *Cluster) RemoveGroup(groupid string) bool {

	cluster.Lock()
	defer cluster.Unlock()
	group, ret := cluster.groups[groupid]
	if !ret {
		logger.WARN("[#cluster#] cluster remove group %s not found.", groupid)
		return false
	}
	logger.INFO("[#cluster#] cluster remove group %s(%d)", groupid, len(group.Servers))
	delete(cluster.groups, groupid)
	return true
}

func (cluster *Cluster) watchHandleFunc(added backends.Entries, removed backends.Entries, err error) {

	if err != nil {
		logger.ERROR("[#cluster#] cluster discovery handlefunc error:%s", err.Error())
		return
	}

	opts := &types.ClusterRegistOptions{}
	for _, entry := range removed {
		if err := json.DeCodeBufferToObject(entry.Data, opts); err != nil {
			logger.ERROR("[#cluster#] cluster discovery handlefunc error: removed, %s", err.Error())
			continue
		}
		cluster.removeEngine(opts)
	}

	for _, entry := range added {
		if err := json.DeCodeBufferToObject(entry.Data, opts); err != nil {
			logger.ERROR("[#cluster#] cluster discovery handlefunc error: added, %s", err.Error())
			continue
		}
		cluster.addEngine(opts)
	}
}

func (cluster *Cluster) addEngine(opts *types.ClusterRegistOptions) bool {

	cluster.Lock()
	defer cluster.Unlock()
	engine, ret := cluster.engines[opts.IP]
	if !ret {
		var err error
		if engine, err = NewEngine(opts.IP); err != nil {
			logger.ERROR("[#cluster#] cluster add engine %s error:%s\n", opts.IP, err.Error())
			return false
		}
		cluster.engines[opts.IP] = engine
		logger.INFO("[#cluster#] cluster add engine %p:%s\n", engine, opts.IP)
	}
	engine.SetRegistOptions(opts)
	engine.SetState(stateHealthy)
	return true
}

func (cluster *Cluster) removeEngine(opts *types.ClusterRegistOptions) bool {

	engine := cluster.GetEngine(opts.IP)
	if engine == nil {
		logger.WARN("[#cluster#] cluster remove engine not found:%s\n", opts.IP)
		return false
	}

	engine.Close() //close engine
	found := false
	groups := cluster.GetGroups()
	for _, group := range groups {
		for _, server := range group.Servers {
			if engine.IP == server {
				found = true
				cluster.Lock()
				engine.SetState(stateUnhealthy)
				cluster.Unlock()
				logger.INFO("[#cluster#] cluster engine state %p:%s %s\n", engine, opts.IP, engine.Status())
				break
			}
		}
	}

	if !found {
		cluster.Lock()
		delete(cluster.engines, engine.IP)
		cluster.Unlock()
		logger.INFO("[#cluster#] cluster remove engine %p:%s\n", engine, opts.IP)
	}
	return true
}
