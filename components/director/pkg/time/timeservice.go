package time

import "time"

//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore
type Service interface {
	Now() time.Time
}

type service struct {
}

func (ts *service) Now() time.Time {
	return time.Now()
}

func NewService() Service {
	return &service{}
}
