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
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

// pendingContainer is exported
type pendingContainer struct {
	GroupID string
	Name    string
	Config  models.Container
}

// Group is exported
// Servers: cluster server ips, correspond engines's key.
// Owners: cluster group owners name.
type Group struct {
	ID      string
	Servers []string
	Owners  []string
}

// Cluster is exported
// Discovery: discovery service
// overcommitRatio: engine overcommit ratio, cpus & momery resources.
// createRetry: create container retry count.
// configCache: container config metadata cache.
// upgradeContainers: upgrading containers.
// pendingContainers: createing containers.
// engines: map[ip]*Engine
// groups:  map[groupid]*Group
type Cluster struct {
	sync.RWMutex
	Discovery *discovery.Discovery

	overcommitRatio   float64
	createRetry       int64
	randSeed          *rand.Rand
	configCache       *ContainersConfigCache
	upgradeContainers *UpgradeContainers
	pendingContainers map[string]*pendingContainer
	engines           map[string]*Engine
	groups            map[string]*Group
	stopCh            chan struct{}
}

// NewCluster is exported
// Make cluster object, set options and discovery service.
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

	cacheRoot := ""
	if val, ret := driverOpts.String("cacheroot", ""); ret {
		cacheRoot = val
	}

	upgradedelay := 10 * time.Second
	if val, ret := driverOpts.String("upgradedelay", ""); ret {
		if dur, err := time.ParseDuration(val); err == nil {
			upgradedelay = dur
		}
	}

	configCache, err := NewContainersConfigCache(cacheRoot)
	if err != nil {
		return nil, err
	}

	return &Cluster{
		Discovery:         discovery,
		overcommitRatio:   overcommitratio,
		createRetry:       createretry,
		randSeed:          rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
		configCache:       configCache,
		upgradeContainers: NewUpgradeContainers(upgradedelay, configCache),
		pendingContainers: make(map[string]*pendingContainer),
		engines:           make(map[string]*Engine),
		groups:            make(map[string]*Group),
		stopCh:            make(chan struct{}),
	}, nil
}

// Start is exported
// Cluster start, open discovery service
func (cluster *Cluster) Start() error {

	cluster.configCache.Init()
	logger.INFO("[#cluster#] discovery service watching...")
	if cluster.Discovery != nil {
		cluster.Discovery.Watch(cluster.stopCh, cluster.watchDiscoveryHandleFunc)
		return nil
	}
	return ErrClusterDiscoveryInvalid
}

// Stop is exported
// Cluster stop, close discovery service
func (cluster *Cluster) Stop() {

	close(cluster.stopCh)
	logger.INFO("[#cluster#] discovery service closed.")
}

// GroupsEngineContains is exported
func (cluster *Cluster) GroupsEngineContains(engine *Engine) bool {

	ret := false
	groups := cluster.GetGroups()
	for _, group := range groups {
		if !ret {
			for _, ip := range group.Servers {
				if ip == engine.IP {
					ret = true
					break
				}
			}
		}
	}
	return ret
}

