package api

import "github.com/gorilla/mux"
import "humpback-center/ctrl"

import (
	"net/http"
)

type handler func(c *Context) error

var routes = map[string]map[string]handler{
	"GET": {
		"/v1/_ping":                            ping,
		"/v1/groups/{groupid}/collections":     getGroupAllContainers,
		"/v1/groups/{groupid}/engines":         getGroupEngines,
		"/v1/groups/collections/{metaid}":      getGroupContainers,
		"/v1/groups/collections/{metaid}/base": getGroupContainersMetaBase,
		"/v1/groups/engines/{server}":          getGroupEngine,
		"/v1/repository/images/catalog":        getRepositoryImagesCatalog,
		"/v1/repository/images/tags/*":         getRepositoryImagesTags,
	},
	"POST": {
		"/v1/groups/event":              postGroupEvent,
		"/v1/groups/collections":        postGroupCreateContainers,
		"/v1/repository/images/migrate": postRepositoryImagesMigrate,
	},
	"PUT": {
		"/v1/groups/collections":         putGroupUpdateContainers,
		"/v1/groups/collections/upgrade": putGroupUpgradeContainers,
		"/v1/groups/collections/action":  putGroupOperateContainers,
		"/v1/groups/container/action":    putGroupOperateContainer,
	},
	"DELETE": {
		"/v1/groups/collections/{metaid}":    deleteGroupRemoveContainers,
		"/v1/groups/container/{containerid}": deleteGroupRemoveContainer,
		"/v1/repository/images/{name:.*}":    deleteRepositoryImages,
	},
}

func NewRouter(controller *ctrl.Controller, enableCors bool) *mux.Router {

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
				c := NewContext(w, r, controller)
				routehandler(c)
			}
			router.Path(routepattern).Methods(routemethod).HandlerFunc(wrap)
			if enableCors {
				optionsmethod := "OPTIONS"
				optionshandler := optionsHandler
				wrap := func(w http.ResponseWriter, r *http.Request) {
					if enableCors {
						writeCorsHeaders(w, r)
					}
					c := NewContext(w, r, controller)
					optionshandler(c)
				}
				router.Path(routepattern).Methods(optionsmethod).HandlerFunc(wrap)
			}
		}
	}
	return router
}

func ping(ctx *Context) error {

	return ctx.JSON(http.StatusOK, "PANG")
}

func optionsHandler(ctx *Context) error {

	ctx.WriteHeader(http.StatusOK)
	return nil
}
