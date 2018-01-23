package cluster

import "sync"

// EnginePriorities is exported
type EnginePriorities struct {
	sync.RWMutex
	Engines map[string]*Engine
}

// NewEnginePriorities is exported
func NewEnginePriorities() *EnginePriorities {

	return &EnginePriorities{
		Engines: make(map[string]*Engine),
	}
}

// Select is exported
func (priorities *EnginePriorities) Select() *Engine {

	var engine *Engine
	priorities.Lock()
	defer priorities.Unlock()
	if len(priorities.Engines) == 0 {
		return nil
	}

	for containerid, e := range priorities.Engines {
		engine = e
		delete(priorities.Engines, containerid)
		break
	}
	return engine
}

// Size is exported
func (priorities *EnginePriorities) Size() int {

	size := 0
	priorities.RLock()
	size = len(priorities.Engines)
	priorities.RUnlock()
	return size
}

// Add is exported
func (priorities *EnginePriorities) Add(containerid string, engine *Engine) {

	priorities.Lock()
	if _, ret := priorities.Engines[containerid]; !ret {
		priorities.Engines[containerid] = engine
	}
	priorities.Unlock()
}

// Remove is exported
func (priorities *EnginePriorities) Remove(containerid string) {

	priorities.Lock()
	delete(priorities.Engines, containerid)
	priorities.Unlock()
}