// GetMetaDataEngines is exported
func (cluster *Cluster) GetMetaDataEngines(metaid string) (*MetaData, []*Engine, error) {

	metaData := cluster.configCache.GetMetaData(metaid)
	if metaData == nil {
		return nil, nil, ErrClusterMetaDataNotFound
	}

	engines := cluster.GetGroupEngines(metaData.GroupID)
	if engines == nil {
		return nil, nil, ErrClusterGroupNotFound
	}
	return metaData, engines, nil
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

// GetGroupEngines is exported
func (cluster *Cluster) GetGroupEngines(groupid string) []*Engine {

	cluster.RLock()
	defer cluster.RUnlock()
	engines := []*Engine{}
	group, ret := cluster.groups[groupid]
	if !ret {
		return nil
	}
	for _, ip := range group.Servers {
		if engine, ret := cluster.engines[ip]; ret {
			engines = append(engines, engine)
		}
	}
	return engines
}

// GetGroupContainers is exported
func (cluster *Cluster) GetGroupContainers(groupid string) *types.GroupContainers {

	cluster.RLock()
	defer cluster.RUnlock()
	if _, ret := cluster.groups[groupid]; !ret {
		return nil
	}

	groupContainers := types.GroupContainers{}
	engines := cluster.GetGroupEngines(groupid)
	groupMetaData := cluster.configCache.GetGroupMetaData(groupid)
	for _, metaData := range groupMetaData {
		groupContainer := &types.GroupContainer{
			MetaID:     metaData.MetaID,
			Config:     metaData.Config,
			Containers: make([]*types.EngineContainer, 0),
		}
		for _, engine := range engines {
			containers := engine.Containers(metaData.MetaID)
			for _, container := range containers {
				groupContainer.Containers = append(groupContainer.Containers, &types.EngineContainer{
					IP:        engine.IP,
					HostName:  engine.Name,
					Container: container.Config.Container,
				})
			}
		}
		groupContainers = append(groupContainers, groupContainer)
	}
	return &groupContainers
}

// GetGroups is exported
func (cluster *Cluster) GetGroups() []*Group {

	groups := []*Group{}
	cluster.RLock()
	for _, group := range cluster.groups {
		groups = append(groups, group)
	}
	cluster.RUnlock()
	return groups
}

// SetGroup is exported
func (cluster *Cluster) SetGroup(groupid string, servers []string, owners []string) {

	cluster.Lock()
	removes := []string{}
	group, ret := cluster.groups[groupid]
	if !ret {
		group = &Group{ID: groupid, Servers: servers, Owners: owners}
		cluster.groups[groupid] = group
		logger.INFO("[#cluster#] created group %s(%d)", groupid, len(servers))
	} else {
		origin := group.Servers
		group.Servers = servers
		group.Owners = owners
		for _, oldip := range origin {
			found := false
			for _, newip := range group.Servers {
				if oldip == newip {
					found = true
					break
				}
			}
			if !found {
				removes = append(removes, oldip)
			}
		}
		logger.INFO("[#cluster#] changed group %s(%d)", groupid, len(servers))
	}
	cluster.Unlock()

	for _, ip := range servers {
		cluster.addEngine(ip)
	}

	for _, ip := range removes {
		cluster.removeEngine(ip)
	}
}

// RemoveGroup is exported
func (cluster *Cluster) RemoveGroup(groupid string) bool {

	cluster.Lock()
	defer cluster.Unlock()
	group, ret := cluster.groups[groupid]
	if !ret {
		logger.WARN("[#cluster#] remove group %s not found.", groupid)
		return false
	}
	logger.INFO("[#cluster#] removed group %s(%d)", groupid, len(group.Servers))
	delete(cluster.groups, groupid)
	return true
}

func (cluster *Cluster) watchDiscoveryHandleFunc(added backends.Entries, removed backends.Entries, err error) {

	if err != nil {
		logger.ERROR("[#cluster#] discovery service watch error:%s", err.Error())
		return
	}

	logger.INFO("[#cluster#] discovery service handlefunc removed:%d added:%d.", len(removed), len(added))
	opts := &types.ClusterRegistOptions{}
	for _, entry := range removed {
		if err := json.DeCodeBufferToObject(entry.Data, opts); err != nil {
			logger.ERROR("[#cluster#] discovery service handlefunc removed decode error: %s", err.Error())
			continue
		}
		ip, _, err := net.SplitHostPort(opts.Addr)
		if err != nil {
			logger.ERROR("[#cluster#] discovery service handlefunc removed resolve addr error: %s", err.Error())
			continue
		}
		logger.INFO("[#cluster#] discovery service removed:%s %s.", entry.Key, opts.Addr)
		if engine := cluster.removeEngine(ip); engine != nil {
			//containers := engine.Containers()
			engine.Close()
			//cluster.migrateContainers(engine, containers)
			logger.INFO("[#cluster#] set engine %s %s", engine.IP, engine.State())
		}
	}

	for _, entry := range added {
		if err := json.DeCodeBufferToObject(entry.Data, opts); err != nil {
			logger.ERROR("[#cluster#] discovery service handlefunc added decode error: %s", err.Error())
			continue
		}
		ip, _, err := net.SplitHostPort(opts.Addr)
		if err != nil {
			logger.ERROR("[#cluster#] discovery service handlefunc added resolve addr error: %s", err.Error())
			continue
		}
		logger.INFO("[#cluster#] discovery service added:%s.", entry.Key)
		if engine := cluster.addEngine(ip); engine != nil {
			engine.Open(opts.Addr)
			//cluster.recoverContainers(engine)
			logger.INFO("[#cluster#] set engine %s %s", engine.IP, engine.State())
		}
	}
}

func (cluster *Cluster) addEngine(ip string) *Engine {

	engine := cluster.GetEngine(ip)
	if engine == nil {
		var err error
		if engine, err = NewEngine(ip, cluster.overcommitRatio, cluster.configCache); err != nil {
			logger.ERROR("[#cluster#] add engine %s error:%s", ip, err.Error())
			return nil
		}
		cluster.Lock()
		cluster.engines[ip] = engine
		cluster.Unlock()
		logger.INFO("[#cluster#] add engine %s", ip)
	}
	return engine
}

func (cluster *Cluster) removeEngine(ip string) *Engine {

	engine := cluster.GetEngine(ip)
	if engine == nil {
		logger.WARN("[#cluster#] remove engine, not found:%s", ip)
		return nil
	}

	if ret := cluster.GroupsEngineContains(engine); !ret {
		cluster.Lock()
		delete(cluster.engines, engine.IP)
		cluster.Unlock()
		logger.INFO("[#cluster#] remove engine %s", engine.IP)
	}
	return engine
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
				operatedContainers = operatedContainers.SetOperatedPair(engine.IP, container.Info.ID, action, err)
			}
			if container.Info.ID == containerid {
				foundContainer = true
				break
			}
		}
	}
	return &operatedContainers, nil
}

