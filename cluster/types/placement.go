package types

// Spread is exported
type Spread struct {
	SpreadDescriptor string `json:"SpreadDescriptor"`
}

// Preference is exported
type Preference struct {
	Spread `json:"Spread"`
}

// Platform is exported
type Platform struct {
	Architecture string `json:"Architecture"`
	OS           string `json:"OS"`
}

// Placement is exported
// Cluster services placement constraints
type Placement struct {
	Constraints []string     `json:"Constraints"`
	Preferences []Preference `json:"Preferences"`
	Platforms   []Platform   `json:"Platforms"`
}
