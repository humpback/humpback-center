package server

import "github.com/humpback/discovery"
import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/etc"

import (
	"flag"
)

/*
ServerCenter
humpback-center service
*/
type ServerCenter struct {
	discovery *discovery.Discovery
	cluster   *cluster.Cluster
}

func NewServerCenter() (*ServerCenter, error) {

	var conf string
	flag.StringVar(&conf, "f", "etc/config.yaml", "humpback-center configuration file.")
	flag.Parse()

	//initialize configuration
	c, err := etc.NewConfiguration(conf)
	if err != nil {
		return nil, err
	}

	//make discovery service
	discovery, err := createDiscovery(c)
	if err != nil {
		return nil, err
	}

	//make cluster
	cluster, err := cluster.NewCluster(discovery)
	if err != nil {
		return nil, err
	}

	return &ServerCenter{
		discovery: discovery,
		cluster:   cluster,
	}, nil
}

func (sc *ServerCenter) Startup() error {

	return nil
}

func (sc *ServerCenter) Stop() error {

	return nil
}
