package timestamp

import "time"

type Generator func() time.Time

func DefaultGenerator() func() time.Time {
	return func() time.Time {
		return time.Now()
	}
}
