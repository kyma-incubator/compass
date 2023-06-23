package bundletest

import "github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"

func UnusedBundleRepository() *automock.BundleRepository {
	return &automock.BundleRepository{}
}

func UnusedBundleInstanceAuthService() *automock.BundleInstanceAuthService {
	return &automock.BundleInstanceAuthService{}
}
