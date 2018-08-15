package types

import "github.com/docker/docker/client"
import "github.com/humpback/discovery"
import "github.com/humpback/discovery/backends"
import "github.com/humpback/gounits/json"
import "github.com/humpback/gounits/network"
import "github.com/humpback/gounits/rand"
import "github.com/humpback/common/models"

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//NodeRegisterOptions is exported
type NodeRegisterOptions struct {
	APIPort    int
	NodeConfig *models.Config
}

// NewNodeRegisterOptions is exported
func NewNodeRegisterOptions(apiPort int, nodeConfig *models.Config) *NodeRegisterOptions {

	return &NodeRegisterOptions{
		APIPort:    apiPort,
		NodeConfig: nodeConfig,
	}
}

//NodeClusterOptions is exported
type NodeClusterOptions struct {
	ClusterName      string
	ClusterURIs      string
	ClusterHeartBeat time.Duration
	ClusterTTL       time.Duration
}

//NodeData is exported
type NodeData struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	IP              string   `json:"ip"`
	APIAddr         string   `json:"apiaddr"`
	Cpus            int64    `json:"cpus"`
	Memory          int64    `json:"memory"`
	StorageDirver   string   `json:"storagedirver"`
	KernelVersion   string   `json:"kernelversion"`
	Architecture    string   `json:"architecture"`
	OperatingSystem string   `json:"operatingsystem"`
	OSType          string   `json:"ostype"`
	EngineLabels    []string `json:"lables"`
	AppVersion      string   `json:"appversion"`
	DockerVersion   string   `json:"dockerversion"`
}

// MapEngineLabels is exported
// covert nodedata's engine labels []string to map
func (nodeData *NodeData) MapEngineLabels() map[string]string {

	labels := map[string]string{}
	for _, label := range nodeData.EngineLabels {
		kv := strings.SplitN(label, "=", 2)
		if len(kv) == 2 {
			labels[kv[0]] = kv[1]
		}
	}
	return labels
}

// NodeOptions is exported
type NodeOptions struct {
	NodeData
	NodeClusterOptions
}

var node *Node

// Node is exported
type Node struct {
	Key       string
	Cluster   string
	discovery *discovery.Discovery
	data      *NodeData
	stopCh    chan struct{}
	quitCh    chan struct{}
}

// GetNodeData is exported
func GetNodeData() *NodeData {

	if node != nil {
		return node.data
	}
	return nil
}

// NodeRegister is exported
// register to cluster discovery
func NodeRegister(options *NodeRegisterOptions) error {

	nodeOptions, err := createNodeOptions(options)
	if err != nil {
		return err
	}

	if _, err := createNode(nodeOptions); err != nil {
		return err
	}

	buf, err := json.EnCodeObjectToBuffer(&nodeOptions.NodeData)
	if err != nil {
		return err
	}

	log.Printf("register to cluster - %s %s [addr:%s]\n", node.Cluster, node.Key, nodeOptions.NodeData.APIAddr)
	node.discovery.Register(node.Key, buf, node.stopCh, func(key string, err error) {
		log.Printf("discovery register %s error:%s\n", key, err.Error())
		if err == backends.ErrRegistLoopQuit {
			close(node.quitCh)
		}
	})
	return nil
}

// NodeClose is exported
// register close
func NodeClose() {

	if node != nil {
		close(node.stopCh) //close register loop
		<-node.quitCh
		node = nil
		log.Printf("register closed.\n")
	}
}

// createNode is exported
// create cluster discovery node
func createNode(nodeOptions *NodeOptions) (*Node, error) {

	if node == nil {
		key, err := rand.UUIDFile("./humpback-agent.key")
		if err != nil {
			return nil, err
		}
		clusterName := nodeOptions.ClusterName
		configOpts := map[string]string{"kv.path": clusterName}
		d, err := discovery.New(nodeOptions.ClusterURIs, nodeOptions.ClusterHeartBeat, nodeOptions.ClusterTTL, configOpts)
		if err != nil {
			return nil, err
		}
		node = &Node{
			Key:       key,
			Cluster:   clusterName,
			discovery: d,
			data:      &nodeOptions.NodeData,
			stopCh:    make(chan struct{}),
			quitCh:    make(chan struct{}),
		}
	}
	return node, nil
}

