package cluster

import "github.com/humpback/discovery"
import "github.com/humpback/discovery/backends"
import "github.com/humpback/gounits/json"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/gounits/system"
import "github.com/humpback/humpback-agent/models"
import "github.com/humpback/humpback-center/cluster/types"

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// pendingContainer is exported
type pendingContainer struct {
	GroupID string
	Name    string
	Config  models.Container
}

// Server is exported
type Server struct {
	Name string `json:"Name"`
	IP   string `json:"IP"`
}

// Group is exported
// Servers: cluster group's servers.
// ContactInfo: cluster manager contactinfo.
type Group struct {
	ID          string   `json:"ID"`
	IsCluster   bool     `json:"IsCluster"`
	Servers     []Server `json:"Servers"`
	ContactInfo string   `json:"ContactInfo"`
}

// Cluster is exported
type Cluster struct {
	sync.RWMutex
	Discovery *discovery.Discovery

	overcommitRatio   float64
	createRetry       int64
	randSeed          *rand.Rand
	nodeCache         *NodeCache
	configCache       *ContainersConfigCache
	upgraderCache     *UpgradeContainersCache
	migtatorCache     *MigrateContainersCache
	enginesPool       *EnginesPool
	restorer          *Restorer
	hooksProcessor    *HooksProcessor
	pendingContainers map[string]*pendingContainer
	engines           map[string]*Engine
	groups            map[string]*Group
	stopCh            chan struct{}
}

// NewCluster is exported
func NewCluster(driverOpts system.DriverOpts, discovery *discovery.Discovery) (*Cluster, error) {

	if discovery == nil {
		return nil, ErrClusterDiscoveryInvalid
	}

	overcommitratio := 0.05
	if val, ret := driverOpts.Float("overcommit", ""); ret {
		if val <= float64(-1) {
			logger.WARN("[#cluster#] set overcommit should be larger than -1, %f is invalid.", val)
		} else if val < float64(0) {
			logger.WARN("[#cluster#] opts, -1 < overcommit < 0 will make center take less resource than docker engine offers.")
			overcommitratio = val
		} else {
			overcommitratio = val
		}
	}

	createretry := int64(0)
	if val, ret := driverOpts.Int("createretry", ""); ret {
		if val < 0 {
			logger.WARN("[#cluster#] set createretry should be larger than or equal to 0, %d is invalid.", val)
		} else {
			createretry = val
		}
	}

	upgradedelay := 10 * time.Second
	if val, ret := driverOpts.String("upgradedelay", ""); ret {
		if dur, err := time.ParseDuration(val); err == nil {
			upgradedelay = dur
		}
	}

	migratedelay := 30 * time.Second
	if val, ret := driverOpts.String("migratedelay", ""); ret {
		if dur, err := time.ParseDuration(val); err == nil {
			migratedelay = dur
		}
	}

	cacheRoot := ""
	if val, ret := driverOpts.String("cacheroot", ""); ret {
		cacheRoot = val
	}

	hooksProcessor := NewHooksProcessor()
	restorer := NewRestorer()
	enginesPool := NewEnginesPool()
	migrateContainersCache := NewMigrateContainersCache(migratedelay)
	configCache, err := NewContainersConfigCache(cacheRoot)
	if err != nil {
		return nil, err
	}

	cluster := &Cluster{
		Discovery:         discovery,
		overcommitRatio:   overcommitratio,
		createRetry:       createretry,
		randSeed:          rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
		nodeCache:         NewNodeCache(),
		configCache:       configCache,
		upgraderCache:     NewUpgradeContainersCache(upgradedelay, configCache),
		migtatorCache:     migrateContainersCache,
		enginesPool:       enginesPool,
		restorer:          restorer,
		hooksProcessor:    hooksProcessor,
		pendingContainers: make(map[string]*pendingContainer),
		engines:           make(map[string]*Engine),
		groups:            make(map[string]*Group),
		stopCh:            make(chan struct{}),
	}

	hooksProcessor.SetCluster(cluster)
	restorer.SetCluster(cluster)
	enginesPool.SetCluster(cluster)
	migrateContainersCache.SetCluster(cluster)
	return cluster, nil
}

