package cluster

import "github.com/humpback/gounits/logger"
import "github.com/humpback/common/models"

import (
	"fmt"
	"sync"
	"time"
)

// UpgradeState is exported
type UpgradeState int

const (
	// UpgradeReady is exported, engine container upgrade is ready
	UpgradeReady = iota + 1
	// UpgradeIgnore is exported, engine healthy is false, can't upgrade, migrate container.
	UpgradeIgnore
	// UpgradeCompleted is exported, engine container upgraded completed.
	UpgradeCompleted
	// UpgradeFailure is exported, engine container upgraded failure.
	UpgradeFailure
	// UpgradeRecovery is exported, engine container recovery upgraded.
	UpgradeRecovery
)

// UpgradeContainer is exported
type UpgradeContainer struct {
	Original *Container
	New      *Container
	State    UpgradeState
}

// Execute is exported
// upgrade originalContainer to image new tag
func (upgradeContainer *UpgradeContainer) Execute(newImageTag string) error {

	engine := upgradeContainer.Original.Engine
	if !engine.IsHealthy() {
		upgradeContainer.State = UpgradeIgnore
		return nil
	}

	originalContainer := upgradeContainer.Original
	containerOperate := models.ContainerOperate{Action: "upgrade", Container: originalContainer.Config.ID, ImageTag: newImageTag}
	newContainer, err := engine.ForceUpgradeContainer(containerOperate)
	if err != nil {
		upgradeContainer.State = UpgradeFailure
		return fmt.Errorf("engine %s %s", engine.IP, err.Error())
	}
	upgradeContainer.New = newContainer
	upgradeContainer.State = UpgradeCompleted
	return nil
}

// Recovery is exported
// upgrade container failure, recovery completed containers to original image tag
func (upgradeContainer *UpgradeContainer) Recovery(originalImageTag string) error {

	engine := upgradeContainer.Original.Engine
	if !engine.IsHealthy() {
		return nil
	}

	upgradeContainer.State = UpgradeFailure
	newContainer := upgradeContainer.New
	containerOperate := models.ContainerOperate{Action: "upgrade", Container: newContainer.Config.ID, ImageTag: originalImageTag}
	newContainer, err := engine.ForceUpgradeContainer(containerOperate)
	if err != nil {
		return fmt.Errorf("engine %s %s", engine.IP, err.Error())
	}
	upgradeContainer.Original = newContainer
	upgradeContainer.State = UpgradeRecovery
	return nil
}

// Upgrader is exported
type Upgrader struct {
	sync.RWMutex
	MetaID        string
	OriginalTag   string
	NewTag        string
	configCache   *ContainersConfigCache
	delayInterval time.Duration
	callback      UpgraderHandleFunc
	containers    []*UpgradeContainer
}

// NewUpgrader is exported
func NewUpgrader(metaid string, originalTag string, newTag string, containers Containers, upgradeDelay time.Duration,
	configCache *ContainersConfigCache, callback UpgraderHandleFunc) *Upgrader {

	upgradeContainers := []*UpgradeContainer{}
	for _, container := range containers {
		if container.BaseConfig != nil && container.Engine != nil {
			upgradeContainers = append(upgradeContainers, &UpgradeContainer{
				Original: container,
				New:      nil,
				State:    UpgradeReady,
			})
		}
	}

	return &Upgrader{
		MetaID:        metaid,
		OriginalTag:   originalTag,
		NewTag:        newTag,
		configCache:   configCache,
		delayInterval: upgradeDelay,
		callback:      callback,
		containers:    upgradeContainers,
	}
}

