package types

// WebHook is exported
type WebHook struct {
	URL         string `json:"Url"`
	SecretToken string `json:"SecretToken"`
}

// WebHooks is exported
type WebHooks []WebHook
