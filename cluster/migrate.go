package cluster

import "github.com/humpback/gounits/logger"

import (
	"sync"
	"time"
)

// MigrateState is exported
type MigrateState int

const (
	MigrateReady = iota + 1
	Migrating
	MigrateFailure
	MigrateCompleted
)

func (state MigrateState) String() string {

	switch state {
	case MigrateReady:
		return "MigrateReady"
	case Migrating:
		return "Migrating"
	case MigrateFailure:
		return "MigrateFailure"
	case MigrateCompleted:
		return "MigrateCompleted"
	}
	return ""
}

// MigrateContainer is exported
type MigrateContainer struct {
	sync.RWMutex
	ContainerID string
	metaData    *MetaData
	baseConfig  *ContainerBaseConfig
	state       MigrateState
}

// NewMigrateContainer is exported
func NewMigrateContainer(containerid string, baseConfig *ContainerBaseConfig) *MigrateContainer {

	if baseConfig == nil || baseConfig.MetaData == nil {
		return nil
	}

	return &MigrateContainer{
		ContainerID: containerid,
		metaData:    baseConfig.MetaData,
		baseConfig:  baseConfig,
		state:       MigrateReady,
	}
}

// GetState is exported
func (mContainer *MigrateContainer) GetState() MigrateState {

	mContainer.RLock()
	defer mContainer.RUnlock()
	return mContainer.state
}

// SetState is exported
func (mContainer *MigrateContainer) SetState(state MigrateState) {

	mContainer.Lock()
	mContainer.state = state
	mContainer.Unlock()
}

// Execute is exported
func (mContainer *MigrateContainer) Execute(cluster *Cluster) {

	mContainer.SetState(Migrating)
	engine, container, err := cluster.createContainer(mContainer.metaData, []string{}, mContainer.baseConfig.Container)
	if err != nil {
		mContainer.SetState(MigrateFailure)
		logger.ERROR("[#cluster] migrator container %s error %s", mContainer.ContainerID[:12], err.Error())
		return
	}

	mContainer.Lock()
	mContainer.state = MigrateCompleted
	logger.INFO("[#cluster] migrator container %s > %s to %s", mContainer.ContainerID[:12], container.Info.ID[:12], engine.IP)
	mContainer.Unlock()
	return
}

// Migrator is exported
type Migrator struct {
	sync.RWMutex
	MetaID       string
	Cluster      *Cluster
	retryCount   int64
	migrateDelay time.Duration
	containers   []*MigrateContainer
	handler      MigratorHandler
}

// NewMigrator is exported
func NewMigrator(metaid string, containers Containers, cluster *Cluster, migrateDelay time.Duration, handler MigratorHandler) *Migrator {

	mContainers := []*MigrateContainer{}
	for _, container := range containers {
		mContainer := NewMigrateContainer(container.Info.ID, container.BaseConfig)
		if mContainer != nil {
			mContainers = append(mContainers, mContainer)
		}
	}

	return &Migrator{
		MetaID:       metaid,
		Cluster:      cluster,
		retryCount:   cluster.createRetry,
		migrateDelay: migrateDelay,
		containers:   mContainers,
		handler:      handler,
	}
}

func (migrator *Migrator) verifyEngines() bool {

	_, engines, err := migrator.Cluster.GetMetaDataEngines(migrator.MetaID)
	if err != nil {
		return false
	}

	for _, engine := range engines {
		if engine.IsHealthy() {
			return true
		}
	}
	return false
}

func (migrator *Migrator) isAllCompleted() bool {

	migrator.RLock()
	defer migrator.RUnlock()
	for _, mContainer := range migrator.containers {
		if mContainer.GetState() != MigrateCompleted {
			return false
		}
	}
	return true
}

func (migrator *Migrator) isPartCompleted() bool {

	migrator.RLock()
	defer migrator.RUnlock()
	nCompletedCount := 0
	for _, mContainer := range migrator.containers {
		if mContainer.GetState() == MigrateCompleted {
			nCompletedCount = nCompletedCount + 1
		}
	}

	if nCompletedCount == 0 {
		return false
	}
	return nCompletedCount != len(migrator.containers)
}

