package ctrl

import "github.com/humpback/humpback-center/models"
import "github.com/humpback/gounits/logger"

import (
	"fmt"
)

func (c *Controller) InitCluster() error {

	c.Cluster.ClearGroups()
	groups, err := c.DataStorage.GetGroups()
	if err != nil {
		return fmt.Errorf("init cluster error:%s", err.Error())
	}
	logger.INFO("[#ctrl#] init cluster groups:%d", len(groups))
	for _, group := range groups {
		c.Cluster.CreateGroup(group.ID, group.Servers)
	}
	return nil
}

func (c *Controller) startCluster() error {

	logger.INFO("[#ctrl#] start cluster.")
	return c.Cluster.Start()
}

func (c *Controller) stopCluster() {

	logger.INFO("[#ctrl#] stop cluster.")
	c.Cluster.Stop()
}

func (c *Controller) GetClusterGroups() []*models.Group {

	return c.Cluster.GetGroups()
}

func (c *Controller) GetClusterGroup(groupid string) *models.Group {

	return c.Cluster.GetGroup(groupid)
}
