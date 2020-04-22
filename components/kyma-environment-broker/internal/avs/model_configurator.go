package avs

type ModelConfigurator interface {
	ProvideSuffix() string
	ProvideTesterAccessId() int64
	ProvideCheckType() string
}
