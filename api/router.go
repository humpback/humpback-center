package api

import "github.com/humpback/humpback-center/cluster"
import "github.com/gorilla/mux"

import (
	"crypto/tls"
	"net/http"
)

type handler func(ctx *Context, w http.ResponseWriter, r *http.Request)

var routes = map[string]map[string]handler{
	"GET": {
		"/v1/_ping":           ping,
		"/v1/cluster/engines": getClusterEngines,
	},
}

func NewRouter(cluster *cluster.Cluster, tlsConfig *tls.Config, enableCors bool) *mux.Router {

	router := mux.NewRouter()
	for method, mappings := range routes {
		for route, handler := range mappings {
			routemethod := method
			routepattern := route
			routehandler := handler
			wrap := func(w http.ResponseWriter, r *http.Request) {
				if enableCors {
					writeCorsHeaders(w, r)
				}
				ctx := NewContext(cluster, tlsConfig)
				routehandler(ctx, w, r)
			}
			router.Path(routepattern).Methods(routemethod).HandlerFunc(wrap)
			if enableCors {
				optionsmethod := "OPTIONS"
				optionshandler := optionsHandler
				wrap := func(w http.ResponseWriter, r *http.Request) {
					if enableCors {
						writeCorsHeaders(w, r)
					}
					ctx := NewContext(cluster, tlsConfig)
					optionshandler(ctx, w, r)
				}
				router.Path(routepattern).Methods(optionsmethod).HandlerFunc(wrap)
			}
		}
	}
	return router
}
