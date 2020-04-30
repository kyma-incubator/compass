package statusupdate

import "time"

func (r *repository) SetTimestampGen(timestampGen func() time.Time) {
	r.timestampGen = timestampGen
}
