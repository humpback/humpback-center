package types

//CreateOption is exported
//cluster create containers option values.
//`IsReCreate` indicates whether to re-create an existing containers each time it is built.
//`ForceRemove` is an attached property. When `IsReCreate` is true, it means to force delete or directly upgrade an existing containers.
//`IsRemoveDelay` delay (8 minutes) remove unused containers for service debounce.
//`IsRecovery` service containers recovery check enable.
type CreateOption struct {
	IsReCreate    bool `json:"IsReCreate"`
	ForceRemove   bool `json:"ForceRemove"`
	IsRemoveDelay bool `json:"IsRemoveDelay"`
	IsRecovery    bool `json:"IsRecovery"`
}

//UpdateOption is exported
type UpdateOption struct {
	IsRemoveDelay bool `json:"IsRemoveDelay"`
	IsRecovery    bool `json:"IsRecovery"`
}
