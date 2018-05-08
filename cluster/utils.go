package cluster

import (
	"sort"
	"strings"
)

func getImageTag(imageName string) string {

	imageTag := "latest"
	values := strings.SplitN(imageName, ":", 2)
	if len(values) == 2 {
		imageTag = values[1]
	}
	return imageTag
}

// searchServerOfEngines is exported
func searchServerOfEngines(server Server, engines map[string]*Engine) *Engine {

	//priority ip
	if server.IP != "" {
		if engine, ret := engines[server.IP]; ret {
			return engine
		}
	} else if server.Name != "" {
		for _, engine := range engines {
			if server.Name == engine.Name {
				return engine
			}
		}
	}
	return nil
}

// selectIPOrName is exported
func selectIPOrName(ip string, name string) string {

	if ip != "" {
		return ip
	}
	return name
}

// compareAddServers is exported
func compareAddServers(nodeCache *NodeCache, originServer Server, newServer Server) bool {

	nodeData1 := nodeCache.Get(selectIPOrName(originServer.IP, originServer.Name))
	nodeData2 := nodeCache.Get(selectIPOrName(newServer.IP, newServer.Name))
	if nodeData1 != nil && nodeData2 != nil {
		if nodeData1 == nodeData2 {
			return true
		}
	}
	if nodeData2 == nil {
		return true
	}
	return false
}

// compareRemoveServers is exported
func compareRemoveServers(nodeCache *NodeCache, originServer Server, newServer Server) bool {

	nodeData1 := nodeCache.Get(selectIPOrName(originServer.IP, originServer.Name))
	nodeData2 := nodeCache.Get(selectIPOrName(newServer.IP, newServer.Name))
	if nodeData1 == nil && nodeData2 == nil {
		return true
	}
	if nodeData1 == nil {
		return true
	}
	if nodeData1 == nodeData2 {
		return true
	}
	return false
}

type rdEngines []*Engine

func (engines rdEngines) Len() int {

	return len(engines)
}

func (engines rdEngines) Swap(i, j int) {

	engines[i], engines[j] = engines[j], engines[i]
}

func (engines rdEngines) Less(i, j int) bool {

	return engines[i].IP < engines[j].IP
}

// removeDuplicatesEngines is exported
func removeDuplicatesEngines(engines []*Engine) []*Engine {

	out := []*Engine{}
	pEngines := rdEngines(engines)
	sort.Sort(pEngines)
	nLen := len(pEngines)
	for i := 0; i < nLen; i++ {
		if i > 0 && pEngines[i-1].IP == pEngines[i].IP {
			continue
		}
		out = append(out, pEngines[i])
	}
	return out
}

type rdGroups []*Group

func (groups rdGroups) Len() int {

	return len(groups)
}

func (groups rdGroups) Swap(i, j int) {

	groups[i], groups[j] = groups[j], groups[i]
}

func (groups rdGroups) Less(i, j int) bool {

	return groups[i].ID < groups[j].ID
}

func removeDuplicatesGroups(groups []*Group) []*Group {

	out := []*Group{}
	pGroups := rdGroups(groups)
	sort.Sort(pGroups)
	nLen := len(pGroups)
	for i := 0; i < nLen; i++ {
		if i > 0 && pGroups[i-1].ID == pGroups[i].ID {
			continue
		}
		out = append(out, pGroups[i])
	}
	return out
}
