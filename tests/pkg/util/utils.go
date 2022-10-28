package util

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

type ApplicationType string

const (
	AuthorizationHeader = "Authorization"
	ContentTypeHeader   = "Content-Type"

	ContentTypeApplicationJSON       = "application/json"
	ContentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"

	ApplicationTypeC4C             ApplicationType = "SAP Cloud for Customer"
	ApplicationTypeS4HANAOnPremise ApplicationType = "SAP S/4HANA On-Premise"
)

func WaitForCache(cache certloader.Cache) error {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(time.Second * 15)
	for {
		select {
		case <-ticker.C:
			if cache.Get() == nil {
				log.D().Info("Waiting for certificate cache to load, sleeping for 1 second")
			} else {
				return nil
			}
		case <-timeout:
			return errors.New("Timeout waiting for cache to load")
		}
	}
}