// Start is exported
// Cluster start, init container config cache watch open discovery service
func (cluster *Cluster) Start() error {

	cluster.configCache.Init()
	if cluster.Discovery != nil {
		logger.INFO("[#cluster#] discovery service watching...")
		cluster.Discovery.Watch(cluster.stopCh, cluster.watchDiscoveryHandleFunc)
		return nil
	}
	return ErrClusterDiscoveryInvalid
}

// Stop is exported
// Cluster stop
// close discovery service
// stop pendEngines loop
func (cluster *Cluster) Stop() {

	close(cluster.stopCh)
	cluster.enginesPool.Release()
	logger.INFO("[#cluster#] discovery service closed.")
}

// GetMetaDataEngines is exported
func (cluster *Cluster) GetMetaDataEngines(metaid string) (*MetaData, []*Engine, error) {

	metaData := cluster.GetMetaData(metaid)
	if metaData == nil {
		return nil, nil, ErrClusterMetaDataNotFound
	}

	engines := cluster.GetGroupEngines(metaData.GroupID)
	if engines == nil {
		return nil, nil, ErrClusterGroupNotFound
	}
	return metaData, engines, nil
}

// GetMetaData is exported
func (cluster *Cluster) GetMetaData(metaid string) *MetaData {

	return cluster.configCache.GetMetaData(metaid)
}

// GetMetaBase is exported
func (cluster *Cluster) GetMetaBase(metaid string) *MetaBase {

	if metaData := cluster.GetMetaData(metaid); metaData != nil {
		return &metaData.MetaBase
	}
	return nil
}

// GetEngine is exported
func (cluster *Cluster) GetEngine(ip string) *Engine {

	cluster.RLock()
	defer cluster.RUnlock()
	if engine, ret := cluster.engines[ip]; ret {
		return engine
	}
	return nil
}

// GetEngineGroups is exported
func (cluster *Cluster) GetEngineGroups(engine *Engine) []*Group {

	cluster.RLock()
	defer cluster.RUnlock()
	groups := []*Group{}
	for _, group := range cluster.groups {
		for _, server := range group.Servers {
			if server.IP != "" && server.IP == engine.IP {
				groups = append(groups, group)
				break
			}
		}
	}
	for _, group := range cluster.groups {
		for _, server := range group.Servers {
			if server.Name != "" && server.Name == engine.Name {
				groups = append(groups, group)
				break
			}
		}
	}
	groups = removeDuplicatesGroups(groups)
	return groups
}

// GetGroupEngines is exported
func (cluster *Cluster) GetGroupEngines(groupid string) []*Engine {

	cluster.RLock()
	defer cluster.RUnlock()
	engines := []*Engine{}
	group, ret := cluster.groups[groupid]
	if !ret {
		return nil
	}

	//priority ip
	for _, server := range group.Servers {
		if server.IP != "" {
			for _, engine := range cluster.engines {
				if server.IP == engine.IP {
					engines = append(engines, engine)
					break
				}
			}
		} else if server.Name != "" {
			for _, engine := range cluster.engines {
				if server.Name == engine.Name {
					engines = append(engines, engine)
					break
				}
			}
		}
	}
	engines = removeDuplicatesEngines(engines)
	return engines
}

// InGroupsContains is exported
func (cluster *Cluster) InGroupsContains(ip string, name string) bool {

	cluster.RLock()
	defer cluster.RUnlock()
	for _, group := range cluster.groups {
		for _, server := range group.Servers {
			if server.IP != "" && server.IP == ip {
				return true
			}
		}
	}
	for _, group := range cluster.groups {
		for _, server := range group.Servers {
			if server.Name != "" && server.Name == name {
				return true
			}
		}
	}
	return false
}

