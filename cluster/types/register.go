package types

import "github.com/humpback/gounits/rand"
import "github.com/humpback/gounits/network"

import (
	"os"
	"strings"
)

// ClusterRegistOptions is exported
// regist to cluster options
// ID: cluster node id.
// Name: local hostname
// IP: humpback website regist server ipaddr.
// Addr: humpback node API addr.
// Labels: humpback node custom values. example labels {"node=wh7", "kernelversion=4.4.0", "os=centos6.8"}
// Version: docker or humpback node version
type ClusterRegistOptions struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	IP      string   `json:"ip"`
	Addr    string   `json:"addr"`
	Labels  []string `json:"labels"`
	Version string   `json:"version"`
}

// NewClusterRegistOptions is exported
func NewClusterRegistOptions(id string, name string, ip string, addr string, labels []string, version string) *ClusterRegistOptions {

	if strings.TrimSpace(id) == "" {
		id = rand.UUID(true)
	}

	if strings.TrimSpace(name) == "" {
		hostname, err := os.Hostname()
		if err == nil {
			name = hostname
		}
	}

	if strings.TrimSpace(ip) == "" {
		ip = network.GetDefaultIP()
	}

	return &ClusterRegistOptions{
		ID:      id,
		Name:    name,
		IP:      ip,
		Addr:    addr,
		Labels:  labels,
		Version: version,
	}
}
