package cluster

import "github.com/humpback/discovery"
import "github.com/humpback/discovery/backends"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/gounits/json"

import (
	"fmt"
	"sync"
)

// Cluster is exported
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
	return fmt.Errorf("cluster discovery watch invalid.")
}

func (cluster *Cluster) Stop() {

	close(cluster.stopCh)
	logger.INFO("[#cluster#] cluster discovery closed.")
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

func (cluster *Cluster) addEngine(id string, opts *RegistClusterOptions) bool {

	if engine := cluster.GetEngineByAddr(opts.Addr); engine != nil {
		logger.WARN("cluster add engine %s, engine %s %s", id, opts.Addr, "is already.")
		return false
	}

	engine, err := NewEngine(id, opts.Addr) //初始状态为unhealthy，engine一次refreshLoop成功后更改状态
	if err != nil {
		logger.ERROR("cluster add engine %s error:%s %s", id, opts.Addr, err.Error())
		return false
	}

	cluster.Lock()
	cluster.engines[id] = engine
	for _, group := range cluster.groups {
		if _, ok := group.Servers[engine.IP]; ok {
			group.Servers[engine.IP] = id
		}
	}
	cluster.Unlock()
	return true
}

func (cluster *Cluster) removeEngine(id string, opts *RegistClusterOptions) bool {

	if engine := cluster.GetEngineByAddr(opts.Addr); engine != nil {
		cluster.Lock()
		for _, group := range cluster.groups {
			if _, ok := group.Servers[engine.IP]; ok {
				group.Servers[engine.IP] = ""
			}
		}
		delete(cluster.engines, id)
		cluster.Unlock()
		engine.Close()
		return true
	}
	return false
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