// GetGroupAllContainers is exported
func (cluster *Cluster) GetGroupAllContainers(groupid string) *types.GroupContainers {

	group := cluster.GetGroup(groupid)
	if group == nil {
		return nil
	}

	groupContainers := types.GroupContainers{}
	groupMetaData := cluster.configCache.GetGroupMetaData(groupid)
	for _, metaData := range groupMetaData {
		if groupContainer := cluster.GetGroupContainers(metaData.MetaID); groupContainer != nil {
			groupContainers = append(groupContainers, groupContainer)
		}
	}
	return &groupContainers
}

// GetGroupContainers is exported
func (cluster *Cluster) GetGroupContainers(metaid string) *types.GroupContainer {

	metaData, engines, err := cluster.GetMetaDataEngines(metaid)
	if err != nil {
		return nil
	}

	for _, engine := range engines {
		if engine.IsHealthy() {
			engine.RefreshContainers()
		}
	}

	groupContainer := &types.GroupContainer{
		MetaID:     metaData.MetaID,
		Instances:  metaData.Instances,
		WebHooks:   metaData.WebHooks,
		Config:     metaData.Config,
		Containers: make([]*types.EngineContainer, 0),
	}

	for _, baseConfig := range metaData.BaseConfigs {
		for _, engine := range engines {
			if engine.IsHealthy() {
				if container := engine.Container(baseConfig.ID); container != nil {
					groupContainer.Containers = append(groupContainer.Containers, &types.EngineContainer{
						IP:        engine.IP,
						HostName:  engine.Name,
						Container: container.Config.Container,
					})
					break
				}
			}
		}
	}
	return groupContainer
}

// GetGroup is exported
func (cluster *Cluster) GetGroup(groupid string) *Group {

	cluster.RLock()
	defer cluster.RUnlock()
	group, ret := cluster.groups[groupid]
	if !ret {
		return nil
	}
	return group
}

// SetGroup is exported
func (cluster *Cluster) SetGroup(group *Group) {

	nSize := len(group.Servers)
	for i := 0; i < nSize; i++ {
		group.Servers[i].Name = strings.ToUpper(group.Servers[i].Name)
	}

	addServers := []Server{}
	removeServers := []Server{}
	cluster.Lock()
	pGroup, ret := cluster.groups[group.ID]
	if !ret {
		pGroup = group
		cluster.groups[group.ID] = pGroup
		logger.INFO("[#cluster#] group created %s(%d)", pGroup.ID, len(pGroup.Servers))
		for _, server := range pGroup.Servers {
			ipOrName := selectIPOrName(server.IP, server.Name)
			if nodeData := cluster.nodeCache.Get(ipOrName); nodeData != nil {
				addServers = append(addServers, server)
			}
		}
	} else {
		origins := pGroup.Servers
		pGroup.Servers = group.Servers
		logger.INFO("[#cluster#] group changed %s(%d)", pGroup.ID, len(pGroup.Servers))
		for _, originServer := range origins {
			found := false
			for _, newServer := range group.Servers {
				if ret := compareRemoveServers(cluster.nodeCache, originServer, newServer); ret {
					found = true
					break
				}
			}
			if !found {
				removeServers = append(removeServers, originServer)
			}
		}
		for _, newServer := range group.Servers {
			found := false
			for _, originServer := range origins {
				if ret := compareAddServers(cluster.nodeCache, originServer, newServer); ret {
					found = true
					break
				}
			}
			if !found {
				addServers = append(addServers, newServer)
			}
		}
	}
	cluster.Unlock()

	for _, server := range removeServers {
		if nodeData := cluster.nodeCache.Get(selectIPOrName(server.IP, server.Name)); nodeData != nil {
			if ret := cluster.InGroupsContains(nodeData.IP, nodeData.Name); !ret {
				logger.INFO("[#cluster#] group changed, remove to pendengines %s\t%s", server.IP, server.Name)
				cluster.enginesPool.RemoveEngine(server.IP, server.Name)
			}
		}
	}

	for _, server := range addServers {
		logger.INFO("[#cluster#] group changed, append to pendengines %s\t%s", server.IP, server.Name)
		cluster.enginesPool.AddEngine(server.IP, server.Name)
	}
}

