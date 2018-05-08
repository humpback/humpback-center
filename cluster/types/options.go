package types

//CreateOption is exported
//cluster create containers option values.
//`IsReCreate` indicates whether to re-create an existing containers each time it is built.
//`ForceRemove` is an attached property. When `IsReCreate` is true, it means to force delete or directly upgrade an existing containers.
type CreateOption struct {
	IsReCreate  bool `json:"IsReCreate"`
	ForceRemove bool `json:"ForceRemove"`
}
