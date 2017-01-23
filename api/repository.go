package api

import "net/http"

func getRepositoryImagesCatalog(c *Context) {

	c.JSON(http.StatusOK, "Catalog.")
}

func getRepositoryImagesTags(c *Context) {

	c.JSON(http.StatusOK, "Tags.")
}

func postRepositoryImagesMigrate(c *Context) {

	c.JSON(http.StatusOK, "Migrate...")
}

func deleteRepositoryImages(c *Context) {

	c.JSON(http.StatusOK, "Remove.")
}
