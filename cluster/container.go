package cluster

import "github.com/docker/docker/api/types"
import units "github.com/docker/go-units"
import "github.com/humpback/humpback-agent/models"
import "github.com/humpback/gounits/convert"
import "github.com/humpback/gounits/rand"

import (
	"fmt"
	"strings"
	"time"
)

// CreateContainerResponse is exported
type CreateContainerResponse struct {
	ID       string   `json:"Id"`
	Name     string   `json:"Name"`
	Warnings []string `json:"Warnings"`
}

// ContainerConfig is exported
type ContainerConfig struct {
	models.Container
}

// Container is exported
type Container struct {
	GroupID    string
	MetaID     string
	BaseConfig *ContainerBaseConfig
	Config     *ContainerConfig
	Info       types.ContainerJSON
	Engine     *Engine
}

// update is exported
// update container info and config
func (c *Container) update(engine *Engine, containerJSON *types.ContainerJSON) {

	config := &models.Container{}
	config.Parse(containerJSON)
	containerConfig := &ContainerConfig{
		Container: *config,
	}
	c.Config = containerConfig

	containerJSON.HostConfig.CPUShares = containerJSON.HostConfig.CPUShares * engine.Cpus / 1024.0
	startAt, _ := time.Parse(time.RFC3339Nano, containerJSON.State.StartedAt)
	finishedAt, _ := time.Parse(time.RFC3339Nano, containerJSON.State.FinishedAt)
	containerJSON.State.StartedAt = startAt.Add(engine.DeltaDuration).Format(time.RFC3339Nano)
	containerJSON.State.FinishedAt = finishedAt.Add(engine.DeltaDuration).Format(time.RFC3339Nano)
	c.Info = *containerJSON

	baseConfig := readConainerBaseConfig(containerJSON.ID, engine.configCache, containerJSON.Config.Env)
	if baseConfig != nil {
		c.GroupID = baseConfig.MetaData.GroupID
		c.MetaID = baseConfig.MetaData.MetaID
		c.BaseConfig = baseConfig
	}
}

// readConainerBaseConfig is exported
// read container metadata baseConfig
func readConainerBaseConfig(containerid string, configCache *ContainersConfigCache, configEnv []string) *ContainerBaseConfig {

	configEnvMap := convert.ConvertKVStringSliceToMap(configEnv)
	groupID := configEnvMap["HUMPBACK_CLUSTER_GROUPID"]
	metaID := configEnvMap["HUMPBACK_CLUSTER_METAID"]
	if len(groupID) > 0 && len(metaID) > 0 {
		baseConfig := configCache.GetContainerBaseConfig(metaID, containerid)
		if baseConfig != nil && baseConfig.MetaData.GroupID == groupID && baseConfig.MetaData.MetaID == metaID {
			return baseConfig
		}
	}
	return nil
}

// Containers represents a list of containers
type Containers []*Container

// StateString returns a single string to describe state
func StateString(state *types.ContainerState) string {

	startedAt, _ := time.Parse(time.RFC3339Nano, state.StartedAt)
	if state.Running {
		if state.Paused {
			return "Paused"
		}
		if state.Restarting {
			return "Restarting"
		}
		return "Running"
	}

	if state.Dead {
		return "Dead"
	}

	if startedAt.IsZero() {
		return "Created"
	}
	return "Exited"
}

// FullStateString returns readable description of the state
func FullStateString(state *types.ContainerState) string {

	startedAt, _ := time.Parse(time.RFC3339Nano, state.StartedAt)
	finishedAt, _ := time.Parse(time.RFC3339Nano, state.FinishedAt)
	if state.Running {
		if state.Paused {
			return fmt.Sprintf("Up %s (Paused)", units.HumanDuration(time.Now().UTC().Sub(startedAt)))
		}
		if state.Restarting {
			return fmt.Sprintf("Restarting (%d) %s ago", state.ExitCode, units.HumanDuration(time.Now().UTC().Sub(finishedAt)))
		}
		healthText := ""
		if h := state.Health; h != nil {
			switch h.Status {
			case types.Starting:
				healthText = "Health: Starting"
			default:
				healthText = h.Status
			}
		}
		if len(healthText) > 0 {
			return fmt.Sprintf("Up %s (%s)", units.HumanDuration(time.Now().UTC().Sub(startedAt)), healthText)
		}
		return fmt.Sprintf("Up %s", units.HumanDuration(time.Now().UTC().Sub(startedAt)))
	}

	if state.Dead {
		return "Dead"
	}

	if startedAt.IsZero() {
		return "Created"
	}

	if finishedAt.IsZero() {
		return ""
	}

	return fmt.Sprintf("Exited (%d) %s ago", state.ExitCode, units.HumanDuration(time.Now().UTC().Sub(finishedAt)))
}

// Get returns a container using its ID or Name
func (containers Containers) Get(IDOrName string) *Container {

	if len(strings.TrimSpace(IDOrName)) == 0 {
		return nil
	}

	for _, container := range containers {
		if container.Info.ID == IDOrName || rand.TruncateID(container.Info.ID) == IDOrName {
			return container
		}
	}

	candidates := []*Container{}
	for _, container := range containers {
		name := container.Info.Name
		if name == IDOrName || name == "/"+IDOrName || container.Engine.ID+name == IDOrName || container.Engine.Name+name == IDOrName {
			candidates = append(candidates, container)
		}
	}

	if size := len(candidates); size == 1 {
		return candidates[0]
	} else if size > 1 {
		return nil
	}

	for _, container := range containers {
		if strings.HasPrefix(container.Info.ID, IDOrName) {
			candidates = append(candidates, container)
		}
	}

	if len(candidates) == 1 {
		return candidates[0]
	}
	return nil
}
