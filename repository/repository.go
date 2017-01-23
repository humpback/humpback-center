package repository

type RepositoryCache struct {
	Name             string
	Host             string
	Root             string
	DirFilter        []string
	AuthRoot         string
	MaxRoutine       int
	Recovery         int
	EngineAPIVersion string
	DockerAPIVersion string
}

func NewRepositoryCache() *RepositoryCache {
	return &RepositoryCache{}
}
