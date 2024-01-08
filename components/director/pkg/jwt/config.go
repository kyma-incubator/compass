package jwt

import (
	"time"
)

// Config is JWT configuration
type Config struct {
	ExpireAfter time.Duration `envconfig:"APP_JWT_EXPIRE_AFTER"`
}
