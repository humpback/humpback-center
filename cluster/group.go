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
		group.Set(servers)
		return true
	}
	return false
}
