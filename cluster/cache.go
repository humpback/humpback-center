package cluster

import "github.com/humpback/gounits/system"
import "github.com/humpback/gounits/rand"
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

// ContainerBaseConfig is exported
type ContainerBaseConfig struct {
	models.Container
	MetaData *MetaData `json:"-"`
}

// MetaData is exported
type MetaData struct {
	GroupID     string                 `json:"GroupId"`
	MetaID      string                 `json:"MetaId"`
	ImageTag    string                 `json:"ImageTag"`
	Name        string                 `json:"Name"`
	BaseConfigs []*ContainerBaseConfig `json:"BaseConfigs"`
}

// ContainersConfigCache is exported
type ContainersConfigCache struct {
	sync.RWMutex
	Root string
	data map[string]*MetaData
}

// NewContainersConfigCache is exported
// Structure ContainersCache
func NewContainersConfigCache(root string) (*ContainersConfigCache, error) {

	if len(strings.TrimSpace(root)) == 0 {
		root = "./cache"
	}

	if err := system.MakeDirectory(root); err != nil {
		return nil, fmt.Errorf("containers cache directory init error:%s", err.Error())
	}

	return &ContainersConfigCache{
		Root: root,
		data: make(map[string]*MetaData),
	}, nil
}

// Init is exported
// Initialize containers baseConfig, load cache directory's metaData
// First clear containers cache
func (cache *ContainersConfigCache) Init() {

	if len(cache.data) > 0 {
		cache.data = make(map[string]*MetaData)
	}

	fis, err := ioutil.ReadDir(cache.Root)
	if err != nil {
		return
	}

	cache.Lock()
	for _, fi := range fis {
		if !fi.IsDir() {
			metaData, err := cache.readMetaData(fi.Name())
			if err == nil {
				for _, baseConfig := range metaData.BaseConfigs {
					baseConfig.MetaData = metaData
				}
				cache.data[metaData.MetaID] = metaData
			}
		}
	}
	cache.Unlock()
}

// MakeUniqueMetaID is exported
// Return a new create unique metaid
func (cache *ContainersConfigCache) MakeUniqueMetaID() string {

	var metaid string
	for {
		metaid = rand.UUID(true)
		if _, ret := cache.data[metaid]; ret {
			continue
		}
		break
	}
	return metaid
}

// SetImageTag is exported
// Return set tag result
func (cache *ContainersConfigCache) SetImageTag(metaid string, imagetag string) bool {

	cache.Lock()
	defer cache.Unlock()
	if metaData, ret := cache.data[metaid]; ret {
		originalTag := metaData.ImageTag
		metaData.ImageTag = imagetag
		if err := cache.writeMetaData(metaData); err != nil {
			metaData.ImageTag = originalTag
			return false
		}
		return true
	}
	return false
}

// GetMetaData is exported
// Return metaid of a metadata
func (cache *ContainersConfigCache) GetMetaData(metaid string) *MetaData {

	cache.RLock()
	defer cache.RUnlock()
	metaData, ret := cache.data[metaid]
	if ret {
		return metaData
	}
	return nil
}

// GetMetaDataOfName is exported
// Return name of a metadata
func (cache *ContainersConfigCache) GetMetaDataOfName(name string) *MetaData {

	cache.RLock()
	defer cache.RUnlock()
	for _, metaData := range cache.data {
		if metaData.Name == name {
			return metaData
		}
	}
	return nil
}

// ContainsMetaData is exported
// Return bool, find metadata name
func (cache *ContainersConfigCache) ContainsMetaData(name string) bool {

	if len(strings.TrimSpace(name)) > 0 {
		metaData := cache.GetMetaDataOfName(name)
		return metaData != nil
	}
	return false
}

// GetGroupMetaData is exported
func (cache *ContainersConfigCache) GetGroupMetaData(groupid string) []*MetaData {

	cache.RLock()
	out := []*MetaData{}
	for _, metaData := range cache.data {
		if metaData.GroupID == groupid {
			out = append(out, metaData)
		}
	}
	cache.RUnlock()
	return out
}

