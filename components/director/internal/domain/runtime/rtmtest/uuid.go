package rtmtest

import "github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"

func UnusedUUIDService() func() *automock.UidService {
	return func() *automock.UidService {
		return &automock.UidService{}
	}
}
