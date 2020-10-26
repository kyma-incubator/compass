package scenario_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/99designs/gqlgen/graphql"
	lbl_mock "github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
	pkg_mock "github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"
	pkg_auth_mock "github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasScenario(t *testing.T) {
	t.Run("could not extract consumer information, should return error", func(t *testing.T) {
		// GIVEN
		directive := scenario.NewDirective(nil, nil, nil)
		// WHEN
		res, err := directive.HasScenario(context.TODO(), nil, nil, "", "")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, consumer.NoConsumerError.Error())
		assert.Equal(t, res, nil)
	})

	t.Run("consumer is of type user, should proceed with next resolver", func(t *testing.T) {
		// GIVEN
		directive := scenario.NewDirective(nil, nil, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.User})
		dummyResolver := &dummyResolver{}
		// WHEN
		res, err := directive.HasScenario(ctx, nil, dummyResolver.SuccessResolve, "", "")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, res, mockedNextOutput())
	})

	t.Run("consumer is of type application, should proceed with next resolver", func(t *testing.T) {
		// GIVEN
		directive := scenario.NewDirective(nil, nil, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.Application})
		dummyResolver := &dummyResolver{}
		// WHEN
		res, err := directive.HasScenario(ctx, nil, dummyResolver.SuccessResolve, "", "")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, res, mockedNextOutput())
	})

	t.Run("consumer is of type integration system, should proceed with next resolver", func(t *testing.T) {
		// GIVEN
		directive := scenario.NewDirective(nil, nil, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.IntegrationSystem})
		dummyResolver := &dummyResolver{}
		// WHEN
		res, err := directive.HasScenario(ctx, nil, dummyResolver.SuccessResolve, "", "")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, res, mockedNextOutput())
	})

	t.Run("could not extract tenant from context, should return error", func(t *testing.T) {
		// GIVEN
		directive := scenario.NewDirective(nil, nil, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.Runtime})
		dummyResolver := &dummyResolver{}
		// WHEN
		res, err := directive.HasScenario(ctx, nil, dummyResolver.SuccessResolve, "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewCannotReadTenantError().Error())
		assert.Equal(t, res, nil)
	})

	t.Run("runtime requests non-existent application", func(t *testing.T) {
		// GIVEN
		const (
			idField       = "id"
			tenantID      = "42"
			applicationID = "24"
		)

		lblRepo := &lbl_mock.LabelRepository{}
		defer lblRepo.AssertExpectations(t)

		directive := scenario.NewDirective(lblRepo, nil, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.ResolverContext{
			Object:   "Application",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{idField: applicationID},
			IsMethod: false,
		}
		ctx = graphql.WithResolverContext(ctx, rCtx)

		notFoundErr := apperrors.NewNotFoundError(resource.Label, model.ScenariosKey)
		lblRepo.On("GetByKey", ctx, tenantID, model.ApplicationLabelableObject, applicationID, model.ScenariosKey).Return(nil, notFoundErr)
		// WHEN
		res, err := directive.HasScenario(ctx, nil, nil, scenario.GetApplicationID, idField)
		// THEN
		require.Error(t, err)
		assert.Error(t, err, notFoundErr)
		assert.Equal(t, res, nil)
	})

	t.Run("runtime requests package instance auth creation for non-existent package", func(t *testing.T) {
		// GIVEN
		const (
			packageIDField = "packageID"
			tenantID       = "42"
			packageID      = "24"
		)

		pkgRepo := &pkg_mock.PackageRepository{}
		defer pkgRepo.AssertExpectations(t)

		directive := scenario.NewDirective(nil, pkgRepo, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.ResolverContext{
			Object:   "PackageInstanceAuth",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{packageIDField: packageID},
			IsMethod: false,
		}
		ctx = graphql.WithResolverContext(ctx, rCtx)

		notFoundErr := apperrors.NewNotFoundErrorWithType(resource.Package)
		pkgRepo.On("GetByID", ctx, tenantID, packageID).Return(nil, notFoundErr)
		// WHEN
		res, err := directive.HasScenario(ctx, nil, nil, scenario.GetApplicationIDByPackage, packageIDField)
		// THEN
		require.Error(t, err)
		assert.Error(t, err, notFoundErr)
		assert.Equal(t, res, nil)
	})

	t.Run("runtime requests package instance auth deletion for non-existent system auth ID", func(t *testing.T) {
		// GIVEN
		const (
			pkgAuthIDField = "authID"
			tenantID       = "42"
			authID         = "24"
		)

		pkgAuthRepo := &pkg_auth_mock.Repository{}
		defer pkgAuthRepo.AssertExpectations(t)

		directive := scenario.NewDirective(nil, nil, pkgAuthRepo)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.ResolverContext{
			Object:   "PackageInstanceAuth",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{pkgAuthIDField: authID},
			IsMethod: false,
		}
		ctx = graphql.WithResolverContext(ctx, rCtx)

		notFoundErr := apperrors.NewNotFoundErrorWithType(resource.PackageInstanceAuth)
		pkgAuthRepo.On("GetByID", ctx, tenantID, authID).Return(nil, notFoundErr)
		// WHEN
		res, err := directive.HasScenario(ctx, nil, nil, scenario.GetApplicationIDByPackageInstanceAuth, pkgAuthIDField)
		// THEN
		require.Error(t, err)
		assert.Error(t, err, notFoundErr)
		assert.Equal(t, res, nil)
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

type dummyResolver struct {
	called bool
}

func (d *dummyResolver) SuccessResolve(_ context.Context) (res interface{}, err error) {
	d.called = true
	return mockedNextOutput(), nil
}

func mockedNextOutput() string {
	return "nextOutput"
}
