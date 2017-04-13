package api

import "humpback-center/api/middleware"
import "humpback-center/ctrl"

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
)

type Dispatcher struct {
	handler http.Handler
}

func (dispatcher *Dispatcher) SetHandler(handler http.Handler) {

	dispatcher.handler = handler
}

func (dispatcher *Dispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if dispatcher.handler == nil {
		httpError(w, "API Dispatcher Invalid.", http.StatusInternalServerError)
		return
	}
	handler := middleware.Logger(dispatcher.handler)
	handler.ServeHTTP(w, r)
}

type Server struct {
	hosts      []string
	tlsConfig  *tls.Config
	dispatcher *Dispatcher
}

func NewServer(hosts []string, tlsConfig *tls.Config, controller *ctrl.Controller, enablecors bool) *Server {

	router := NewRouter(controller, enablecors)
	return &Server{
		hosts:     hosts,
		tlsConfig: tlsConfig,
		dispatcher: &Dispatcher{
			handler: router,
		},
	}
}

func (server *Server) ListenHosts() []string {

	return server.hosts
}

func (server *Server) SetHandler(handler http.Handler) {

	server.dispatcher.SetHandler(handler)
}

func (server *Server) Startup() error {

	errorsCh := make(chan error, len(server.hosts))
	for _, host := range server.hosts {
		protoAddrParts := strings.SplitN(host, "://", 2)
		if len(protoAddrParts) == 1 {
			protoAddrParts = append([]string{"tcp"}, protoAddrParts...)
		}

		go func() {
			var (
				err error
				l   net.Listener
				s   = http.Server{
					Addr:    protoAddrParts[1],
					Handler: server.dispatcher,
				}
			)

			switch protoAddrParts[0] {
			case "unix":
				l, err = newUnixListener(protoAddrParts[1], server.tlsConfig)
			case "tcp":
				l, err = newListener("tcp", protoAddrParts[1], server.tlsConfig)
			default:
				err = fmt.Errorf("API UnSupported Protocol:%q", protoAddrParts[0])
			}
			if err != nil {
				errorsCh <- err
			} else {
				errorsCh <- s.Serve(l)
			}
		}()
	}

	for i := 0; i < len(server.hosts); i++ {
		err := <-errorsCh
		if err != nil {
			return err
		}
	}
	return nil
}

func newListener(proto string, addr string, tlsConfig *tls.Config) (net.Listener, error) {

	l, err := net.Listen(proto, addr)
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		tlsConfig.NextProtos = []string{"http/1.1"}
		l = tls.NewListener(l, tlsConfig)
	}
	return l, nil
}
