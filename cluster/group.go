package cluster

import "github.com/humpback/humpback-center/models"

func (cluster *Cluster) GetGroups() []*models.Group {

	cluster.RLock()
	groups := []*models.Group{}
	for _, group := range cluster.groups {
		groups = append(groups, group)
	}
	cluster.RUnlock()
	return groups
}

func (cluster *Cluster) GetGroup(groupid string) *models.Group {

	cluster.RLock()
	defer cluster.RUnlock()
	if group, ret := cluster.groups[groupid]; ret {
		return group
	}
	return nil
}

func (cluster *Cluster) CreateGroup(groupid string, servers []string) bool {

	cluster.Lock()
	defer cluster.Unlock()
	if _, ret := cluster.groups[groupid]; !ret {
		cluster.groups[groupid] = &models.Group{
			ID:      groupid,
			Servers: servers,
		}
		return true
	}
	return false
}

func (cluster *Cluster) RemoveGroup(groupid string) bool {

	cluster.Lock()
	defer cluster.Unlock()
	if _, ret := cluster.groups[groupid]; ret {
		delete(cluster.groups, groupid)
		return true
	}
	return false
}

func (cluster *Cluster) ClearGroups() {

	cluster.Lock()
	for groupid, group := range cluster.groups {
		group.Servers = group.Servers[0:0]
		delete(cluster.groups, groupid)
	}
	cluster.Unlock()
}

func (cluster *Cluster) InsertGroupServer(groupid string, server string) bool {

	cluster.Lock()
	defer cluster.Unlock()
	if group, ret := cluster.groups[groupid]; ret {
		return group.Insert(server)
	}
	return false
}

func (cluster *Cluster) RemoveGroupServer(groupid string, server string) bool {

	cluster.Lock()
	defer cluster.Unlock()
	if group, ret := cluster.groups[groupid]; ret {
		return group.Remove(server)
	}
	return false
}

func (cluster *Cluster) SetGroupServers(groupid string, servers []string) bool {

	cluster.Lock()
	defer cluster.Unlock()
	if group, ret := cluster.groups[groupid]; ret {
		group.Servers = servers
		return true
	}
	return false
}

func (cluster *Cluster) ClearGroupServers(groupid string) {

	cluster.Lock()
	if group, ret := cluster.groups[groupid]; ret {
		group.Servers = group.Servers[0:0]
	}
	cluster.Unlock()
}
