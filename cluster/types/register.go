package types

import "github.com/humpback/gounits/network"

import (
	"net"
)

// ClusterRegistOptions is exported
// regist to cluster options
// IP: humpback node host ipaddr.
// APIAddr: humpback node API addr. (ip:port)
// Labels: humpback node custom values.
// example labels {"node=wh7", "kernelversion=4.4.0", "os=centos6.8"}
type ClusterRegistOptions struct {
	IP      string   `json:"ip"`
	APIAddr string   `json:"apiaddr"`
	Labels  []string `json:"labels"`
}

// NewClusterRegistOptions is exported
func NewClusterRegistOptions(apiaddr string, labels []string) (*ClusterRegistOptions, error) {

	ip := network.GetDefaultIP()
	host, port, err := net.SplitHostPort(apiaddr)
	if err != nil {
		return nil, err
	}

	if len(host) == 0 {
		host = ip //set default api ipaddr
		apiaddr = host + ":" + port
	}

	if _, err := net.ResolveIPAddr("ip4", host); err != nil {
		return nil, err
	}

	return &ClusterRegistOptions{
		IP:      ip,
		APIAddr: apiaddr,
		Labels:  labels,
	}, nil
}
