package ptr

import "time"

func Bool(in bool) *bool {
	return &in
}

func String(str string) *string {
	return &str
}

func Integer(in int) *int {
	return &in
}

func Time(in time.Time) *time.Time {
	return &in
}