func (migrator *Migrator) resetMigrateContainers() {

	migrator.Lock()
	for _, mContainer := range migrator.containers {
		if mContainer.GetState() == MigrateFailure {
			mContainer.SetState(MigrateReady)
		}
	}
	migrator.retryCount = migrator.retryCount - 1
	migrator.Unlock()
}

func (migrator *Migrator) selectMigrateContainer() *MigrateContainer {

	migrator.RLock()
	defer migrator.RUnlock()
	for _, mContainer := range migrator.containers {
		if mContainer.GetState() == MigrateReady {
			return mContainer
		}
	}
	return nil
}

func (migrator *Migrator) removeMigrateContainer(mContainer *MigrateContainer) {

	migrator.Lock()
	defer migrator.Unlock()
	for i, pmContainer := range migrator.containers {
		if pmContainer == mContainer {
			migrator.containers = append(migrator.containers[:i], migrator.containers[i+1:]...)
			return
		}
	}
}

// Containers is exported
func (migrator *Migrator) Containers() []*MigrateContainer {

	migrator.RLock()
	mContainers := []*MigrateContainer{}
	for _, mContainer := range migrator.containers {
		mContainers = append(mContainers, mContainer)
	}
	migrator.RUnlock()
	return mContainers
}

// Container is exported
func (migrator *Migrator) Container(containerid string) *MigrateContainer {

	migrator.RLock()
	defer migrator.RUnlock()
	for _, mContainer := range migrator.containers {
		if mContainer.ContainerID == containerid {
			return mContainer
		}
	}
	return nil
}

// Start is exported
func (migrator *Migrator) Start() {

	time.Sleep(migrator.migrateDelay)
	for {
		migrator.RLock()
		mContainers := migrator.containers
		migrator.RUnlock()
		if len(mContainers) == 0 {
			break
		}

		if !migrator.isAllCompleted() {
			if !migrator.verifyEngines() {
				logger.INFO("[#cluster] migrator %s containers, verify engines no healthy.", migrator.MetaID)
				migrator.Cluster.configCache.ClearContainerBaseConfig(migrator.MetaID)
				migrator.handler.OnMigratorNotifyHandleFunc(migrator)
				break
			}
		}

		mContainer := migrator.selectMigrateContainer()
		if mContainer != nil {
			mContainer.Execute(migrator.Cluster)
			if mContainer.GetState() == MigrateCompleted {
				migrator.Cluster.configCache.RemoveContainerBaseConfig(migrator.MetaID, mContainer.ContainerID)
			}
			continue
		}

		if migrator.isAllCompleted() {
			migrator.handler.OnMigratorNotifyHandleFunc(migrator)
			break
		}

		if migrator.retryCount <= 0 {
			migrator.Lock()
			for _, mContainer := range migrator.containers {
				if mContainer.GetState() == MigrateFailure {
					migrator.Cluster.configCache.RemoveContainerBaseConfig(migrator.MetaID, mContainer.ContainerID)
				}
			}
			migrator.Unlock()
			migrator.handler.OnMigratorNotifyHandleFunc(migrator)
			break
		}

		if migrator.isPartCompleted() {
			migrator.handler.OnMigratorNotifyHandleFunc(migrator)
		}
		migrator.resetMigrateContainers()
		continue
	}
	migrator.handler.OnMigratorQuitHandleFunc(migrator)
}

// Update is exported
func (migrator *Migrator) Update(metaid string, containers Containers) {

}

// Cancel is exported
func (migrator *Migrator) Cancel(metaid string, containers Containers) {

	for _, container := range containers {
		if mContainer := migrator.Container(container.Info.ID); mContainer != nil {
			state := mContainer.GetState()
			if state == MigrateReady || state == MigrateFailure {
				migrator.removeMigrateContainer(mContainer)
			} else if state == MigrateCompleted || state == Migrating {
				go migrator.Cluster.RemoveContainer(mContainer.ContainerID)
			}
		}
	}
}

// MigratorHandler is exported
type MigratorHandler interface {
	OnMigratorQuitHandleFunc(migrator *Migrator)
	OnMigratorNotifyHandleFunc(migrator *Migrator)
}

