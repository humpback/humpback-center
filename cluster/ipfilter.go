package cluster

import (
	"sync"
)

// EnginesFilter is exported
type EnginesFilter struct {
	sync.Mutex
	engines map[string]*Engine
}

// NewEnginesFilter is exported
func NewEnginesFilter() *EnginesFilter {

	return &EnginesFilter{
		engines: make(map[string]*Engine),
	}
}

// Size is exported
func (filter *EnginesFilter) Size() int {

	filter.Lock()
	defer filter.Unlock()
	return len(filter.engines)
}

// Set is exported
func (filter *EnginesFilter) Set(engine *Engine) {

	filter.Lock()
	if engine != nil {
		if _, ret := filter.engines[engine.IP]; !ret {
			filter.engines[engine.IP] = engine
		}
	}
	filter.Unlock()
}

// Filter is exported
func (filter *EnginesFilter) Filter(engines []*Engine) []*Engine {

	filter.Lock()
	defer filter.Unlock()
	if len(filter.engines) == 0 {
		return engines
	}

	out := []*Engine{}
	for _, engine := range engines {
		if _, ret := filter.engines[engine.IP]; !ret {
			out = append(out, engine)
		}
	}
	return out
}
