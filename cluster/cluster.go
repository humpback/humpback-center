package cluster

import "github.com/humpback/discovery"
import "github.com/humpback/discovery/backends"
import "github.com/humpback/gounits/json"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/gounits/system"
import "github.com/humpback/humpback-agent/models"
import "github.com/humpback/humpback-center/cluster/types"

import (
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
type Group struct {
	ID      string
	Servers []string
}

// Cluster is exported
// Discovery: discovery service
// overcommitRatio: engine overcommit ratio, cpus & momery resources.
// createRetry: create container retry count.
// configCache: container config metadata cache.
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

	cacheRoot := ""
	if val, ret := driverOpts.String("cacheroot", ""); ret {
		cacheRoot = val
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
		pendingContainers: make(map[string]*pendingContainer),
		engines:           make(map[string]*Engine),
		groups:            make(map[string]*Group),
		stopCh:            make(chan struct{}),
	}, nil
}

func (cluster *Cluster) Start() error {

	cluster.configCache.Init()
	logger.INFO("[#cluster#] discovery service watching...")
	if cluster.Discovery != nil {
		cluster.Discovery.Watch(cluster.stopCh, cluster.watchDiscoveryHandleFunc)
		return nil
	}
	return ErrClusterDiscoveryInvalid
}

func (cluster *Cluster) Stop() {

	close(cluster.stopCh)
	logger.INFO("[#cluster#] discovery service closed.")
}

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

func (cluster *Cluster) GetEngine(ip string) *Engine {

	cluster.RLock()
	defer cluster.RUnlock()
	if engine, ret := cluster.engines[ip]; ret {
		return engine
	}
	return nil
}

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

func (cluster *Cluster) GetGroups() []*Group {

	groups := []*Group{}
	cluster.RLock()
	for _, group := range cluster.groups {
		groups = append(groups, group)
	}
	cluster.RUnlock()
	return groups
}

func (cluster *Cluster) GetGroup(groupid string) *Group {

	cluster.RLock()
	defer cluster.RUnlock()
	if group, ret := cluster.groups[groupid]; ret {
		return group
	}
	return nil
}

func (cluster *Cluster) SetGroup(groupid string, servers []string) {

	cluster.Lock()
	removes := []string{}
	group, ret := cluster.groups[groupid]
	if !ret {
		group = &Group{ID: groupid, Servers: servers}
		cluster.groups[groupid] = group
		logger.INFO("[#cluster#] created group %s(%d)", groupid, len(servers))
	} else {
		origin := group.Servers
		group.Servers = servers
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

// SetContainers is exported
func (cluster *Cluster) SetContainers(groupid string, metaid string, instances int) {
	//在现有metaid上调整实例数(增加或删除实例), metaid不会变，配置不会变
}

// CreateContainers is exported
func (cluster *Cluster) CreateContainers(groupid string, instances int, config models.Container) (map[string]string, error) {

	engines := cluster.GetGroupEngines(groupid)
	if engines == nil {
		logger.ERROR("[#cluster#] create container error %s : %s", groupid, ErrClusterGroupNotFound)
		return nil, ErrClusterGroupNotFound
	}

	if len(engines) == 0 {
		logger.ERROR("[#cluster#] create container error %s : %s", groupid, ErrClusterNoEngineAvailable)
		return nil, ErrClusterNoEngineAvailable
	}

	if ret := cluster.cehckContainerNameUniqueness(groupid, config.Name); !ret {
		logger.ERROR("[#cluster#] create container error %s : %s", groupid, ErrClusterCreateContainerNameConflict)
		return nil, ErrClusterCreateContainerNameConflict
	}

	cluster.Lock()
	cluster.pendingContainers[config.Name] = &pendingContainer{
		GroupID: groupid,
		Name:    config.Name,
		Config:  config,
	}
	cluster.Unlock()

	metaid := cluster.configCache.MakeUniqueMetaID()
	createdParis := map[string]string{}
	ipList := []string{}
	index := 1
	for ; instances > 0; instances-- {
		containerConfig := config
		containerConfig.Env = append(containerConfig.Env, "HUMPBACK_CLUSTER_GROUPID="+groupid, "HUMPBACK_CLUSTER_METAID="+metaid)
		containerConfig.Name = groupid[:8] + "-" + containerConfig.Name + "-" + strconv.Itoa(index)
		engine, container, err := cluster.createContainer(engines, ipList, containerConfig)
		if err != nil {
			logger.ERROR("[#cluster#] create container %s %s, error:%s", groupid, containerConfig.Name, err.Error())
			if err == ErrClusterNoEngineAvailable {
				continue
			}
			ipList = filterAppendIPList(engine, ipList)
			var retries int64
			for ; retries < cluster.createRetry && err != nil; retries++ {
				engine, container, err = cluster.createContainer(engines, ipList, containerConfig)
				ipList = filterAppendIPList(engine, ipList)
			}
			if err != nil {
				continue
			}
		}
		index++
		ipList = filterAppendIPList(engine, ipList)
		ipList = append(ipList, engine.IP)
		createdParis[container.Info.ID] = engine.IP
		containerConfig.ID = container.Info.ID
		baseConfig := &ContainerBaseConfig{Container: containerConfig}
		cluster.configCache.SetContainerBaseConfig(metaid, groupid, config.Name, baseConfig)
	}

	cluster.Lock()
	delete(cluster.pendingContainers, config.Name)
	cluster.Unlock()
	if instances > 0 && len(createdParis) == 0 {
		return nil, ErrClusterCreateContainerFailure
	}
	return createdParis, nil
}

func (cluster *Cluster) createContainer(engines []*Engine, ipList []string, config models.Container) (*Engine, *Container, error) {

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
			return filterEngines
		} else {
			for i := len(selectEngines) - 1; i > 0; i-- {
				j := cluster.randSeed.Intn(i + 1)
				selectEngines[i], selectEngines[j] = selectEngines[j], selectEngines[i]
			}
		}
	}
	return selectEngines
}

func (cluster *Cluster) cehckContainerNameUniqueness(groupid string, name string) bool {

	cluster.RLock()
	defer cluster.RUnlock()
	for _, pendingContainer := range cluster.pendingContainers {
		if pendingContainer.GroupID == groupid && pendingContainer.Name == name {
			return false
		}
	}
	metaData := cluster.configCache.GetMetaDataOfName(name)
	if metaData != nil && metaData.GroupID == groupid {
		return false
	}
	return true
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
