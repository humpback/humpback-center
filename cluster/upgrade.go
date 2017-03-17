package cluster

import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-agent/models"

import (
	"fmt"
	"sync"
	"time"
)

type UpgradeState int

const (
	UpgradeReady = iota + 1
	UpgradeIgnore
	UpgradeCompleted
	UpgradeFailure
	UpgradeRecovery
)

type UpgradeContainer struct {
	Original *Container
	New      *Container
	State    UpgradeState
}

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
	return nil
}

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
	return nil
}

type Upgrader struct {
	sync.RWMutex
	MetaID           string
	OriginalImageTag string
	NewImageTag      string
	delayInterval    time.Duration
	callback         UpgraderHandlerFunc
	configCache      *ContainersConfigCache
	containers       []*UpgradeContainer
}

func NewUpgrader(metaid string, imageTag string, containers Containers, upgradeDelay time.Duration,
	configCache *ContainersConfigCache, callback UpgraderHandlerFunc) *Upgrader {

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

func (upgrader *Upgrader) Start() {

	upgrader.Lock()
	defer upgrader.Unlock()
	var err error
	errMsgs := []string{}
	upgrader.configCache.SetImageTag(upgrader.MetaID, upgrader.NewImageTag)
	for _, upgradeContainer := range upgrader.containers {
		if err = upgradeContainer.Execute(upgrader.NewImageTag); err != nil {
			logger.ERROR("[#cluster#] upgrade execute %s", err.Error())
			errMsgs = append(errMsgs, "upgrade container execute, "+err.Error())
			break
		}
		if upgradeContainer.State == UpgradeCompleted {
			baseConfig := ContainerBaseConfig{}
			baseConfig = *(upgradeContainer.Original.BaseConfig)
			baseConfig.ID = upgradeContainer.New.Config.ID
			baseConfig.Image = upgradeContainer.New.Config.Image
			upgradeContainer.New.GroupID = baseConfig.MetaData.GroupID
			upgradeContainer.New.MetaID = baseConfig.MetaData.MetaID
			upgradeContainer.New.BaseConfig = &baseConfig
			upgrader.configCache.CreateContainerBaseConfig(baseConfig.MetaData.MetaID, &baseConfig)
			upgrader.configCache.RemoveContainerBaseConfig(baseConfig.MetaData.MetaID, upgradeContainer.Original.Config.ID)
			time.Sleep(upgrader.delayInterval)
		}
	}

	if err != nil { //recovery upgrade completed containers
		upgrader.configCache.SetImageTag(upgrader.MetaID, upgrader.OriginalImageTag)
		for _, upgradeContainer := range upgrader.containers {
			if upgradeContainer.State == UpgradeCompleted {
				if err := upgradeContainer.Recovery(upgrader.OriginalImageTag); err != nil {
					logger.ERROR("[#cluster#] upgrade recovery %s", err.Error())
					errMsgs = append(errMsgs, "upgrade container recovery, "+err.Error())
				}
				if upgradeContainer.State == UpgradeRecovery {
					baseConfig := ContainerBaseConfig{}
					baseConfig = *(upgradeContainer.New.BaseConfig)
					baseConfig.ID = upgradeContainer.Original.Config.ID
					baseConfig.Image = upgradeContainer.Original.Config.Image
					upgradeContainer.Original.GroupID = baseConfig.MetaData.GroupID
					upgradeContainer.Original.MetaID = baseConfig.MetaData.MetaID
					upgradeContainer.Original.BaseConfig = &baseConfig
					upgrader.configCache.CreateContainerBaseConfig(baseConfig.MetaData.MetaID, &baseConfig)
					upgrader.configCache.RemoveContainerBaseConfig(baseConfig.MetaData.MetaID, upgradeContainer.New.Config.ID)
					time.Sleep(upgrader.delayInterval)
				}
			}
		}
	}
	upgrader.callback(upgrader, errMsgs)
}

type UpgraderHandlerFunc func(upgrader *Upgrader, errMsgs []string)

type UpgradeContainers struct {
	sync.RWMutex
	delayInterval time.Duration
	configCache   *ContainersConfigCache
	upgraders     map[string]*Upgrader
}

func NewUpgradeContainers(upgradeDelay time.Duration, configCache *ContainersConfigCache) *UpgradeContainers {

	return &UpgradeContainers{
		delayInterval: upgradeDelay,
		configCache:   configCache,
		upgraders:     make(map[string]*Upgrader),
	}
}

func (upgradeContainers *UpgradeContainers) Upgrade(metaid string, imageTag string, containers Containers) {

	upgradeContainers.Lock()
	if _, ret := upgradeContainers.upgraders[metaid]; !ret {
		upgrader := NewUpgrader(metaid, imageTag, containers, upgradeContainers.delayInterval, upgradeContainers.configCache, upgradeContainers.UpgraderHandlerFunc)
		if upgrader != nil {
			upgradeContainers.upgraders[metaid] = upgrader
			logger.INFO("[#cluster] upgrade start %s > %s", metaid, imageTag)
			go upgrader.Start()
		}
	}
	upgradeContainers.Unlock()
}

func (upgradeContainers *UpgradeContainers) Contains(metaid string) bool {

	upgradeContainers.RLock()
	defer upgradeContainers.RUnlock()
	_, ret := upgradeContainers.upgraders[metaid]
	return ret
}

func (upgradeContainers *UpgradeContainers) UpgraderHandlerFunc(upgrader *Upgrader, errMsgs []string) {

	logger.INFO("[#cluster] upgrade done %s > %s", upgrader.MetaID, upgrader.NewImageTag)
	upgradeContainers.Lock()
	delete(upgradeContainers.upgraders, upgrader.MetaID)
	upgradeContainers.Unlock()
}
