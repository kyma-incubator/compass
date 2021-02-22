package tokens

import "time"

type Config struct {
	Length                int           `envconfig:"default=64"`
	RuntimeExpiration     time.Duration `envconfig:"default=60m"`
	ApplicationExpiration time.Duration `envconfig:"default=5m"`
	CSRExpiration         time.Duration `envconfig:"default=5m"`
}
