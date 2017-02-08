package cluster

import "github.com/humpback/gounits/http"

import (
	"net"
)

type Engine struct {
	ID         string
	IP         string
	Addr       string
	Cpus       int64 //定期获取
	Memory     int64 //定期获取
	httpClient *http.HttpClient
	stopCh     chan struct{}
}

func NewEngine(id string, addr string) (*Engine, error) {

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ipaddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, err
	}

	return &Engine{
		ID:     id,
		IP:     ipaddr.IP.String(),
		Addr:   addr,
		stopCh: make(chan struct{}),
	}, nil
}

func (engine *Engine) String() string {

	return engine.ID
}

//func (e *Engine) refreshLoop() {
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