// RemoveGroup is exported
func (cluster *Cluster) RemoveGroup(groupid string) bool {

	engines := cluster.GetGroupEngines(groupid)
	if engines == nil {
		logger.WARN("[#cluster#] remove group %s not found.", groupid)
		return false
	}

	groupMetaData := cluster.configCache.GetGroupMetaData(groupid)
	for _, metaData := range groupMetaData {
		for _, engine := range engines {
			if engine.IsHealthy() {
				containers := engine.Containers(metaData.MetaID)
				for _, container := range containers {
					if err := engine.RemoveContainer(container.Info.ID); err != nil {
						logger.ERROR("[#cluster#] engine %s, remove container error:%s", engine.IP, err.Error())
					}
				}
			}
		}
		cluster.hooksProcessor.Hook(metaData, RemoveMetaEvent)
	}
	cluster.configCache.RemoveGroupMetaData(groupid)
	cluster.Lock()
	delete(cluster.groups, groupid)
	logger.INFO("[#cluster#] removed group %s", groupid)
	cluster.Unlock()
	return true
}

func (cluster *Cluster) watchDiscoveryHandleFunc(added backends.Entries, removed backends.Entries, err error) {

	if err != nil {
		logger.ERROR("[#cluster#] discovery watch error:%s", err.Error())
		return
	}

	logger.INFO("[#cluster#] discovery watch removed:%d added:%d.", len(removed), len(added))
	for _, entry := range removed {
		nodeData := &NodeData{}
		if err := json.DeCodeBufferToObject(entry.Data, nodeData); err != nil {
			logger.ERROR("[#cluster#] discovery watch removed decode error: %s", err.Error())
			continue
		}
		nodeData.Name = strings.ToUpper(nodeData.Name)
		logger.INFO("[#cluster#] discovery watch, remove to pendengines %s\t%s", nodeData.IP, nodeData.Name)
		cluster.enginesPool.RemoveEngine(nodeData.IP, nodeData.Name)
		cluster.nodeCache.Remove(entry.Key)
	}

	for _, entry := range added {
		nodeData := &NodeData{}
		if err := json.DeCodeBufferToObject(entry.Data, nodeData); err != nil {
			logger.ERROR("[#cluster#] discovery service watch added decode error: %s", err.Error())
			continue
		}
		nodeData.Name = strings.ToUpper(nodeData.Name)
		logger.INFO("[#cluster#] discovery watch, append to pendengines %s\t%s", nodeData.IP, nodeData.Name)
		cluster.nodeCache.Add(entry.Key, nodeData)
		cluster.enginesPool.AddEngine(nodeData.IP, nodeData.Name)
	}
}

// OperateContainer is exported
func (cluster *Cluster) OperateContainer(containerid string, action string) (string, *types.OperatedContainers, error) {

	metaData := cluster.configCache.GetMetaDataOfContainer(containerid)
	if metaData == nil {
		return "", nil, ErrClusterContainerNotFound
	}
	operatedContainers, err := cluster.OperateContainers(metaData.MetaID, containerid, action)
	return metaData.MetaID, operatedContainers, err
}

