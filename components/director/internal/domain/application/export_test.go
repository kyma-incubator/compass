package application

import "time"

func (r *Resolver) SetConverter(converter ApplicationConverter) {
	r.appConverter = converter
}

func (s *service) SetTimestampGen(timestampGen func() time.Time) {
	s.timestampGen = timestampGen
}
