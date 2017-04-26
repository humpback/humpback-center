package cluster

import "github.com/docker/docker/api/types"
import "github.com/humpback/gounits/utils"
import "github.com/humpback/gounits/convert"
import "github.com/humpback/gounits/http"
import "github.com/humpback/gounits/logger"
import ctypes "humpback-center/cluster/types"
import "common/models"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// engine send request timeout
	requestTimeout = 3 * time.Minute
	// engine refresh loop interval
	refreshInterval = 30 * time.Second
)

// Engine state define
type engineState int

// State enum value
const (
	//pending: engine added to cluster engines pool, but not been validated.
	StatePending engineState = iota
	//unhealthy: engine is unreachable.
	StateUnhealthy
	//healthy: engine is ready reachable.
	StateHealthy
	//disconnected: engine is removed from discovery
	StateDisconnected
)

// Engine state content mapping
var stateText = map[engineState]string{
	StatePending:      "Pending",
	StateUnhealthy:    "Unhealthy",
	StateHealthy:      "Healthy",
	StateDisconnected: "Disconnected",
}

// GetStateText is exported return a state typed text.
func GetStateText(state engineState) string {
	return stateText[state]
}

// Engine is exported
type Engine struct {
	sync.RWMutex
	ID        string            `json:"ID"`
	Name      string            `json:"Name"`
	IP        string            `json:"IP"`
	APIAddr   string            `json:"APIAddr"`
	Cpus      int64             `json:"Cpus"`
	Memory    int64             `json:"Memory"`
	Labels    map[string]string `json:"Labels"`
	StateText string            `json:"StateText"`

	overcommitRatio int64
	client          *http.HttpClient
	configCache     *ContainersConfigCache
	containers      map[string]*Container
	stopCh          chan struct{}
	state           engineState
}

// NewEngine is exported
func NewEngine(nodeData *NodeData, overcommitRatio float64, configCache *ContainersConfigCache) (*Engine, error) {

	ipAddr, err := net.ResolveIPAddr("ip4", nodeData.IP)
	if err != nil {
		return nil, err
	}

	return &Engine{
		ID:              nodeData.ID,
		Name:            nodeData.Name,
		IP:              ipAddr.IP.String(),
		APIAddr:         nodeData.APIAddr,
		Cpus:            nodeData.Cpus,
		Memory:          int64(math.Ceil(float64(nodeData.Memory) / 1024.0 / 1024.0)),
		Labels:          nodeData.MapLabels(),
		StateText:       stateText[StatePending],
		overcommitRatio: int64(overcommitRatio * 100),
		client:          http.NewWithTimeout(requestTimeout),
		configCache:     configCache,
		containers:      make(map[string]*Container),
		state:           StatePending,
	}, nil
}

// Open is exported
// Engine start refresh containers loop
func (engine *Engine) Open() {

	engine.Lock()
	if engine.state != StateHealthy {
		engine.state = StateHealthy
		engine.StateText = stateText[engine.state]
		engine.stopCh = make(chan struct{})
		logger.INFO("[#cluster#] engine %s open.", engine.IP)
		go func() {
			engine.ValidateContainers()
			engine.refreshLoop()
		}()
	}
	engine.Unlock()
}

// Close is exported
// Engine stop refresh containers loop
func (engine *Engine) Close() {

	engine.Lock()
	if engine.state == StateHealthy {
		close(engine.stopCh)
		engine.state = StateDisconnected
		engine.StateText = stateText[engine.state]
		engine.client.Close()
		logger.INFO("[#cluster#] engine %s closed.", engine.IP)
	}
	engine.Unlock()
}

// Update is exported
// Engine update info
func (engine *Engine) Update(nodeData *NodeData) {

	engine.Lock()
	engine.ID = nodeData.ID
	engine.Name = nodeData.Name
	engine.APIAddr = nodeData.APIAddr
	engine.Cpus = nodeData.Cpus
	engine.Memory = int64(math.Ceil(float64(nodeData.Memory) / 1024.0 / 1024.0))
	engine.Labels = nodeData.MapLabels()
	engine.Unlock()
}

