package cluster

import "github.com/humpback/gounits/logger"
import "github.com/humpback/gounits/system"
import "github.com/humpback/humpback-agent/models"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ContainerBaseConfig struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	GroupID string            `json:"groupid"`
	Config  *models.Container `json:"config"`
}

type ContainerConfigCache struct {
	sync.RWMutex
	Root  string
	Pairs map[string]*ContainerBaseConfig
}

func NewContainerConfigCache(root string) *ContainerConfigCache {

	if len(strings.TrimSpace(root)) == 0 {
		root = "./cache"
	}

	system.MakeDirectory(root)
	return &ContainerConfigCache{
		Root:  root,
		Pairs: make(map[string]*ContainerBaseConfig),
	}
}

func (cache *ContainerConfigCache) Init() {

	if len(cache.Pairs) > 0 {
		cache.Pairs = make(map[string]*ContainerBaseConfig)
	}

	fis, err := ioutil.ReadDir(cache.Root)
	if err != nil {
		return
	}

	cache.Lock()
	for _, fi := range fis {
		if !fi.IsDir() {
			fpath, _ := filepath.Abs(cache.Root + "/" + fi.Name())
			baseConfig := cache.readContainerConfig(fpath)
			if baseConfig != nil {
				cache.Pairs[baseConfig.ID] = baseConfig
				logger.INFO("[#cluster#] cache baseConfig:%s", baseConfig.ID)
			}
		}
	}
	cache.Unlock()
}

func (cache *ContainerConfigCache) Get(containerid string) *ContainerBaseConfig {

	cache.RLock()
	defer cache.RUnlock()
	baseConfig, ret := cache.Pairs[containerid]
	if !ret {
		fpath, err := filepath.Abs(cache.Root + "/" + containerid + ".json")
		if err != nil {
			return nil
		}
		return cache.readContainerConfig(fpath)
	}
	return baseConfig
}

func (cache *ContainerConfigCache) Write(containerid string, groupid string, name string, config *models.Container) error {

	cache.Lock()
	defer cache.Unlock()

	var (
		ret        bool
		baseConfig *ContainerBaseConfig
	)

	baseConfig, ret = cache.Pairs[containerid]
	if ret {
		return fmt.Errorf("container %s config is already exists", containerid)
	}

	baseConfig = &ContainerBaseConfig{
		ID:      containerid,
		Name:    name,
		GroupID: groupid,
		Config:  config,
	}

	fpath, err := filepath.Abs(cache.Root + "/" + containerid + ".json")
	if err != nil {
		return nil
	}

	if err := cache.writeContainerConfig(fpath, baseConfig); err != nil {
		return err
	}
	cache.Pairs[containerid] = baseConfig
	return nil
}

func (cache *ContainerConfigCache) Remove(containerid string) error {

	cache.Lock()
	defer cache.Unlock()
	if _, ret := cache.Pairs[containerid]; !ret {
		return fmt.Errorf("container %s config not found", containerid)
	}

	containerPath, _ := filepath.Abs(cache.Root + "/" + containerid + ".json")
	if err := os.Remove(containerPath); err != nil {
		return err
	}
	delete(cache.Pairs, containerid)
	return nil
}

func (cache *ContainerConfigCache) readContainerConfig(fpath string) *ContainerBaseConfig {

	if ret := system.FileExist(fpath); !ret {
		return nil
	}

	fd, err := os.OpenFile(fpath, os.O_RDONLY, 0777)
	if err != nil {
		return nil
	}

	baseConfig := &ContainerBaseConfig{}
	if err := json.NewDecoder(fd).Decode(baseConfig); err != nil {
		return nil
	}
	return baseConfig
}

func (cache *ContainerConfigCache) writeContainerConfig(fpath string, baseConfig *ContainerBaseConfig) error {

	buffer := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buffer).Encode(baseConfig); err != nil {
		return err
	}
	return ioutil.WriteFile(fpath, buffer.Bytes(), 0777)
}
