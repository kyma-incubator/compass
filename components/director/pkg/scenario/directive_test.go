package scenario_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/99designs/gqlgen/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
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
			idField           = "id"
			resoverContextKey = "resolver_context"
		)
		notFoundErr := apperrors.NewNotFoundError(resource.Label, model.ScenariosKey)

		labelRepo := &automock.LabelRepository{}
		defer labelRepo.AssertExpectations(t)
		labelRepo.On("GetByKey").Return(nil, notFoundErr)

		directive := scenario.NewDirective(labelRepo, nil, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: "42"})
		rCtx := &graphql.ResolverContext{
			Object:   "Application",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{idField: "42"},
			IsMethod: false,
		}
		ctx = context.WithValue(ctx, resoverContextKey, rCtx)
		dummyResolver := &dummyResolver{}
		// WHEN
		res, err := directive.HasScenario(ctx, nil, dummyResolver.SuccessResolve, scenario.GetApplicationID, idField)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), notFoundErr)
		assert.Equal(t, res, nil)
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

type dummyResolver struct {
	called bool
}

func (d *dummyResolver) SuccessResolve(ctx context.Context) (res interface{}, err error) {
	d.called = true
	return mockedNextOutput(), nil
}

func mockedNextOutput() string {
	return "nextOutput"
}
