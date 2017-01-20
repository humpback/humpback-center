package api

import "net/http"

func ping(ctx *Context, w http.ResponseWriter, r *http.Request) {

	ctx.JSON(w, http.StatusOK, "PANG")
}

func getClusterEngines(ctx *Context, w http.ResponseWriter, r *http.Request) {

	//engines := ctx.Cluster.GetEngines()
	ctx.JSON(w, http.StatusOK, "engines")
}

func optionsHandler(ctx *Context, w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
}
