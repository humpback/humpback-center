package cluster

import (
	"time"
)

// MetaRestorer is exported
type MetaRestorer struct {
	Cluster          *Cluster
	recoveryInterval time.Duration
	stopCh           chan struct{}
}

// NewMetaRestorer is exported
func NewMetaRestorer(recoveryInterval time.Duration) *MetaRestorer {

	return &MetaRestorer{
		recoveryInterval: recoveryInterval,
		stopCh:           make(chan struct{}),
	}
}

// SetCluster is exported
func (restorer *MetaRestorer) SetCluster(cluster *Cluster) {

	restorer.Cluster = cluster
}

// Start is exported
func (restorer *MetaRestorer) Start() {

	if restorer.Cluster != nil {
		go restorer.doLoop()
	}
}

// Stop is exported
func (restorer *MetaRestorer) Stop() {

	close(restorer.stopCh)
}

// doLoop is exported
func (restorer *MetaRestorer) doLoop() {

	for {
		ticker := time.NewTicker(restorer.recoveryInterval)
		select {
		case <-ticker.C:
			{
				ticker.Stop()
				restorer.recovery()
			}
		case <-restorer.stopCh:
			{
				ticker.Stop()
				return
			}
		}
	}
}

func (restorer *MetaRestorer) recovery() {

	if restorer.Cluster != nil {
		metaids := []string{}
		metaEngines := make(map[string]*Engine)
		groups := restorer.Cluster.GetGroups()
		for _, group := range groups {
			groupMetaData := restorer.Cluster.configCache.GetGroupMetaData(group.ID)
			for _, metaData := range groupMetaData {
				metaids = append(metaids, metaData.MetaID)
				if _, engines, err := restorer.Cluster.GetMetaDataEngines(metaData.MetaID); err == nil {
					for _, engine := range engines {
						if engine.IsHealthy() && engine.HasMeta(metaData.MetaID) {
							metaEngines[engine.IP] = engine
						}
					}
				}
			}
		}
		if len(metaids) > 0 {
			restorer.Cluster.RefreshEnginesContainers(metaEngines)
			for _, metaid := range metaids {
				restorer.Cluster.RecoveryContainers(metaid)
			}
		}
	}
}
