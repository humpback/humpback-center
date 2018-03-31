package cluster

import "github.com/humpback/gounits/logger"
import "github.com/humpback/common/models"

// WeightedEngine is exported
type WeightedEngine struct {
	engine *Engine
	weight int64
}

// Containers is exported
// Return engine's containers
func (weighted *WeightedEngine) Containers() Containers {

	if weighted.engine != nil {
		return weighted.engine.Containers("")
	}
	return Containers{}
}

// Engine is exported
func (weighted *WeightedEngine) Engine() *Engine {

	return weighted.engine
}

// Weight is exported
func (weighted *WeightedEngine) Weight() int64 {

	return weighted.weight
}

type weightedEngines []*WeightedEngine

func (engines weightedEngines) Len() int {

	return len(engines)
}

func (engines weightedEngines) Swap(i, j int) {

	engines[i], engines[j] = engines[j], engines[i]
}

func (engines weightedEngines) Less(i, j int) bool {

	if engines[i].Weight() == engines[j].Weight() {
		return len(engines[i].Containers()) < len(engines[j].Containers())
	}
	return engines[i].Weight() < engines[j].Weight()
}

func (engines weightedEngines) Engines() []*Engine {

	out := []*Engine{}
	for _, weightedEngine := range engines {
		out = append(out, weightedEngine.Engine())
	}
	return out
}

func selectWeightdEngines(engines []*Engine, config models.Container) weightedEngines {

	out := weightedEngines{}
	for _, engine := range engines {
		totalCpus := engine.TotalCpus()
		totalMemory := engine.TotalMemory()
		if totalMemory < config.Memory || totalCpus < config.CPUShares {
			logger.INFO("[#cluster#] weighted engine %s filter.", engine.IP)
			continue
		}

		var cpuScore int64 = 100
		var memoryScore int64 = 100

		if config.CPUShares > 0 {
			cpuScore = (engine.UsedCpus() + config.CPUShares) * 100 / totalCpus
		}

		if config.Memory > 0 {
			memoryScore = (engine.UsedMemory()/1024/1024 + config.Memory) * 100 / totalMemory
		}

		//logger.INFO("[#cluster#] weighted engine %s cpuScore:%d memorySocre:%d weight:%d", engine.IP, cpuScore, memoryScore, cpuScore+memoryScore)
		if cpuScore <= 100 && memoryScore <= 100 {
			out = append(out, &WeightedEngine{
				engine: engine,
				weight: cpuScore + memoryScore,
			})
		}
	}
	return out
}
