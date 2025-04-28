package core

// Agent represents a generic AI-powered entity in Blockchain
type Agent struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Role             string                 `json:"role"`
	ValidatorAddress string                 `json:"validator_address,omitempty"`
	IsValidator      bool                   `json:"is_validator"`
	Metadata         map[string]interface{} `json:"metadata"`
}
