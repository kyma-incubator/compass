package model

type Version struct {
	// for example 4.6
	Value      string
	Deprecated *bool
	// for example 4.5
	DeprecatedSince *string
	// if true, will be removed in the next version
	ForRemoval *bool
}

type VersionInput struct {
	Value           string
	Deprecated      *bool
	DeprecatedSince *string
	ForRemoval      *bool
}

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
