package storage

import "github.com/boltdb/bolt"
import "github.com/humpback/gounits/system"
import "github.com/humpback/humpback-center/cluster/storage/node"

import (
	"fmt"
	"path"
	"path/filepath"
	"time"
)

const (
	databaseFileName = "data.db"
)

// DataStorage defines the implementation of datastore using
// BoltDB as the storage system.
type DataStorage struct {
	path        string
	driver      *bolt.DB
	NodeStorage *node.NodeStorage
}

// NewDataStorage is exported
func NewDataStorage(storePath string) (*DataStorage, error) {

	var err error
	storePath, err = filepath.Abs(storePath)
	if err != nil {
		return nil, fmt.Errorf("storage driver path invalid, %s", err)
	}

	storePath = filepath.Clean(storePath)
	if err = system.MakeDirectory(storePath); err != nil {
		return nil, fmt.Errorf("storage driver make directory failure, %s", err)
	}

	databasePath := path.Join(storePath, databaseFileName)
	databasePath = filepath.Clean(databasePath)
	return &DataStorage{
		path: databasePath,
	}, nil
}

// Open is exported
// open storage driver file.
func (storage *DataStorage) Open() error {

	if storage.driver == nil {
		driver, err := bolt.Open(storage.path, 0600, &bolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			return err
		}

		nodeStorage, err := node.NewNodeStorage(driver)
		if err != nil {
			return err
		}

		storage.NodeStorage = nodeStorage
		storage.driver = driver
	}
	return nil
}

// Close is exported
// Close storage driver file.
func (storage *DataStorage) Close() error {

	if storage.driver != nil {
		return storage.driver.Close()
	}
	return nil
}
