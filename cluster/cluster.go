package cluster

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/humpback/common/models"
	"github.com/humpback/discovery"
	"github.com/humpback/discovery/backends"
	"github.com/humpback/gounits/logger"
	"github.com/humpback/gounits/system"
	"github.com/humpback/humpback-center/cluster/storage"
	"github.com/humpback/humpback-center/cluster/types"
	"github.com/humpback/humpback-center/notify"
)

var pendingWaitForInterval = time.Duration(time.Second * 5)

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
	Name        string   `json:"Name"`
	IsCluster   bool     `json:"IsCluster"`
	Location    string   `json:"ClusterLocation"`
	Servers     []Server `json:"Servers"`
	ContactInfo string   `json:"ContactInfo"`
}

// Cluster is exported
type Cluster struct {
	sync.RWMutex
	Location          string
	NotifySender      *notify.NotifySender
	Discovery         *discovery.Discovery
	overcommitRatio   float64
	createRetry       int64
	removeDelay       time.Duration
	recoveryInterval  time.Duration
	randSeed          *rand.Rand
	nodeCache         *types.NodeCache
	configCache       *ContainersConfigCache
	migtatorCache     *MigrateContainersCache
	enginesPool       *EnginesPool
	hooksProcessor    *HooksProcessor
	storageDriver     *storage.DataStorage
	pendingContainers map[string]*pendingContainer
	engines           map[string]*Engine
	groups            map[string]*Group
	stopCh            chan struct{}
}

// NewCluster is exported
func NewCluster(driverOpts system.DriverOpts, notifySender *notify.NotifySender, discovery *discovery.Discovery) (*Cluster, error) {

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

	removedelay := time.Duration(0)
	if val, ret := driverOpts.String("removedelay", ""); ret {
		if dur, err := time.ParseDuration(val); err == nil {
			removedelay = dur
		}
	}

	migratedelay := 30 * time.Second
	if val, ret := driverOpts.String("migratedelay", ""); ret {
		if dur, err := time.ParseDuration(val); err == nil {
			migratedelay = dur
		}
	}

	recoveryInterval := 150 * time.Second
	if val, ret := driverOpts.String("recoveryinterval", ""); ret {
		if dur, err := time.ParseDuration(val); err == nil {
			recoveryInterval = dur
		}
	}

	clusterLocation := ""
	if val, ret := driverOpts.String("location", ""); ret {
		clusterLocation = strings.TrimSpace(val)
	}

	cacheRoot := ""
	if val, ret := driverOpts.String("cacheroot", ""); ret {
		cacheRoot = val
	}

	enginesPool := NewEnginesPool()
	migrateContainersCache := NewMigrateContainersCache(migratedelay)
	configCache, err := NewContainersConfigCache(cacheRoot)
	if err != nil {
		return nil, err
	}

	dataPath := "./data"
	if val, ret := driverOpts.String("datapath", ""); ret {
		dataPath = strings.TrimSpace(val)
	}

	storageDriver, err := storage.NewDataStorage(dataPath)
	if err != nil {
		return nil, err
	}

	cluster := &Cluster{
		Location:          clusterLocation,
		NotifySender:      notifySender,
		Discovery:         discovery,
		overcommitRatio:   overcommitratio,
		createRetry:       createretry,
		removeDelay:       removedelay,
		recoveryInterval:  recoveryInterval,
		randSeed:          rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
		nodeCache:         types.NewNodeCache(),
		configCache:       configCache,
		migtatorCache:     migrateContainersCache,
		enginesPool:       enginesPool,
		hooksProcessor:    NewHooksProcessor(),
		storageDriver:     storageDriver,
		pendingContainers: make(map[string]*pendingContainer),
		engines:           make(map[string]*Engine),
		groups:            make(map[string]*Group),
		stopCh:            make(chan struct{}),
	}

	enginesPool.SetCluster(cluster)
	migrateContainersCache.SetCluster(cluster)
	return cluster, nil
}

