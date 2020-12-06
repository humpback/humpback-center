package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/humpback/common/models"
	"github.com/humpback/gounits/rand"
	"github.com/humpback/gounits/system"
	"github.com/humpback/humpback-center/cluster/types"
)

// ContainerBaseConfig is exported
type ContainerBaseConfig struct {
	Index int `json:"Index"`
	models.Container
	MetaData *MetaData `json:"-"`
}

// SortContainerBaseConfigs is exported
type SortContainerBaseConfigs []*ContainerBaseConfig

func (containers SortContainerBaseConfigs) Len() int {

	return len(containers)
}

func (containers SortContainerBaseConfigs) Swap(i, j int) {

	containers[i], containers[j] = containers[j], containers[i]
}

func (containers SortContainerBaseConfigs) Less(i, j int) bool {

	return containers[i].Index < containers[j].Index
}

// MetaBase is exported
type MetaBase struct {
	GroupID               string           `json:"GroupId"`
	MetaID                string           `json:"MetaId"`
	IsRemoveDelay         bool             `json:"IsRemoveDelay"`
	IsRecovery            bool             `json:"IsRecovery"`
	Instances             int              `json:"Instances"`
	WebHooks              types.WebHooks   `json:"WebHooks"`
	Placement             types.Placement  `json:"Placement"`
	ImageTag              string           `json:"ImageTag"`
	Config                models.Container `json:"Config"`
	CreateAt              int64            `json:"CreateAt"`
	LastUpdateAt          int64            `json:"LastUpdateAt"`
	AvailableNodesChanged bool             `json:"AvailableNodesChanged"`
}

// MetaData is exported
type MetaData struct {
	MetaBase
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

	var (
		metaid string
	)

	for {
		metaid = rand.UUID(true)
		if _, ret := cache.data[metaid]; ret {
			continue
		}
		break
	}
	return metaid
}

// MakeContainerIdleIndex is exported
// Return a baseContainerConfig idle index
func (cache *ContainersConfigCache) MakeContainerIdleIndex(metaid string) int {

	cache.Lock()
	defer cache.Unlock()
	metaData, ret := cache.data[metaid]
	if !ret {
		return -1
	}

	index := 1
	sortContainerBaseConfigs := SortContainerBaseConfigs(metaData.BaseConfigs)
	if len(sortContainerBaseConfigs) == 0 {
		return index
	}

	sort.Sort(sortContainerBaseConfigs)
	for {
		for i := 0; i < len(sortContainerBaseConfigs); i++ {
			if sortContainerBaseConfigs[i].Index != index {
				return index
			}
			index++
		}
		break
	}
	return index
}

