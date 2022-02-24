package scenario_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/99designs/gqlgen/graphql"
	bndl_mock "github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"
	bndl_auth_mock "github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth/automock"
	lbl_mock "github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/scenario"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasScenario(t *testing.T) {
	t.Run("could not extract consumer information, should return error", func(t *testing.T) {
		// GIVEN
		directive := scenario.NewDirective(nil, nil, nil, nil)
		// WHEN
		res, err := directive.HasScenario(context.TODO(), nil, nil, "", "")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, consumer.NoConsumerError.Error())
		assert.Equal(t, res, nil)
	})

	t.Run("consumer is of type user, should proceed with next resolver", func(t *testing.T) {
		// GIVEN
		directive := scenario.NewDirective(nil, nil, nil, nil)
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
		directive := scenario.NewDirective(nil, nil, nil, nil)
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
		directive := scenario.NewDirective(nil, nil, nil, nil)
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
		directive := scenario.NewDirective(nil, nil, nil, nil)
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
			runtimeID     = "23"
		)

		lblRepo := &lbl_mock.LabelRepository{}
		defer lblRepo.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := scenario.NewDirective(mockedTransactioner, lblRepo, nil, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.Runtime, ConsumerID: runtimeID})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.FieldContext{
			Object:   "Application",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{idField: applicationID},
			IsMethod: false,
		}
		ctx = graphql.WithFieldContext(ctx, rCtx)
		ctxWithTx := persistence.SaveToContext(ctx, mockedTx)

		notFoundErr := apperrors.NewNotFoundError(resource.Label, model.ScenariosKey)
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.ApplicationLabelableObject, applicationID, model.ScenariosKey).Return(nil, notFoundErr)
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(nil, notFoundErr)
		// WHEN
		res, err := directive.HasScenario(ctx, nil, nil, scenario.GetApplicationID, idField)
		// THEN
		require.Error(t, err)
		assert.Error(t, err, notFoundErr)
		assert.Equal(t, res, nil)
	})

	t.Run("runtime requests bundle instance auth creation for non-existent bundle", func(t *testing.T) {
		// GIVEN
		const (
			bundleIDField = "bundleID"
			tenantID      = "42"
			bundleID      = "24"
		)

		bndlRepo := &bndl_mock.BundleRepository{}
		defer bndlRepo.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := scenario.NewDirective(mockedTransactioner, nil, bndlRepo, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.FieldContext{
			Object:   "BundleInstanceAuth",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{bundleIDField: bundleID},
			IsMethod: false,
		}
		ctx = graphql.WithFieldContext(ctx, rCtx)
		ctxWithTx := persistence.SaveToContext(ctx, mockedTx)

		notFoundErr := apperrors.NewNotFoundErrorWithType(resource.Bundle)
		bndlRepo.On("GetByID", ctxWithTx, tenantID, bundleID).Return(nil, notFoundErr)
		// WHEN
		res, err := directive.HasScenario(ctx, nil, nil, scenario.GetApplicationIDByBundle, bundleIDField)
		// THEN
		require.Error(t, err)
		assert.Error(t, err, notFoundErr)
		assert.Equal(t, res, nil)
	})

	t.Run("runtime requests bundle instance auth deletion for non-existent system auth ID", func(t *testing.T) {
		// GIVEN
		const (
			bndlAuthIDField = "authID"
			tenantID        = "42"
			bndlAuthID      = "24"
		)

		bndlAuthRepo := &bndl_auth_mock.Repository{}
		defer bndlAuthRepo.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := scenario.NewDirective(mockedTransactioner, nil, nil, bndlAuthRepo)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.FieldContext{
			Object:   "BundleInstanceAuth",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{bndlAuthIDField: bndlAuthID},
			IsMethod: false,
		}
		ctx = graphql.WithFieldContext(ctx, rCtx)
		ctxWithTx := persistence.SaveToContext(ctx, mockedTx)

		notFoundErr := apperrors.NewNotFoundErrorWithType(resource.BundleInstanceAuth)
		bndlAuthRepo.On("GetByID", ctxWithTx, tenantID, bndlAuthID).Return(nil, notFoundErr)
		// WHEN
		res, err := directive.HasScenario(ctx, nil, nil, scenario.GetApplicationIDByBundleInstanceAuth, bndlAuthIDField)
		// THEN
		require.Error(t, err)
		assert.Error(t, err, notFoundErr)
		assert.Equal(t, res, nil)
	})

	t.Run("runtime is in formation with application in application query", func(t *testing.T) {
		// GIVEN
		const (
			idField       = "id"
			tenantID      = "42"
			runtimeID     = "23"
			applicationID = "24"
		)

		lblRepo := &lbl_mock.LabelRepository{}
		defer lblRepo.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := scenario.NewDirective(mockedTransactioner, lblRepo, nil, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerID: runtimeID, ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.FieldContext{
			Object:   "Application",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{idField: applicationID},
			IsMethod: false,
		}
		ctx = graphql.WithFieldContext(ctx, rCtx)
		ctxWithTx := persistence.SaveToContext(ctx, mockedTx)

		mockedLabel := &model.Label{Value: []interface{}{"DEFAULT"}}
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.ApplicationLabelableObject, applicationID, model.ScenariosKey).Return(mockedLabel, nil)
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(mockedLabel, nil)

		dummyResolver := &dummyResolver{}
		// WHEN
		res, err := directive.HasScenario(ctx, nil, dummyResolver.SuccessResolve, scenario.GetApplicationID, idField)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, res, mockedNextOutput())
	})

	t.Run("runtime is NOT in formation with application in application query", func(t *testing.T) {
		// GIVEN
		const (
			idField       = "id"
			tenantID      = "42"
			runtimeID     = "23"
			applicationID = "24"
		)

		lblRepo := &lbl_mock.LabelRepository{}
		defer lblRepo.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := scenario.NewDirective(mockedTransactioner, lblRepo, nil, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerID: runtimeID, ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.FieldContext{
			Object:   "Application",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{idField: applicationID},
			IsMethod: false,
		}
		ctx = graphql.WithFieldContext(ctx, rCtx)
		ctxWithTx := persistence.SaveToContext(ctx, mockedTx)

		mockedAppLabel := &model.Label{Value: []interface{}{"DEFAULT"}}
		mockedRuntimeLabel := &model.Label{Value: []interface{}{"TEST"}}
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.ApplicationLabelableObject, applicationID, model.ScenariosKey).Return(mockedAppLabel, nil)
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(mockedRuntimeLabel, nil)
		// WHEN
		res, err := directive.HasScenario(ctx, nil, nil, scenario.GetApplicationID, idField)
		// THEN
		require.Error(t, err)
		assert.Error(t, err, scenario.ErrMissingScenario)
		assert.Equal(t, res, nil)
	})

	t.Run("runtime is in formation with owning application in request bundle instance auth flow ", func(t *testing.T) {
		// GIVEN
		const (
			bundleIDField = "bundleID"
			tenantID      = "42"
			bundleID      = "24"
			runtimeID     = "23"
			applicationID = "22"
		)

		bndlRepo := &bndl_mock.BundleRepository{}
		defer bndlRepo.AssertExpectations(t)

		lblRepo := &lbl_mock.LabelRepository{}
		defer lblRepo.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := scenario.NewDirective(mockedTransactioner, lblRepo, bndlRepo, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerID: runtimeID, ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.FieldContext{
			Object:   "BundleInstanceAuth",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{bundleIDField: bundleID},
			IsMethod: false,
		}
		ctx = graphql.WithFieldContext(ctx, rCtx)
		ctxWithTx := persistence.SaveToContext(ctx, mockedTx)

		mockedBndl := &model.Bundle{ApplicationID: applicationID}
		bndlRepo.On("GetByID", ctxWithTx, tenantID, bundleID).Return(mockedBndl, nil)

		mockedLabel := &model.Label{Value: []interface{}{"DEFAULT"}}
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.ApplicationLabelableObject, mockedBndl.ApplicationID, model.ScenariosKey).Return(mockedLabel, nil)
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(mockedLabel, nil)

		dummyResolver := &dummyResolver{}
		// WHEN
		res, err := directive.HasScenario(ctx, nil, dummyResolver.SuccessResolve, scenario.GetApplicationIDByBundle, bundleIDField)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, res, mockedNextOutput())
	})
	t.Run("runtime is NOT in formation with owning application in request bundle instance auth flow ", func(t *testing.T) {
		// GIVEN
		const (
			bundleIDField = "bundleID"
			tenantID      = "42"
			bundleID      = "24"
			runtimeID     = "23"
			applicationID = "22"
		)

		bndlRepo := &bndl_mock.BundleRepository{}
		defer bndlRepo.AssertExpectations(t)

		lblRepo := &lbl_mock.LabelRepository{}
		defer lblRepo.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := scenario.NewDirective(mockedTransactioner, lblRepo, bndlRepo, nil)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerID: runtimeID, ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.FieldContext{
			Object:   "BundleInstanceAuth",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{bundleIDField: bundleID},
			IsMethod: false,
		}
		ctx = graphql.WithFieldContext(ctx, rCtx)
		ctxWithTx := persistence.SaveToContext(ctx, mockedTx)

		mockedBndl := &model.Bundle{ApplicationID: applicationID}
		bndlRepo.On("GetByID", ctxWithTx, tenantID, bundleID).Return(mockedBndl, nil)

		mockedAppLabel := &model.Label{Value: []interface{}{"DEFAULT"}}
		mockedRuntimeLabel := &model.Label{Value: []interface{}{"TEST"}}
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.ApplicationLabelableObject, applicationID, model.ScenariosKey).Return(mockedAppLabel, nil)
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(mockedRuntimeLabel, nil)
		// WHEN
		res, err := directive.HasScenario(ctx, nil, nil, scenario.GetApplicationIDByBundle, bundleIDField)
		// THEN
		require.Error(t, err)
		assert.Error(t, err, scenario.ErrMissingScenario)
		assert.Equal(t, res, nil)
	})

	t.Run("runtime is in formation with owning application in delete bundle instance auth flow", func(t *testing.T) {
		// GIVEN
		const (
			bndlAuthIDField = "authID"
			tenantID        = "42"
			bndlAuthID      = "24"
			runtimeID       = "23"
			applicationID   = "22"
			bundleID        = "21"
		)

		bndlAuthRepo := &bndl_auth_mock.Repository{}
		defer bndlAuthRepo.AssertExpectations(t)

		bndlRepo := &bndl_mock.BundleRepository{}
		defer bndlRepo.AssertExpectations(t)

		lblRepo := &lbl_mock.LabelRepository{}
		defer lblRepo.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := scenario.NewDirective(mockedTransactioner, lblRepo, bndlRepo, bndlAuthRepo)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerID: runtimeID, ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.FieldContext{
			Object:   "BundleInstanceAuth",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{bndlAuthIDField: bndlAuthID},
			IsMethod: false,
		}
		ctx = graphql.WithFieldContext(ctx, rCtx)
		ctxWithTx := persistence.SaveToContext(ctx, mockedTx)

		mockedBndlAuth := &model.BundleInstanceAuth{BundleID: bundleID}
		bndlAuthRepo.On("GetByID", ctxWithTx, tenantID, bndlAuthID).Return(mockedBndlAuth, nil)

		mockedBndl := &model.Bundle{ApplicationID: applicationID}
		bndlRepo.On("GetByID", ctxWithTx, tenantID, mockedBndlAuth.BundleID).Return(mockedBndl, nil)

		mockedLabel := &model.Label{Value: []interface{}{"DEFAULT"}}
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.ApplicationLabelableObject, mockedBndl.ApplicationID, model.ScenariosKey).Return(mockedLabel, nil)
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(mockedLabel, nil)

		dummyResolver := &dummyResolver{}
		// WHEN
		res, err := directive.HasScenario(ctx, nil, dummyResolver.SuccessResolve, scenario.GetApplicationIDByBundleInstanceAuth, bndlAuthIDField)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, res, mockedNextOutput())
	})
	t.Run("runtime is NOT in formation with owning application in delete bundle instance auth flow", func(t *testing.T) {
		// GIVEN
		const (
			bndlAuthIDField = "authID"
			tenantID        = "42"
			bndlAuthID      = "24"
			runtimeID       = "23"
			applicationID   = "22"
			bundleID        = "21"
		)

		bndlAuthRepo := &bndl_auth_mock.Repository{}
		defer bndlAuthRepo.AssertExpectations(t)

		bndlRepo := &bndl_mock.BundleRepository{}
		defer bndlRepo.AssertExpectations(t)

		lblRepo := &lbl_mock.LabelRepository{}
		defer lblRepo.AssertExpectations(t)

		mockedTx, mockedTransactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer mockedTransactioner.AssertExpectations(t)

		directive := scenario.NewDirective(mockedTransactioner, lblRepo, bndlRepo, bndlAuthRepo)
		ctx := context.WithValue(context.TODO(), consumer.ConsumerKey, consumer.Consumer{ConsumerID: runtimeID, ConsumerType: consumer.Runtime})
		ctx = context.WithValue(ctx, tenant.TenantContextKey, tenant.TenantCtx{InternalID: tenantID})
		rCtx := &graphql.FieldContext{
			Object:   "BundleInstanceAuth",
			Field:    graphql.CollectedField{},
			Args:     map[string]interface{}{bndlAuthIDField: bndlAuthID},
			IsMethod: false,
		}
		ctx = graphql.WithFieldContext(ctx, rCtx)
		ctxWithTx := persistence.SaveToContext(ctx, mockedTx)

		mockedBndlAuth := &model.BundleInstanceAuth{BundleID: bundleID}
		bndlAuthRepo.On("GetByID", ctxWithTx, tenantID, bndlAuthID).Return(mockedBndlAuth, nil)

		mockedBndl := &model.Bundle{ApplicationID: applicationID}
		bndlRepo.On("GetByID", ctxWithTx, tenantID, mockedBndlAuth.BundleID).Return(mockedBndl, nil)

		mockedAppLabel := &model.Label{Value: []interface{}{"DEFAULT"}}
		mockedRuntimeLabel := &model.Label{Value: []interface{}{"TEST"}}
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.ApplicationLabelableObject, mockedBndl.ApplicationID, model.ScenariosKey).Return(mockedAppLabel, nil)
		lblRepo.On("GetByKey", ctxWithTx, tenantID, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(mockedRuntimeLabel, nil)
		// WHEN
		res, err := directive.HasScenario(ctx, nil, nil, scenario.GetApplicationIDByBundleInstanceAuth, bndlAuthIDField)
		// THEN
		require.Error(t, err)
		assert.Error(t, err, scenario.ErrMissingScenario)
		assert.Equal(t, res, nil)
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