// IsHealthy is exported
// Determine if the engine is in healthy state
func (engine *Engine) IsHealthy() bool {

	engine.RLock()
	defer engine.RUnlock()
	return engine.state == StateHealthy
}

// IsPending is exported
// Determine if the engine is in pending state
func (engine *Engine) IsPending() bool {

	engine.RLock()
	defer engine.RUnlock()
	return engine.state == StatePending
}

// IsDisconnected is exported
// Determine if the engine is in disconnected state
func (engine *Engine) IsDisconnected() bool {

	engine.RLock()
	defer engine.RUnlock()
	return engine.state == StateDisconnected
}

// State is exported
// Return a engine status string result.
func (engine *Engine) State() string {

	engine.RLock()
	defer engine.RUnlock()
	return stateText[engine.state]
}

// SetState is exported
// Set engine state.
func (engine *Engine) SetState(state engineState) {

	engine.Lock()
	engine.state = state
	engine.StateText = stateText[state]
	engine.Unlock()
}

// MetaIds is exported
// Return engine containers all metaids array.
func (engine *Engine) MetaIds() []string {

	ids := []string{}
	engine.RLock()
	for _, container := range engine.containers {
		if metaid := container.MetaID(); metaid != "" {
			if ret := utils.Contains(metaid, ids); !ret {
				ids = append(ids, metaid)
			}
		}
	}
	engine.RUnlock()
	return ids
}

// HasMeta is exported
// find metaid in engine containers
func (engine *Engine) HasMeta(metaid string) bool {

	engine.RLock()
	defer engine.RUnlock()
	for _, container := range engine.containers {
		if container.MetaID() == metaid {
			return true
		}
	}
	return false
}

// HasContainer is exported
// find containerid in engine containers
func (engine *Engine) HasContainer(containerid string) bool {

	engine.RLock()
	defer engine.RUnlock()
	for _, container := range engine.containers {
		if container.Info.ID == containerid {
			return true
		}
	}
	return false
}

// Containers is exported
// Return engine containers.
// if metaid is empty string so return engine's all containers
func (engine *Engine) Containers(metaid string) Containers {

	engine.RLock()
	containers := Containers{}
	for _, container := range engine.containers {
		if metaid == "" || container.MetaID() == metaid {
			containers = append(containers, container)
		}
	}
	engine.RUnlock()
	return containers
}

// Container is exported
// Return engine container of containerid.
func (engine *Engine) Container(containerid string) *Container {

	engine.RLock()
	defer engine.RUnlock()
	for _, container := range engine.containers {
		if container.Info.ID == containerid {
			return container
		}
	}
	return nil
}

// UsedMemory is exported
// Return engine all containers used memory size.
func (engine *Engine) UsedMemory() int64 {

	var used int64
	engine.RLock()
	for _, container := range engine.containers {
		used += container.Info.HostConfig.Memory
	}
	engine.RUnlock()
	return used
}

// UsedCpus is exported
// Return engine all containers used cpus size.
func (engine *Engine) UsedCpus() int64 {

	var used int64
	engine.RLock()
	for _, container := range engine.containers {
		used += container.Info.HostConfig.CPUShares
	}
	engine.RUnlock()
	return used
}

// TotalMemory is exported
// Return engine total memory size.
func (engine *Engine) TotalMemory() int64 {

	engine.RLock()
	defer engine.RUnlock()
	return engine.Memory + (engine.Memory * engine.overcommitRatio / 100)
}

// TotalCpus is exported
// Return engine total cpus size.
func (engine *Engine) TotalCpus() int64 {

	engine.RLock()
	defer engine.RUnlock()
	return engine.Cpus + (engine.Cpus * engine.overcommitRatio / 100)
}

