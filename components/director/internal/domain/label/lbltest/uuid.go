package lbltest

import "github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"

// UnusedUUIDService returns a mock uid service that does not expect to get called
func UnusedUUIDService() func() *automock.UIDService {
	return func() *automock.UIDService {
		return &automock.UIDService{}
	}
}
