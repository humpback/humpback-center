package cluster

import "github.com/humpback/common/models"
import "github.com/humpback/gounits/container"
import "github.com/humpback/gounits/httpx"
import "github.com/humpback/gounits/logger"

import (
	"context"
	"net"
	"net/http"
	"strings"
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
	client *httpx.HttpClient
}

// Submit is exported
func (hook *Hook) Submit() {

	webHooks := hook.MetaBase.WebHooks
	for _, webHook := range webHooks {
		headers := map[string][]string{}
		secretToken := strings.TrimSpace(webHook.SecretToken)
		if secretToken != "" {
			headers["X-Humpback-Token"] = []string{secretToken}
		}
		hookURL := strings.TrimSpace(webHook.URL)
		respWebHook, err := hook.client.PostJSON(context.Background(), hookURL, nil, hook, headers)
		if err != nil {
			logger.ERROR("[#cluster#] webhook %s post %s to %s, http error:%s", hook.Event, hook.MetaBase.MetaID, hookURL, err.Error())
			continue
		}
		if respWebHook.StatusCode() >= http.StatusBadRequest {
			logger.ERROR("[#cluster#] webhook %s post %s to %s, http code %d", hook.Event, hook.MetaBase.MetaID, hookURL, respWebHook.StatusCode())
		}
		respWebHook.Close()
	}
}

// HooksProcessor is exported
type HooksProcessor struct {
	bStart     bool
	client     *httpx.HttpClient
	hooksQueue *container.SyncQueue
	stopCh     chan struct{}
}

// NewHooksProcessor is exported
func NewHooksProcessor() *HooksProcessor {

	client := httpx.NewClient().
		SetTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   45 * time.Second,
				KeepAlive: 90 * time.Second,
			}).DialContext,
			DisableKeepAlives:     false,
			MaxIdleConns:          50,
			MaxIdleConnsPerHost:   65,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout,
			ExpectContinueTimeout: http.DefaultTransport.(*http.Transport).ExpectContinueTimeout,
		})

	return &HooksProcessor{
		bStart:     false,
		client:     client,
		hooksQueue: container.NewSyncQueue(),
	}
}

// SubmitHook is exported
func (processor *HooksProcessor) SubmitHook(metaBase MetaBase, hookContainers HookContainers, hookEvent HookEvent) {

	hook := &Hook{
		client:         processor.client,
		Timestamp:      time.Now().UnixNano(),
		Event:          hookEvent.String(),
		MetaBase:       metaBase,
		HookContainers: hookContainers,
	}
	processor.hooksQueue.Push(hook)
}

func (processor *HooksProcessor) Start() {

	if !processor.bStart {
		processor.bStart = true
		go processor.eventPopLoop()
	}
}

func (processor *HooksProcessor) Close() {

	if processor.bStart {
		processor.hooksQueue.Close()
		processor.bStart = false
	}
}

func (processor *HooksProcessor) eventPopLoop() {

	for processor.bStart {
		value := processor.hooksQueue.Pop()
		if value != nil {
			hook := value.(*Hook)
			go hook.Submit()
		}
	}
}
