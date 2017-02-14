package models

// Engine is exported
type Engine struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	IP      string   `json:"ip"`
	Addr    string   `json:"addr"`
	Version string   `json:"version"`
	Labels  []string `json:"labels"`
	State   string   `json:"state"`
}

// NewEngine is exported
func NewEngine(id string, name string, ip string, addr string, version string, labels []string, state string) *Engine {

	return &Engine{
		ID:      id,
		Name:    name,
		IP:      ip,
		Addr:    addr,
		Version: version,
		Labels:  labels,
		State:   state,
	}
}
