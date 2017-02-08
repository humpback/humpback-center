package ctrl

import "github.com/humpback/discovery"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-center/api/request"
import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/etc"
import "github.com/humpback/humpback-center/models"

import (
	"fmt"
	"time"
)

func createCluster(configuration *etc.Configuration) (*cluster.Cluster, error) {

	heartbeat, err := time.ParseDuration(configuration.Discovery.Heartbeat)
	if err != nil {
		return nil, fmt.Errorf("discovery heartbeat invalid.")
	}

	if heartbeat < 1*time.Second {
		return nil, fmt.Errorf("discovery heartbeat should be at least 1s.")
	}

	configopts := map[string]string{"kv.path": configuration.Discovery.Cluster}
	discovery, err := discovery.New(configuration.Discovery.URIs, heartbeat, 0, configopts)
	if err != nil {
		return nil, err
	}

	cluster, err := cluster.NewCluster(discovery)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (c *Controller) initCluster() error {

	groups, err := c.DataStorage.GetGroups()
	if err != nil {
		return fmt.Errorf("init cluster groups error:%s", err.Error())
	}
	logger.INFO("[#ctrl#] init cluster groups:%d", len(groups))
	for _, group := range groups {
		c.Cluster.SetGroup(group.ID, group.Servers)
	}
	return nil
}

func (c *Controller) startCluster() error {

	logger.INFO("[#ctrl#] start cluster.")
	return c.Cluster.Start()
}

func (c *Controller) stopCluster() {

	c.Cluster.Stop()
	logger.INFO("[#ctrl#] stop cluster.")
}

func (c *Controller) SetCluster(cluster *cluster.Cluster) {

	if cluster != nil {
		logger.INFO("[#ctrl#] set cluster %p.", cluster)
		c.Cluster = cluster
	}
}

func (c *Controller) GetClusterGroups() []*models.Group {

	groups := []*models.Group{}
	cgroups := c.Cluster.GetGroups()
	for _, group := range cgroups {
		it := models.NewGroup(group.ID)
		for server := range group.Servers {
			it.Insert(server)
		}
		groups = append(groups, it)
	}
	return groups
}

func (c *Controller) GetClusterGroup(groupid string) *models.Group {

	group := c.Cluster.GetGroup(groupid)
	if group != nil {
		it := models.NewGroup(group.ID)
		for server := range group.Servers {
			it.Insert(server)
		}
		return it
	}
	return nil
}

func (c *Controller) SetClusterGroupEvent(groupid string, event string) {

	switch event {
	case request.GROUP_CREATE_EVENT, request.GROUP_CHANGE_EVENT:
		{
			group, err := c.DataStorage.GetGroup(groupid)
			if err != nil {
				logger.ERROR("[#ctrl#] set cluster groupevent %s %s, error:%s", groupid, event, err.Error())
				return
			}
			c.Cluster.SetGroup(group.ID, group.Servers)
		}
	case request.GROUP_REMOVE_EVENT:
		{
			c.Cluster.RemoveGroup(groupid)
		}
	}
}
