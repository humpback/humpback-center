package cluster

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/humpback/discovery/backends"
	"github.com/humpback/gounits/json"
	"github.com/humpback/humpback-center/cluster/storage/entry"
	"github.com/humpback/humpback-center/cluster/storage/node"
	"github.com/humpback/humpback-center/cluster/types"
)

//ParseServer is exported
func ParseServer(ipOrName string) Server {

	server := Server{}
	ip := net.ParseIP(ipOrName)
	if ip != nil {
		server.IP = ipOrName
	} else {
		server.Name = ipOrName
	}
	return server
}

func getImageTag(imageName string) string {

	imageTag := "latest"
	tempPaths := strings.Split(imageName, "/")
	values := strings.SplitN(tempPaths[len(tempPaths)-1], ":", 2)
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

func searchServerOfStorage(server Server, nodeStorage *node.NodeStorage) *Engine {

	var node *entry.Node
	if server.IP != "" {
		node, _ = nodeStorage.NodeByIP(server.IP)
	} else if server.Name != "" {
		node, _ = nodeStorage.NodeByName(server.Name)
	}

	if node != nil {
		engine := &Engine{}
		engine.Update(node.NodeData)
		engine.NodeLabels = node.NodeLabels
		engine.AvailabilityText = node.Availability
		return engine
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
func compareAddServers(nodeCache *types.NodeCache, originServer Server, newServer Server) bool {

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
func compareRemoveServers(nodeCache *types.NodeCache, originServer Server, newServer Server) bool {

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

func deCodeEntry(entry *backends.Entry) (*types.NodeData, error) {

	if entry == nil {
		return nil, fmt.Errorf("decode entry invalid")
	}

	nodeData := &types.NodeData{}
	err := json.DeCodeBufferToObject(entry.Data, nodeData)
	if err != nil {
		return nil, err
	}

	nodeData.Name = strings.ToUpper(nodeData.Name)
	return nodeData, nil
}
