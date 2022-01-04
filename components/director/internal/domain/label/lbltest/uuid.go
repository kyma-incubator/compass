package lbltest

import "github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"

func UnusedUUIDService() func() *automock.UIDService {
	return func() *automock.UIDService {
		return &automock.UIDService{}
	}
}
