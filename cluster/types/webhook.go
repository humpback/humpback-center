package types

// WebHook is exported
type WebHook struct {
	URL         string `json:"Url"`
	SecretToken string `json:"SecretToken"`
}

// WebHooks is exported
type WebHooks []WebHook

// BindConfig is exported
type BindConfig struct {
	Category   string   `json:"Category"`
	OwnerToken string   `json:"OwnerToken"`
	Sandbox    bool     `json:"Sandbox"`
	Location   string   `json:"Location"`
	Protocol   string   `json:"Protocol"`
	Port       int      `json:"Port"`
	Health     string   `json:"Health"`
	APIIDs     []string `json:"APIIDs"`
}
