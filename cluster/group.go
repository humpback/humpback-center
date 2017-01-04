package cluster

import (
	"fmt"
	"sync"
)

// Group is exported
type Group struct {
	sync.RWMutex
	ID      string
	Name    string
	engines map[string]*Engine
}

// NewGroup is exported
func NewGroup(id string, name string) *Group {

	return &Group{
		ID:   id,
		Name: name,
	}
}

func (g *Group) AddEngine(ip string) (*Engine, error) {

	g.Lock()
	defer g.Unlock()
	if _, ret := g.engines[ip]; !ret {
		engine := NewEngine(ip)
		//go monitor engine state
		//engine.connect()
		return engine, nil
	}
	return nil, fmt.Errorf("engine %s exists already.")
}

func (g *Group) RemoveEngine(ip string) bool {

	g.Lock()
	defer g.Unlock()
	if _, ret := g.engines[ip]; ret {
		//engine.disconect()
		delete(g.engines, ip)
		return true
	}
	return false
}