// OperateContainers is exported
// if containerid is empty string so operate metaid's all containers
func (cluster *Cluster) OperateContainers(metaid string, containerid string, action string) (*types.OperatedContainers, error) {

	metaData, engines, err := cluster.validateMetaData(metaid)
	if err != nil {
		logger.ERROR("[#cluster#] %s containers %s error, %s", action, metaid, err.Error())
		return nil, err
	}

	foundContainer := false
	operatedContainers := types.OperatedContainers{}
	for _, engine := range engines {
		if foundContainer {
			break
		}
		containers := engine.Containers(metaData.MetaID)
		for _, container := range containers {
			if containerid == "" || container.Info.ID == containerid {
				var err error
				if engine.IsHealthy() {
					if err = engine.OperateContainer(models.ContainerOperate{Action: action, Container: container.Info.ID}); err != nil {
						logger.ERROR("[#cluster#] engine %s, %s container error:%s", engine.IP, action, err.Error())
					}
				} else {
					err = fmt.Errorf("engine state is %s", engine.State())
				}
				operatedContainers = operatedContainers.SetOperatedPair(engine.IP, engine.Name, container.Info.ID, action, err)
			}
			if container.Info.ID == containerid {
				foundContainer = true
				break
			}
		}
	}
	cluster.hooksProcessor.Hook(metaData, OperateMetaEvent)
	return &operatedContainers, nil
}

// UpgradeContainers is exported
func (cluster *Cluster) UpgradeContainers(metaid string, imagetag string) (*types.UpgradeContainers, error) {

	metaData, engines, err := cluster.validateMetaData(metaid)
	if err != nil {
		logger.ERROR("[#cluster#] upgrade containers %s error, %s", metaid, err.Error())
		return nil, err
	}

	containers := Containers{}
	for _, engine := range engines {
		for _, container := range engine.Containers(metaData.MetaID) {
			containers = append(containers, container)
		}
	}

	upgradeContainers := types.UpgradeContainers{}
	if len(containers) > 0 {
		upgradeCh := make(chan bool)
		cluster.upgraderCache.Upgrade(upgradeCh, metaData.MetaID, imagetag, containers)
		<-upgradeCh
		close(upgradeCh)
		cluster.hooksProcessor.Hook(metaData, UpgradeMetaEvent)
		for _, engine := range engines {
			if engine.IsHealthy() {
				containers := engine.Containers(metaData.MetaID)
				for _, container := range containers {
					upgradeContainers = upgradeContainers.SetUpgradePair(engine.IP, engine.Name, container.Config.Container)
				}
			}
		}
	}
	return &upgradeContainers, nil
}

// RemoveContainer is exported
func (cluster *Cluster) RemoveContainer(containerid string) (string, *types.RemovedContainers, error) {

	metaData := cluster.configCache.GetMetaDataOfContainer(containerid)
	if metaData == nil {
		return "", nil, ErrClusterContainerNotFound
	}
	removedContainers, err := cluster.RemoveContainers(metaData.MetaID, containerid)
	return metaData.MetaID, removedContainers, err
}

// RemoveContainers is exported
// if containerid is empty string so remove metaid's all containers
func (cluster *Cluster) RemoveContainers(metaid string, containerid string) (*types.RemovedContainers, error) {

	metaData, engines, err := cluster.validateMetaData(metaid)
	if err != nil {
		logger.ERROR("[#cluster#] remove containers %s error, %s", metaid, err.Error())
		return nil, err
	}

	foundContainer := false
	removedContainers := types.RemovedContainers{}
	for _, engine := range engines {
		if foundContainer {
			break
		}
		containers := engine.Containers(metaData.MetaID)
		for _, container := range containers {
			if containerid == "" || container.Info.ID == containerid {
				var err error
				if engine.IsHealthy() {
					if err = engine.RemoveContainer(container.Info.ID); err != nil {
						logger.ERROR("[#cluster#] engine %s, remove container error:%s", engine.IP, err.Error())
					}
				} else {
					err = fmt.Errorf("engine state is %s", engine.State())
				}
				removedContainers = removedContainers.SetRemovedPair(engine.IP, engine.Name, container.Info.ID, err)
			}
			if container.Info.ID == containerid {
				foundContainer = true
				break
			}
		}
	}

	cluster.hooksProcessor.Hook(metaData, RemoveMetaEvent)
	if metaData := cluster.configCache.GetMetaData(metaData.MetaID); metaData != nil {
		if len(metaData.BaseConfigs) == 0 {
			cluster.configCache.RemoveMetaData(metaData.MetaID)
		}
	}
	return &removedContainers, nil
}

