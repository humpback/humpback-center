package types

import "github.com/humpback/gounits/network"

import (
	"net"
)

// ClusterRegistOptions is exported
// regist to cluster options
// APIAddr: humpback node API addr. (IP:Port)
type ClusterRegistOptions struct {
	Addr string `json:"Addr"`
}

// NewClusterRegistOptions is exported
func NewClusterRegistOptions(addr string) (*ClusterRegistOptions, error) {

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	if len(host) == 0 {
		host = network.GetDefaultIP() // use defaule ipaddr
		addr = host + ":" + port
	}

	if _, err := net.ResolveIPAddr("ip4", host); err != nil {
		return nil, err
	}

	return &ClusterRegistOptions{
		Addr: addr,
	}, nil
}
