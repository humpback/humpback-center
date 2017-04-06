package cluster

import "github.com/humpback/gounits/logger"

import (
	"sync"
	"time"
)

// PendEngines is exported
type PendEngines struct {
	sync.RWMutex
	Cluster *Cluster
	stopCh  chan struct{}
	engines map[string]*Engine
}

// NewPendEngines is exported
func NewPendEngines() *PendEngines {

	pend := &PendEngines{
		stopCh:  make(chan struct{}),
		engines: make(map[string]*Engine),
	}
	go pend.doLoop()
	return pend
}

// SetCluster is exported
func (pend *PendEngines) SetCluster(cluster *Cluster) {

	pend.Cluster = cluster
}

// Close is exported
func (pend *PendEngines) Close() {

	close(pend.stopCh)
	pend.Lock()
	for _, engine := range pend.engines {
		engine.Close()
		delete(pend.engines, engine.IP)
	}
	pend.Unlock()
}

// AddEngine is exported
func (pend *PendEngines) AddEngine(ip string, name string) {

	ipOrName := selectIPOrName(ip, name)
	nodeData := pend.Cluster.nodeCache.Get(ipOrName)
	if nodeData == nil {
		return
	}

	if engine := pend.Cluster.GetEngine(nodeData.IP); engine != nil {
		return
	}

	if ret := pend.Cluster.InGroupsContains(nodeData.IP, nodeData.Name); !ret {
		return
	}

	pend.Lock()
	engine, ret := pend.engines[nodeData.IP]
	if !ret {
		var err error
		if engine, err = NewEngine(nodeData, pend.Cluster.overcommitRatio, pend.Cluster.configCache); err == nil {
			pend.engines[nodeData.IP] = engine
		}
	}
	pend.Unlock()
}

// RemoveEngine is exported
func (pend *PendEngines) RemoveEngine(ip string, name string) {

	ipOrName := selectIPOrName(ip, name)
	nodeData := pend.Cluster.nodeCache.Get(ipOrName)
	if nodeData == nil {
		return
	}

	engine := pend.Cluster.GetEngine(nodeData.IP)
	if engine != nil {
		pend.Cluster.Lock()
		delete(pend.Cluster.engines, engine.IP)
		pend.Cluster.Unlock()
		pend.Lock()
		pend.engines[engine.IP] = engine
		pend.Unlock()
	}
}

func (pend *PendEngines) doLoop() {

	for {
		ticker := time.NewTicker(3 * time.Second)
		select {
		case <-ticker.C:
			{
				ticker.Stop()
				pend.Lock()
				wgroup := sync.WaitGroup{}
				for _, engine := range pend.engines {
					if engine.IsPending() {
						wgroup.Add(1)
						go func(e *Engine) {
							if err := e.RefreshContainers(); err == nil {
								////检查meta实例数，若不对应，则创建差值
								e.Open()
								pend.Cluster.Lock()
								pend.Cluster.engines[e.IP] = e
								pend.Cluster.Unlock()
								pend.Cluster.migtatorCache.Cancel(e)
								logger.INFO("[#cluster#] engine %s %s", e.IP, e.State())
							}
							wgroup.Done()
						}(engine)
					} else if engine.IsHealthy() {
						wgroup.Add(1)
						go func(e *Engine) {
							pend.Cluster.migtatorCache.Start(e)
							e.Close()
							logger.INFO("[#cluster#] engine %s %s", e.IP, e.State())
							wgroup.Done()
						}(engine)
					}
				}
				wgroup.Wait()
				for _, engine := range pend.engines {
					delete(pend.engines, engine.IP)
				}
				pend.Unlock()
			}
		case <-pend.stopCh:
			{
				ticker.Stop()
				return
			}
		}
	}
}
