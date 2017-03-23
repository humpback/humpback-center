package ctrl

import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/etc"
import "github.com/humpback/humpback-center/repository"
import "github.com/humpback/gounits/logger"

type Controller struct {
	Configuration   *etc.Configuration
	Cluster         *cluster.Cluster
	RepositoryCache *repository.RepositoryCache
}

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
		Configuration:   configuration,
		Cluster:         cluster,
		RepositoryCache: repositorycache,
	}, nil
}

func (c *Controller) Initialize() error {

	logger.INFO("[#ctrl#] controller initialize.....")
	return c.startCluster()
}

func (c *Controller) UnInitialize() {

	c.stopCluster()
	logger.INFO("[#ctrl#] controller uninitialized.")
}
