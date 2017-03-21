package ctrl

import "github.com/humpback/discovery"
import "github.com/humpback/gounits/convert"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-center/api/request"
import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/cluster/types"
import "github.com/humpback/humpback-center/etc"
import "github.com/humpback/humpback-center/models"
import agentmodels "github.com/humpback/humpback-agent/models"

import (
	"fmt"
	"time"
)

func createCluster(configuration *etc.Configuration) (*cluster.Cluster, error) {

	clusterOpts := configuration.Cluster
	heartbeat, err := time.ParseDuration(clusterOpts.Discovery.Heartbeat)
	if err != nil {
		return nil, fmt.Errorf("discovery heartbeat invalid.")
	}

	if heartbeat < 1*time.Second {
		return nil, fmt.Errorf("discovery heartbeat should be at least 1s.")
	}

	configopts := map[string]string{"kv.path": clusterOpts.Discovery.Cluster}
	discovery, err := discovery.New(clusterOpts.Discovery.URIs, heartbeat, 0, configopts)
	if err != nil {
		return nil, err
	}

	cluster, err := cluster.NewCluster(clusterOpts.DriverOpts, discovery)
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
		c.Cluster.SetGroup(group.ID, group.Servers, group.Owners)
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

func (c *Controller) getEngineState(server string) string {

	state := cluster.GetStateText(cluster.StateDisconnected)
	if engine := c.Cluster.GetEngine(server); engine != nil {
		state = engine.State()
	}
	return state
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
		for _, server := range group.Servers {
			status := c.getEngineState(server)
			it.Insert(server, status)
		}
		groups = append(groups, it)
	}
	return groups
}

func (c *Controller) GetClusterGroup(groupid string) *models.Group {

	if group := c.Cluster.GetGroup(groupid); group != nil {
		it := models.NewGroup(group.ID)
		for _, server := range group.Servers {
			status := c.getEngineState(server)
			it.Insert(server, status)
		}
		return it
	}
	return nil
}

func (c *Controller) GetClusterGroupEngines(groupid string) []*models.Engine {

	engines := c.Cluster.GetGroupEngines(groupid)
	if engines == nil {
		return nil
	}

	result := []*models.Engine{}
	for _, engine := range engines {
		labels := convert.ConvertMapToKVStringSlice(engine.Labels)
		it := models.NewEngine(engine.ID, engine.Name, engine.IP, engine.Addr, labels, engine.State())
		result = append(result, it)
	}
	return result
}

func (c *Controller) GetClusterEngine(server string) *models.Engine {

	if engine := c.Cluster.GetEngine(server); engine != nil {
		labels := convert.ConvertMapToKVStringSlice(engine.Labels)
		return models.NewEngine(engine.ID, engine.Name, engine.IP, engine.Addr, labels, engine.State())
	}
	return nil
}

func (c *Controller) SetClusterGroupEvent(groupid string, event string) {

	logger.INFO("[#ctrl#] set cluster groupevent %s.", event)
	switch event {
	case request.GROUP_CREATE_EVENT, request.GROUP_CHANGE_EVENT:
		{
			group, err := c.DataStorage.GetGroup(groupid)
			if err != nil {
				logger.ERROR("[#ctrl#] set cluster groupevent %s %s, error:%s", groupid, event, err.Error())
				return
			}
			c.Cluster.SetGroup(group.ID, group.Servers, group.Owners)
		}
	case request.GROUP_REMOVE_EVENT:
		{
			c.Cluster.RemoveGroup(groupid)
		}
	}
}

func (c *Controller) CreateClusterContainers(groupid string, instances int, config agentmodels.Container) (string, *types.CreatedContainers, error) {

	return c.Cluster.CreateContainers(groupid, instances, config)
}

func (c *Controller) SetClusterContainers(metaid string, instances int) (*types.CreatedContainers, error) {

	return c.Cluster.SetContainers(metaid, instances)
}

func (c *Controller) OperateContainers(metaid string, action string) (*types.OperatedContainers, error) {

	return c.Cluster.OperateContainers(metaid, "", action)
}

func (c *Controller) OperateContainer(containerid string, action string) (string, *types.OperatedContainers, error) {

	return c.Cluster.OperateContainer(containerid, action)
}

func (c *Controller) UpgradeContainers(metaid string, imagetag string) error {

	return c.Cluster.UpgradeContainers(metaid, imagetag)
}

func (c *Controller) RemoveContainers(metaid string) (*types.RemovedContainers, error) {

	return c.Cluster.RemoveContainers(metaid, "")
}

func (c *Controller) RemoveContainer(containerid string) (string, *types.RemovedContainers, error) {

	return c.Cluster.RemoveContainer(containerid)
}
