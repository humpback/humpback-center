package node

import "github.com/boltdb/bolt"
import "github.com/humpback/humpback-center/cluster/types"
import "github.com/humpback/humpback-center/cluster/storage/dao"
import "github.com/humpback/humpback-center/cluster/storage/entry"

import (
	"strings"
)

const (
	// BucketName represents the name of the bucket where this stores data.
	BucketName = "nodes"
)

// NodeStorage is exported
type NodeStorage struct {
	driver *bolt.DB
}

// NewNodeStorage is exported
func NewNodeStorage(driver *bolt.DB) (*NodeStorage, error) {

	err := dao.CreateBucket(driver, BucketName)
	if err != nil {
		return nil, err
	}

	return &NodeStorage{
		driver: driver,
	}, nil
}

// NodeByIP is exported
func (nodeStorage *NodeStorage) NodeByIP(ip string) (*entry.Node, error) {

	var node entry.Node
	err := dao.GetObject(nodeStorage.driver, BucketName, []byte(ip), &node)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// NodeByID is exported
func (nodeStorage *NodeStorage) NodeByID(id string) (*entry.Node, error) {

	var node *entry.Node
	err := nodeStorage.driver.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var value entry.Node
			err := dao.UnmarshalObject(v, &value)
			if err != nil {
				return err
			}
			if strings.ToUpper(value.ID) == strings.ToUpper(id) {
				node = &value
				break
			}
		}
		if node == nil {
			return dao.ErrStorageObjectNotFound
		}
		return nil
	})
	return node, err
}

// NodeByName is exported
func (nodeStorage *NodeStorage) NodeByName(name string) (*entry.Node, error) {

	var node *entry.Node
	err := nodeStorage.driver.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var value entry.Node
			err := dao.UnmarshalObject(v, &value)
			if err != nil {
				return err
			}
			if strings.ToUpper(value.Name) == strings.ToUpper(name) {
				node = &value
				break
			}
		}
		if node == nil {
			return dao.ErrStorageObjectNotFound
		}
		return nil
	})
	return node, err
}

// SetNodeData set a node entry.
func (nodeStorage *NodeStorage) SetNodeData(nodeData *types.NodeData) error {

	var node *entry.Node
	node, _ = nodeStorage.NodeByIP(nodeData.IP)
	if node == nil {
		node = &entry.Node{
			NodeLabels:   map[string]string{},
			Availability: "Active",
		}
	}

	node.NodeData = nodeData
	return nodeStorage.driver.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		data, err := dao.MarshalObject(node)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(node.IP), data)
	})
}

// SetNodeLabels set a node labels.
func (nodeStorage *NodeStorage) SetNodeLabels(ip string, labels map[string]string) error {

	var node *entry.Node
	node, err := nodeStorage.NodeByIP(ip)
	if err != nil {
		return err
	}

	node.NodeLabels = labels
	return nodeStorage.driver.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		data, err := dao.MarshalObject(node)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(node.IP), data)
	})
}

// DeleteNode deletes a node entry.
func (nodeStorage *NodeStorage) DeleteNode(ip string) error {

	return dao.DeleteObject(nodeStorage.driver, BucketName, []byte(ip))
}
