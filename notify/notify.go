package notify

import (
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

//notify template string
var templateBody string

//NotifySender is exported
type NotifySender struct {
	sync.RWMutex
	SiteURL   string
	initWatch bool
	endPoints []IEndPoint
	events    map[string]*Event
}

//NewNotifySender is exported
func NewNotifySender(siteurl string, endPoints []EndPoint) *NotifySender {

	sender := &NotifySender{
		SiteURL:   siteurl,
		initWatch: true,
		endPoints: []IEndPoint{},
		events:    make(map[string]*Event),
	}

	if buf, err := ioutil.ReadFile("./notify/template.html"); err == nil {
		templateBody = string(buf)
	}

	factory := &NotifyEndPointFactory{}
	sender.Lock()
	for _, endPoint := range endPoints {
		switch strings.ToUpper(endPoint.Name) {
		case "API":
			apiEndPoint := factory.CreateAPIEndPoint(endPoint)
			sender.endPoints = append(sender.endPoints, apiEndPoint)
		case "SMTP":
			smtpEndPoint := factory.CreateSMTPEndPoint(endPoint)
			sender.endPoints = append(sender.endPoints, smtpEndPoint)
		}
	}
	sender.Unlock()

	go func() {
		time.Sleep(30 * time.Second)
		sender.initWatch = false
	}()
	return sender
}

//AddGroupEnginesWatchEvent is exported
func (sender *NotifySender) AddGroupEnginesWatchEvent(description string, watchGroup *WatchGroup) {

	event := NewEvent(GroupEnginesWatchEvent, description, nil, watchGroup.ContactInfo, sender.SiteURL, sender.endPoints)
	event.data["WatchGroup"] = watchGroup
	sender.Lock()
	sender.events[event.ID] = event
	sender.Unlock()
	go sender.dispatchEvents()
}

//AddGroupMetaContainersEvent is exported
func (sender *NotifySender) AddGroupMetaContainersEvent(description string, err error, groupMeta *GroupMeta) {

	event := NewEvent(GroupMetaContainersEvent, description, err, groupMeta.ContactInfo, sender.SiteURL, sender.endPoints)
	event.data["GroupMeta"] = groupMeta
	sender.Lock()
	sender.events[event.ID] = event
	sender.Unlock()
	go sender.dispatchEvents()
}

//dispatchEvents is exported
//dispatch all events.
func (sender *NotifySender) dispatchEvents() {

	sender.Lock()
	for {
		if len(sender.events) == 0 {
			break
		}
		if !sender.initWatch {
			wgroup := sync.WaitGroup{}
			for _, event := range sender.events {
				wgroup.Add(1)
				go func(e *Event) {
					e.dispatch(templateBody)
					wgroup.Done()
				}(event)
			}
			wgroup.Wait()
		}
		for _, event := range sender.events {
			delete(sender.events, event.ID)
		}
	}
	sender.Unlock()
}
