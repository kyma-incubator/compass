package credloader

import (
	"errors"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// WaitForCertCache waits for a CertCache to get populated with data
func WaitForCertCache(cache CertCache) error {
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
			return errors.New("timeout waiting for cache to load")
		}
	}
}

// WaitForKeyCache waits for a KeysCache to get populated with data
func WaitForKeyCache(cache KeysCache) error {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(time.Second * 15)
	for {
		select {
		case <-ticker.C:
			if cache.Get() == nil {
				log.D().Info("Waiting for key cache to load, sleeping for 1 second")
			} else {
				return nil
			}
		case <-timeout:
			return errors.New("timeout waiting for cache to load")
		}
	}
}
