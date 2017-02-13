package models

import (
	"sync"
)

// Server is exported
// IP, cluster server ipaddr.
// Status, cluster engine status.
type Server struct {
	IP     string `json:"ip"`
	Status string `json:"status"`
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

func (group *Group) Contains(ip string) bool {

	group.RLock()
	defer group.RUnlock()
	if len(group.Servers) > 0 {
		for _, server := range group.Servers {
			if server.IP == ip {
				return true
			}
		}
	}
	return false
}

func (group *Group) Set(ip string, status string) bool {

	group.Lock()
	defer group.Unlock()
	for i, server := range group.Servers {
		if server.IP == ip {
			group.Servers[i].Status = status
			return true
		}
	}
	return false
}

func (group *Group) Insert(ip string, status string) bool {

	ret := group.Contains(ip)
	if ret {
		return false
	}

	group.Lock()
	group.Servers = append(group.Servers, Server{
		IP:     ip,
		Status: status,
	})
	group.Unlock()
	return true
}

func (group *Group) Remove(ip string) bool {

	group.Lock()
	defer group.Unlock()
	for i, server := range group.Servers {
		if server.IP == ip {
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
