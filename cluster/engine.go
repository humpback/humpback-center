package cluster

import "github.com/humpback/gounits/http"
import "github.com/humpback/humpback-center/cluster/types"

import (
	"net"
	"sync"
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
		IP:     ipaddr.IP.String(),
		Labels: make(map[string]string),
		stopCh: make(chan struct{}),
		state:  StateUnhealthy,
	}, nil
}

func (engine *Engine) SetRegistOptions(opts *types.ClusterRegistOptions) *Engine {

	if opts != nil {
		engine.Addr = opts.Addr
	} else {
		engine.Addr = ""
	}
	return engine
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

func (engine *Engine) Close() {

	if engine.httpClient != nil {
		engine.httpClient.Close()
	}
	// close the chan, exit refreshLoop.
	close(engine.stopCh)
}

//func (e *Engine) refreshLoop() {
//定期获取如下信息：
//Cpus       int64
//Memory     int64
//Containers 容器列表信息
//select {
//	case <-e.stopCh:
//		return
//	}
//}
