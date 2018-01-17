package ctrl

import "github.com/humpback/gounits/httpx"
import "github.com/humpback/gounits/logger"
import "humpback-center/cluster"
import "humpback-center/etc"

import (
	"net"
	"net/http"
	"time"
)

// Controller is exprted
type Controller struct {
	client        *httpx.HttpClient
	Configuration *etc.Configuration
	Cluster       *cluster.Cluster
}

// NewController is exported
func NewController(configuration *etc.Configuration) (*Controller, error) {

	cluster, err := createCluster(configuration)
	if err != nil {
		return nil, err
	}

	client := httpx.NewClient().
		SetTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   45 * time.Second,
				KeepAlive: 90 * time.Second,
			}).DialContext,
			DisableKeepAlives:     false,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout,
			ExpectContinueTimeout: http.DefaultTransport.(*http.Transport).ExpectContinueTimeout,
		})

	return &Controller{
		client:        client,
		Configuration: configuration,
		Cluster:       cluster,
	}, nil
}

// Initialize is exported
// init cluster
func (c *Controller) Initialize() error {

	logger.INFO("[#ctrl#] controller initialize.....")
	logger.INFO("[#ctrl#] configuration %+v", c.Configuration)
	return c.startCluster()
}

// UnInitialize is exported
// uninit cluster
func (c *Controller) UnInitialize() {

	c.stopCluster()
	logger.INFO("[#ctrl#] controller uninitialized.")
}
