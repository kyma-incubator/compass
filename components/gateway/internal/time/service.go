package time

import "time"

type TimeService struct {
}

func (ts *TimeService) Now() time.Time {
	return time.Now().UTC()
}
