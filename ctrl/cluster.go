package ctrl

import "common/models"
import "humpback-center/api/request"
import "humpback-center/cluster"
import "humpback-center/cluster/types"
import "humpback-center/etc"
import "humpback-center/notify"
import "github.com/humpback/discovery"
import "github.com/humpback/gounits/logger"

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
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

	configopts := map[string]string{"kv.path": strings.TrimSpace(clusterOpts.Discovery.Cluster)}
	discovery, err := discovery.New(clusterOpts.Discovery.URIs, heartbeat, 0, configopts)
	if err != nil {
		return nil, err
	}

	siteURL := strings.SplitN(configuration.SiteAPI, "/api", 2)
	notifySender := notify.NewNotifySender(siteURL[0], configuration.GetNotificationsEndPoints())
	cluster, err := cluster.NewCluster(clusterOpts.DriverOpts, notifySender, discovery)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func (c *Controller) initCluster() {

	if groups := c.getClusterGroupStoreData(""); groups != nil {
		logger.INFO("[#ctrl#] init cluster groups:%d", len(groups))
		for _, group := range groups {
			if group.IsCluster {
				if c.Cluster.Location == "" || strings.ToUpper(group.Location) == strings.ToUpper(c.Cluster.Location) {
					c.Cluster.SetGroup(group)
				}
			}
		}
	}
}

func (c *Controller) startCluster() error {

	c.initCluster()
	logger.INFO("[#ctrl#] start cluster.")
	return c.Cluster.Start()
}

func (c *Controller) stopCluster() {

	c.Cluster.Stop()
	logger.INFO("[#ctrl#] stop cluster.")
}

func (c *Controller) getClusterGroupStoreData(groupid string) []*cluster.Group {

	query := map[string][]string{}
	groupid = strings.TrimSpace(groupid)
	if groupid != "" {
		query["groupid"] = []string{groupid}
	}

	t := time.Now().UnixNano() / int64(time.Millisecond)
	value := fmt.Sprintf("HUMPBACK_CENTER%d", t)
	code := base64.StdEncoding.EncodeToString([]byte(value))
	headers := map[string][]string{"x-get-cluster": []string{code}}
	respGroups, err := c.client.Get(context.Background(), c.Configuration.SiteAPI+"/groups/getclusters", query, headers)
	if err != nil {
		logger.ERROR("[#ctrl#] get cluster group storedata error:%s", err.Error())
		return nil
	}

	defer respGroups.Close()
	if respGroups.StatusCode() != http.StatusOK {
		logger.ERROR("[#ctrl#] get cluster group storedata error:%d %s", respGroups.StatusCode(), respGroups.String())
		return nil
	}

	groups := []*cluster.Group{}
	if err := respGroups.JSON(&groups); err != nil {
		logger.ERROR("[#ctrl#] get cluster group storedata error:%s", err.Error())
		return nil
	}
	return groups
}

func (c *Controller) SetCluster(cluster *cluster.Cluster) {

	if cluster != nil {
		logger.INFO("[#ctrl#] set cluster %p.", cluster)
		c.Cluster = cluster
	}
}

func (c *Controller) GetClusterGroupAllContainers(groupid string) *types.GroupContainers {

	return c.Cluster.GetGroupAllContainers(groupid)
}

func (c *Controller) GetClusterGroupContainers(metaid string) *types.GroupContainer {

	return c.Cluster.GetGroupContainers(metaid)
}

func (c *Controller) GetClusterGroupContainersMetaBase(metaid string) *cluster.MetaBase {

	return c.Cluster.GetMetaBase(metaid)
}

func (c *Controller) GetClusterGroupAllEngines(groupid string) []*cluster.Engine {

	return c.Cluster.GetGroupAllEngines(groupid)
}

func (c *Controller) GetClusterEngine(server string) *cluster.Engine {

	return c.Cluster.GetEngine(server)
}

func (c *Controller) SetClusterGroupEvent(groupid string, event string) {

	logger.INFO("[#ctrl#] set cluster groupevent %s.", event)
	switch event {
	case request.GROUP_CREATE_EVENT, request.GROUP_CHANGE_EVENT:
		{
			if groups := c.getClusterGroupStoreData(groupid); groups != nil {
				logger.INFO("[#ctrl#] get cluster groups:%d", len(groups))
				if len(groups) > 0 {
					for _, group := range groups {
						if group.IsCluster {
							if c.Cluster.Location == "" || strings.ToUpper(group.Location) == strings.ToUpper(c.Cluster.Location) {
								c.Cluster.SetGroup(group)
							} else { // group location changed
								if c.Cluster.GetGroup(groupid) != nil {
									c.Cluster.RemoveGroup(groupid)
								}
							}
						}
					}
				} else { // group iscluster change to false
					if c.Cluster.GetGroup(groupid) != nil {
						c.Cluster.RemoveGroup(groupid)
					}
				}
			}
		}
	case request.GROUP_REMOVE_EVENT:
		{ // group removed
			if c.Cluster.GetGroup(groupid) != nil {
				c.Cluster.RemoveGroup(groupid)
			}
		}
	}
}

func (c *Controller) CreateClusterContainers(groupid string, instances int, webhooks types.WebHooks, config models.Container) (string, *types.CreatedContainers, error) {

	return c.Cluster.CreateContainers(groupid, instances, webhooks, config)
}

func (c *Controller) UpdateClusterContainers(metaid string, instances int, webhooks types.WebHooks) (*types.CreatedContainers, error) {

	return c.Cluster.UpdateContainers(metaid, instances, webhooks)
}

func (c *Controller) OperateContainers(metaid string, action string) (*types.OperatedContainers, error) {

	return c.Cluster.OperateContainers(metaid, "", action)
}

func (c *Controller) OperateContainer(containerid string, action string) (string, *types.OperatedContainers, error) {

	return c.Cluster.OperateContainer(containerid, action)
}

func (c *Controller) UpgradeContainers(metaid string, imagetag string) (*types.UpgradeContainers, error) {

	return c.Cluster.UpgradeContainers(metaid, imagetag)
}

func (c *Controller) RemoveContainers(metaid string) (*types.RemovedContainers, error) {

	return c.Cluster.RemoveContainers(metaid, "")
}

func (c *Controller) RemoveContainer(containerid string) (string, *types.RemovedContainers, error) {

	return c.Cluster.RemoveContainer(containerid)
}
