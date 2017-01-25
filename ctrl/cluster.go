package ctrl

import "github.com/humpback/discovery"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/etc"
import "github.com/humpback/humpback-center/models"

import (
	"fmt"
	"time"
)

func CreateCluster(configuration *etc.Configuration) (*cluster.Cluster, error) {

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

func (c *Controller) SetCluster(cluster *cluster.Cluster) {

	if cluster != nil {
		c.Cluster = cluster
	}
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