// CreateContainer is exported
// Engine create a container
func (engine *Engine) CreateContainer(config models.Container) (*Container, error) {

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(config); err != nil {
		return nil, err
	}

	header := map[string][]string{"Content-Type": []string{"application/json"}}
	respCreated, err := engine.client.Post("http://"+engine.APIAddr+"/v1/containers", nil, buf, header)
	if err != nil {
		return nil, err
	}

	defer respCreated.Close()
	if respCreated.StatusCode() != 200 {
		return nil, fmt.Errorf("engine %s, create container %s failure, %s", engine.IP, config.Name, ctypes.ParseHTTPResponseError(respCreated))
	}

	createContainerResponse := &ctypes.CreateContainerResponse{}
	if err := respCreated.JSON(createContainerResponse); err != nil {
		return nil, err
	}

	config.ID = createContainerResponse.ID
	configEnvMap := convert.ConvertKVStringSliceToMap(config.Env)
	containerIndex, _ := strconv.Atoi(configEnvMap["HUMPBACK_CLUSTER_CONTAINER_INDEX"])
	metaData := engine.configCache.GetMetaData(configEnvMap["HUMPBACK_CLUSTER_METAID"])
	baseConfig := &ContainerBaseConfig{Index: containerIndex, Container: config, MetaData: metaData}
	engine.configCache.CreateContainerBaseConfig(metaData.MetaID, baseConfig)
	logger.INFO("[#cluster#] engine %s, create container %s:%s", engine.IP, createContainerResponse.ID[:12], config.Name)
	containers, err := engine.updateContainer(createContainerResponse.ID, engine.containers)
	if err != nil {
		return nil, err
	}

	engine.Lock()
	engine.containers = containers
	container, ret := engine.containers[createContainerResponse.ID]
	if !ret {
		engine.Unlock()
		return nil, fmt.Errorf("created container, update container info failure")
	}
	engine.Unlock()
	return container, nil
}

// RemoveContainer is exported
// Engine remove a container.
func (engine *Engine) RemoveContainer(containerid string) error {

	query := map[string][]string{"force": []string{"true"}}
	respRemoved, err := engine.client.Delete("http://"+engine.APIAddr+"/v1/containers/"+containerid, query, nil)
	if err != nil {
		return err
	}

	defer respRemoved.Close()
	if respRemoved.StatusCode() != 200 {
		return fmt.Errorf("engine %s, remove container %s failure, %s", engine.IP, containerid[:12], ctypes.ParseHTTPResponseError(respRemoved))
	}

	logger.INFO("[#cluster#] engine %s, remove container %s", engine.IP, containerid[:12])
	engine.Lock()
	if container, ret := engine.containers[containerid]; ret {
		engine.configCache.RemoveContainerBaseConfig(container.MetaID(), containerid)
		delete(engine.containers, containerid)
	}
	engine.Unlock()
	return nil
}

// OperateContainer is exported
// Engine operate a container.
func (engine *Engine) OperateContainer(operate models.ContainerOperate) error {

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(operate); err != nil {
		return err
	}

	header := map[string][]string{"Content-Type": []string{"application/json"}}
	respOperated, err := engine.client.Put("http://"+engine.APIAddr+"/v1/containers", nil, buf, header)
	if err != nil {
		return err
	}

	defer respOperated.Close()
	if respOperated.StatusCode() != 200 {
		return fmt.Errorf("engine %s, %s container %s failure, %s", engine.IP, operate.Action, operate.Container[:12], ctypes.ParseHTTPResponseError(respOperated))
	}

	logger.INFO("[#cluster#] engine %s, %s container %s", engine.IP, operate.Action, operate.Container[:12])
	if containers, err := engine.updateContainer(operate.Container, engine.containers); err == nil {
		engine.Lock()
		engine.containers = containers
		engine.Unlock()
	}
	return nil
}

