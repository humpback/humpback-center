package cluster

import "github.com/humpback/gounits/convert"
import "github.com/humpback/gounits/http"
import "github.com/humpback/humpback-center/cluster/types"

import (
	"net"
	"sync"
)

// Engine state define
type engineState int

const (
	//pending: engine added to cluster, but not been validated.
	statePending engineState = iota
	//unhealthy: engine is unreachable.
	stateUnhealthy
	//healthy: engine is ready reachable.
	stateHealthy
	//disconnected: engine is removed from discovery
	stateDisconnected
)

// Engine state mapping
var stateText = map[engineState]string{
	statePending:      "Pending",
	stateUnhealthy:    "Unhealthy",
	stateHealthy:      "Healthy",
	stateDisconnected: "Disconnected",
}

// Engine is exported
//Labels, engine labels values {"node","wh7", "kernelversion":"4.4.0", "os":"centos6.8"}
type Engine struct {
	sync.RWMutex
	ID      string
	Name    string
	IP      string
	Addr    string
	Cpus    int64
	Memory  int64
	Version string
	Labels  map[string]string

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
		state:  stateUnhealthy,
	}, nil
}

func (engine *Engine) SetRegistOptions(opts *types.ClusterRegistOptions) {

	engine.ID = opts.ID
	engine.Name = opts.Name
	engine.IP = opts.IP
	engine.Addr = opts.Addr
	engine.Version = opts.Version
	engine.Labels = convert.ConvertKVStringSliceToMap(opts.Labels)
}

func (engine *Engine) IsHealthy() bool {

	engine.Lock()
	defer engine.Unlock()
	return engine.state == stateHealthy
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
