package frmtest

import "github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"

func UnusedUUIDService() func() *automock.UidService {
	return func() *automock.UidService {
		return &automock.UidService{}
	}
}
