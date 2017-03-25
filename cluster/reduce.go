package cluster

// ReduceEngine is exported
type ReduceEngine struct {
	metaid    string
	engine    *Engine
	container *Container
}

// Containers is exported
// Return engine's containers of metaid
func (reduce *ReduceEngine) Containers() Containers {

	if reduce.engine != nil {
		return reduce.engine.Containers(reduce.metaid)
	}
	return Containers{}
}

// ReduceContainer is exported
func (reduce *ReduceEngine) ReduceContainer() *Container {

	return reduce.container
}

// Engine is exported
func (reduce *ReduceEngine) Engine() *Engine {

	return reduce.engine
}

type reduceEngines []*ReduceEngine

func (engines reduceEngines) Len() int {

	return len(engines)
}

func (engines reduceEngines) Swap(i, j int) {

	engines[i], engines[j] = engines[j], engines[i]
}

func (engines reduceEngines) Less(i, j int) bool {

	return len(engines[i].Containers()) > len(engines[j].Containers())
}

func selectReduceEngines(metaid string, engines []*Engine) reduceEngines {

	out := reduceEngines{}
	for _, engine := range engines {
		if engine.IsHealthy() {
			containers := engine.Containers(metaid)
			if len(containers) > 0 {
				out = append(out, &ReduceEngine{
					engine:    engine,
					metaid:    metaid,
					container: containers[0],
				})
			}
		}
	}
	return out
}