// MigratorQuitHandleFunc is exported
type MigratorQuitHandleFunc func(migrator *Migrator)

// OnMigratorQuitHandleFunc is exported
func (fn MigratorQuitHandleFunc) OnMigratorQuitHandleFunc(migrator *Migrator) {
	fn(migrator)
}

// MigratorNotifyHandleFunc is exported
type MigratorNotifyHandleFunc func(migrator *Migrator)

// OnMigratorNotifyHandleFunc is exported
func (fn MigratorNotifyHandleFunc) OnMigratorNotifyHandleFunc(migrator *Migrator) {
	fn(migrator)
}

// MigrateContainersCache is exported
type MigrateContainersCache struct {
	sync.RWMutex
	MigratorHandler
	Cluster      *Cluster
	migrateDelay time.Duration
	migrators    map[string]*Migrator
}

// NewMigrateContainersCache is exported
func NewMigrateContainersCache(migrateDelay time.Duration) *MigrateContainersCache {

	return &MigrateContainersCache{
		migrateDelay: migrateDelay,
		migrators:    make(map[string]*Migrator),
	}
}

// SetCluster is exported
func (cache *MigrateContainersCache) SetCluster(cluster *Cluster) {

	cache.Cluster = cluster
}

// Contains is exported
func (cache *MigrateContainersCache) Contains(metaid string) bool {

	cache.RLock()
	defer cache.RUnlock()
	_, ret := cache.migrators[metaid]
	return ret
}

// Start is exported
// engine offline, start migrate containers.
// engine parameter is offline engine pointer.
func (cache *MigrateContainersCache) Start(engine *Engine) {

	metaids := engine.MetaIds()
	if len(metaids) == 0 {
		return
	}

	cache.Lock()
	for _, metaid := range metaids {
		containers := engine.Containers(metaid)
		if len(containers) == 0 {
			continue
		}
		migrator, ret := cache.migrators[metaid]
		if !ret {
			migrator = NewMigrator(metaid, containers, cache.Cluster, cache.migrateDelay, cache)
			cache.migrators[metaid] = migrator
			logger.INFO("[#cluster] migrator start %s %s", engine.IP, metaid)
			go migrator.Start()
		} else {
			logger.INFO("[#cluster] migrator update %s %s", engine.IP, metaid)
			migrator.Update(metaid, containers)
		}
	}
	cache.Unlock()
}

// Cancel is exported
// engine online, cancel migrate containers of state is MigrateReady.
// engine parameter is online engine pointer.
func (cache *MigrateContainersCache) Cancel(engine *Engine) {

	engine.RefreshContainers()
	metaids := engine.MetaIds()
	if len(metaids) == 0 {
		return
	}

	cache.RLock()
	defer cache.RUnlock()
	waitgroup := sync.WaitGroup{}
	for _, metaid := range metaids {
		if migrator, ret := cache.migrators[metaid]; ret {
			containers := engine.Containers(metaid)
			if len(containers) == 0 {
				continue
			}
			waitgroup.Add(1)
			go func(id string, c Containers) {
				defer waitgroup.Done()
				logger.INFO("[#cluster] migrator cancel %s %s", engine.IP, id)
				migrator.Cancel(id, c)
			}(metaid, containers)
		}
	}
	waitgroup.Wait()
}

// OnMigratorQuitHandleFunc is exported
func (cache *MigrateContainersCache) OnMigratorQuitHandleFunc(migrator *Migrator) {

	logger.INFO("[#cluster] migrator quited %s", migrator.MetaID)
	cache.Lock()
	delete(cache.migrators, migrator.MetaID)
	cache.Unlock()
}

// OnMigratorNotifyHandleFunc is exported
func (cache *MigrateContainersCache) OnMigratorNotifyHandleFunc(migrator *Migrator) {

	logger.INFO("[#cluster] migrator notify %s", migrator.MetaID)
	mContainers := migrator.Containers()
	for _, mContainer := range mContainers {
		logger.INFO("[#cluster] migrator container %s %s", mContainer.ContainerID[:12], mContainer.state.String())
	}
}