// RecoveryContainers is exported
func (cluster *Cluster) RecoveryContainers(metaid string) error {

	metaData, engines, err := cluster.validateMetaData(metaid)
	if err != nil {
		logger.ERROR("[#cluster#] recovery containers %s error, %s", metaid, err.Error())
		return err
	}

	if ret := cluster.containsPendingContainers(metaData.GroupID, metaData.Config.Name); ret {
		logger.ERROR("[#cluster#] recovery containers %s error, %s", metaData.MetaID, ErrClusterContainersSetting)
		return ErrClusterContainersSetting
	}

	if len(engines) > 0 {
		baseConfigsCount := cluster.configCache.GetMetaDataBaseConfigsCount(metaData.MetaID)
		if baseConfigsCount != -1 && metaData.Instances != baseConfigsCount {
			if metaData.Instances > baseConfigsCount {
				cluster.createContainers(metaData, metaData.Instances-baseConfigsCount, metaData.Config)
			} else {
				cluster.reduceContainers(metaData, baseConfigsCount-metaData.Instances)
			}
			cluster.hooksProcessor.Hook(metaData, RecoveryMetaEvent)
		}
	}
	return nil
}

// UpdateContainers is exported
func (cluster *Cluster) UpdateContainers(metaid string, instances int, webhooks types.WebHooks) (*types.CreatedContainers, error) {

	if instances <= 0 {
		logger.ERROR("[#cluster#] update containers %s error, %s", metaid, ErrClusterContainersInstancesInvalid)
		return nil, ErrClusterContainersInstancesInvalid
	}

	metaData, engines, err := cluster.validateMetaData(metaid)
	if err != nil {
		logger.ERROR("[#cluster#] update containers %s error, %s", metaid, err.Error())
		return nil, err
	}

	if ret := cluster.containsPendingContainers(metaData.GroupID, metaData.Config.Name); ret {
		logger.ERROR("[#cluster#] update containers %s error, %s", metaData.MetaID, ErrClusterContainersSetting)
		return nil, ErrClusterContainersSetting
	}

	cluster.configCache.SetMetaData(metaid, instances, webhooks)
	if len(engines) > 0 {
		originalInstances := len(metaData.BaseConfigs)
		if originalInstances < instances {
			cluster.createContainers(metaData, instances-originalInstances, metaData.Config)
		} else {
			cluster.reduceContainers(metaData, originalInstances-instances)
		}
	}

	cluster.hooksProcessor.Hook(metaData, UpdateMetaEvent)
	createdContainers := types.CreatedContainers{}
	for _, engine := range engines {
		if engine.IsHealthy() {
			containers := engine.Containers(metaData.MetaID)
			for _, container := range containers {
				createdContainers = createdContainers.SetCreatedPair(engine.IP, engine.Name, container.Config.Container)
			}
		}
	}
	return &createdContainers, nil
}

// CreateContainers is exported
func (cluster *Cluster) CreateContainers(groupid string, instances int, webhooks types.WebHooks, config models.Container) (string, *types.CreatedContainers, error) {

	if instances <= 0 {
		return "", nil, ErrClusterContainersInstancesInvalid
	}

	engines := cluster.GetGroupEngines(groupid)
	if engines == nil {
		logger.ERROR("[#cluster#] create containers error %s : %s", groupid, ErrClusterGroupNotFound)
		return "", nil, ErrClusterGroupNotFound
	}

	if len(engines) == 0 {
		logger.ERROR("[#cluster#] create containers error %s : %s", groupid, ErrClusterNoEngineAvailable)
		return "", nil, ErrClusterNoEngineAvailable
	}

	if ret := cluster.cehckContainerNameUniqueness(groupid, config.Name); !ret {
		logger.ERROR("[#cluster#] create containers error %s : %s", groupid, ErrClusterCreateContainerNameConflict)
		return "", nil, ErrClusterCreateContainerNameConflict
	}

	metaData := cluster.configCache.CreateMetaData(groupid, instances, webhooks, config)
	createdContainers := cluster.createContainers(metaData, instances, config)
	if len(createdContainers) == 0 {
		cluster.configCache.RemoveMetaData(metaData.MetaID)
		return "", nil, ErrClusterCreateContainerFailure
	}
	cluster.hooksProcessor.Hook(metaData, CreateMetaEvent)
	return metaData.MetaID, &createdContainers, nil
}

