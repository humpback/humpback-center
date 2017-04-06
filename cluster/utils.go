package cluster

import (
	"sort"
)

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

	ret := []*Engine{}
	pEngines := rdEngines(engines)
	sort.Sort(pEngines)
	nLen := len(pEngines)
	for i := 0; i < nLen; i++ {
		if i > 0 && pEngines[i-1].IP == pEngines[i].IP {
			continue
		}
		ret = append(ret, pEngines[i])
	}
	return ret
}
