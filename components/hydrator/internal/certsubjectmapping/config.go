package certsubjectmapping

import "time"

type Config struct {
	ResyncInterval      time.Duration `envconfig:"APP_CERT_SUBJECT_MAPPING_RESYNC_INTERVAL"`
	EnvironmentMappings string        `envconfig:"default=[],APP_SUBJECT_CONSUMER_MAPPING_CONFIG"`
}
