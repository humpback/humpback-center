package api

import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/repository"
import "github.com/gorilla/mux"

import (
	"net/http"
)

type handler func(ctx *Context, w http.ResponseWriter, r *http.Request)

var routes = map[string]map[string]handler{
	"GET": {
		"/v1/_ping":                     ping,
		"/v1/cluster/engines":           getClusterEngines,
		"/v1/repository/images/refresh": getImagesRefresh,
		"/v1/repository/images/catalog": getImagesCatalog,
		"/v1/repository/images/tags/*":  getImagesTags,
	},
	"POST": {
		"v1/repository/images/migrate": postImagesMigrate,
	},
	"DELETE": {
		"v1/repository/images/{name:.*}": deleteImages,
	},
}

func NewRouter(cluster *cluster.Cluster, repositorycache *repository.RepositoryCache, enableCors bool) *mux.Router {

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
				ctx := NewContext(cluster, repositorycache)
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
					ctx := NewContext(cluster, repositorycache)
					optionshandler(ctx, w, r)
				}
				router.Path(routepattern).Methods(optionsmethod).HandlerFunc(wrap)
			}
		}
	}
	return router
}

func ping(ctx *Context, w http.ResponseWriter, r *http.Request) {

	ctx.JSON(w, http.StatusOK, "PANG")
}

func optionsHandler(ctx *Context, w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
}
