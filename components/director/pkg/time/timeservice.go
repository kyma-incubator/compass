package time

import "time"

// Service missing godoc
//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type Service interface {
	Now() time.Time
}

type service struct {
}

// Now missing godoc
func (ts *service) Now() time.Time {
	return time.Now()
}

// NewService missing godoc
func NewService() Service {
	return &service{}
}