// Start is exported
func (upgrader *Upgrader) Start(upgradeCh chan<- bool) {

	var (
		err     error
		ret     bool
		errMsgs []string
	)

	ret = true
	errMsgs = []string{}
	upgrader.Lock()
	defer upgrader.Unlock()
	upgrader.configCache.SetImageTag(upgrader.MetaID, upgrader.NewTag)
	for _, upgradeContainer := range upgrader.containers {
		if err = upgradeContainer.Execute(upgrader.NewTag); err != nil {
			upgrader.configCache.RemoveContainerBaseConfig(upgrader.MetaID, upgradeContainer.Original.Config.ID)
			errMsgs = append(errMsgs, "upgrade container execute, "+err.Error())
			logger.ERROR("[#cluster#] upgrade container %s execute %s", ShortContainerID(upgradeContainer.Original.Config.ID), err.Error())
			break
		}
	}

	if err != nil { //recovery upgrade completed containers
		ret = false
		upgrader.configCache.SetImageTag(upgrader.MetaID, upgrader.OriginalTag)
		for _, upgradeContainer := range upgrader.containers {
			if upgradeContainer.State == UpgradeCompleted {
				if err := upgradeContainer.Recovery(upgrader.OriginalTag); err != nil {
					upgrader.configCache.RemoveContainerBaseConfig(upgrader.MetaID, upgradeContainer.New.Config.ID)
					errMsgs = append(errMsgs, "upgrade container recovery, "+err.Error())
					logger.ERROR("[#cluster#] upgrade container %s recovery %s", ShortContainerID(upgradeContainer.New.Config.ID), err.Error())
				}
			}
		}
	}
	upgrader.callback(upgrader, errMsgs)
	upgradeCh <- ret
}

// UpgraderHandleFunc exported
type UpgraderHandleFunc func(upgrader *Upgrader, errMsgs []string)

// UpgradeContainersCache is exported
type UpgradeContainersCache struct {
	sync.RWMutex
	Cluster       *Cluster
	delayInterval time.Duration
	upgraders     map[string]*Upgrader
}

// NewUpgradeContainersCache is exported
func NewUpgradeContainersCache(upgradeDelay time.Duration) *UpgradeContainersCache {

	return &UpgradeContainersCache{
		delayInterval: upgradeDelay,
		upgraders:     make(map[string]*Upgrader),
	}
}

// SetCluster is exported
func (cache *UpgradeContainersCache) SetCluster(cluster *Cluster) {

	cache.Cluster = cluster
}

// Upgrade is exported
func (cache *UpgradeContainersCache) Upgrade(upgradeCh chan<- bool, metaid string, newTag string, containers Containers) {

	if cache.Cluster == nil || cache.Cluster.configCache == nil {
		return
	}

	configCache := cache.Cluster.configCache
	metaData := configCache.GetMetaData(metaid)
	if metaData == nil {
		return
	}

	cache.Lock()
	if _, ret := cache.upgraders[metaid]; !ret {
		upgrader := NewUpgrader(metaData.MetaID, metaData.ImageTag, newTag, containers, cache.delayInterval, configCache, cache.UpgraderHandleFunc)
		if upgrader != nil {
			cache.upgraders[metaData.MetaID] = upgrader
			logger.INFO("[#cluster#] upgrade start %s > %s", upgrader.MetaID, upgrader.NewTag)
			go upgrader.Start(upgradeCh)
		}
	}
	cache.Unlock()
}

// Contains is exported
func (cache *UpgradeContainersCache) Contains(metaid string) bool {

	cache.RLock()
	defer cache.RUnlock()
	_, ret := cache.upgraders[metaid]
	return ret
}

// UpgraderHandleFunc is exported
func (cache *UpgradeContainersCache) UpgraderHandleFunc(upgrader *Upgrader, errMsgs []string) {

	cache.Lock()
	delete(cache.upgraders, upgrader.MetaID)
	cache.Unlock()
	if cache.Cluster != nil {
		if _, engines, err := cache.Cluster.GetMetaDataEngines(upgrader.MetaID); err == nil {
			metaEngines := make(map[string]*Engine)
			for _, engine := range engines {
				if engine.IsHealthy() && engine.HasMeta(upgrader.MetaID) {
					metaEngines[engine.IP] = engine
				}
			}
			cache.Cluster.RefreshEnginesContainers(metaEngines)
		}
	}
	if len(errMsgs) > 0 {
		logger.ERROR("[#cluster#] upgrade failure %s > %s", upgrader.MetaID, upgrader.NewTag)
	} else {
		logger.INFO("[#cluster#] upgrade done %s > %s", upgrader.MetaID, upgrader.NewTag)
	}
}
