package api

import (
	"fmt"
	"net/http"
)

func getClusterEngines(ctx *Context) {

	p := ctx.Query("callback")
	fmt.Printf("pcal:%s\n", p)
	ctx.JSON(http.StatusOK, "engines...")
}
