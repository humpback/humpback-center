package cluster

import "github.com/humpback/gounits/utils"

import (
	"sync"
)

// Restorer is exported
type Restorer struct {
	sync.RWMutex
	Cluster  *Cluster
	isRuning bool
	metaIds  []string
}

// NewRestorer is exported
func NewRestorer() *Restorer {

	return &Restorer{
		isRuning: false,
		metaIds:  []string{},
	}
}

// SetCluster is exported
func (restorer *Restorer) SetCluster(cluster *Cluster) {

	restorer.Cluster = cluster
}

// RecoveryMetaID is exported
func (restorer *Restorer) RecoveryMetaID(metaid string) {

	restorer.Lock()
	if ret := utils.Contains(metaid, restorer.metaIds); !ret {
		restorer.metaIds = append(restorer.metaIds, metaid)
	}
	restorer.Unlock()

	restorer.RLock()
	defer restorer.RUnlock()
	if !restorer.isRuning {
		go restorer.recovery()
	}
}

// RecoveryGroup is exported
func (restorer *Restorer) RecoveryGroup(group *Group) {

	groupMetaData := restorer.Cluster.configCache.GetGroupMetaData(group.ID)
	for _, metaData := range groupMetaData {
		restorer.RecoveryMetaID(metaData.MetaID)
	}
}

// RecoveryGroups is exported
func (restorer *Restorer) RecoveryGroups(groups []*Group) {

	for _, group := range groups {
		restorer.RecoveryGroup(group)
	}
}

// RecoveryEngine is exported
func (restorer *Restorer) RecoveryEngine(engine *Engine) {

	groups := restorer.Cluster.GetEngineGroups(engine)
	if len(groups) > 0 {
		restorer.RecoveryGroups(groups)
	}
}

func (restorer *Restorer) recovery() {

	restorer.Lock()
	defer restorer.Unlock()
	if restorer.isRuning {
		return
	}

	restorer.isRuning = true
	nLen := len(restorer.metaIds)
	for i := 0; i < nLen; i++ {
		restorer.Cluster.RecoveryContainers(restorer.metaIds[i])
	}
	restorer.metaIds = []string{}
	restorer.isRuning = false
}
