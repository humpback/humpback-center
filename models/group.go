package models

// Group is exported
type Group struct {
	ID      string   `json:"id" bson:"ID"`
	Servers []string `json:"servers" bson:"Servers"`
}

func NewGroup(id string) *Group {

	return &Group{
		ID:      id,
		Servers: []string{},
	}
}

func (group *Group) Contains(server string) bool {

	if len(group.Servers) > 0 {
		for _, it := range group.Servers {
			if it == server {
				return true
			}
		}
	}
	return false
}

func (group *Group) Insert(server string) bool {

	if ret := group.Contains(server); !ret {
		group.Servers = append(group.Servers, server)
		return true
	}
	return false
}

func (group *Group) Remove(server string) bool {

	for i, it := range group.Servers {
		if it == server {
			group.Servers = append(group.Servers[:i], group.Servers[i+1:]...)
			return true
		}
	}
	return false
}

func (group *Group) Clear() {

	if len(group.Servers) > 0 {
		group.Servers = group.Servers[0:0]
	}
}
