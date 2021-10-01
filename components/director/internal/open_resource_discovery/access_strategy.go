package ord

// AccessStrategy is an ORD object
type AccessStrategy struct {
	Type              AccessStrategyType `json:"type"`
	CustomType        AccessStrategyType `json:"customType"`
	CustomDescription string             `json:"customDescription"`
}

// AccessStrategyType represents the possible type of the AccessStrategy
type AccessStrategyType string

// IsSupported checks if the given AccessStrategy is supported by CMP
func (a AccessStrategyType) IsSupported() bool {
	return supportedAccessStrategies[a]
}

const (
	// OpenAccessStrategy is one of the available AccessStrategy types
	OpenAccessStrategy AccessStrategyType = "open"
	// CustomAccessStrategy is one of the available AccessStrategy types
	CustomAccessStrategy AccessStrategyType = "custom"
)

var supportedAccessStrategies = map[AccessStrategyType]bool{
	OpenAccessStrategy: true,
}

// AccessStrategies is a slice of AccessStrategy objects
type AccessStrategies []AccessStrategy

// GetSupported returns the first AccessStrategy in the slice that is supported by CMP
func (as AccessStrategies) GetSupported() (AccessStrategyType, bool) {
	for _, v := range as {
		if supportedAccessStrategies[v.Type] {
			return v.Type, true
		}
		if v.Type == CustomAccessStrategy && supportedAccessStrategies[v.CustomType] {
			return v.CustomType, true
		}
	}
	return "", false
}
