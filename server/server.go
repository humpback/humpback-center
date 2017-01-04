package server

import (
	"github.com/humpback/discovery"
	"github.com/humpback/humpback-center/cluster"
)

type Server struct {
	discovery *discovery.Discovery
	cluster   *cluster.Cluster
}

func NewServer() (*Server, error) {

	return nil, nil
}

func (s *Server) Startup() error {

	return nil
}

func (s *Server) Stop() error {

	return nil
}
