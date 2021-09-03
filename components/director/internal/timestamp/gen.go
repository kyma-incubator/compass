package timestamp

import "time"

type Generator func() time.Time

func DefaultGenerator() time.Time {
	return time.Now()
}
