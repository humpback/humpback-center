package server

import "github.com/humpback/discovery"
import "github.com/humpback/humpback-center/etc"

import (
	"fmt"
	"time"
)

func createDiscovery(c *etc.Configuration) (*discovery.Discovery, error) {

	heartbeat, err := time.ParseDuration(c.Discovery.Heartbeat)
	if err != nil {
		return nil, fmt.Errorf("discovery heartbeat invalid.")
	}

	configopts := map[string]string{"kv.path": c.Discovery.SysPath}
	d, err := discovery.New(c.Discovery.URIs, heartbeat, 0, configopts)
	if err != nil {
		return nil, err
	}
	return d, nil
}
