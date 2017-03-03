package cluster

import "github.com/docker/docker/api/types"
import units "github.com/docker/go-units"
import "github.com/humpback/humpback-agent/models"
import "github.com/humpback/gounits/rand"

import (
	"fmt"
	"strings"
	"time"
)

// Container name prefix
const ContaierPrefix = "HumpbackC-"

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
	types.Container
	BaseConfig *ContainerBaseConfig
	Config     *ContainerConfig
	Info       types.ContainerJSON
	Engine     *Engine
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
		if container.ID == IDOrName || rand.TruncateID(container.ID) == IDOrName {
			return container
		}
	}

	candidates := []*Container{}
	for _, container := range containers {
		found := false
		for _, name := range container.Names {
			if name == IDOrName || name == "/"+IDOrName || container.Engine.ID+name == IDOrName || container.Engine.Name+name == IDOrName {
				found = true
				break
			}
		}
		if found {
			candidates = append(candidates, container)
		}
	}

	if size := len(candidates); size == 1 {
		return candidates[0]
	} else if size > 1 {
		return nil
	}

	for _, container := range containers {
		if strings.HasPrefix(container.ID, IDOrName) {
			candidates = append(candidates, container)
		}
	}

	if len(candidates) == 1 {
		return candidates[0]
	}
	return nil
}
