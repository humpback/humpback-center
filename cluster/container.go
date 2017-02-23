package cluster

import "github.com/docker/engine-api/types"
import "github.com/humpback/humpback-agent/models"

// ContainerConfig is exported
type ContainerConfig struct {
	models.Container
}

// Container is exported
type Container struct {
	types.Container
	Config *ContainerConfig
	Info   types.ContainerJSON
	Engine *Engine
}
