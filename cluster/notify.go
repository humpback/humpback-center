package cluster

import "humpback-center/notify"

//WatchEngines is exported
type WatchEngines []*Engine

//NewWatchEngine is exported
func NewWatchEngine(ip string, name string, state engineState) *Engine {

	return &Engine{
		IP:    ip,
		Name:  name,
		state: state,
	}
}

//NotifyGroupEnginesWatchEvent is exported
func (cluster *Cluster) NotifyGroupEnginesWatchEvent(description string, watchEngines WatchEngines) {

	watchGroups := make(notify.WatchGroups)
	for _, engine := range watchEngines {
		e := &notify.Engine{
			IP:    engine.IP,
			Name:  engine.Name,
			State: stateText[engine.state],
		}
		groups := cluster.GetEngineGroups(engine)
		for _, group := range groups {
			if watchGroup, ret := watchGroups[group.ID]; !ret {
				watchGroup = &notify.WatchGroup{
					GroupID:     group.ID,
					GroupName:   group.Name,
					Location:    group.Location,
					ContactInfo: group.ContactInfo,
					Engines:     []*notify.Engine{e},
				}
				watchGroups[group.ID] = watchGroup
			} else {
				watchGroup.Engines = append(watchGroup.Engines, e)
			}
		}
	}
	for _, watchGroup := range watchGroups {
		cluster.NotifySender.AddGroupEnginesWatchEvent(description, watchGroup)
	}
}

//NotifyGroupMetaContainersEvent is exported
func (cluster *Cluster) NotifyGroupMetaContainersEvent(description string, exception error, metaid string) {

	metaData, engines, err := cluster.GetMetaDataEngines(metaid)
	if err != nil {
		return
	}

	group := cluster.GetGroup(metaData.GroupID)
	if group == nil {
		return
	}

	containers := []notify.Container{}
	for _, baseConfig := range metaData.BaseConfigs {
		for _, engine := range engines {
			if engine.IsHealthy() && engine.HasContainer(baseConfig.ID) {
				state := "Unkonw"
				if c := engine.Container(baseConfig.ID); c != nil {
					state = StateString(c.Info.State)
				}
				containers = append(containers, notify.Container{
					ID:     baseConfig.ID[:12],
					Name:   baseConfig.Name,
					Server: engine.IP,
					State:  state,
				})
			}
		}
	}

	groupMeta := &notify.GroupMeta{
		MetaID:      metaData.MetaID,
		MetaName:    metaData.Config.Name,
		Location:    group.Location,
		GroupID:     group.ID,
		GroupName:   group.Name,
		Instances:   metaData.Instances,
		Image:       metaData.Config.Image,
		ContactInfo: group.ContactInfo,
		Containers:  containers,
	}
	cluster.NotifySender.AddGroupMetaContainersEvent(description, exception, groupMeta)
}
