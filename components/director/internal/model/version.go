package model

// Version missing godoc
type Version struct {
	// for example 4.6
	Value      string
	Deprecated *bool
	// for example 4.5
	DeprecatedSince *string
	// if true, will be removed in the next version
	ForRemoval *bool
}

// VersionInput missing godoc
type VersionInput struct {
	Value           string  `json:"version"`
	Deprecated      *bool   `json:",omitempty"`
	DeprecatedSince *string `json:",omitempty"`
	ForRemoval      *bool   `json:",omitempty"`
}

// ToVersion missing godoc
func (v *VersionInput) ToVersion() *Version {
	if v == nil {
		return nil
	}

	return &Version{
		Value:           v.Value,
		Deprecated:      v.Deprecated,
		DeprecatedSince: v.DeprecatedSince,
		ForRemoval:      v.ForRemoval,
	}
}
