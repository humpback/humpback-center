package cluster

import "github.com/humpback/discovery"
import "github.com/humpback/discovery/backends"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/gounits/json"

import (
	"fmt"
	"sync"
)

// Group is exported
// Servers map:[ip]engineid, value is engineid.
type Group struct {
	ID      string
	Servers map[string]string
}

// Cluster is exported
// engines: map[engineid]*Engine
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
	return fmt.Errorf("cluster discovery watch fail.")
}

func (cluster *Cluster) Stop() {

	close(cluster.stopCh)
	logger.INFO("[#cluster#] cluster discovery closed.")
}

func (cluster *Cluster) GetGroups() []*Group {

	cluster.RLock()
	groups := []*Group{}
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

func (cluster *Cluster) GetEngine(engineid string) *Engine {

	cluster.RLock()
	defer cluster.RUnlock()
	if engine, ret := cluster.engines[engineid]; ret {
		return engine
	}
	return nil
}

func (cluster *Cluster) GetEngineByAddr(addr string) *Engine {

	cluster.RLock()
	defer cluster.RUnlock()
	for _, engine := range cluster.engines {
		if engine.Addr == addr {
			return engine
		}
	}
	return nil
}

func (cluster *Cluster) GetEngineByIP(ip string) *Engine {

	cluster.RLock()
	defer cluster.RUnlock()
	for _, engine := range cluster.engines {
		if engine.IP == ip {
			return engine
		}
	}
	return nil
}

func (cluster *Cluster) SetGroup(groupid string, servers []string) {

	cluster.Lock()
	var group *Group
	if _, ret := cluster.groups[groupid]; !ret {
		group = &Group{ID: groupid}
		cluster.groups[groupid] = group
	} else {
		group = cluster.groups[groupid]
	}

	//reset group servers, unbind all engine.
	group.Servers = make(map[string]string)
	for _, server := range servers {
		group.Servers[server] = "" //unbind
	}
	logger.INFO("[#cluster#] cluster set group %s(%d)", groupid, len(servers))
	cluster.Unlock()

	//bind engine to group.
	for _, server := range servers {
		if engine := cluster.GetEngineByIP(server); engine != nil {
			logger.INFO("[#cluster#] cluster engine %s %s bind to group %s > %s", engine.ID, engine.Addr, group.ID, server)
			cluster.Lock()
			group.Servers[server] = engine.ID //bind
			cluster.Unlock()
		}
	}
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

func (cluster *Cluster) addEngine(engineid string, opts *RegistClusterOptions) bool {

	if engine := cluster.GetEngine(engineid); engine != nil {
		logger.WARN("[#cluster#] cluster add engine %s conflict, engine %s is already.", engineid, opts.Addr)
		return false
	}

	engine, err := NewEngine(engineid, opts) //初始状态为unhealthy，engine一次refreshLoop成功后更改状态
	if err != nil {
		logger.ERROR("[#cluster#] cluster add engine %s error:%s %s.", engineid, opts.Addr, err.Error())
		return false
	}

	logger.INFO("[#cluster#] cluster add engine %s > %s", engineid, opts.Addr)
	cluster.Lock()
	cluster.engines[engineid] = engine
	//bind engine to groups.
	for _, group := range cluster.groups {
		if _, ret := group.Servers[engine.IP]; ret {
			group.Servers[engine.IP] = engineid //bind
			logger.INFO("[#cluster#] cluster engine %s %s bind to group %s > %s", engineid, opts.Addr, group.ID, engine.IP)
		}
	}
	cluster.Unlock()
	return true
}

func (cluster *Cluster) removeEngine(engineid string, opts *RegistClusterOptions) bool {

	engine := cluster.GetEngine(engineid)
	if engine == nil {
		logger.WARN("[#cluster#] cluster remove engine %s invalid, engine %s %s", engineid, opts.Addr, "not found.")
		return false
	}

	logger.INFO("[#cluster#] cluster remove engine %s > %s", engineid, opts.Addr)
	cluster.Lock()
	delete(cluster.engines, engineid)
	//unbind engine to groups.
	for _, group := range cluster.groups {
		if _, ret := group.Servers[engine.IP]; ret {
			group.Servers[engine.IP] = "" //unbind
			logger.INFO("[#cluster#] cluster engine %s %s unbind to group %s > %s", engineid, opts.Addr, group.ID, engine.IP)
		}
	}
	engine.Close()
	cluster.Unlock()
	return true
}

func (cluster *Cluster) watchHandleFunc(added backends.Entries, removed backends.Entries, err error) {

	if err != nil {
		logger.ERROR("[#cluster#] cluster discovery handlefunc error:%s", err.Error())
		return
	}

	opts := &RegistClusterOptions{}
	for _, entry := range removed {
		if err := json.DeCodeBufferToObject(entry.Data, opts); err != nil {
			logger.ERROR("[#cluster#] cluster discovery handlefunc error: removed, %s", err.Error())
			continue
		}
		cluster.removeEngine(entry.Key, opts)
	}

	for _, entry := range added {
		if err := json.DeCodeBufferToObject(entry.Data, opts); err != nil {
			logger.ERROR("[#cluster#] cluster discovery handlefunc error: added, %s", err.Error())
			continue
		}
		cluster.addEngine(entry.Key, opts)
	}
}
