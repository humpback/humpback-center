package cluster

import "github.com/humpback/gounits/http"
import "github.com/humpback/gounits/logger"
import "github.com/humpback/humpback-agent/models"

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"
)

// HookEvent is exported
type HookEvent int

const (
	CreateMetaEvent HookEvent = iota + 1
	RemoveMetaEvent
	OperateMetaEvent
	UpdateMetaEvent
	UpgradeMetaEvent
	MigrateMetaEvent
	RecoveryMetaEvent
)

func (event HookEvent) String() string {

	switch event {
	case CreateMetaEvent:
		return "CreateMetaEvent"
	case RemoveMetaEvent:
		return "RemoveMetaEvent"
	case OperateMetaEvent:
		return "OperateMetaEvent"
	case UpdateMetaEvent:
		return "UpdateMetaEvent"
	case UpgradeMetaEvent:
		return "UpgradeMetaEvent"
	case MigrateMetaEvent:
		return "MigrateMetaEvent"
	case RecoveryMetaEvent:
		return "RecoveryMetaEvent"
	}
	return ""
}

// HookContainer is exported
type HookContainer struct {
	IP        string           `json:"IP"`
	Name      string           `json:"Name"`
	Container models.Container `json:"Container"`
}

// HookContainers is exported
type HookContainers []*HookContainer

// Hook is exported
type Hook struct {
	Timestamp int64    `json:"Timestamp"`
	Event     string   `json:"Event"`
	MetaBase  MetaBase `json:"MetaBase"`
	HookContainers
}

// NewHook is exported
func NewHook(cluster *Cluster, metaData *MetaData, timeStamp int64, hookEvent HookEvent) *Hook {

	hookContainers := HookContainers{}
	engines := cluster.GetGroupEngines(metaData.GroupID)
	for _, engine := range engines {
		if engine.IsHealthy() {
			containers := engine.Containers(metaData.MetaID)
			for _, container := range containers {
				hookContainers = append(hookContainers, &HookContainer{
					IP:        engine.IP,
					Name:      engine.Name,
					Container: container.Config.Container,
				})
			}
		}
	}

	return &Hook{
		Timestamp:      timeStamp,
		Event:          hookEvent.String(),
		MetaBase:       metaData.MetaBase,
		HookContainers: hookContainers,
	}
}

// Post is exported
func (hook *Hook) Post(pool *sync.Pool) {

	metaWebHooks := hook.MetaBase.WebHooks
	for _, metaWebHook := range metaWebHooks {
		buf := pool.Get().(*bytes.Buffer)
		buf.Reset()
		if err := json.NewEncoder(buf).Encode(hook); err != nil {
			pool.Put(buf)
			logger.ERROR("[#cluster#] webhook %s post %s to %s, encode error:%s", hook.Event, hook.MetaBase.MetaID, metaWebHook.URL, err.Error())
			continue
		}
		header := map[string][]string{"Content-Type": []string{"application/json"}}
		if metaWebHook.SecretToken != "" {
			header["X-Humpback-Token"] = []string{metaWebHook.SecretToken}
		}
		resp, err := http.NewWithTimeout(requestTimeout).Post(metaWebHook.URL, nil, buf, header)
		if err != nil {
			pool.Put(buf)
			logger.ERROR("[#cluster#] webhook %s post %s to %s, http error:%s", hook.Event, hook.MetaBase.MetaID, metaWebHook.URL, err.Error())
			continue
		}
		resp.Close()
		pool.Put(buf)
		logger.INFO("[#cluster#] webhook %s post %s to %s, http code %d", hook.Event, hook.MetaBase.MetaID, metaWebHook.URL, resp.StatusCode())
	}
}

// HooksProcessor is exported
type HooksProcessor struct {
	Cluster *Cluster
	pool    *sync.Pool
}

// NewHooksProcessor is exported
func NewHooksProcessor() *HooksProcessor {

	return &HooksProcessor{
		pool: &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 64<<10))
			},
		},
	}
}

// SetCluster is exported
func (processor *HooksProcessor) SetCluster(cluster *Cluster) {

	processor.Cluster = cluster
}

// Hook is exported
func (processor *HooksProcessor) Hook(metaData *MetaData, hookEvent HookEvent) {

	if metaData != nil && len(metaData.WebHooks) > 0 {
		timeStamp := time.Now().UnixNano()
		hook := NewHook(processor.Cluster, metaData, timeStamp, hookEvent)
		go hook.Post(processor.pool)
	}
}
