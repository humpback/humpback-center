package api

import "net/http"

func getImagesRefresh(ctx *Context, w http.ResponseWriter, r *http.Request) {

	ctx.JSON(w, http.StatusOK, "Refresh...")
}

func getImagesCatalog(ctx *Context, w http.ResponseWriter, r *http.Request) {

	ctx.JSON(w, http.StatusOK, "Catalog.")
}

func getImagesTags(ctx *Context, w http.ResponseWriter, r *http.Request) {

	ctx.JSON(w, http.StatusOK, "Tags.")
}

func postImagesMigrate(ctx *Context, w http.ResponseWriter, r *http.Request) {

	ctx.JSON(w, http.StatusOK, "Migrate...")
}

func deleteImages(ctx *Context, w http.ResponseWriter, r *http.Request) {

	ctx.JSON(w, http.StatusOK, "Remove.")
}
