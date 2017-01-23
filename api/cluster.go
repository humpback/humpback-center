package api

import (
	"fmt"
	"net/http"
)

func getClusterEngines(c *Context) {

	p1 := c.Query("callback")
	fmt.Printf("p1:%s\n", p1)
	p2 := c.Query("name")
	fmt.Printf("p2:%s\n", p2)
	p3 := c.Query("value")
	fmt.Printf("p3:%s\n", p3)

	c.JSON(http.StatusOK, "engines...")
}
