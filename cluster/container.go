package cluster

import "github.com/docker/docker/pkg/stringid"
import "github.com/docker/docker/api/types"
import units "github.com/docker/go-units"
import "github.com/humpback/gounits/convert"
import "github.com/humpback/gounits/rand"
import "github.com/humpback/common/models"

import (
	"fmt"
	"strings"
	"time"
)

// ShortContainerID is exported
// return a short containerid string.
func ShortContainerID(containerid string) string {
	return stringid.TruncateID(containerid)
}

// ContainerConfig is exported
type ContainerConfig struct {
	models.Container
}

// Container is exported
type Container struct {
	BaseConfig *ContainerBaseConfig
	Config     *ContainerConfig
	Info       types.ContainerJSON
	Engine     *Engine
}

// GroupID is exported
// Return Container GroupID
func (c *Container) GroupID() string {

	if c.BaseConfig != nil && c.BaseConfig.MetaData != nil {
		return c.BaseConfig.MetaData.GroupID
	}
	return ""
}

// MetaID is exported
// Return Container MetaID
func (c *Container) MetaID() string {

	if c.BaseConfig != nil && c.BaseConfig.MetaData != nil {
		return c.BaseConfig.MetaData.MetaID
	}
	return ""
}

// Index is exported
// Return Container Index
func (c *Container) Index() int {

	if c.BaseConfig != nil {
		return c.BaseConfig.Index
	}
	return -1
}

// OriginalName is exported
// Return Container OriginalName
func (c *Container) OriginalName() string {

	if c.BaseConfig != nil {
		configEnvMap := convert.ConvertKVStringSliceToMap(c.BaseConfig.Env)
		if originalName, ret := configEnvMap["HUMPBACK_CLUSTER_CONTAINER_ORIGINALNAME"]; ret {
			return originalName
		}
	}
	return ""
}

// ValidateConfig is exported
func (c *Container) ValidateConfig() bool {

	configEnvMap := convert.ConvertKVStringSliceToMap(c.Info.Config.Env)
	groupID := configEnvMap["HUMPBACK_CLUSTER_GROUPID"]
	metaID := configEnvMap["HUMPBACK_CLUSTER_METAID"]
	if len(groupID) == 0 && len(metaID) == 0 {
		return true //true, general container that do not called scheduling of cluster.
	}

	if len(groupID) > 0 || len(metaID) > 0 {
		if c.BaseConfig != nil { // valid cluster container
			return true
		}
	}
	return false // invalid cluster container
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
	//startAt, _ := time.Parse(time.RFC3339Nano, containerJSON.State.StartedAt)
	//finishedAt, _ := time.Parse(time.RFC3339Nano, containerJSON.State.FinishedAt)
	//containerJSON.State.StartedAt = startAt.Add(engine.DeltaDuration).Format(time.RFC3339Nano)
	//containerJSON.State.FinishedAt = finishedAt.Add(engine.DeltaDuration).Format(time.RFC3339Nano)
	c.Info = *containerJSON

	configEnvMap := convert.ConvertKVStringSliceToMap(containerJSON.Config.Env)
	groupID := configEnvMap["HUMPBACK_CLUSTER_GROUPID"]
	metaID := configEnvMap["HUMPBACK_CLUSTER_METAID"]
	if len(groupID) > 0 && len(metaID) > 0 {
		c.BaseConfig = engine.configCache.GetContainerBaseConfig(metaID, containerJSON.ID)
	} else {
		c.BaseConfig = nil
	}
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
