package api

import "net/http"

func getRepositoryImagesCatalog(c *Context) error {

	return c.JSON(http.StatusOK, "Catalog.")
}

func getRepositoryImagesTags(c *Context) error {

	return c.JSON(http.StatusOK, "Tags.")
}

func postRepositoryImagesMigrate(c *Context) error {

	return c.JSON(http.StatusOK, "Migrate...")
}

func deleteRepositoryImages(c *Context) error {

	return c.JSON(http.StatusOK, "Remove.")
}
