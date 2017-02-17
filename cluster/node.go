package cluster

import "github.com/humpback/discovery"
import "github.com/humpback/discovery/backends"
import "github.com/humpback/gounits/json"
import "github.com/humpback/gounits/rand"
import "github.com/humpback/humpback-center/cluster/types"

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

// NodeArgs is exported
type NodeArgs struct {
	APIAddr          string
	ClusterName      string
	ClusterURIs      string
	ClusterHeartBeat time.Duration
	ClusterTTL       time.Duration
}

func NewNodeArgs(apiPort int, clusterName, clusterURIs, clusterHeartBeat, clusterTTL string) (*NodeArgs, error) {

	portStr := strconv.Itoa(apiPort)
	apiAddr := net.JoinHostPort("", portStr)

	heartbeat, err := time.ParseDuration(clusterHeartBeat)
	if err != nil {
		return nil, err
	}

	ttl, err := time.ParseDuration(clusterTTL)
	if err != nil {
		return nil, err
	}

	return &NodeArgs{
		APIAddr:          apiAddr,
		ClusterName:      clusterName,
		ClusterURIs:      clusterURIs,
		ClusterHeartBeat: heartbeat,
		ClusterTTL:       ttl,
	}, nil
}

// Node is exported
type Node struct {
	Key       string
	Cluster   string
	discovery *discovery.Discovery
	stopCh    chan struct{}
	quitCh    chan struct{}
}

// cluster node
var node *Node

// createNode is exported, create cluster discovery node
func createNode(args *NodeArgs) (*Node, error) {

	key, err := rand.UUIDFile("./humpback-agent.key")
	if err != nil {
		return nil, err
	}

	clusterName := args.ClusterName
	configOpts := map[string]string{"kv.path": clusterName}
	d, err := discovery.New(args.ClusterURIs, args.ClusterHeartBeat, args.ClusterTTL, configOpts)
	if err != nil {
		return nil, err
	}

	return &Node{
		Key:       key,
		Cluster:   clusterName,
		discovery: d,
		stopCh:    make(chan struct{}),
		quitCh:    make(chan struct{}),
	}, nil
}

// Register is exported, register to cluster discovery
func Register(args *NodeArgs) error {

	if node != nil {
		return fmt.Errorf("cluster node is register already.")
	}

	var err error
	if node, err = createNode(args); err != nil {
		return err
	}

	registOpts, err := types.NewClusterRegistOptions(args.APIAddr)
	if err != nil {
		return err
	}

	buf, err := json.EnCodeObjectToBuffer(registOpts)
	if err != nil {
		return err
	}

	log.Printf("register to cluster - %s %s [addr:%s]\n", node.Cluster, node.Key, registOpts.Addr)
	node.discovery.Register(node.Key, buf, node.stopCh, func(key string, err error) {
		log.Printf("discovery register %s error:%s\n", key, err.Error())
		if err == backends.ErrRegistLoopQuit {
			close(node.quitCh)
		}
	})
	return nil
}

// Close is exported, register close
func Close() {

	if node != nil {
		close(node.stopCh)
		<-node.quitCh
		node = nil
		log.Printf("register closed.\n")
	}
}
