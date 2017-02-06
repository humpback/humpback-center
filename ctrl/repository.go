package ctrl

import "github.com/humpback/humpback-center/etc"
import "github.com/humpback/humpback-center/repository"

func createRepositoryCache(configuration *etc.Configuration) (*repository.RepositoryCache, error) {

	repositorycache := repository.NewRepositoryCache()
	return repositorycache, nil
}

func (c *Controller) SetRepositoryCache(repositorycache *repository.RepositoryCache) {

	if repositorycache != nil {
		c.RepositoryCache = repositorycache
	}
}
