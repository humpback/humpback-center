package cluster

type Engine struct {
	Key string
}

func NewEngine(key string) *Engine {

	return &Engine{
		Key: key,
	}
}
