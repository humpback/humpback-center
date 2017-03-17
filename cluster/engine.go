package cluster

import "github.com/docker/docker/api/types"
import "github.com/humpback/gounits/http"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-agent/models"
import ctypes "github.com/humpback/humpback-center/cluster/types"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	// engine send request timeout min value
	requestMinTimeout = 15 * time.Second
	// engine send request timeout max value
	requestMaxTimeout = 45 * time.Second
	// engine refresh loop interval
	refreshInterval = 20 * time.Second
	// threshold of delta duration between humpback-center and humpback-agent's systime
	thresholdTime = 2 * time.Second
)

// Engine state define
type engineState int

// State enum value
const (
	//pending: engine added to cluster, but not been validated.
	StatePending engineState = iota
	//unhealthy: engine is unreachable.
	StateUnhealthy
	//healthy: engine is ready reachable.
	StateHealthy
	//disconnected: engine is removed from discovery
	StateDisconnected
)

// Engine state mapping
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
	ID            string
	Name          string
	IP            string
	Addr          string
	Cpus          int64
	Memory        int64
	Labels        map[string]string //docker daemon labels
	DeltaDuration time.Duration     //humpback-center's systime - humpback-agent's systime

	overcommitRatio int64
	configCache     *ContainersConfigCache
	containers      map[string]*Container
	stopCh          chan struct{}
	state           engineState
}

// NewEngine is exported
// Make new engine object
func NewEngine(ip string, overcommitRatio float64, configCache *ContainersConfigCache) (*Engine, error) {

	ipAddr, err := net.ResolveIPAddr("ip4", ip)
	if err != nil {
		return nil, err
	}

	return &Engine{
		IP:              ipAddr.IP.String(),
		overcommitRatio: int64(overcommitRatio * 100),
		Labels:          make(map[string]string),
		containers:      make(map[string]*Container),
		configCache:     configCache,
		state:           StateDisconnected,
	}, nil
}

// Open is exported
// Open a engine refresh loop, update engine and refresh containers.
func (engine *Engine) Open(addr string) {

	engine.Lock()
	engine.Addr = addr
	if engine.state != StateHealthy {
		engine.state = StateHealthy
		engine.stopCh = make(chan struct{})
		logger.INFO("[#cluster#] engine %s open.", engine.IP)
		go func() {
			engine.updateSpecs()
			engine.RefreshContainers()
			engine.refreshLoop()
		}()
	}
	engine.Unlock()
}

// Close is exported
// Close a engine, reset engine info.
func (engine *Engine) Close() {

	//engine.cleanupContainers()
	engine.Lock()
	defer engine.Unlock()
	if engine.state == StateDisconnected {
		return
	}
	close(engine.stopCh) //quit refreshLoop.
	engine.ID = ""
	engine.Name = ""
	engine.Addr = ""
	engine.Labels = make(map[string]string)
	engine.state = StateDisconnected
	logger.INFO("[#cluster#] engine %s closed.", engine.IP)
}

// IsHealthy is exported
// Return a engine ishealthy status bool result.
func (engine *Engine) IsHealthy() bool {

	engine.RLock()
	defer engine.RUnlock()
	return engine.state == StateHealthy
}

// SetState is exported
// Set a engine status.
func (engine *Engine) SetState(state engineState) {

	engine.Lock()
	engine.state = state
	engine.Unlock()
}

// State is exported
// Return a engine status string result.
func (engine *Engine) State() string {

	engine.RLock()
	defer engine.RUnlock()
	return stateText[engine.state]
}

// Containers is exported
// Return a engine all containers.
func (engine *Engine) Containers() Containers {

	engine.RLock()
	containers := Containers{}
	for _, container := range engine.containers {
		containers = append(containers, container)
	}
	engine.RUnlock()
	return containers
}

// UsedMemory is exported
// Return a engine all containers used memory size.
func (engine *Engine) UsedMemory() int64 {

	var used int64
	engine.RLock()
	for _, c := range engine.containers {
		used += c.Info.HostConfig.Memory
	}
	engine.RUnlock()
	return used
}

