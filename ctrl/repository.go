package ctrl

import "humpback-center/etc"
import "humpback-center/repository"

func createRepositoryCache(configuration *etc.Configuration) (*repository.RepositoryCache, error) {

	repositorycache := repository.NewRepositoryCache()
	return repositorycache, nil
}

func (c *Controller) SetRepositoryCache(repositorycache *repository.RepositoryCache) {

	if repositorycache != nil {
		c.RepositoryCache = repositorycache
	}
}