// UpgradeContainers is exported
func (cluster *Cluster) UpgradeContainers(metaid string, imagetag string) error {

	metaData, engines, err := cluster.validateMetaData(metaid)
	if err != nil {
		logger.ERROR("[#cluster#] upgrade containers %s error, %s", metaid, err.Error())
		return err
	}

	upgradecontainers := Containers{}
	for _, engine := range engines {
		containers := engine.Containers(metaData.MetaID)
		for _, container := range containers {
			upgradecontainers = append(upgradecontainers, container)
		}
	}

	if len(upgradecontainers) > 0 {
		cluster.upgradeContainers.Upgrade(metaData.MetaID, imagetag, upgradecontainers)
	}
	return nil
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
				removedContainers = removedContainers.SetRemovedPair(engine.IP, container.Info.ID, err)
			}
			if container.Info.ID == containerid {
				foundContainer = true
				break
			}
		}
	}

	if metaData := cluster.configCache.GetMetaData(metaData.MetaID); metaData != nil {
		if len(metaData.BaseConfigs) == 0 {
			cluster.configCache.RemoveMetaData(metaData.MetaID)
		}
	}
	return &removedContainers, nil
}

// SetContainers is exported
func (cluster *Cluster) SetContainers(metaid string, instances int) (*types.CreatedContainers, error) {

	if instances <= 0 {
		return nil, ErrClusterContainersInstancesInvalid
	}

	metaData, engines, err := cluster.validateMetaData(metaid)
	if err != nil {
		logger.ERROR("[#cluster#] set containers %s error, %s", metaid, err.Error())
		return nil, err
	}

	if len(engines) == 0 {
		logger.ERROR("[#cluster#] set containers error %s : %s", metaData.GroupID, ErrClusterNoEngineAvailable)
		return nil, ErrClusterNoEngineAvailable
	}

	originalInstances := len(metaData.BaseConfigs)
	if originalInstances == instances {
		return nil, ErrClusterContainersInstancesNoChange
	}

	if ret := cluster.containsPendingContainers(metaData.GroupID, metaData.Config.Name); ret {
		return nil, ErrClusterContainersSetting
	}

	if originalInstances < instances {
		createdContainers := cluster.createContainers(metaData, originalInstances+1, instances-originalInstances, metaData.Config)
		if len(createdContainers) == 0 {
			return nil, ErrClusterContainersInstancesNoChange
		}
	} else {
		cluster.reduceContainers(metaData, engines, originalInstances-instances)
	}

	createdContainers := types.CreatedContainers{}
	for _, engine := range engines {
		containers := engine.Containers(metaData.MetaID)
		for _, container := range containers {
			createdContainers = createdContainers.SetCreatedPair(engine.IP, container.Config.Container)
		}
	}
	return &createdContainers, nil
}

// CreateContainers is exported
func (cluster *Cluster) CreateContainers(groupid string, instances int, config models.Container) (string, *types.CreatedContainers, error) {

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

	metaData := cluster.configCache.CreateMetaData(groupid, config)
	createdContainers := cluster.createContainers(metaData, 1, instances, config)
	if len(createdContainers) == 0 {
		cluster.configCache.RemoveMetaData(metaData.MetaID)
		return "", nil, ErrClusterCreateContainerFailure
	}
	return metaData.MetaID, &createdContainers, nil
}