// reduceContainers is exported
func (cluster *Cluster) reduceContainers(metaData *MetaData, instances int) {

	cluster.Lock()
	cluster.pendingContainers[metaData.Config.Name] = &pendingContainer{
		GroupID: metaData.GroupID,
		Name:    metaData.Config.Name,
		Config:  metaData.Config,
	}
	cluster.Unlock()

	for ; instances > 0; instances-- {
		if _, _, err := cluster.reduceContainer(metaData); err != nil {
			logger.ERROR("[#cluster#] reduce container %s, error:%s", metaData.Config.Name, err.Error())
		}
	}

	cluster.Lock()
	delete(cluster.pendingContainers, metaData.Config.Name)
	cluster.Unlock()
}

// reduceContainer is exported
func (cluster *Cluster) reduceContainer(metaData *MetaData) (*Engine, *Container, error) {

	engines := cluster.GetGroupEngines(metaData.GroupID)
	if engines == nil || len(engines) == 0 {
		return nil, nil, ErrClusterNoEngineAvailable
	}

	reduceEngines := selectReduceEngines(metaData.MetaID, engines)
	if len(reduceEngines) == 0 {
		return nil, nil, ErrClusterNoEngineAvailable
	}

	sort.Sort(reduceEngines)
	reduceEngine := reduceEngines[0]
	engine := reduceEngine.Engine()
	container := reduceEngine.ReduceContainer()
	if err := engine.RemoveContainer(container.Info.ID); err != nil {
		return nil, nil, err
	}
	return engine, container, nil
}

// createContainers is exported
func (cluster *Cluster) createContainers(metaData *MetaData, instances int, config models.Container) types.CreatedContainers {

	cluster.Lock()
	cluster.pendingContainers[config.Name] = &pendingContainer{
		GroupID: metaData.GroupID,
		Name:    config.Name,
		Config:  config,
	}
	cluster.Unlock()

	createdContainers := types.CreatedContainers{}
	ipList := []string{}
	for ; instances > 0; instances-- {
		index := cluster.configCache.MakeContainerIdleIndex(metaData.MetaID)
		if index < 0 {
			continue
		}
		indexStr := strconv.Itoa(index)
		containerConfig := config
		containerConfig.Name = "CLUSTER-" + metaData.GroupID[:8] + "-" + containerConfig.Name + "-" + indexStr
		containerConfig.Env = append(containerConfig.Env, "HUMPBACK_CLUSTER_GROUPID="+metaData.GroupID)
		containerConfig.Env = append(containerConfig.Env, "HUMPBACK_CLUSTER_METAID="+metaData.MetaID)
		containerConfig.Env = append(containerConfig.Env, "HUMPBACK_CLUSTER_CONTAINER_INDEX="+indexStr)
		containerConfig.Env = append(containerConfig.Env, "HUMPBACK_CLUSTER_CONTAINER_ORIGINALNAME="+containerConfig.Name)
		engine, container, err := cluster.createContainer(metaData, ipList, containerConfig)
		if err != nil {
			if err == ErrClusterNoEngineAvailable {
				logger.ERROR("[#cluster#] create container %s, error:%s", containerConfig.Name, err.Error())
				continue
			}
			logger.ERROR("[#cluster#] engine %s, create container %s, error:%s", engine.IP, containerConfig.Name, err.Error())
			ipList = filterAppendIPList(engine, ipList)
			var retries int64
			for ; retries < cluster.createRetry && err != nil; retries++ {
				engine, container, err = cluster.createContainer(metaData, ipList, containerConfig)
				ipList = filterAppendIPList(engine, ipList)
			}
			if err != nil {
				if err == ErrClusterNoEngineAvailable {
					logger.ERROR("[#cluster#] create container %s, error:%s", containerConfig.Name, err.Error())
				} else {
					logger.ERROR("[#cluster#] engine %s, create container %s, error:%s", engine.IP, containerConfig.Name, err.Error())
				}
				continue
			}
		}
		ipList = filterAppendIPList(engine, ipList)
		createdContainers = createdContainers.SetCreatedPair(engine.IP, engine.Name, container.Config.Container)
	}

	cluster.Lock()
	delete(cluster.pendingContainers, config.Name)
	cluster.Unlock()
	return createdContainers
}