// UsedCpus is exported
// Return a engine all containers used cpus size.
func (engine *Engine) UsedCpus() int64 {

	var used int64
	engine.RLock()
	for _, c := range engine.containers {
		used += c.Info.HostConfig.CPUShares
	}
	engine.RUnlock()
	return used
}

// TotalMemory is exported
// Return a engine total memory size.
func (engine *Engine) TotalMemory() int64 {

	engine.RLock()
	defer engine.RUnlock()
	return engine.Memory + (engine.Memory * engine.overcommitRatio / 100)
}

// TotalCpus is exported
// Return a engine total cpus size.
func (engine *Engine) TotalCpus() int64 {

	engine.RLock()
	defer engine.RUnlock()
	return engine.Cpus + (engine.Cpus * engine.overcommitRatio / 100)
}

// CreateContainer is exported
// Engine create a container.
func (engine *Engine) CreateContainer(config models.Container) (*Container, error) {

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(config); err != nil {
		return nil, err
	}

	header := map[string][]string{"Content-Type": []string{"application/json"}}
	respCreated, err := http.NewWithTimeout(requestMaxTimeout).Post("http://"+engine.Addr+"/v1/containers", nil, buf, header)
	if err != nil {
		return nil, err
	}

	defer respCreated.Close()
	if respCreated.StatusCode() != 200 {
		return nil, fmt.Errorf("engine %s, create container %s failure %d", engine.IP, config.Name, respCreated.StatusCode())
	}

	createContainerResponse := &ctypes.CreateContainerResponse{}
	if err := respCreated.JSON(createContainerResponse); err != nil {
		return nil, err
	}

	logger.INFO("[#cluster#] engine %s, create container %s:%s", engine.IP, createContainerResponse.ID[:12], config.Name)
	containers, err := engine.updateContainer(createContainerResponse.ID, engine.containers)
	if err != nil {
		return nil, err
	}

	engine.Lock()
	defer engine.Unlock()
	engine.containers = containers
	container, ret := engine.containers[createContainerResponse.ID]
	if !ret {
		return nil, fmt.Errorf("created container, update container info failure")
	}
	return container, nil
}

