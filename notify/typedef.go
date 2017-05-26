package notify

import "net/http"

//EndPoint is exported
type EndPoint struct {
	Name     string      `yaml:"name"`
	URL      string      `yaml:"url"`
	Enabled  bool        `yaml:"enabled"`
	Sender   string      `yaml:"sender"`
	Headers  http.Header `yaml:"headers"`
	Host     string      `yaml:"host"`
	Port     int         `yaml:"port"`
	User     string      `yaml:"user"`
	Password string      `yaml:"password"`
}

//Notifications is exported
type Notifications struct {
	EndPoints []EndPoint `yaml:"endpoints,omitempty"`
}

//Engine is exported
type Engine struct {
	IP    string
	Name  string
	State string
}

//WatchGroup is exported
type WatchGroup struct {
	GroupID     string
	GroupName   string
	Location    string
	ContactInfo string
	Engines     []*Engine
}

//WatchGroups is exported
type WatchGroups map[string]*WatchGroup

//Container is exported
type Container struct {
	ID     string
	Name   string
	Server string
	State  string
}

//GroupMeta is exported
type GroupMeta struct {
	MetaID      string
	MetaName    string
	Location    string
	GroupID     string
	GroupName   string
	Instances   int
	Image       string
	ContactInfo string
	Containers  []Container
}
