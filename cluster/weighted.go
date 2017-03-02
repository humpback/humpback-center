package cluster

import "github.com/humpback/humpback-agent/models"
import "github.com/humpback/gounits/logger"

type WeightedEngine struct {
	Engine *Engine
	Weight int64
}

type weightedEngineList []*WeightedEngine

func (weight weightedEngineList) Len() int {

	return len(weight)
}

func (weight weightedEngineList) Swap(i, j int) {

	weight[i], weight[j] = weight[j], weight[i]
}

func (weight weightedEngineList) Less(i, j int) bool {

	var iengine = weight[i]
	var jengine = weight[j]
	if iengine.Weight == jengine.Weight {
		return len(iengine.Engine.Containers()) < len(jengine.Engine.Containers())
	}
	return iengine.Weight < jengine.Weight
}

func (weight weightedEngineList) Engines() []*Engine {

	engines := []*Engine{}
	for _, it := range weight {
		engines = append(engines, it.Engine)
	}
	return engines
}

func weightEngines(engines []*Engine, config models.Container) weightedEngineList {

	weightedEngines := weightedEngineList{}
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

		logger.INFO("[#cluster#] weighted engine %s cpuScore:%d memorySocre:%d weight:%d", engine.IP, cpuScore, memoryScore, cpuScore+memoryScore)
		if cpuScore <= 100 && memoryScore <= 100 {
			weightedEngines = append(weightedEngines, &WeightedEngine{
				Engine: engine,
				Weight: cpuScore + memoryScore,
			})
		}
	}
	return weightedEngines
}
