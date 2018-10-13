package cluster

import "github.com/humpback/gounits/logger"

import (
	"sync"
	"time"
)

// EnginesPool is exported
type EnginesPool struct {
	sync.RWMutex
	Cluster     *Cluster
	poolEngines map[string]*Engine
	pendEngines map[string]*Engine
	stopCh      chan struct{}
}

// NewEnginesPool is exported
func NewEnginesPool() *EnginesPool {

	pool := &EnginesPool{
		poolEngines: make(map[string]*Engine),
		pendEngines: make(map[string]*Engine),
		stopCh:      make(chan struct{}),
	}
	go pool.doLoop()
	return pool
}

// SetCluster is exported
func (pool *EnginesPool) SetCluster(cluster *Cluster) {

	pool.Cluster = cluster
}

// Release is exported
func (pool *EnginesPool) Release() {

	close(pool.stopCh)
	pool.Lock()
	for _, engine := range pool.pendEngines {
		delete(pool.pendEngines, engine.IP)
	}
	for _, engine := range pool.poolEngines {
		delete(pool.poolEngines, engine.IP)
	}
	pool.Unlock()
}

// InitEngineNodeLabels is exported
func (pool *EnginesPool) InitEngineNodeLabels(engine *Engine) {

	node, _ := pool.Cluster.storageDriver.NodeStorage.NodeByIP(engine.IP)
	if node != nil {
		engine.SetNodeLabelsPairs(node.NodeLabels)
	}
}

// AddEngine is exported
func (pool *EnginesPool) AddEngine(ip string, name string) {

	ipOrName := selectIPOrName(ip, name)
	nodeData := pool.Cluster.nodeCache.Get(ipOrName)
	if nodeData == nil {
		return
	}

	pool.Cluster.storageDriver.NodeStorage.SetNodeData(nodeData)
	engine := pool.Cluster.GetEngine(nodeData.IP)
	if engine != nil {
		pool.InitEngineNodeLabels(engine)
		engine.Update(nodeData)
		return
	}

	if ret := pool.Cluster.InGroupsContains(nodeData.IP, nodeData.Name); !ret {
		return
	}

	pool.Lock()
	defer pool.Unlock()
	if pendEngine, ret := pool.pendEngines[nodeData.IP]; ret {
		pool.InitEngineNodeLabels(pendEngine)
		if pendEngine.IsHealthy() {
			delete(pool.pendEngines, pendEngine.IP)
			pendEngine.Update(nodeData)
			pool.Cluster.Lock()
			pool.Cluster.engines[pendEngine.IP] = pendEngine
			pool.Cluster.Unlock()
			logger.INFO("[#cluster#] addengine, pool engine reused %s %s %s.", pendEngine.IP, pendEngine.Name, pendEngine.State())
		} else {
			logger.INFO("[#cluster#] addengine, pool pending engine %s %s %s is already.", pendEngine.IP, pendEngine.Name, pendEngine.State())
		}
		return
	}

	poolEngine, ret := pool.poolEngines[nodeData.IP]
	if ret {
		poolEngine.Update(nodeData)
		poolEngine.SetState(StatePending)
		logger.INFO("[#cluster#] addengine, pool engine reused %s %s %s.", poolEngine.IP, poolEngine.Name, poolEngine.State())
	} else {
		var err error
		poolEngine, err = NewEngine(nodeData, pool.Cluster.overcommitRatio, pool.Cluster.removeDelay, pool.Cluster.configCache)
		if err != nil {
			return
		}
		pool.poolEngines[poolEngine.IP] = poolEngine
		logger.INFO("[#cluster#] addengine, pool engine create %s %s %s.", poolEngine.IP, poolEngine.Name, poolEngine.State())
	}
	pool.InitEngineNodeLabels(poolEngine)
	pool.pendEngines[poolEngine.IP] = poolEngine
}

// RemoveEngine is exported
func (pool *EnginesPool) RemoveEngine(ip string, name string) {

	ipOrName := selectIPOrName(ip, name)
	nodeData := pool.Cluster.nodeCache.Get(ipOrName)
	if nodeData == nil {
		return
	}

	pool.Lock()
	if engine := pool.Cluster.GetEngine(nodeData.IP); engine != nil {
		pool.Cluster.Lock()
		delete(pool.Cluster.engines, engine.IP)
		pool.Cluster.Unlock()
		pool.pendEngines[engine.IP] = engine
	}
	pool.Unlock()
}

func (pool *EnginesPool) doLoop() {

	for {
		ticker := time.NewTicker(2 * time.Second)
		select {
		case <-ticker.C:
			{
				ticker.Stop()
				pool.Lock()
				wgroup := sync.WaitGroup{}
				for _, pendEngine := range pool.pendEngines {
					if pendEngine.IsPending() {
						wgroup.Add(1)
						go func(engine *Engine) {
							if err := engine.RefreshContainers(); err == nil {
								engine.Open()
								pool.Cluster.migtatorCache.Cancel(engine)
								pool.Cluster.Lock()
								pool.Cluster.engines[engine.IP] = engine
								pool.Cluster.Unlock()
								logger.INFO("[#cluster#] engine %s %s %s", engine.IP, engine.Name, engine.State())
							}
							wgroup.Done()
						}(pendEngine)
					} else if pendEngine.IsHealthy() {
						wgroup.Add(1)
						go func(engine *Engine) {
							pool.Cluster.migtatorCache.Start(engine)
							engine.Close()
							logger.INFO("[#cluster#] engine %s %s %s", engine.IP, engine.Name, engine.State())
							wgroup.Done()
						}(pendEngine)
					}
				}
				wgroup.Wait()
				for _, pendEngine := range pool.pendEngines {
					delete(pool.pendEngines, pendEngine.IP)
				}
				pool.Unlock()
			}
		case <-pool.stopCh:
			{
				ticker.Stop()
				return
			}
		}
	}
}
