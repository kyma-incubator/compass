package packageinstanceauth

import "time"

func (s *service) SetTimestampGen(timestampGen func() time.Time) {
	s.timestampGen = timestampGen
}
