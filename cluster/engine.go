package cluster

import "github.com/humpback/gounits/http"
import "github.com/humpback/gounits/convert"

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

	containers map[string]*Container
	httpClient *http.HttpClient
	stopCh     chan struct{}
	state      engineState
}

func NewEngine(id string, opts *RegistClusterOptions) (*Engine, error) {

	host, _, err := net.SplitHostPort(opts.Addr)
	if err != nil {
		return nil, err
	}

	ipaddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return nil, err
	}

	return &Engine{
		ID:         id,
		Name:       opts.Name,
		IP:         ipaddr.IP.String(),
		Addr:       opts.Addr,
		Version:    opts.Version,
		Labels:     convert.ConvertKVStringSliceToMap(opts.Labels),
		containers: make(map[string]*Container),
		stopCh:     make(chan struct{}),
		state:      stateUnhealthy,
	}, nil
}

func (engine *Engine) String() string {

	return engine.ID
}

func (engine *Engine) IsHealthy() bool {

	engine.Lock()
	defer engine.Unlock()
	return engine.state == stateHealthy
}

func (engine *Engine) setState(state engineState) {

	engine.Lock()
	engine.state = state
	engine.Unlock()
}

func (engine *Engine) Status() string {

	engine.Lock()
	defer engine.Unlock()
	return stateText[engine.state]
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

func (engine *Engine) Close() {

	if engine.httpClient != nil {
		engine.httpClient.Close()
	}
	// close the chan, exit refreshLoop.
	close(engine.stopCh)
}
