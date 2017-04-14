package cluster

import (
	"sync"
)

// EnginesFilter is exported
type EnginesFilter struct {
	sync.RWMutex
	allocEngines map[string]*Engine
	failEngines  map[string]*Engine
}

// NewEnginesFilter is exported
func NewEnginesFilter() *EnginesFilter {

	return &EnginesFilter{
		allocEngines: make(map[string]*Engine),
		failEngines:  make(map[string]*Engine),
	}
}

// Size is exported
func (filter *EnginesFilter) Size() int {

	filter.RLock()
	defer filter.RUnlock()
	return len(filter.allocEngines) + len(filter.failEngines)
}

// SetAllocEngine is exported
func (filter *EnginesFilter) SetAllocEngine(engine *Engine) {

	filter.Lock()
	if engine != nil {
		if _, ret := filter.allocEngines[engine.IP]; !ret {
			filter.allocEngines[engine.IP] = engine
		}
	}
	filter.Unlock()
}

// SetFailEngine is exported
func (filter *EnginesFilter) SetFailEngine(engine *Engine) {

	filter.Lock()
	if engine != nil {
		if _, ret := filter.failEngines[engine.IP]; !ret {
			filter.failEngines[engine.IP] = engine
		}
	}
	filter.Unlock()
}

// AllocEngines is exported
func (filter *EnginesFilter) AllocEngines() []*Engine {

	filter.RLock()
	defer filter.RUnlock()
	engines := []*Engine{}
	for _, engine := range filter.allocEngines {
		engines = append(engines, engine)
	}
	return engines
}

// FailEngines is exported
func (filter *EnginesFilter) FailEngines() []*Engine {

	filter.RLock()
	defer filter.RUnlock()
	engines := []*Engine{}
	for _, engine := range filter.failEngines {
		engines = append(engines, engine)
	}
	return engines
}

// Filter is exported
func (filter *EnginesFilter) Filter(engines []*Engine) []*Engine {

	if filter.Size() == 0 {
		return engines
	}

	filter.RLock()
	filterEngines := make(map[string]*Engine)
	for _, engine := range filter.allocEngines {
		filterEngines[engine.IP] = engine
	}

	for _, engine := range filter.failEngines {
		filterEngines[engine.IP] = engine
	}

	out := []*Engine{}
	for _, engine := range engines {
		if _, ret := filterEngines[engine.IP]; !ret {
			out = append(out, engine)
		}
	}
	filter.RUnlock()
	return out
}
