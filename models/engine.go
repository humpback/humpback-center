package models

// Engine is exported
type Engine struct {
	ID      string   `json:"id"`
	IP      string   `json:"ip"`
	APIAddr string   `json:"apiaddr"`
	Labels  []string `json:"labels"`
	State   string   `json:"state"`
}

// NewEngine is exported
func NewEngine(id string, ip string, apiaddr string, labels []string, state string) *Engine {

	return &Engine{
		ID:      id,
		IP:      ip,
		APIAddr: apiaddr,
		Labels:  labels,
		State:   state,
	}
}
