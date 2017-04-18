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
				if restorer.Cluster != nil {
					metaids := []string{}
					groups := restorer.Cluster.GetGroups()
					for _, group := range groups {
						groupMetaData := restorer.Cluster.configCache.GetGroupMetaData(group.ID)
						for _, metaData := range groupMetaData {
							metaids = append(metaids, metaData.MetaID)
						}
					}
					for _, metaid := range metaids {
						restorer.Cluster.RecoveryContainers(metaid)
					}
				}
			}
		case <-restorer.stopCh:
			{
				ticker.Stop()
				return
			}
		}
	}
}
