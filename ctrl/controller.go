package ctrl

import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/repository"
import "github.com/humpback/humpback-center/storage"
import "github.com/humpback/gounits/logger"

type Controller struct {
	Cluster         *cluster.Cluster
	DataStorage     *storage.DataStorage
	RepositoryCache *repository.RepositoryCache
}

func NewController(cluster *cluster.Cluster, repositorycache *repository.RepositoryCache,
	dataStorage *storage.DataStorage) *Controller {

	return &Controller{
		Cluster:         cluster,
		DataStorage:     dataStorage,
		RepositoryCache: repositorycache,
	}
}

func (c *Controller) SetCluster(cluster *cluster.Cluster) {

	if cluster != nil {
		c.Cluster = cluster
	}
}

func (c *Controller) SetRepositoryCache(repositorycache *repository.RepositoryCache) {

	if repositorycache != nil {
		c.RepositoryCache = repositorycache
	}
}

func (c *Controller) Initialize() error {

	logger.INFO("[#ctrl#] controller initialize.....")
	if err := c.InitCluster(); err != nil {
		return err
	}
	return c.startCluster()
}

func (c *Controller) UnInitialize() {

	c.stopCluster()
	logger.INFO("[#ctrl#] controller uninitialized.")
}
