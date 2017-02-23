package cluster

import "github.com/docker/engine-api/types"
import "github.com/humpback/gounits/http"

import (
	"net"
	"strings"
	"sync"
	"time"
)

const (
	// timeout for request send out to the engine
	requestTimeout = 10 * time.Second
	// engine refresh loop interval
	refreshInterval = 10 * time.Second
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
	ID     string
	Name   string
	IP     string
	Addr   string
	Cpus   int64
	Memory int64
	Labels map[string]string //docker daemon labels

	containers map[string]*Container
	httpClient *http.HttpClient
	stopCh     chan struct{}
	state      engineState
}

func NewEngine(ip string) (*Engine, error) {

	ipaddr, err := net.ResolveIPAddr("ip4", ip)
	if err != nil {
		return nil, err
	}

	return &Engine{
		IP:         ipaddr.IP.String(),
		Labels:     make(map[string]string),
		containers: make(map[string]*Container),
		httpClient: http.NewWithTimeout(requestTimeout),
		stopCh:     make(chan struct{}),
		state:      StateDisconnected,
	}, nil
}

func (engine *Engine) Open(addr string) {

	engine.Lock()
	engine.Addr = addr
	if engine.state != StateHealthy {
		engine.state = StateHealthy
		go func() {
			engine.updateSpecs() //first update specs
			engine.refreshLoop() //loop
		}()
	}
	engine.Unlock()
}

func (engine *Engine) Close() {

	engine.Lock()
	defer engine.Unlock()
	if engine.state == StateDisconnected {
		return
	}
	close(engine.stopCh) //quit refreshLoop.
	engine.httpClient.Close()
	engine.state = StateDisconnected
	engine.Addr = ""
}

func (engine *Engine) IsHealthy() bool {

	engine.Lock()
	defer engine.Unlock()
	return engine.state == StateHealthy
}

func (engine *Engine) SetState(state engineState) {

	engine.Lock()
	engine.state = state
	engine.Unlock()
}

func (engine *Engine) State() string {

	engine.Lock()
	defer engine.Unlock()
	return stateText[engine.state]
}

func (engine *Engine) refreshLoop() {

	for {
		tick := time.NewTicker(refreshInterval)
		select {
		case <-tick.C:
			{
				tick.Stop()
				engine.updateSpecs()
			}
		case <-engine.stopCh:
			{
				tick.Stop()
				return
			}
		}
	}
}

func (engine *Engine) updateSpecs() error {

	resp, err := engine.httpClient.Get("http://"+engine.Addr+"/v1/dockerinfo", nil, nil)
	if err != nil {
		return err
	}

	defer resp.Close()
	info := &types.Info{}
	if err := resp.JSON(info); err != nil {
		return err
	}

	engine.ID = info.ID
	engine.Name = info.Name
	engine.Cpus = int64(info.NCPU)
	engine.Memory = info.MemTotal
	if info.Driver != "" {
		engine.Labels["storagedirver"] = info.Driver
	}

	if info.KernelVersion != "" {
		engine.Labels["kernelversion"] = info.KernelVersion
	}

	if info.OperatingSystem != "" {
		engine.Labels["operatingsystem"] = info.OperatingSystem
	}

	for _, label := range info.Labels {
		kv := strings.SplitN(label, "=", 2)
		if len(kv) != 2 {
			continue
		}
		engine.Labels[kv[0]] = kv[1]
	}
	return nil
}
