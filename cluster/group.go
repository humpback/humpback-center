package cluster

/*
func (cluster *Cluster) CreateGroup(groupid string, servers []string) bool {

	cluster.Lock()
	defer cluster.Unlock()
	if _, ret := cluster.groups[groupid]; !ret {
		group := &Group{
			ID: groupid,
		}
		for _, server := range servers {
			//在engines中检查ip是否存在，若存在，则取到engine->key，否则为空，engine不存在.
			group.Servers[server] = ""
		}
		cluster.groups[groupid] = group
		return true
	}
	return false
}

func (cluster *Cluster) SetGroupServer(server string, engineid string) bool {

	return false
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
*/