// UpgradeContainer is exported
// Engine upgrade a container.
func (engine *Engine) UpgradeContainer(operate models.ContainerOperate) (*Container, error) {

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(operate); err != nil {
		return nil, err
	}

	header := map[string][]string{"Content-Type": []string{"application/json"}}
	respUpgraded, err := engine.client.Put("http://"+engine.APIAddr+"/v1/containers", nil, buf, header)
	if err != nil {
		return nil, err
	}

	defer respUpgraded.Close()
	if respUpgraded.StatusCode() != 200 {
		return nil, fmt.Errorf("engine %s, %s container %s failure, %s", engine.IP, operate.Action, operate.Container[:12], ctypes.ParseHTTPResponseError(respUpgraded))
	}

	upgradeContainerResponse := &ctypes.UpgradeContainerResponse{}
	if err := respUpgraded.JSON(upgradeContainerResponse); err != nil {
		return nil, err
	}

	logger.INFO("[#cluster#] engine %s, %s container %s to %s", engine.IP, operate.Action, operate.Container[:12], upgradeContainerResponse.ID[:12])
	containers, err := engine.updateContainer(upgradeContainerResponse.ID, engine.containers)
	if err != nil {
		return nil, err
	}

	engine.Lock()
	engine.containers = containers
	container, ret := engine.containers[upgradeContainerResponse.ID]
	if !ret {
		engine.Unlock()
		return nil, fmt.Errorf("upgrade container, update container info failure")
	}
	engine.Unlock()
	return container, nil
}

// RefreshContainers is exported
// Engine refresh all containers.
func (engine *Engine) RefreshContainers() error {

	query := map[string][]string{"all": []string{"true"}}
	respContainers, err := engine.client.Get("http://"+engine.APIAddr+"/v1/containers", query, nil)
	if err != nil {
		return err
	}

	defer respContainers.Close()
	if respContainers.StatusCode() != 200 {
		return fmt.Errorf("engine %s, refresh containers failure, %s", engine.IP, ctypes.ParseHTTPResponseError(respContainers))
	}

	dockerContainers := []types.Container{}
	if err := respContainers.JSON(&dockerContainers); err != nil {
		return err
	}

	//logger.INFO("[#cluster#] engine %s %s refresh containers.", engine.IP, engine.State())
	merged := make(map[string]*Container)
	for _, container := range dockerContainers {
		mergedUpdate, err := engine.updateContainer(container.ID, merged)
		if err != nil {
			logger.ERROR("[#cluster#] engine %s update container error, %s", engine.IP, err.Error())
		} else {
			merged = mergedUpdate
		}
	}

	engine.Lock()
	engine.containers = merged
	engine.Unlock()
	return nil
}

// ValidateContainers is exported
// clear engine invalid cluster containers
func (engine *Engine) ValidateContainers() {

	//get expel containers
	expelContainers := Containers{}
	containers := engine.Containers("")
	for _, container := range containers {
		if !container.ValidateConfig() {
			expelContainers = append(expelContainers, container)
		}
	}

	// rename all expel containers, prevent container name conflicts.
	for _, container := range expelContainers {
		expelName := strings.TrimSuffix(container.Info.Name, "-expel") + "-expel"
		if container.Info.Name != expelName {
			operate := models.ContainerOperate{
				Action:    "rename",
				Container: container.Info.ID,
				NewName:   expelName,
			}
			if err := engine.OperateContainer(operate); err != nil {
				logger.ERROR("[#cluster#] engine %s container %s expel, rename error:%s", engine.IP, container.Info.ID[:12], err.Error())
				continue
			}
			logger.WARN("[#cluster#] engine %s container %s expel, rename to %s.", engine.IP, container.Info.ID[:12], operate.NewName)
		}
	}

	// remove all expel containers
	for _, container := range expelContainers {
		if err := engine.RemoveContainer(container.Info.ID); err != nil {
			logger.WARN("[#cluster#] engine %s container %s expel, remove error %s.", engine.IP, container.Info.ID[:12], err.Error())
			continue
		}
		logger.WARN("[#cluster#] engine %s container %s expel, remove.", engine.IP, container.Info.ID[:12])
	}
}

