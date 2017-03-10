package cluster

import "github.com/docker/docker/api/types"
import "github.com/humpback/gounits/http"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-agent/models"

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
	// timeout for request send out to the engine
	requestTimeout = 15 * time.Second
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
	httpClient      *http.HttpClient
	stopCh          chan struct{}
	state           engineState
}

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

func (engine *Engine) Open(addr string) {

	engine.Lock()
	engine.Addr = addr
	if engine.state != StateHealthy {
		engine.state = StateHealthy
		engine.httpClient = http.NewWithTimeout(requestTimeout)
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

func (engine *Engine) Close() {

	engine.cleanupContainers()
	engine.Lock()
	defer engine.Unlock()
	if engine.state == StateDisconnected {
		return
	}
	close(engine.stopCh) //quit refreshLoop.
	engine.httpClient.Close()
	engine.ID = ""
	engine.Name = ""
	engine.Addr = ""
	engine.Labels = make(map[string]string)
	engine.state = StateDisconnected
	logger.INFO("[#cluster#] engine %s closed.", engine.IP)
}

func (engine *Engine) IsHealthy() bool {

	engine.RLock()
	defer engine.RUnlock()
	return engine.state == StateHealthy
}

func (engine *Engine) SetState(state engineState) {

	engine.Lock()
	engine.state = state
	engine.Unlock()
}

func (engine *Engine) State() string {

	engine.RLock()
	defer engine.RUnlock()
	return stateText[engine.state]
}

func (engine *Engine) Containers() Containers {

	engine.RLock()
	containers := Containers{}
	for _, container := range engine.containers {
		containers = append(containers, container)
	}
	engine.RUnlock()
	return containers
}

func (engine *Engine) UsedMemory() int64 {

	var used int64
	engine.RLock()
	for _, c := range engine.containers {
		used += c.Info.HostConfig.Memory
	}
	engine.RUnlock()
	return used
}

func (engine *Engine) UsedCpus() int64 {

	var used int64
	engine.RLock()
	for _, c := range engine.containers {
		used += c.Info.HostConfig.CPUShares
	}
	engine.RUnlock()
	return used
}

func (engine *Engine) TotalMemory() int64 {

	engine.RLock()
	defer engine.RUnlock()
	return engine.Memory + (engine.Memory * engine.overcommitRatio / 100)
}

func (engine *Engine) TotalCpus() int64 {

	engine.RLock()
	defer engine.RUnlock()
	return engine.Cpus + (engine.Cpus * engine.overcommitRatio / 100)
}

func (engine *Engine) CreateContainer(config models.Container) (*Container, error) {

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(config); err != nil {
		return nil, err
	}

	logger.INFO("[#cluster#] engine %s create container %s", engine.IP, config.Name)
	header := map[string][]string{"Content-Type": []string{"application/json"}}
	respCreated, err := engine.httpClient.Post("http://"+engine.Addr+"/v1/containers", nil, buf, header)
	if err != nil {
		return nil, err
	}

	defer respCreated.Close()
	if respCreated.StatusCode() != 200 {
		logger.ERROR("[#cluster#] engine %s create container %s error:%d %s", engine.IP, config.Name, respCreated.StatusCode(), respCreated.String())
		return nil, fmt.Errorf("create container failure")
	}

	createContainerResponse := &CreateContainerResponse{}
	if err := respCreated.JSON(createContainerResponse); err != nil {
		return nil, err
	}

	containers, err := engine.updateContainer(createContainerResponse.ID, engine.containers)
	if err != nil {
		return nil, err
	}

	engine.Lock()
	defer engine.Unlock()
	engine.containers = containers
	container, ret := engine.containers[createContainerResponse.ID]
	if !ret {
		return nil, fmt.Errorf("create container update info failure")
	}
	return container, nil
}

func (engine *Engine) RefreshContainers() error {

	query := map[string][]string{"all": []string{"true"}}
	respContainers, err := engine.httpClient.Get("http://"+engine.Addr+"/v1/containers", query, nil)
	if err != nil {
		return err
	}

	defer respContainers.Close()
	if respContainers.StatusCode() != 200 {
		return fmt.Errorf("GET(%d) %s", respContainers.StatusCode(), respContainers.RawURL())
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
			logger.ERROR("[#cluster#] engine %s update container error:%s", engine.Addr, err.Error())
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

	respSpecs, err := engine.httpClient.Get("http://"+engine.Addr+"/v1/dockerinfo", nil, nil)
	if err != nil {
		return err
	}

	defer respSpecs.Close()
	if respSpecs.StatusCode() != 200 {
		return fmt.Errorf("GET(%d) %s", respSpecs.StatusCode(), respSpecs.RawURL())
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
	respContainer, err := engine.httpClient.Get("http://"+engine.Addr+"/v1/containers/"+containerid, query, nil)
	if err != nil {
		return nil, err
	}

	defer respContainer.Close()
	if respContainer.StatusCode() != 200 {
		return nil, fmt.Errorf("GET(%d) %s", respContainer.StatusCode(), respContainer.RawURL())
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