// reduceContainers is exported
func (cluster *Cluster) reduceContainers(metaData *MetaData, engines []*Engine, instances int) {

	cluster.Lock()
	cluster.pendingContainers[metaData.Config.Name] = &pendingContainer{
		GroupID: metaData.GroupID,
		Name:    metaData.Config.Name,
		Config:  metaData.Config,
	}
	cluster.Unlock()

	containers := reduceContainers{}
	for _, engine := range engines {
		if engine.IsHealthy() {
			containers = append(containers, engine.Containers(metaData.MetaID)...)
		}
	}

	sort.Sort(containers)
	for i := 0; i < len(containers) && i < instances; i++ {
		engine := containers[i].Engine
		if err := engine.RemoveContainer(containers[i].Info.ID); err != nil {
			logger.ERROR("[#cluster#] engine %s, remove container error:%s", engine.IP, err.Error())
			break
		}
	}

	cluster.Lock()
	delete(cluster.pendingContainers, metaData.Config.Name)
	cluster.Unlock()
}

// createContainers is exported
func (cluster *Cluster) createContainers(metaData *MetaData, index int, instances int, config models.Container) types.CreatedContainers {

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
		indexStr := strconv.Itoa(index)
		containerConfig := config
		containerConfig.Name = metaData.GroupID[:8] + "-" + containerConfig.Name + "-" + indexStr
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
		createdContainers = createdContainers.SetCreatedPair(engine.IP, container.Config.Container)
		index++
	}

	cluster.Lock()
	delete(cluster.pendingContainers, config.Name)
	cluster.Unlock()
	return createdContainers
}

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

	weightedEngines := weightEngines(selectEngines, config)
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

func (cluster *Cluster) cehckContainerNameUniqueness(groupid string, name string) bool {

	if ret := cluster.containsPendingContainers(groupid, name); ret {
		return false
	}

	metaData := cluster.configCache.GetMetaDataOfName(name)
	if metaData != nil && metaData.GroupID == groupid {
		return false
	}
	return true
}

func (cluster *Cluster) validateMetaData(metaid string) (*MetaData, []*Engine, error) {

	if ret := cluster.upgradeContainers.Contains(metaid); ret {
		return nil, nil, ErrClusterContainersUpgrading
	}

	metaData, engines, err := cluster.GetMetaDataEngines(metaid)
	if err != nil {
		return nil, nil, err
	}
	return metaData, engines, nil
}

/*
func (cluster *Cluster) migrateContainers(engine *Engine, containers Containers) {

	//分配的原始engine记录在baseConfig中，当原始engine启动后，需删除上次迁移成功的engine上的容器.
	//非原始engine启动，需要还原？或不做迁移,防止抖动.
	//迁移流程会触发多协程并发掉线情况，可考虑做一个迁移任务池，逐步处理迁移工作.
	// add to container pool, 迁移池考虑按groupid分组.
	// go migrate containers.
	// 掉线节点考虑将容器加入池中，上线节点先启动本地，然后若池中存在则删除，放弃迁移.
	logger.INFO("[#cluster#] migrate containers %s", engine.IP)
	for _, container := range containers {
		if container.BaseConfig == nil {
			continue
		}
		groupid := container.BaseConfig.GroupID
		engines := cluster.GetGroupEngines(groupid)
		if engines == nil || len(engines) == 0 {
			continue
		}
		ipList := []string{}
		_, c, err := cluster.createContainer(engines, ipList, *container.BaseConfig.Config)
		if err != nil {
			continue
		}
		if err := cluster.configCache.Set(c.ID, groupid, container.BaseConfig.IP, container.BaseConfig.Name, container.BaseConfig.Config); err != nil {
			logger.INFO("[#cluster#] config cache error:%s", err.Error())
		}
		logger.INFO("[#cluster#] migrate successed.")
	}
}

func (cluster *Cluster) recoverContainers(engine *Engine) {

	logger.INFO("[#cluster#] recover containers %s", engine.IP)
	time.Sleep(time.Second * 3)
	containers := engine.Containers()
	for _, container := range containers {
		if container.BaseConfig == nil {
			continue
		}

		engines := cluster.GetGroupEngines(container.BaseConfig.GroupID)
		if engines == nil || len(engines) == 0 {
			continue
		}

		for _, e := range engines {
			if e.IP != engine.IP {
				cs := e.Containers()
				for _, c := range cs {
					if c.BaseConfig != nil && c.BaseConfig.Config.Name == container.BaseConfig.Config.Name {
						if container.BaseConfig.IP == engine.IP {
							e.RemoveContainer(c.BaseConfig.ID, c.BaseConfig.Config.Name)
							cluster.configCache.Remove(c.BaseConfig.ID)
						} else {
							engine.RemoveContainer(container.BaseConfig.ID, container.BaseConfig.Name)
							cluster.configCache.Remove(container.BaseConfig.ID)
						}

					}
				}
			}
		}
	}
	logger.INFO("[#cluster#] recover containers successed")
}
*/