// refreshLoop is exported
// Get container information regularly and engine performance collection
func (engine *Engine) refreshLoop() {

	const perfUpdateInterval = 5 * time.Minute //engine performance collection interval
	lastPrefUpdateAt := time.Now()
	for {
		runTicker := time.NewTicker(refreshInterval)
		select {
		case <-runTicker.C:
			{
				runTicker.Stop()
				if engine.IsHealthy() {
					if time.Since(lastPrefUpdateAt) > perfUpdateInterval {
						//engine.updatePerformance()
						lastPrefUpdateAt = time.Now()
					}
					if err := engine.RefreshContainers(); err != nil {
						logger.ERROR("[#cluster#] engine %s refresh containers error:%s", engine.IP, err.Error())
					}
					engine.ValidateContainers()
				}
			}
		case <-engine.stopCh:
			{
				runTicker.Stop()
				return
			}
		}
	}
}

// updateSpecs exported
func (engine *Engine) updateSpecs() error {

	respSpecs, err := engine.client.Get("http://"+engine.APIAddr+"/v1/dockerinfo", nil, nil)
	if err != nil {
		return err
	}

	defer respSpecs.Close()
	if respSpecs.StatusCode() != 200 {
		return fmt.Errorf("engine %s, update specs failure, %s", engine.IP, ctypes.ParseHTTPResponseError(respSpecs))
	}

	dockerInfo := &types.Info{}
	if err := respSpecs.JSON(dockerInfo); err != nil {
		return err
	}

	//logger.INFO("[#cluster#] engine %s update specs.", engine.IP)
	engine.Lock()
	engine.ID = dockerInfo.ID
	engine.Name = strings.ToUpper(dockerInfo.Name)
	engine.Cpus = int64(dockerInfo.NCPU)
	engine.Memory = int64(math.Ceil(float64(dockerInfo.MemTotal) / 1024.0 / 1024.0))
	if dockerInfo.Driver != "" {
		engine.Labels["storagedirver"] = dockerInfo.Driver
	}

	if dockerInfo.KernelVersion != "" {
		engine.Labels["kernelversion"] = dockerInfo.KernelVersion
	}

	if dockerInfo.OperatingSystem != "" {
		engine.Labels["operatingsystem"] = dockerInfo.OperatingSystem
	}

	for _, label := range dockerInfo.Labels {
		kv := strings.SplitN(label, "=", 2)
		if len(kv) != 2 {
			continue
		}
		engine.Labels[kv[0]] = kv[1]
	}
	engine.Unlock()
	return nil
}

// updateContainer exported
func (engine *Engine) updateContainer(containerid string, containers map[string]*Container) (map[string]*Container, error) {

	var container *Container
	engine.RLock()
	if current, ret := engine.containers[containerid]; ret {
		container = current
	} else {
		container = &Container{
			Engine: engine,
		}
	}
	engine.RUnlock()

	query := map[string][]string{"originaldata": []string{"true"}}
	respContainer, err := engine.client.Get("http://"+engine.APIAddr+"/v1/containers/"+containerid, query, nil)
	if err != nil {
		return nil, err
	}

	defer respContainer.Close()
	if respContainer.StatusCode() != 200 {
		return nil, fmt.Errorf("engine %s, update container %s failure, %s", engine.IP, containerid[:12], ctypes.ParseHTTPResponseError(respContainer))
	}

	containerJSON := &types.ContainerJSON{}
	if err := respContainer.JSON(containerJSON); err != nil {
		return nil, err
	}
	engine.Lock()
	container.update(engine, containerJSON)
	containers[containerid] = container
	engine.Unlock()
	return containers, nil
}
