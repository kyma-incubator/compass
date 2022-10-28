package formationmapping

// Config holds the configuration available for the formation mapping
type Config struct {
	AsyncAPIPathPrefix string `envconfig:"APP_FORMATION_MAPPING_API_PATH_PREFIX"`
	AsyncAPIEndpoint   string `envconfig:"APP_FORMATION_MAPPING_API_ENDPOINT"`
}
