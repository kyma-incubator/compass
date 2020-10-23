package scenario_test

import (
	"testing"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	packageMock "github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth/automock"
)

func TestHasScenario(t *testing.T) {
	t.Run("consumer is of type user", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})

	t.Run("consumer is of type application", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})

	t.Run("consumer is of type integration system", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})

	t.Run("runtime requests non-existent application", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})

	t.Run("runtime requests package instance auth creation for non-existent package", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})

	t.Run("runtime requests package instance auth deletion for non-existent system auth ID", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})

	t.Run("runtime is in formation with application in application query", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})

	t.Run("runtime is NOT in formation with application in application query", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})

	t.Run("runtime is in formation with package in request package instance auth flow ", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})
	t.Run("runtime is NOT in formation with package in request package instance auth flow ", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})

	t.Run("runtime is in formation with package in delete package instance auth flow", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})
	t.Run("runtime is NOT in formation with package in delete package instance auth flow", func(t *testing.T) {
		// GIVEN
		// WHEN
		// THEN
	})

}

func fakeRepoBuilder() (mp_package.PackageRepository, packageinstanceauth.Repository) {
	return &packageMock.PackageRepository{}, &automock.Repository{}
}
