package api

import "net/http"

func getImagesRefresh(c *Context) {

	c.JSON(http.StatusOK, "Refresh...")
}

func getImagesCatalog(c *Context) {

	c.JSON(http.StatusOK, "Catalog.")
}

func getImagesTags(c *Context) {

	c.JSON(http.StatusOK, "Tags.")
}

func postImagesMigrate(c *Context) {

	c.JSON(http.StatusOK, "Migrate...")
}

func deleteImages(c *Context) {

	c.JSON(http.StatusOK, "Remove.")
}