// SetImageTag is exported
// Return set tag result
func (cache *ContainersConfigCache) SetImageTag(metaid string, imagetag string) bool {

	cache.Lock()
	defer cache.Unlock()
	if strings.TrimSpace(imagetag) == "" {
		imagetag = "latest"
	}
	if metaData, ret := cache.data[metaid]; ret {
		originalTag := metaData.ImageTag
		metaData.ImageTag = imagetag
		originalImage := metaData.Config.Image
		tempPaths := strings.Split(originalImage, "/")
		nPos := strings.LastIndex(tempPaths[len(tempPaths)-1], ":")
		if nPos == -1 {
			nPos = len(originalImage)
		}
		tempPaths[len(tempPaths)-1] = strings.Split(tempPaths[len(tempPaths)-1], ":")[0] + ":" + imagetag
		metaData.Config.Image = strings.Join(tempPaths, "/")
		if err := cache.writeMetaData(metaData); err != nil {
			metaData.ImageTag = originalTag
			metaData.Config.Image = originalImage
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
func (cache *ContainersConfigCache) GetMetaDataOfName(groupid string, name string) *MetaData {

	cache.RLock()
	defer cache.RUnlock()
	for _, metaData := range cache.data {
		if metaData.GroupID == groupid && metaData.Config.Name == name {
			return metaData
		}
	}
	return nil
}

// GetMetaDataOfContainer is exported
// Return containerid of a metadata
func (cache *ContainersConfigCache) GetMetaDataOfContainer(containerid string) *MetaData {

	cache.RLock()
	defer cache.RUnlock()
	for _, metaData := range cache.data {
		for _, baseConfig := range metaData.BaseConfigs {
			if baseConfig.ID == containerid {
				return metaData
			}
		}
	}
	return nil
}

// GetMetaDataBaseConfigs is exported
func (cache *ContainersConfigCache) GetMetaDataBaseConfigs(metaid string) []*ContainerBaseConfig {

	cache.RLock()
	defer cache.RUnlock()
	baseConfigs := []*ContainerBaseConfig{}
	if metaData, ret := cache.data[metaid]; ret {
		baseConfigs = metaData.BaseConfigs
	}
	return baseConfigs
}

// GetMetaDataBaseConfigsCount is exported
// result < 0, not found metadata
func (cache *ContainersConfigCache) GetMetaDataBaseConfigsCount(metaid string) int {

	cache.RLock()
	defer cache.RUnlock()
	if metaData, ret := cache.data[metaid]; ret {
		return len(metaData.BaseConfigs)
	}
	return -1
}

// ContainsMetaData is exported
// Return bool, find metadata name
func (cache *ContainersConfigCache) ContainsMetaData(groupid string, name string) bool {

	if len(strings.TrimSpace(name)) > 0 {
		metaData := cache.GetMetaDataOfName(groupid, name)
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

// SetAvailableNodesChanged is exported
func (cache *ContainersConfigCache) SetAvailableNodesChanged(metaid string, changed bool) {

	cache.Lock()
	if metaData, ret := cache.data[metaid]; ret {
		metaData.AvailableNodesChanged = changed
		metaData.LastUpdateAt = time.Now().Unix()
		cache.writeMetaData(metaData)
	}
	cache.Unlock()
}

// SetMetaData is exported
func (cache *ContainersConfigCache) SetMetaData(metaid string, instances int, webhooks types.WebHooks, placement types.Placement, config models.Container, isremovedelay bool, isrecovery bool) {

	cache.Lock()
	if metaData, ret := cache.data[metaid]; ret {
		metaData.IsRemoveDelay = isremovedelay
		metaData.IsRecovery = isrecovery
		metaData.Instances = instances
		metaData.WebHooks = webhooks
		metaData.Placement = placement
		metaData.Config = config
		metaData.LastUpdateAt = time.Now().Unix()
		cache.writeMetaData(metaData)
	}
	cache.Unlock()
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

// CreateMetaData is exported
func (cache *ContainersConfigCache) CreateMetaData(groupid string, instances int, webhooks types.WebHooks, placement types.Placement, config models.Container, isremovedelay bool, isrecovery bool) (*MetaData, error) {

	cache.Lock()
	defer cache.Unlock()
	for _, pMetaData := range cache.data {
		if pMetaData.GroupID == groupid && pMetaData.Config.Name == config.Name {
			return pMetaData, fmt.Errorf("create meta conflict, %s", config.Name)
		}
	}
	metaid := cache.MakeUniqueMetaID()
	imageTag := getImageTag(config.Image)
	metaData := &MetaData{
		MetaBase: MetaBase{
			GroupID:       groupid,
			MetaID:        metaid,
			IsRemoveDelay: isremovedelay,
			IsRecovery:    isrecovery,
			Instances:     instances,
			WebHooks:      webhooks,
			Placement:     placement,
			ImageTag:      imageTag,
			Config:        config,
			CreateAt:      time.Now().Unix(),
		},
		BaseConfigs: []*ContainerBaseConfig{},
	}

	if err := cache.writeMetaData(metaData); err != nil {
		return nil, err
	}
	cache.data[metaid] = metaData
	return metaData, nil
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

// CreateContainerBaseConfig is exported
func (cache *ContainersConfigCache) CreateContainerBaseConfig(metaid string, baseConfig *ContainerBaseConfig) {

	cache.Lock()
	defer cache.Unlock()
	if metaData, ret := cache.data[metaid]; ret {
		for i := 0; i < len(metaData.BaseConfigs); i++ {
			if metaData.BaseConfigs[i].ID == baseConfig.ID {
				return
			}
		}
		baseConfig.MetaData = metaData
		metaData.BaseConfigs = append(metaData.BaseConfigs, baseConfig)
		cache.writeMetaData(metaData)
	}
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

// ClearContainerBaseConfig is exported
func (cache *ContainersConfigCache) ClearContainerBaseConfig(metaid string) {

	cache.Lock()
	metaData, ret := cache.data[metaid]
	if ret {
		metaData.BaseConfigs = []*ContainerBaseConfig{}
		cache.writeMetaData(metaData)
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

	metaData := &MetaData{
		MetaBase: MetaBase{
			IsRemoveDelay: true, //isremovedelay default is enabled.
			IsRecovery:    true, //isrecovery default is enabled.
		},
	}

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
	return os.Remove(metaPath)
}
