package cluster

import "github.com/humpback/gounits/logger"
import "common/models"

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
func (upgradeContainer *UpgradeContainer) Execute(imageTag string) error {

	engine := upgradeContainer.Original.Engine
	if !engine.IsHealthy() {
		upgradeContainer.State = UpgradeIgnore
		return nil
	}

	originalContainer := upgradeContainer.Original
	operateContainer := models.ContainerOperate{Action: "upgrade", Container: originalContainer.Config.ID, ImageTag: imageTag}
	newContainer, err := engine.UpgradeContainer(operateContainer)
	if err != nil {
		upgradeContainer.State = UpgradeFailure
		return fmt.Errorf("engine %s %s", engine.IP, err.Error())
	}
	upgradeContainer.New = newContainer
	upgradeContainer.State = UpgradeCompleted
	engine.RefreshContainers()
	return nil
}

// Recovery is exported
// upgrade container failure, recovery completed containers to original image tag
func (upgradeContainer *UpgradeContainer) Recovery(imageTag string) error {

	engine := upgradeContainer.Original.Engine
	if !engine.IsHealthy() {
		return nil
	}

	upgradeContainer.State = UpgradeFailure
	newContainer := upgradeContainer.New
	operateContainer := models.ContainerOperate{Action: "upgrade", Container: newContainer.Config.ID, ImageTag: imageTag}
	newContainer, err := engine.UpgradeContainer(operateContainer)
	if err != nil {
		return fmt.Errorf("engine %s %s", engine.IP, err.Error())
	}
	upgradeContainer.Original = newContainer
	upgradeContainer.State = UpgradeRecovery
	engine.RefreshContainers()
	return nil
}

// Upgrader is exported
type Upgrader struct {
	sync.RWMutex
	MetaID           string
	OriginalImageTag string
	NewImageTag      string
	delayInterval    time.Duration
	callback         UpgraderHandleFunc
	configCache      *ContainersConfigCache
	containers       []*UpgradeContainer
}

// NewUpgrader is exported
func NewUpgrader(metaid string, imageTag string, containers Containers, upgradeDelay time.Duration,
	configCache *ContainersConfigCache, callback UpgraderHandleFunc) *Upgrader {

	metaData := configCache.GetMetaData(metaid)
	if metaData == nil {
		return nil
	}

	upgradeContainers := []*UpgradeContainer{}
	for _, container := range containers {
		upgradeContainers = append(upgradeContainers, &UpgradeContainer{
			Original: container,
			New:      nil,
			State:    UpgradeReady,
		})
	}

	return &Upgrader{
		MetaID:           metaid,
		OriginalImageTag: metaData.ImageTag,
		NewImageTag:      imageTag,
		delayInterval:    upgradeDelay,
		callback:         callback,
		configCache:      configCache,
		containers:       upgradeContainers,
	}
}

// Start is exported
func (upgrader *Upgrader) Start(upgradeCh chan<- bool) {

	upgrader.Lock()
	defer upgrader.Unlock()
	var err error
	ret := true
	errMsgs := []string{}

	upgrader.configCache.SetImageTag(upgrader.MetaID, upgrader.NewImageTag)
	for _, upgradeContainer := range upgrader.containers {
		if err = upgradeContainer.Execute(upgrader.NewImageTag); err != nil {
			upgrader.configCache.RemoveContainerBaseConfig(upgrader.MetaID, upgradeContainer.Original.Config.ID)
			errMsgs = append(errMsgs, "upgrade container execute, "+err.Error())
			logger.ERROR("[#cluster#] upgrade container %s execute %s", upgradeContainer.Original.Config.ID[:12], err.Error())
			break
		}
		if upgradeContainer.State == UpgradeCompleted {
			baseConfig := ContainerBaseConfig{}
			baseConfig = *(upgradeContainer.Original.BaseConfig)
			baseConfig.ID = upgradeContainer.New.Config.ID
			baseConfig.Image = upgradeContainer.New.Config.Image
			upgradeContainer.New.BaseConfig = &baseConfig
			upgrader.configCache.CreateContainerBaseConfig(upgrader.MetaID, &baseConfig)
			upgrader.configCache.RemoveContainerBaseConfig(upgrader.MetaID, upgradeContainer.Original.Config.ID)
			time.Sleep(upgrader.delayInterval)
		}
	}

	if err != nil { //recovery upgrade completed containers
		ret = false
		upgrader.configCache.SetImageTag(upgrader.MetaID, upgrader.OriginalImageTag)
		for _, upgradeContainer := range upgrader.containers {
			if upgradeContainer.State == UpgradeCompleted {
				if err := upgradeContainer.Recovery(upgrader.OriginalImageTag); err != nil {
					upgrader.configCache.RemoveContainerBaseConfig(upgrader.MetaID, upgradeContainer.New.Config.ID)
					errMsgs = append(errMsgs, "upgrade container recovery, "+err.Error())
					logger.ERROR("[#cluster#] upgrade container %s recovery %s", upgradeContainer.New.Config.ID[:12], err.Error())
				}
				if upgradeContainer.State == UpgradeRecovery {
					baseConfig := ContainerBaseConfig{}
					baseConfig = *(upgradeContainer.New.BaseConfig)
					baseConfig.ID = upgradeContainer.Original.Config.ID
					baseConfig.Image = upgradeContainer.Original.Config.Image
					upgradeContainer.Original.BaseConfig = &baseConfig
					upgrader.configCache.CreateContainerBaseConfig(upgrader.MetaID, &baseConfig)
					upgrader.configCache.RemoveContainerBaseConfig(upgrader.MetaID, upgradeContainer.New.Config.ID)
					time.Sleep(upgrader.delayInterval)
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
	delayInterval time.Duration
	configCache   *ContainersConfigCache
	upgraders     map[string]*Upgrader
}

// NewUpgradeContainersCache is exported
func NewUpgradeContainersCache(upgradeDelay time.Duration, configCache *ContainersConfigCache) *UpgradeContainersCache {

	return &UpgradeContainersCache{
		delayInterval: upgradeDelay,
		configCache:   configCache,
		upgraders:     make(map[string]*Upgrader),
	}
}

// Upgrade is exported
func (cache *UpgradeContainersCache) Upgrade(upgradeCh chan<- bool, metaid string, imageTag string, containers Containers) {

	cache.Lock()
	if _, ret := cache.upgraders[metaid]; !ret {
		upgrader := NewUpgrader(metaid, imageTag, containers, cache.delayInterval, cache.configCache, cache.UpgraderHandleFunc)
		if upgrader != nil {
			cache.upgraders[metaid] = upgrader
			logger.INFO("[#cluster#] upgrade start %s > %s", metaid, imageTag)
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
	if len(errMsgs) > 0 {
		logger.ERROR("[#cluster#] upgrade failure %s > %s", upgrader.MetaID, upgrader.NewImageTag)
	} else {
		logger.INFO("[#cluster#] upgrade done %s > %s", upgrader.MetaID, upgrader.NewImageTag)
	}
}
