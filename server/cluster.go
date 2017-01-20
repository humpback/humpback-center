package server

import "github.com/humpback/discovery"
import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/etc"

import (
	"fmt"
	"time"
)

func createCluster(c *etc.Configuration) (*cluster.Cluster, error) {

	heartbeat, err := time.ParseDuration(c.Discovery.Heartbeat)
	if err != nil {
		return nil, fmt.Errorf("discovery heartbeat invalid.")
	}

	if heartbeat < 1*time.Second {
		return nil, fmt.Errorf("discovery heartbeat should be at least 1s.")
	}

	configopts := map[string]string{"kv.path": c.Discovery.SysPath}
	discovery, err := discovery.New(c.Discovery.URIs, heartbeat, 0, configopts)
	if err != nil {
		return nil, err
	}

	cluster, err := cluster.NewCluster(discovery)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}
