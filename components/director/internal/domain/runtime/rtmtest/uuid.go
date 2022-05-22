package rtmtest

import "github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"

// UnusedUUIDService returns a mock uid service that does not expect to get called
func UnusedUUIDService() *automock.UidService {
	return &automock.UidService{}
}
