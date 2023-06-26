package bundletest

import "github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"

// UnusedBundleRepository returns BundleRepository mock that does not expect to be invoked
func UnusedBundleRepository() *automock.BundleRepository {
	return &automock.BundleRepository{}
}

// UnusedBundleInstanceAuthService returns BundleInstanceAuthService mock that does not expect to be invoked
func UnusedBundleInstanceAuthService() *automock.BundleInstanceAuthService {
	return &automock.BundleInstanceAuthService{}
}
