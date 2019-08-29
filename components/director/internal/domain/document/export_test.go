package document

import "time"

func (r *Resolver) SetConverter(converter DocumentConverter) {
	r.converter = converter
}

func (s *service) SetTimestampGen(timestampGen func() time.Time) {
	s.timestampGen = timestampGen
}
