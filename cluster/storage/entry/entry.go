package entry

import "github.com/humpback/humpback-center/cluster/types"

//Node is exported
type Node struct {
	*types.NodeData
	NodeLabels   map[string]string `json:"nodelabels"`
	Availability string            `json:"availability"`
}