// RemoveContainer is exported
// Engine remove a container.
func (engine *Engine) RemoveContainer(containerid string) error {

	query := map[string][]string{"force": []string{"true"}}
	respRemoved, err := http.NewWithTimeout(requestMinTimeout).Delete("http://"+engine.Addr+"/v1/containers/"+containerid, query, nil)
	if err != nil {
		return err
	}

	defer respRemoved.Close()
	if respRemoved.StatusCode() != 200 {
		return fmt.Errorf("engine %s, remove container %s failure %d", engine.IP, containerid[:12], respRemoved.StatusCode())
	}

	logger.INFO("[#cluster#] engine %s, remove container %s", engine.IP, containerid[:12])
	engine.Lock()
	delete(engine.containers, containerid)
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
	respOperated, err := http.NewWithTimeout(requestMinTimeout).Put("http://"+engine.Addr+"/v1/containers", nil, buf, header)
	if err != nil {
		return err
	}

	defer respOperated.Close()
	if respOperated.StatusCode() != 200 {
		return fmt.Errorf("engine %s, %s container %s failure %d", engine.IP, operate.Action, operate.Container[:12], respOperated.StatusCode())
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
	respUpgraded, err := http.NewWithTimeout(requestMaxTimeout).Put("http://"+engine.Addr+"/v1/containers", nil, buf, header)
	if err != nil {
		return nil, err
	}

	defer respUpgraded.Close()
	if respUpgraded.StatusCode() != 200 {
		return nil, fmt.Errorf("%s container %s failure %d", operate.Action, operate.Container[:12], respUpgraded.StatusCode())
	}

	upgradeContainerResponse := &ctypes.UpgradeContainerResponse{}
	if err := respUpgraded.JSON(upgradeContainerResponse); err != nil {
		return nil, err
	}

	logger.INFO("[#cluster#] engine %s %s container %s to %s", engine.IP, operate.Action, operate.Container[:12], upgradeContainerResponse.ID[:12])
	if err := engine.RefreshContainers(); err != nil {
		return nil, err
	}

	engine.Lock()
	defer engine.Unlock()
	container, ret := engine.containers[upgradeContainerResponse.ID]
	if !ret {
		return nil, fmt.Errorf("upgrade container, update container info failure")
	}
	return container, nil
}

// RefreshContainers is exported
// Engine refresh all containers.
func (engine *Engine) RefreshContainers() error {

	query := map[string][]string{"all": []string{"true"}}
	respContainers, err := http.NewWithTimeout(requestMinTimeout).Get("http://"+engine.Addr+"/v1/containers", query, nil)
	if err != nil {
		return err
	}

	defer respContainers.Close()
	if respContainers.StatusCode() != 200 {
		return fmt.Errorf("http GET response statuscode %d", respContainers.StatusCode())
	}

	dockerContainers := []types.Container{}
	if err := respContainers.JSON(&dockerContainers); err != nil {
		return err
	}

	//logger.INFO("[#cluster#] engine %s refresh containers.", engine.Addr)
	merged := make(map[string]*Container)
	for _, c := range dockerContainers {
		mergedUpdate, err := engine.updateContainer(c.ID, merged)
		if err != nil {
			logger.ERROR("[#cluster#] engine refresh containers error, %s", err.Error())
		} else {
			merged = mergedUpdate
		}
	}

	engine.Lock()
	engine.containers = merged
	engine.Unlock()
	return nil
}

func (engine *Engine) cleanupContainers() {

	engine.Lock()
	engine.containers = make(map[string]*Container)
	engine.Unlock()
}

func (engine *Engine) refreshLoop() {

	const specsUpdateInterval = 5 * time.Minute
	lastSpecsUpdateAt := time.Now()
	for {
		runTicker := time.NewTicker(refreshInterval)
		select {
		case <-runTicker.C:
			{
				runTicker.Stop()
				isHealthy := engine.IsHealthy()
				if !isHealthy || time.Since(lastSpecsUpdateAt) > specsUpdateInterval {
					if err := engine.updateSpecs(); err != nil {
						logger.ERROR("[#cluster#] engine %s update specs error:%s", engine.Addr, err.Error())
						continue
					}
					lastSpecsUpdateAt = time.Now()
				}
				if err := engine.RefreshContainers(); err != nil {
					logger.ERROR("[#cluster#] engine %s refresh containers error:%s", engine.Addr, err.Error())
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

func (engine *Engine) updateSpecs() error {

	respSpecs, err := http.NewWithTimeout(requestMinTimeout).Get("http://"+engine.Addr+"/v1/dockerinfo", nil, nil)
	if err != nil {
		return err
	}

	defer respSpecs.Close()
	if respSpecs.StatusCode() != 200 {
		return fmt.Errorf("http GET response statuscode %d", respSpecs.StatusCode())
	}

	dockerInfo := &types.Info{}
	if err := respSpecs.JSON(dockerInfo); err != nil {
		return err
	}

	//logger.INFO("[#cluster#] engine %s update specs.", engine.Addr)
	engine.Lock()
	defer engine.Unlock()
	engine.ID = dockerInfo.ID
	engine.Name = dockerInfo.Name
	engine.Cpus = int64(dockerInfo.NCPU)
	engine.Memory = int64(math.Ceil(float64(dockerInfo.MemTotal) / 1024.0 / 1024.0))

	var delta time.Duration
	if dockerInfo.SystemTime != "" {
		engineSystime, _ := time.Parse(time.RFC3339Nano, dockerInfo.SystemTime)
		delta = time.Now().UTC().Sub(engineSystime)
	} else {
		delta = time.Duration(0)
	}

	absDelta := delta
	if delta.Seconds() < 0 {
		absDelta = time.Duration(-1*delta.Seconds()) * time.Second
	}

	if absDelta < thresholdTime {
		engine.DeltaDuration = 0
	} else {
		engine.DeltaDuration = delta
	}

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
	return nil
}

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
	respContainer, err := http.NewWithTimeout(requestMinTimeout).Get("http://"+engine.Addr+"/v1/containers/"+containerid, query, nil)
	if err != nil {
		return nil, err
	}

	defer respContainer.Close()
	if respContainer.StatusCode() != 200 {
		return nil, fmt.Errorf("engine %s, update container %s failure %d", engine.IP, containerid[:12], respContainer.StatusCode())
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
