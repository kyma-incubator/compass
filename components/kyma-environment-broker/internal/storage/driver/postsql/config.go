package postsql

import "time"

const (
	defaultRetryTimeout  = time.Second * 3
	defaultRetryInterval = time.Millisecond * 500
)