// Start is exported
// Cluster start, init container config cache watch open discovery service
func (cluster *Cluster) Start() error {

	if err := cluster.storageDriver.Open(); err != nil {
		return err
	}

	cluster.configCache.Init()
	if cluster.Discovery != nil {
		if cluster.Location != "" {
			logger.INFO("[#cluster#] cluster location: %s", cluster.Location)
		}
		logger.INFO("[#cluster#] discovery service watching...")
		cluster.Discovery.WatchNodes(cluster.stopCh, cluster.watchDiscoveryHandleFunc)
		cluster.hooksProcessor.Start()
		go cluster.recoveryContainersLoop()
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
	cluster.hooksProcessor.Close()
	cluster.storageDriver.Close()
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

// GetGroups is exported
func (cluster *Cluster) GetGroups() []*Group {

	cluster.RLock()
	defer cluster.RUnlock()
	groups := []*Group{}
	for _, group := range cluster.groups {
		groups = append(groups, group)
	}
	return groups
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

// GetServerOfEngines is exported
func (cluster *Cluster) GetServerOfEngines(server Server) *Engine {

	engine := searchServerOfEngines(server, cluster.engines)
	if engine == nil {
		engine = searchServerOfStorage(server, cluster.storageDriver.NodeStorage)
		if engine == nil {
			engine = &Engine{
				Name:         server.Name,
				IP:           server.IP,
				EngineLabels: map[string]string{},
				NodeLabels:   map[string]string{},
			}
		}
		engine.StateText = stateText[StateDisconnected]
		engine.state = StateDisconnected
	}
	return engine
}

// GetGroupAllEngines is exported
// Returns all engine under group and contains offline (cluster engines not exists.)
func (cluster *Cluster) GetGroupAllEngines(groupid string) []*Engine {

	cluster.RLock()
	defer cluster.RUnlock()
	group, ret := cluster.groups[groupid]
	if !ret {
		return nil
	}

	engines := []*Engine{}
	for _, server := range group.Servers {
		engine := searchServerOfEngines(server, cluster.engines)
		if engine == nil {
			engine = searchServerOfStorage(server, cluster.storageDriver.NodeStorage)
			if engine == nil {
				engine = &Engine{
					Name:         server.Name,
					IP:           server.IP,
					EngineLabels: map[string]string{},
					NodeLabels:   map[string]string{},
				}
			}
			engine.StateText = stateText[StateDisconnected]
			engine.state = StateDisconnected
		}
		engines = append(engines, engine)
	}
	engines = removeDuplicatesEngines(engines)
	return engines
}

// GetGroupEngines is exported
// Returns pairs engine under group and cluster engines is exists
func (cluster *Cluster) GetGroupEngines(groupid string) []*Engine {

	cluster.RLock()
	defer cluster.RUnlock()
	group, ret := cluster.groups[groupid]
	if !ret {
		return nil
	}

	engines := []*Engine{}
	for _, server := range group.Servers {
		if engine := searchServerOfEngines(server, cluster.engines); engine != nil {
			engines = append(engines, engine)
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

// GetMetaEnginesContainers is exported
func (cluster *Cluster) GetMetaEnginesContainers(metaData *MetaData, metaEngines map[string]*Engine) *types.GroupContainer {

	groupContainer := &types.GroupContainer{
		GroupID:       metaData.GroupID,
		MetaID:        metaData.MetaID,
		IsRemoveDelay: metaData.IsRemoveDelay,
		IsRecovery:    metaData.IsRecovery,
		Instances:     metaData.Instances,
		Placement:     metaData.Placement,
		WebHooks:      metaData.WebHooks,
		Config:        metaData.Config,
		Containers:    make([]*types.EngineContainer, 0),
		CreateAt:      metaData.CreateAt,
		LastUpdateAt:  metaData.LastUpdateAt,
	}

	baseConfigs := cluster.configCache.GetMetaDataBaseConfigs(metaData.MetaID)
	for _, baseConfig := range baseConfigs {
		for _, engine := range metaEngines {
			if engine.IsHealthy() && engine.HasMeta(metaData.MetaID) {
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

// RefreshEnginesContainers is exported
func (cluster *Cluster) RefreshEnginesContainers(metaEngines map[string]*Engine) {

	waitGroup := sync.WaitGroup{}
	for _, engine := range metaEngines {
		if engine.IsHealthy() {
			waitGroup.Add(1)
			go func(e *Engine) {
				e.RefreshContainers()
				waitGroup.Done()
			}(engine)
		}
	}
	waitGroup.Wait()
}

// GetGroupAllContainers is exported
func (cluster *Cluster) GetGroupAllContainers(groupid string) *types.GroupContainers {

	metaEngines := make(map[string]*Engine)
	groupMetaData := cluster.configCache.GetGroupMetaData(groupid)
	for _, metaData := range groupMetaData {
		if _, engines, err := cluster.GetMetaDataEngines(metaData.MetaID); err == nil {
			for _, engine := range engines {
				if engine.IsHealthy() && engine.HasMeta(metaData.MetaID) {
					metaEngines[engine.IP] = engine
				}
			}
		}
	}

	//cluster.RefreshEnginesContainers(metaEngines)
	groupContainers := types.GroupContainers{}
	for _, metaData := range groupMetaData {
		if groupContainer := cluster.GetMetaEnginesContainers(metaData, metaEngines); groupContainer != nil {
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

	metaEngines := make(map[string]*Engine)
	for _, engine := range engines {
		if engine.IsHealthy() && engine.HasMeta(metaid) {
			metaEngines[engine.IP] = engine
		}
	}
	//cluster.RefreshEnginesContainers(metaEngines)
	return cluster.GetMetaEnginesContainers(metaData, metaEngines)
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
		logger.INFO("[#cluster#] group created %s %s (%d)", pGroup.ID, pGroup.Name, len(pGroup.Servers))
		for _, server := range pGroup.Servers {
			ipOrName := selectIPOrName(server.IP, server.Name)
			if nodeData := cluster.nodeCache.Get(ipOrName); nodeData != nil {
				addServers = append(addServers, server)
			}
		}
	} else {
		origins := pGroup.Servers
		pGroup.Name = group.Name
		pGroup.Location = group.Location
		pGroup.Servers = group.Servers
		pGroup.IsCluster = group.IsCluster
		pGroup.ContactInfo = group.ContactInfo
		logger.INFO("[#cluster#] group changed %s %s (%d)", pGroup.ID, pGroup.Name, len(pGroup.Servers))
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
				logger.INFO("[#cluster#] group %s remove server to pendengines %s\t%s", pGroup.ID, server.IP, server.Name)
				cluster.enginesPool.RemoveEngine(server.IP, server.Name)
			} else {
				// after recovery containers, need to clear migrator cache of meta container ?
				// Migrator StartEngineContainers(groupid, engine)... ?
			}
		}
	}

	for _, server := range addServers {
		logger.INFO("[#cluster#] group %s append server to pendengines %s\t%s", pGroup.ID, server.IP, server.Name)
		cluster.enginesPool.AddEngine(server.IP, server.Name)
		/*
			if cluster is engine exists ? {
				// Migrator CancelEngineContainers(groupid, engine)... // add to group, this group migrator cancel.
			}
		*/
	}
}

// RemoveGroup is exported
func (cluster *Cluster) RemoveGroup(groupid string) bool {

	engines := cluster.GetGroupEngines(groupid)
	if engines == nil {
		logger.WARN("[#cluster#] remove group %s not found.", groupid)
		return false
	}

	// remove group migrator's all meta.
	cluster.migtatorCache.RemoveGroup(groupid)
	// get group all metaData and clean metaData containers.
	wgroup := sync.WaitGroup{}
	groupMetaData := cluster.configCache.GetGroupMetaData(groupid)
	for _, metaData := range groupMetaData {
		wgroup.Add(1)
		go func(mdata *MetaData) {
			cluster.removeContainers(mdata, "")
			cluster.configCache.RemoveMetaData(mdata.MetaID)
			cluster.submitHookEvent(mdata, RemoveMetaEvent)
			wgroup.Done()
		}(metaData)
	}
	wgroup.Wait()

	// remove metadata and group to cluster.
	cluster.configCache.RemoveGroupMetaData(groupid)
	cluster.Lock()
	delete(cluster.groups, groupid) // remove group
	logger.INFO("[#cluster#] removed group %s", groupid)
	cluster.Unlock()

	// remove engine to engines pool.
	for _, engine := range engines {
		if engine.IsHealthy() {
			if ret := cluster.InGroupsContains(engine.IP, engine.Name); !ret {
				// engine does not belong to the any groups, remove to cluster.
				logger.INFO("[#cluster#] group %s remove server to pendengines %s\t%s", groupid, engine.IP, engine.Name)
				cluster.enginesPool.RemoveEngine(engine.IP, engine.Name)
			}
		}
	}
	return true
}

//SetServerNodeLabels is exported
func (cluster *Cluster) SetServerNodeLabels(server Server, labels map[string]string) error {

	engine := searchServerOfStorage(server, cluster.storageDriver.NodeStorage)
	if engine == nil {
		return ErrClusterServerNotFound
	}

	err := cluster.storageDriver.NodeStorage.SetNodeLabels(engine.IP, labels)
	if err == nil {
		//update engine node-labels in memory
		if engine = cluster.GetEngine(engine.IP); engine != nil {
			originalLabels := engine.NodeLabelsPairs()
			if !reflect.DeepEqual(originalLabels, labels) {
				logger.INFO("[#cluster#] set %s(%s) node-labels, %+v", engine.IP, engine.Name, labels)
				engine.SetNodeLabelsPairs(labels)
				metaids := engine.MetaIds()
				for _, metaid := range metaids {
					if metaData := cluster.GetMetaData(metaid); metaData != nil {
						if metaData.Placement.Constraints != nil && len(metaData.Placement.Constraints) > 0 {
							logger.INFO("[#cluster#] meta %s enable available nodes changed.", metaid)
							cluster.configCache.SetAvailableNodesChanged(metaid, true)
						}
					}
				}
			}
		}
	}
	return err
}

func (cluster *Cluster) watchDiscoveryHandleFunc(added backends.Entries, removed backends.Entries, err error) {

	if err != nil {
		logger.ERROR("[#cluster#] discovery watch error:%s", err.Error())
		return
	}

	if len(added) == 0 && len(removed) == 0 {
		return
	}

	watchEngines := WatchEngines{}
	logger.INFO("[#cluster#] discovery watch removed:%d added:%d.", len(removed), len(added))
	for _, entry := range removed {
		nodeData, err := deCodeEntry(entry)
		if err != nil {
			logger.ERROR("[#cluster#] discovery watch removed decode error: %s", err.Error())
			continue
		}
		logger.INFO("[#cluster#] discovery watch, remove to pendengines %s\t%s", nodeData.IP, nodeData.Name)

		if cluster.nodeCache.ContainsOtherKey(entry.Key, nodeData.IP) == false {
			watchEngines = append(watchEngines, NewWatchEngine(nodeData.IP, nodeData.Name, StateDisconnected))
			cluster.enginesPool.RemoveEngine(nodeData.IP, nodeData.Name)
		}
		cluster.nodeCache.Remove(entry.Key)
	}

	for _, entry := range added {
		nodeData, err := deCodeEntry(entry)
		if err != nil {
			logger.ERROR("[#cluster#] discovery service watch added decode error: %s", err.Error())
			continue
		}
		logger.INFO("[#cluster#] discovery watch, append to pendengines %s\t%s", nodeData.IP, nodeData.Name)
		watchEngines = append(watchEngines, NewWatchEngine(nodeData.IP, nodeData.Name, StateHealthy))
		cluster.nodeCache.Add(entry.Key, nodeData)
		cluster.enginesPool.AddEngine(nodeData.IP, nodeData.Name)
	}
	cluster.NotifyGroupEnginesWatchEvent("cluster discovery some engines state changed.", watchEngines)
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
		logger.ERROR("[#cluster#] %s meta %s error, %s", action, metaid, err.Error())
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
	cluster.submitHookEvent(metaData, OperateMetaEvent)
	return &operatedContainers, nil
}

func (cluster *Cluster) actionContainers(action string, engineContainers map[string]*Engine) (string, error) {

	for container, engine := range engineContainers {
		if engine != nil {
			operate := models.ContainerOperate{Action: action, Container: container}
			if err := engine.OperateContainer(operate); err != nil {
				return container, err
			}
		}
	}
	return "", nil
}

func (cluster *Cluster) upgradeContainers(metaData *MetaData, engines []*Engine, config models.Container) (*types.UpgradeContainers, error) {

	priorities := NewEnginePriorities(metaData, engines)
	logger.INFO("[#cluster#] upgrade %s containers, priorities %s", config.Name, priorities.EngineStrings())
	engineContainers := map[string]*Engine{}
	for _, baseConfig := range metaData.BaseConfigs {
		var e *Engine
		for _, engine := range engines {
			if engine.IsHealthy() && engine.HasContainer(baseConfig.ID) {
				e = engine
				break
			}
		}
		engineContainers[baseConfig.ID] = e
	}

	afterStop := false
	if config.NetworkMode != "bridge" && config.NetworkMode != "nat" {
		afterStop = true
	}

	if !afterStop {
		for _, portBindings := range metaData.Config.Ports {
			if portBindings.PublicPort != 0 {
				afterStop = true
				break
			}
		}
	}

	if afterStop { //after stop old tag containers.
		if containerid, err := cluster.actionContainers("stop", engineContainers); err != nil {
			cluster.configCache.RemoveContainerBaseConfig(metaData.MetaID, containerid)
			delete(engineContainers, containerid)
			logger.ERROR("[#cluster#] upgrade %s after-stop containers error, %s", metaData.MetaID, err.Error())
			cluster.actionContainers("start", engineContainers) //restart old tag containers.
			return nil, err
		}
	}

	createdContainers, err := cluster.createContainers(metaData, metaData.Instances, priorities, config)
	if err != nil {
		for _, container := range createdContainers {
			if engine := cluster.GetEngine(container.IP); engine != nil {
				engine.RemoveContainer(container.ID)
			}
		}
		if afterStop { //restart old tag containers.
			cluster.actionContainers("start", engineContainers)
		}
		return nil, err
	}

	//remove old tag containers.
	for container, engine := range engineContainers {
		if engine != nil {
			engine.RemoveContainer(container)
		} else {
			cluster.configCache.RemoveContainerBaseConfig(metaData.MetaID, container)
		}
	}

	upgradeContainers := types.UpgradeContainers{}
	for _, container := range createdContainers {
		upgradeContainers = upgradeContainers.SetUpgradePair(container.IP, container.HostName, container.Container)
	}
	return &upgradeContainers, nil
}

// UpgradeContainers is exported
func (cluster *Cluster) UpgradeContainers(metaid string, imagetag string) (*types.UpgradeContainers, error) {

	metaData, engines, err := cluster.validateMetaData(metaid)
	if err != nil {
		logger.ERROR("[#cluster#] upgrade meta %s error, %s", metaid, err.Error())
		return nil, err
	}

	if metaData.ImageTag == imagetag {
		return nil, fmt.Errorf("upgrade meta %s cancel, this tag has already in cluster", metaid)
	}

	config := metaData.Config
	tempPaths := strings.Split(config.Image, "/")
	tempPaths[len(tempPaths)-1] = strings.Split(tempPaths[len(tempPaths)-1], ":")[0] + ":" + imagetag
	config.Image = strings.Join(tempPaths, "/")
	upgradeContainers, err := cluster.upgradeContainers(metaData, engines, config)
	if err != nil {
		return nil, fmt.Errorf("upgrade %s failure, %s", metaid, err.Error())
	}
	//save new tag to meta file.
	cluster.configCache.SetImageTag(metaid, imagetag)
	cluster.submitHookEvent(metaData, UpgradeMetaEvent)
	return upgradeContainers, nil
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

// RemoveContainersOfMetaName is exported
// remove meta's all containers
func (cluster *Cluster) RemoveContainersOfMetaName(groupid string, metaname string) (string, *types.RemovedContainers, error) {

	metaData := cluster.configCache.GetMetaDataOfName(groupid, metaname)
	if metaData == nil {
		return "", nil, ErrClusterMetaDataNotFound
	}

	removedContainers, err := cluster.RemoveContainers(metaData.MetaID, "")
	if err != nil {
		return "", nil, err
	}
	return metaData.MetaID, removedContainers, nil
}

// RemoveContainers is exported
// if containerid is empty string so remove metaid's all containers
func (cluster *Cluster) RemoveContainers(metaid string, containerid string) (*types.RemovedContainers, error) {

	metaData, _, err := cluster.validateMetaData(metaid)
	if err != nil {
		logger.ERROR("[#cluster#] remove meta %s error, %s", metaid, err.Error())
		return nil, err
	}

	logger.INFO("[#cluster#] remove meta %s %s", metaid, containerid)
	removedContainers := cluster.removeContainers(metaData, containerid)
	cluster.submitHookEvent(metaData, RemoveMetaEvent)
	metaData = cluster.configCache.GetMetaData(metaData.MetaID)
	if metaData != nil {
		if containerid == "" || len(metaData.BaseConfigs) == 0 {
			cluster.configCache.RemoveMetaData(metaData.MetaID)
		}
	}
	return removedContainers, nil
}

// RecoveryContainers is exported
func (cluster *Cluster) RecoveryContainers(metaid string) error {

	metaData, engines, err := cluster.validateMetaData(metaid)
	if err != nil {
		return fmt.Errorf("recovery meta %s %s", metaid, err)
	}

	if !metaData.IsRecovery {
		return fmt.Errorf("recovery meta %s is disabled", metaData.MetaID)
	}

	baseConfigs := cluster.configCache.GetMetaDataBaseConfigs(metaData.MetaID)
	for _, baseConfig := range baseConfigs {
		if baseConfig.ID != "" {
			found := false
			for _, engine := range engines {
				if engine.IsHealthy() && engine.HasContainer(baseConfig.ID) {
					found = true
					break
				}
			}
			if !found { //clean meta invalid container.
				cluster.configCache.RemoveContainerBaseConfig(metaData.MetaID, baseConfig.ID)
				logger.WARN("[#cluster#] recovery meta %s remove invalid container %s", metaData.MetaID, ShortContainerID(baseConfig.ID))
			}
		}
	}

	if len(engines) > 0 {
		baseConfigsCount := cluster.configCache.GetMetaDataBaseConfigsCount(metaData.MetaID)
		if baseConfigsCount != -1 && metaData.Instances != baseConfigsCount {
			var err error
			if metaData.Instances > baseConfigsCount {
				_, err = cluster.createContainers(metaData, metaData.Instances-baseConfigsCount, nil, metaData.Config)
			} else {
				cluster.reduceContainers(metaData, baseConfigsCount-metaData.Instances)
			}
			cluster.submitHookEvent(metaData, RecoveryMetaEvent)
			cluster.NotifyGroupMetaContainersEvent("Cluster Meta Containers Recovered.", err, metaData.MetaID)
		}
	}
	return nil
}

// UpdateContainers is exported
func (cluster *Cluster) UpdateContainers(metaid string, instances int, webhooks types.WebHooks, placement types.Placement, config models.Container, updateOption types.UpdateOption) (*types.CreatedContainers, error) {

	if instances < 0 {
		logger.ERROR("[#cluster#] update meta %s error, %s", metaid, ErrClusterContainersInstancesInvalid)
		return nil, ErrClusterContainersInstancesInvalid
	}

	metaData, engines, err := cluster.validateMetaData(metaid)
	if err != nil {
		logger.ERROR("[#cluster#] update meta %s error, %s", metaid, err.Error())
		return nil, err
	}

	if config.Name == "" {
		config = metaData.Config
	}

	originalConfig := metaData.Config
	originalPlacement := metaData.Placement
	imageTag := getImageTag(config.Image)
	cluster.configCache.SetMetaData(metaid, instances, webhooks, placement, config, updateOption.IsRemoveDelay, updateOption.IsRecovery)
	cluster.configCache.SetImageTag(metaid, imageTag)
	metaData = cluster.configCache.GetMetaData(metaid)
	if metaData == nil {
		return nil, fmt.Errorf("update invalid meta %s", metaid)
	}

	if len(engines) > 0 {
		originalInstances := len(metaData.BaseConfigs)
		if instances == 0 { //clear containers.
			logger.INFO("[#cluster#] update %s containers, reduce instances to %d.", config.Name, instances)
			cluster.reduceContainers(metaData, originalInstances)
		} else {
			placementCompared := reflect.DeepEqual(originalPlacement, placement)
			availableNodesChanged := metaData.AvailableNodesChanged
			if !reflect.DeepEqual(originalConfig, config) || !placementCompared || availableNodesChanged {
				//config or placement changed, re-create all containers.
				logger.INFO("[#cluster#] update %s containers, re-create %d instances.", config.Name, instances)
				var priorities *EnginePriorities
				if originalInstances == instances && placementCompared && !availableNodesChanged {
					priorities = NewEnginePriorities(metaData, engines)
					logger.INFO("[#cluster#] update %s containers, priorities %s", config.Name, priorities.EngineStrings())
				}
				cluster.reduceContainers(metaData, originalInstances)
				_, err = cluster.createContainers(metaData, instances, priorities, metaData.Config)
			} else { //instances changed only.
				if originalInstances < instances {
					logger.INFO("[#cluster#] update %s containers, instances changed only, append %d instances.", config.Name, instances-originalInstances)
					_, err = cluster.createContainers(metaData, instances-originalInstances, nil, metaData.Config)
				} else {
					logger.INFO("[#cluster#] update %s containers, instances changed only, reduce %d containers.", config.Name, originalInstances-instances)
					cluster.reduceContainers(metaData, originalInstances-instances)
				}
			}
		}
	}

	if metaData.AvailableNodesChanged {
		cluster.configCache.SetAvailableNodesChanged(metaid, false)
		logger.INFO("[#cluster#] meta %s disable available nodes changed.", metaid)
	}

	cluster.submitHookEvent(metaData, UpdateMetaEvent)
	if err == nil {
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
	return nil, err
}

// CreateContainers is exported
func (cluster *Cluster) CreateContainers(groupid string, instances int, webhooks types.WebHooks, placement types.Placement, config models.Container, createOption types.CreateOption) (string, *types.CreatedContainers, error) {

	if instances <= 0 {
		return "", nil, ErrClusterContainersInstancesInvalid
	}

	group := cluster.GetGroup(groupid)
	engines := cluster.GetGroupEngines(groupid)
	if group == nil || engines == nil {
		logger.ERROR("[#cluster#] create containers %s error, %s", config.Name, ErrClusterGroupNotFound)
		return "", nil, ErrClusterGroupNotFound
	}

	if len(engines) == 0 {
		logger.ERROR("[#cluster#] create containers %s error, %s", config.Name, ErrClusterNoEngineAvailable)
		return "", nil, ErrClusterNoEngineAvailable
	}

	var (
		metaID  string
		bCreate bool = true
	)

	if !createOption.IsReCreate {
		if ret := cluster.cehckContainerNameUniqueness(groupid, config.Name); !ret {
			logger.ERROR("[#cluster#] create containers %s error, %s", config.Name, ErrClusterCreateContainerNameConflict)
			return "", nil, ErrClusterCreateContainerNameConflict
		}
	} else {
		if metaData := cluster.configCache.GetMetaDataOfName(groupid, config.Name); metaData != nil {
			if len(webhooks) == 0 {
				webhooks = metaData.WebHooks
			}
			if createOption.ForceRemove {
				cluster.RemoveContainers(metaData.MetaID, "")
				time.Sleep(pendingWaitForInterval)
			} else {
				imageTag := getImageTag(config.Image)
				if metaData.ImageTag == imageTag {
					logger.WARN("[#cluster#] re-create %s containers %s tag %s eq.", metaData.MetaID, config.Name, imageTag)
					for {
						time.Sleep(pendingWaitForInterval)
						if !cluster.containsPendingContainers(groupid, config.Name) {
							break
						}
						logger.WARN("[#cluster#] re-create %s containers %s pending...", metaData.MetaID, config.Name)
					}
					cluster.RemoveContainers(metaData.MetaID, "")
					time.Sleep(pendingWaitForInterval)
				} else {
					metaID = metaData.MetaID
					bCreate = false
				}
			}
		}
	}

	createdContainers := types.CreatedContainers{}
	if bCreate {
		metaData, err := cluster.configCache.CreateMetaData(groupid, instances, webhooks, placement, config, createOption.IsRemoveDelay, createOption.IsRecovery)
		if err != nil {
			if strings.Contains(err.Error(), "create meta conflict") {
				newMetaID, containers, err := cluster.reCreateContainers(groupid, metaData.MetaID, instances, webhooks, placement, config, createOption)
				if err != nil {
					return "", nil, err
				}
				createdContainers = *containers
				return newMetaID, &createdContainers, nil
			}
			logger.ERROR("[#cluster#] create containers %s error, %s", config.Name, ErrClusterContainersMetaCreateFailure)
			return "", nil, ErrClusterContainersMetaCreateFailure
		}
		createdContainers, err = cluster.createContainers(metaData, instances, nil, config)
		if len(createdContainers) == 0 {
			cluster.configCache.RemoveMetaData(metaData.MetaID)
			var resultErr string
			if err != nil {
				resultErr = err.Error()
			}
			logger.ERROR("[#cluster#] create containers %s error, %s", config.Name, ErrClusterCreateContainerFailure)
			return "", nil, fmt.Errorf("%s, %s\n", ErrClusterCreateContainerFailure.Error(), resultErr)
		}
		metaID = metaData.MetaID
		cluster.submitHookEvent(metaData, CreateMetaEvent)
	} else {
		newMetaID, containers, err := cluster.reCreateContainers(groupid, metaID, instances, webhooks, placement, config, createOption)
		if err != nil {
			return "", nil, err
		}
		metaID = newMetaID
		createdContainers = *containers
	}
	return metaID, &createdContainers, nil
}

// reContainers is exported
func (cluster *Cluster) reCreateContainers(groupid string, metaID string, instances int, webhooks types.WebHooks, placement types.Placement, config models.Container, createOption types.CreateOption) (string, *types.CreatedContainers, error) {

	retries := cluster.createRetry

RECREATE:
	for {
		if !cluster.containsPendingContainers(groupid, config.Name) {
			break
		}
		time.Sleep(pendingWaitForInterval)
		logger.WARN("[#cluster#] re-create %s containers %s pending...", metaID, config.Name)
	}

	updateOption := types.UpdateOption{
		IsRemoveDelay: createOption.IsRemoveDelay,
		IsRecovery:    createOption.IsRecovery,
	}
	containers, err := cluster.UpdateContainers(metaID, instances, webhooks, placement, config, updateOption)
	if err != nil || len(*containers) == 0 {
		if err != nil {
			if strings.Contains(err.Error(), "cluster containers state is setting") {
				retries = retries - 1
				if retries > 0 {
					time.Sleep(pendingWaitForInterval)
					goto RECREATE
				}
			} else if strings.Contains(err.Error(), "cluster metadata not found") {
				retries = retries - 1
				if retries > 0 {
					pMetaData := cluster.configCache.GetMetaDataOfName(groupid, config.Name)
					if pMetaData != nil && pMetaData.MetaID != metaID {
						metaID = pMetaData.MetaID
						time.Sleep(pendingWaitForInterval)
						goto RECREATE
					}
				}
			}
		}
		logger.ERROR("[#cluster#] re-create %s containers %s error, %s", metaID, config.Name, ErrClusterCreateContainerFailure)
		return "", nil, ErrClusterCreateContainerFailure
	}
	return metaID, containers, nil
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

// removeContainers is exported
func (cluster *Cluster) removeContainers(metaData *MetaData, containerid string) *types.RemovedContainers {

	cluster.Lock()
	cluster.pendingContainers[metaData.Config.Name] = &pendingContainer{
		GroupID: metaData.GroupID,
		Name:    metaData.Config.Name,
		Config:  metaData.Config,
	}
	cluster.Unlock()

	removedContainers := types.RemovedContainers{}
	if engines := cluster.GetGroupEngines(metaData.GroupID); engines != nil {
		foundContainer := false
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
	}

	cluster.Lock()
	delete(cluster.pendingContainers, metaData.Config.Name)
	cluster.Unlock()
	return &removedContainers
}

// createContainers is exported
func (cluster *Cluster) createContainers(metaData *MetaData, instances int, priorities *EnginePriorities, config models.Container) (types.CreatedContainers, error) {

	cluster.Lock()
	cluster.pendingContainers[config.Name] = &pendingContainer{
		GroupID: metaData.GroupID,
		Name:    config.Name,
		Config:  config,
	}
	cluster.Unlock()

	var resultErr error
	createdContainers := types.CreatedContainers{}
	filter := NewEnginesFilter()
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
		if cluster.Location != "" {
			containerConfig.Env = append(containerConfig.Env, "HUMPBACK_CLUSTER_LOCATION="+cluster.Location)
		}
		engine, container, err := cluster.createContainer(metaData, filter, priorities, containerConfig)
		if err != nil {
			if err == ErrClusterNoEngineAvailable || strings.Index(err.Error(), " not found") >= 0 {
				resultErr = err
				logger.ERROR("[#cluster#] create container %s, error:%s", containerConfig.Name, err.Error())
				continue
			}
			logger.ERROR("[#cluster#] engine %s, create container %s, error:%s", engine.IP, containerConfig.Name, err.Error())
			var retries int64
			for ; retries < cluster.createRetry && err != nil; retries++ {
				engine, container, err = cluster.createContainer(metaData, filter, nil, containerConfig)
			}
			if err != nil {
				resultErr = err
				if err == ErrClusterNoEngineAvailable {
					logger.ERROR("[#cluster#] create container %s, error:%s", containerConfig.Name, err.Error())
				} else {
					logger.ERROR("[#cluster#] engine %s, create container %s, error:%s", engine.IP, containerConfig.Name, err.Error())
				}
				continue
			}
		}
		createdContainers = createdContainers.SetCreatedPair(engine.IP, engine.Name, container.Config.Container)
	}

	cluster.Lock()
	delete(cluster.pendingContainers, config.Name)
	cluster.Unlock()
	return createdContainers, resultErr
}

// createContainer is exported
func (cluster *Cluster) createContainer(metaData *MetaData, filter *EnginesFilter, priorities *EnginePriorities, config models.Container) (*Engine, *Container, error) {

	engines := cluster.GetGroupEngines(metaData.GroupID)
	if engines == nil || len(engines) == 0 {
		return nil, nil, ErrClusterNoEngineAvailable
	}

	for _, engine := range engines {
		if engine.IsHealthy() && engine.HasMeta(metaData.MetaID) {
			filter.SetAllocEngine(engine)
		}
	}

	var engine *Engine
	if priorities != nil {
		engine = priorities.Select()
	}

	if engine == nil {
		selectEngines := cluster.selectEngines(engines, filter, config)
		if len(selectEngines) == 0 {
			return nil, nil, ErrClusterNoEngineAvailable
		}
		selectEngines = cluster.selectPlacementEngines(selectEngines, filter, &metaData.Placement)
		if len(selectEngines) == 0 {
			return nil, nil, ErrClusterNoEngineAvailable
		}
		engine = selectEngines[0]
	}

	container, err := engine.CreateContainer(config)
	if err != nil {
		filter.SetFailEngine(engine)
		return engine, nil, err
	}
	return engine, container, nil
}

// selectEngines is exported
func (cluster *Cluster) selectEngines(engines []*Engine, filter *EnginesFilter, config models.Container) []*Engine {

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

	if len(selectEngines) > 0 {
		filterEngines := filter.Filter(selectEngines)
		if len(filterEngines) > 0 {
			selectEngines = filterEngines
		} else {
			filterEngines = filter.AllocEngines()
			if len(filterEngines) > 0 {
				selectEngines = filterEngines
			}
			for i := len(selectEngines) - 1; i > 0; i-- {
				j := cluster.randSeed.Intn(i + 1)
				selectEngines[i], selectEngines[j] = selectEngines[j], selectEngines[i]
			}
		}
	}
	return selectEngines
}

// selectPlacementEngines is exported
func (cluster *Cluster) selectPlacementEngines(engines []*Engine, filter *EnginesFilter, placement *types.Placement) []*Engine {

	selectEngines := []*Engine{}
	if placement.Constraints != nil && len(placement.Constraints) > 0 {
		constraints, err := ParseConstraints(placement.Constraints)
		if err != nil {
			logger.ERROR("[#cluster#] placement engines error, %s", err.Error())
			return selectEngines //return empty engines
		}
		for _, engine := range engines {
			ret := MatchConstraints(constraints, engine)
			if ret {
				selectEngines = append(selectEngines, engine)
				logger.INFO("[#cluster#] placement engines, %s(%s)", engine.IP, engine.Name)
			} else {
				logger.INFO("[#cluster#] placement engines filter, %s(%s)", engine.IP, engine.Name)
			}
		}
		if len(selectEngines) == 0 {
			filterEngines := filter.AllocEngines()
			if len(filterEngines) > 0 {
				logger.INFO("[#cluster#] placement select alloc engines")
				selectEngines = filterEngines
			} else {
				//last retry fail! select a fail engine, throw cretae error,
				//prevent return cluster no engine available error.
				logger.INFO("[#cluster#] placement select fail engines")
				selectEngines = filter.FailEngines()
			}
			for i := len(selectEngines) - 1; i > 0; i-- {
				j := cluster.randSeed.Intn(i + 1)
				selectEngines[i], selectEngines[j] = selectEngines[j], selectEngines[i]
			}
		}
	} else {
		selectEngines = engines
		logger.INFO("[#cluster#] skip placement engines.")
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

	metaData, engines, err := cluster.GetMetaDataEngines(metaid)
	if err != nil {
		return nil, nil, err
	}

	if ret := cluster.migtatorCache.Contains(metaData.MetaID); ret {
		return nil, nil, ErrClusterContainersMigrating
	}

	if ret := cluster.containsPendingContainers(metaData.GroupID, metaData.Config.Name); ret {
		return nil, nil, ErrClusterContainersSetting
	}
	return metaData, engines, nil
}

func (cluster *Cluster) submitHookEvent(metaData *MetaData, hookEvent HookEvent) {

	if len(metaData.WebHooks) == 0 {
		return
	}

	hookContainers := HookContainers{}
	engines := cluster.GetGroupEngines(metaData.GroupID)
	for _, engine := range engines {
		if engine.IsHealthy() {
			containers := engine.Containers(metaData.MetaID)
			for _, container := range containers {
				hookContainers = append(hookContainers, &HookContainer{
					IP:        engine.IP,
					Name:      engine.Name,
					Container: container.Config.Container,
				})
			}
		}
	}
	cluster.hooksProcessor.SubmitHook(metaData.MetaBase, hookContainers, hookEvent)
}

func (cluster *Cluster) recoveryContainersLoop() {

	for {
		ticker := time.NewTicker(cluster.recoveryInterval)
		select {
		case <-ticker.C:
			{
				ticker.Stop()
				metaids := []string{}
				metaEngines := make(map[string]*Engine)
				groups := cluster.GetGroups()
				for _, group := range groups {
					groupMetaData := cluster.configCache.GetGroupMetaData(group.ID)
					for _, metaData := range groupMetaData {
						metaids = append(metaids, metaData.MetaID)
						if _, engines, err := cluster.GetMetaDataEngines(metaData.MetaID); err == nil {
							for _, engine := range engines {
								if engine.IsHealthy() && engine.HasMeta(metaData.MetaID) {
									metaEngines[engine.IP] = engine
								}
							}
						}
					}
				}
				if len(metaids) > 0 {
					cluster.RefreshEnginesContainers(metaEngines)
					for _, metaid := range metaids {
						if err := cluster.RecoveryContainers(metaid); err != nil {
							logger.ERROR("[#cluster#] recovery containers error, %s", err.Error())
						}
					}
				}
			}
		case <-cluster.stopCh:
			{
				ticker.Stop()
				return
			}
		}
	}
}
