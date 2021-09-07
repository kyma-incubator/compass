package timestamp

import "time"

// Generator missing godoc
type Generator func() time.Time

// DefaultGenerator missing godoc
func DefaultGenerator() time.Time {
	return time.Now()
}
