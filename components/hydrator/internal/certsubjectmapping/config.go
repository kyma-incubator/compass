package certsubjectmapping

import "time"

type Config struct {
	resyncInterval      time.Duration
	environmentMappings string `envconfig:"default=[],APP_SUBJECT_CONSUMER_MAPPING_CONFIG"`
}
