package avs

type ModelConfigurator interface {
	ProvideSuffix() string
	ProvideTesterAccessId() int64
	ProvideTags() []*Tag
	ProvideNewOrDefaultServiceName(defaultServiceName string) string
	ProvideCheckType() string
}
