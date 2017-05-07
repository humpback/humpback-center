package ctrl

import "github.com/humpback/gounits/http"
import "github.com/humpback/gounits/logger"
import "humpback-center/cluster"
import "humpback-center/etc"
import "humpback-center/repository"

import (
	"time"
)

const (
	// humpback-api site request timeout value
	requestAPITimeout = 15 * time.Second
)

// Controller is exprted
type Controller struct {
	httpClient      *http.HttpClient
	Configuration   *etc.Configuration
	Cluster         *cluster.Cluster
	RepositoryCache *repository.RepositoryCache
}

// NewController is exported
func NewController(configuration *etc.Configuration) (*Controller, error) {

	cluster, err := createCluster(configuration)
	if err != nil {
		return nil, err
	}

	repositorycache, err := createRepositoryCache(configuration)
	if err != nil {
		return nil, err
	}

	return &Controller{
		httpClient:      http.NewWithTimeout(requestAPITimeout),
		Configuration:   configuration,
		Cluster:         cluster,
		RepositoryCache: repositorycache,
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
