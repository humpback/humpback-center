package api

import "net/http"

func getClusterEngines(ctx *Context, w http.ResponseWriter, r *http.Request) {

	ctx.JSON(w, http.StatusOK, "engines...")
}
