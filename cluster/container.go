package cluster

import "github.com/docker/docker/api/types"
import units "github.com/docker/go-units"
import "github.com/humpback/humpback-agent/models"

import (
	"fmt"
	"time"
)

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