// createNodeOptions is exported
// create cluster node options
func createNodeOptions(options *NodeRegisterOptions) (*NodeOptions, error) {

	heartbeat, err := time.ParseDuration(options.NodeConfig.DockerClusterHeartBeat)
	if err != nil {
		return nil, err
	}

	ttl, err := time.ParseDuration(options.NodeConfig.DockerClusterTTL)
	if err != nil {
		return nil, err
	}

	var hostIP string
	if options.NodeConfig.DockerAgentIPAddr == "0.0.0.0" {
		hostIP = network.GetDefaultIP()
	} else {
		hostIP = options.NodeConfig.DockerAgentIPAddr
	}

	if _, err := net.ResolveIPAddr("ip4", hostIP); err != nil {
		return nil, err
	}

	apiAddr := hostIP + net.JoinHostPort("", strconv.Itoa(options.APIPort))
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	dockerClient, err := client.NewClient(options.NodeConfig.DockerEndPoint, options.NodeConfig.DockerAPIVersion, nil, defaultHeaders)
	if err != nil {
		return nil, err
	}

	engineInfo, err := dockerClient.Info(context.Background())
	if err != nil {
		return nil, err
	}

	hostName, err := os.Hostname()
	if err != nil {
		hostName = engineInfo.Name
	}

	return &NodeOptions{
		NodeData: NodeData{
			ID:              engineInfo.ID,
			Name:            hostName,
			IP:              hostIP,
			APIAddr:         apiAddr,
			Cpus:            (int64)(engineInfo.NCPU),
			Memory:          engineInfo.MemTotal,
			StorageDirver:   engineInfo.Driver,
			KernelVersion:   engineInfo.KernelVersion,
			Architecture:    engineInfo.Architecture,
			OperatingSystem: engineInfo.OperatingSystem,
			OSType:          engineInfo.OSType,
			EngineLabels:    engineInfo.Labels,
			AppVersion:      options.NodeConfig.AppVersion,
			DockerVersion:   engineInfo.ServerVersion,
		},
		NodeClusterOptions: NodeClusterOptions{
			ClusterName:      options.NodeConfig.DockerClusterName,
			ClusterURIs:      options.NodeConfig.DockerClusterURIs,
			ClusterHeartBeat: heartbeat,
			ClusterTTL:       ttl,
		},
	}, nil
}

// NodeCache is exported
type NodeCache struct {
	sync.RWMutex
	nodes map[string]*NodeData
}

// NewNodeCache is exported
func NewNodeCache() *NodeCache {

	return &NodeCache{
		nodes: make(map[string]*NodeData),
	}
}

// Add is exported
// nodeCache add online nodeData.
func (cache *NodeCache) Add(key string, nodeData *NodeData) {

	cache.Lock()
	if _, ret := cache.nodes[key]; !ret {
		cache.nodes[key] = nodeData
	}
	cache.Unlock()
}

// Remove is exported
// nodeCache remove offline nodeData.
func (cache *NodeCache) Remove(key string) {

	cache.Lock()
	delete(cache.nodes, key)
	cache.Unlock()
}

// Node is exported
// nodeCache get nodedata of key
func (cache *NodeCache) Node(key string) *NodeData {

	cache.RLock()
	defer cache.RUnlock()
	if nodeData, ret := cache.nodes[key]; ret {
		return nodeData
	}
	return nil
}

// Get is exported
// nodeCache get nodeData of server ip or server hostname
func (cache *NodeCache) Get(IPOrName string) *NodeData {

	cache.RLock()
	defer cache.RUnlock()
	for _, nodeData := range cache.nodes {
		if nodeData.IP == IPOrName {
			return nodeData
		}
	}

	for _, nodeData := range cache.nodes {
		if nodeData.Name == IPOrName {
			return nodeData
		}
	}
	return nil
}
