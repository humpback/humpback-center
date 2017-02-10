package models

import (
	"sync"
)

// Server is exported
// EngineID == "", docker engine is offline.
type Server struct {
	IPAddr   string `json:"ipaddr"`
	EngineID string `json:"engineid"`
}

// Group is exported
type Group struct {
	sync.RWMutex
	ID      string   `json:"id"`
	Servers []Server `json:"servers"`
}

// NewGroup is exported
func NewGroup(id string) *Group {

	return &Group{
		ID:      id,
		Servers: []Server{},
	}
}

func (group *Group) Contains(ipaddr string) bool {

	group.RLock()
	defer group.RUnlock()
	if len(group.Servers) > 0 {
		for _, server := range group.Servers {
			if server.IPAddr == ipaddr {
				return true
			}
		}
	}
	return false
}

func (group *Group) Set(ipaddr string, engineid string) bool {

	group.Lock()
	defer group.Unlock()
	for i, server := range group.Servers {
		if server.IPAddr == ipaddr {
			group.Servers[i].EngineID = engineid
			return true
		}
	}
	return false
}

func (group *Group) Insert(ipaddr string, engineid string) bool {

	ret := group.Contains(ipaddr)
	if ret {
		return false
	}

	group.Lock()
	group.Servers = append(group.Servers, Server{
		IPAddr:   ipaddr,
		EngineID: engineid,
	})
	group.Unlock()
	return true
}

func (group *Group) Remove(ipaddr string) bool {

	group.Lock()
	defer group.Unlock()
	for i, server := range group.Servers {
		if server.IPAddr == ipaddr {
			group.Servers = append(group.Servers[:i], group.Servers[i+1:]...)
			return true
		}
	}
	return false
}

func (group *Group) Clear() {

	group.Lock()
	if len(group.Servers) > 0 {
		group.Servers = group.Servers[0:0]
	}
	group.Unlock()
}