// RemoveMetaData is exported
// Remove metaid of a metadata
func (cache *ContainersConfigCache) RemoveMetaData(metaid string) bool {

	cache.Lock()
	defer cache.Unlock()
	if _, ret := cache.data[metaid]; ret {
		if err := cache.removeMeteData(metaid); err == nil {
			delete(cache.data, metaid)
			return true
		}
	}
	return false
}

// RemoveGroupMetaData is exported
func (cache *ContainersConfigCache) RemoveGroupMetaData(groupid string) bool {

	cache.Lock()
	removed := false
	for _, metaData := range cache.data {
		if metaData.GroupID == groupid {
			if err := cache.removeMeteData(metaData.MetaID); err == nil {
				delete(cache.data, metaData.MetaID)
				removed = true
			}
		}
	}
	cache.Unlock()
	return removed
}

// GetContainerBaseConfig is exported
func (cache *ContainersConfigCache) GetContainerBaseConfig(metaid string, containerid string) *ContainerBaseConfig {

	cache.RLock()
	defer cache.RUnlock()
	metaData, ret := cache.data[metaid]
	if ret {
		for _, baseConfig := range metaData.BaseConfigs {
			if baseConfig.ID == containerid {
				return baseConfig
			}
		}
	}
	return nil
}

// SetContainerBaseConfig is exported
func (cache *ContainersConfigCache) SetContainerBaseConfig(metaid string, groupid string, name string, baseConfig *ContainerBaseConfig) {

	var (
		ret      bool
		metaData *MetaData
	)

	cache.Lock()
	defer cache.Unlock()
	metaData, ret = cache.data[metaid]
	if !ret {
		imageTag := "latest"
		imageName := strings.SplitN(baseConfig.Image, ":", 2)
		if len(imageName) == 2 {
			imageTag = imageName[1]
		}
		metaData = &MetaData{
			GroupID:     groupid,
			MetaID:      metaid,
			ImageTag:    imageTag,
			Name:        name,
			BaseConfigs: []*ContainerBaseConfig{baseConfig},
		}
		baseConfig.MetaData = metaData
		cache.data[metaData.MetaID] = metaData
	} else {
		for i := 0; i < len(metaData.BaseConfigs); i++ {
			if metaData.BaseConfigs[i].ID == baseConfig.ID {
				return
			}
		}
		baseConfig.MetaData = metaData
		metaData.BaseConfigs = append(metaData.BaseConfigs, baseConfig)
	}
	cache.writeMetaData(metaData)
}

// RemoveContainerBaseConfig is exported
func (cache *ContainersConfigCache) RemoveContainerBaseConfig(metaid string, containerid string) {

	cache.Lock()
	metaData, ret := cache.data[metaid]
	if ret {
		for i, baseConfig := range metaData.BaseConfigs {
			if baseConfig.ID == containerid {
				metaData.BaseConfigs = append(metaData.BaseConfigs[:i], metaData.BaseConfigs[i+1:]...)
				cache.writeMetaData(metaData)
				break
			}
		}
	}
	cache.Unlock()
}

// readMetaData is exported
func (cache *ContainersConfigCache) readMetaData(metaid string) (*MetaData, error) {

	metaPath, err := filepath.Abs(cache.Root + "/" + metaid)
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	metaData := &MetaData{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(metaData); err != nil {
		return nil, nil
	}
	return metaData, nil
}

// writeMetaData is exported
func (cache *ContainersConfigCache) writeMetaData(metaData *MetaData) error {

	metaPath, err := filepath.Abs(cache.Root + "/" + metaData.MetaID)
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buffer).Encode(metaData); err != nil {
		return err
	}
	return ioutil.WriteFile(metaPath, buffer.Bytes(), 0777)
}

// removeMeteData is exported
func (cache *ContainersConfigCache) removeMeteData(metaid string) error {

	metaPath, err := filepath.Abs(cache.Root + "/" + metaid)
	if err != nil {
		return err
	}

	if err := os.Remove(metaPath); err != nil {
		return err
	}
	return nil
}
