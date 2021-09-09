package ord

// AccessStrategy missing godoc
type AccessStrategy struct {
	Type              AccessStrategyType `json:"type"`
	CustomType        AccessStrategyType `json:"customType"`
	CustomDescription string             `json:"customDescription"`
}

// AccessStrategyType missing godoc
type AccessStrategyType string

// IsSupported missing godoc
func (a AccessStrategyType) IsSupported() bool {
	return supportedAccessStrategies[a]
}

const (
	// OpenAccessStrategy missing godoc
	OpenAccessStrategy AccessStrategyType = "open"
	// CustomAccessStrategy missing godoc
	CustomAccessStrategy AccessStrategyType = "custom"
)

var supportedAccessStrategies = map[AccessStrategyType]bool{
	OpenAccessStrategy: true,
}

// AccessStrategies missing godoc
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