// createContainer is exported
func (cluster *Cluster) createContainer(metaData *MetaData, ipList []string, config models.Container) (*Engine, *Container, error) {

	engines := cluster.GetGroupEngines(metaData.GroupID)
	if engines == nil || len(engines) == 0 {
		return nil, nil, ErrClusterNoEngineAvailable
	}

	selectEngines := cluster.selectEngines(engines, ipList, config)
	if len(selectEngines) == 0 {
		return nil, nil, ErrClusterNoEngineAvailable
	}

	engine := selectEngines[0]
	container, err := engine.CreateContainer(config)
	if err != nil {
		return engine, nil, err
	}
	return engine, container, nil
}

// selectEngines is exported
func (cluster *Cluster) selectEngines(engines []*Engine, ipList []string, config models.Container) []*Engine {

	selectEngines := []*Engine{}
	for _, engine := range engines {
		if engine.IsHealthy() {
			selectEngines = append(selectEngines, engine)
		}
	}

	if len(selectEngines) == 0 {
		return selectEngines //return empty engines
	}

	weightedEngines := selectWeightdEngines(selectEngines, config)
	if len(weightedEngines) > 0 {
		sort.Sort(weightedEngines)
		selectEngines = weightedEngines.Engines()
	}

	if len(selectEngines) > 0 && len(ipList) > 0 {
		filterEngines := filterIPList(selectEngines, ipList)
		if len(filterEngines) > 0 {
			selectEngines = filterEngines
		} else {
			for i := len(selectEngines) - 1; i > 0; i-- {
				j := cluster.randSeed.Intn(i + 1)
				selectEngines[i], selectEngines[j] = selectEngines[j], selectEngines[i]
			}
		}
	}
	return selectEngines
}

// containsPendingContainers is exported
func (cluster *Cluster) containsPendingContainers(groupid string, name string) bool {

	cluster.RLock()
	defer cluster.RUnlock()
	for _, pendingContainer := range cluster.pendingContainers {
		if pendingContainer.GroupID == groupid && pendingContainer.Name == name {
			return true
		}
	}
	return false
}

// cehckContainerNameUniqueness is exported
func (cluster *Cluster) cehckContainerNameUniqueness(groupid string, name string) bool {

	if ret := cluster.containsPendingContainers(groupid, name); ret {
		return false
	}

	metaData := cluster.configCache.GetMetaDataOfName(groupid, name)
	if metaData != nil {
		return false
	}
	return true
}

// validateMetaData is exported
func (cluster *Cluster) validateMetaData(metaid string) (*MetaData, []*Engine, error) {

	if ret := cluster.upgraderCache.Contains(metaid); ret {
		return nil, nil, ErrClusterContainersUpgrading
	}

	if ret := cluster.migtatorCache.Contains(metaid); ret {
		return nil, nil, ErrClusterContainersMigrating
	}

	metaData, engines, err := cluster.GetMetaDataEngines(metaid)
	if err != nil {
		return nil, nil, err
	}
	return metaData, engines, nil
}
