package cluster

import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-agent/models"

import (
	"fmt"
	"sync"
)

type UpgradeState int

const (
	UpgradeReady = iota + 1
	UpgradeIgnore
	UpgradeCompleted
	UpgradeFailure
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

	//baseConfig := &ContainerBaseConfig{}
	//*baseConfig = *originalContainer.BaseConfig
	//baseConfig.ID = newContainer.Config
	//设置baseconfig
	return nil
}

func (upgradeContainer *UpgradeContainer) Recovery() error {

	//还原container
	//还原baseconfig
	return nil
}

type upgrader struct {
	sync.RWMutex
	MetaID      string
	ImageTag    string
	configCache *ContainersConfigCache
	containers  []*UpgradeContainer
}

func NewUpgrader(metaid string, imageTag string, containers Containers, configCache *ContainersConfigCache) *upgrader {

	upgradeContainers := []*UpgradeContainer{}
	for _, container := range containers {
		upgradeContainers = append(upgradeContainers, &UpgradeContainer{
			Original: container,
			New:      nil,
			State:    UpgradeReady,
		})
	}

	return &upgrader{
		MetaID:      metaid,
		ImageTag:    imageTag,
		configCache: configCache,
		containers:  upgradeContainers,
	}
}

func (u *upgrader) Start() {

	u.configCache.SetImageTag(u.MetaID, u.ImageTag)
	for _, upgradeContainer := range u.containers {
		if err := upgradeContainer.Execute(u.ImageTag); err != nil {
			logger.ERROR("[#cluster#] %s", err.Error())
			//upgradeContainer.Recovery()
			return
		}
	}

	for _, upgradeContainer := range u.containers {
		if upgradeContainer.State == UpgradeIgnore {
			if err := upgradeContainer.Execute(u.ImageTag); err != nil {
				logger.ERROR("[#cluster#] %s", err.Error())
				//upgradeContainer.Recovery()
				return
			}
		}
	}
}

func (u *upgrader) Recovery() {

	//u.configCache.SetImageTag(u.MetaID, u.ImageTag) 设置回老版本
	//恢复所有为UpgradeCompleted的容器, 重新调用Execute，给入老版本
	//事件回调，删除升级map，
	//发送邮件警告升级失败或成功.
}

type UpgradeContainers struct {
	sync.RWMutex
	configCache *ContainersConfigCache
	upgraders   map[string]*upgrader
}

func NewUpgradeContainers(configCache *ContainersConfigCache) *UpgradeContainers {

	return &UpgradeContainers{
		configCache: configCache,
		upgraders:   make(map[string]*upgrader),
	}
}

func (cache *UpgradeContainers) Upgrade(metaid string, imageTag string, containers Containers) {

	cache.Lock()
	if _, ret := cache.upgraders[metaid]; !ret {
		upgrader := NewUpgrader(metaid, imageTag, containers, cache.configCache)
		cache.upgraders[metaid] = upgrader
		go upgrader.Start()
	}
	cache.Unlock()
}

func (cache *UpgradeContainers) Contains(metaid string) bool {

	cache.RLock()
	defer cache.RUnlock()
	_, ret := cache.upgraders[metaid]
	return ret
}
