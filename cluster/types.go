package cluster

// regist to cluster options
// Labels: humpback node custum values.
// example labels {"node=wh7", "kernelversion=4.4.0", "os=centos6.8"}
type RegistClusterOptions struct {
	Name    string   `json:"name"`
	Addr    string   `json:"addr"`
	Labels  []string `json:"labels"`
	Version string   `json:"version"`
}
