package api

import "net/http"

func getImagesRefresh(ctx *Context) {

	ctx.JSON(http.StatusOK, "Refresh...")
}

func getImagesCatalog(ctx *Context) {

	ctx.JSON(http.StatusOK, "Catalog.")
}

func getImagesTags(ctx *Context) {

	ctx.JSON(http.StatusOK, "Tags.")
}

func postImagesMigrate(ctx *Context) {

	ctx.JSON(http.StatusOK, "Migrate...")
}

func deleteImages(ctx *Context) {

	ctx.JSON(http.StatusOK, "Remove.")
}
