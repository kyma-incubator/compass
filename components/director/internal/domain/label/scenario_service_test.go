package label_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	tenantID         = "tenantID"
	externalTenantID = "externalTenantID"
)

func TestScenarioService_GetScenarioNamesForApplication(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		appID := "appID"
		scenarios := []interface{}{"scenario1", "scenario2"}

		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

		labelRepo := &automock.LabelRepository{}
		objLabel := &model.Label{
			Value: interface{}(scenarios),
		}
		labelRepo.On("GetByKey", ctx, tenantID, model.ApplicationLabelableObject, appID, model.ScenariosKey).Return(objLabel, nil).Once()
		scenarioService := label.NewScenarioService(labelRepo)

		//WHEN
		actual, err := scenarioService.GetScenarioNamesForApplication(ctx, appID)
		//THEN
		assert.NoError(t, err)
		assert.Equal(t, []string{"scenario1", "scenario2"}, actual)
		labelRepo.AssertExpectations(t)
	})

	t.Run("error when cannot load tenant", func(t *testing.T) {
		// GIVEN
		appID := "appID"
		ctx := context.TODO()

		labelRepo := &automock.LabelRepository{}
		scenarioService := label.NewScenarioService(labelRepo)

		//WHEN
		actual, err := scenarioService.GetScenarioNamesForApplication(ctx, appID)
		//THEN
		assert.Error(t, err)
		assert.Nil(t, actual)
		labelRepo.AssertExpectations(t)
	})

	t.Run("error when cannot get label by key", func(t *testing.T) {
		// GIVEN
		appID := "appID"
		testError := errors.New("error")
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetByKey", ctx, tenantID, model.ApplicationLabelableObject, appID, model.ScenariosKey).Return(&model.Label{}, testError).Once()
		scenarioService := label.NewScenarioService(labelRepo)

		//WHEN
		actual, err := scenarioService.GetScenarioNamesForApplication(ctx, appID)
		//THEN
		assert.Error(t, err)
		assert.Nil(t, actual)
		labelRepo.AssertExpectations(t)
	})

	t.Run("error when cannot convert value to string slice", func(t *testing.T) {
		// GIVEN
		appID := "appID"
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

		labelRepo := &automock.LabelRepository{}
		objLabel := &model.Label{
			Value: []interface{}{1, 2},
		}
		labelRepo.On("GetByKey", ctx, tenantID, model.ApplicationLabelableObject, appID, model.ScenariosKey).Return(objLabel, nil).Once()
		scenarioService := label.NewScenarioService(labelRepo)

		//WHEN
		actual, err := scenarioService.GetScenarioNamesForApplication(ctx, appID)
		//THEN
		assert.Error(t, err)
		assert.Nil(t, actual)
		labelRepo.AssertExpectations(t)
	})
}

func TestScenarioService_GetScenarioNamesForRuntime(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		appID := "appID"
		scenarios := []interface{}{"scenario1", "scenario2"}

		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

		labelRepo := &automock.LabelRepository{}
		objLabel := &model.Label{
			Value: interface{}(scenarios),
		}
		labelRepo.On("GetByKey", ctx, tenantID, model.RuntimeLabelableObject, appID, model.ScenariosKey).Return(objLabel, nil).Once()
		scenarioService := label.NewScenarioService(labelRepo)

		//WHEN
		actual, err := scenarioService.GetScenarioNamesForRuntime(ctx, appID)
		//THEN
		assert.NoError(t, err)
		assert.Equal(t, []string{"scenario1", "scenario2"}, actual)
		labelRepo.AssertExpectations(t)
	})
}

func TestScenarioService_GetRuntimeScenarioLabelsForAnyMatchingScenario(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		scenarios := []string{"scenario1"}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

		labelRepo := &automock.LabelRepository{}
		labels := []model.Label{
			{
				ID:         "1",
				Key:        model.ScenariosKey,
				Value:      []string{"scenario1", "scenario2"},
				ObjectType: model.RuntimeLabelableObject,
			},
			{
				ID:         "2",
				Key:        model.ScenariosKey,
				Value:      []string{"scenario1", "scenario3"},
				ObjectType: model.RuntimeLabelableObject,
			}}
		labelRepo.On("ListByObjectTypeAndMatchAnyScenario", ctx, tenantID, model.RuntimeLabelableObject, scenarios).Return(labels, nil).Once()
		scenarioService := label.NewScenarioService(labelRepo)

		//WHEN
		actual, err := scenarioService.GetRuntimeScenarioLabelsForAnyMatchingScenario(ctx, scenarios)
		//THEN
		assert.NoError(t, err)
		assert.Equal(t, labels, actual)
		labelRepo.AssertExpectations(t)
	})

	t.Run("error when cannot load tenant", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		scenarioService := label.NewScenarioService(labelRepo)
		//WHEN
		actual, err := scenarioService.GetRuntimeScenarioLabelsForAnyMatchingScenario(ctx, []string{})
		//THEN
		assert.Error(t, err)
		assert.Nil(t, actual)
		labelRepo.AssertExpectations(t)
	})
}

func TestScenarioService_GetApplicationScenarioLabelsForAnyMatchingScenario(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		scenarios := []string{"scenario1"}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

		labelRepo := &automock.LabelRepository{}
		labels := []model.Label{
			{
				ID:         "1",
				Key:        model.ScenariosKey,
				Value:      []string{"scenario1", "scenario2"},
				ObjectType: model.ApplicationLabelableObject,
			},
			{
				ID:         "2",
				Key:        model.ScenariosKey,
				Value:      []string{"scenario1", "scenario3"},
				ObjectType: model.ApplicationLabelableObject,
			}}
		labelRepo.On("ListByObjectTypeAndMatchAnyScenario", ctx, tenantID, model.ApplicationLabelableObject, scenarios).Return(labels, nil).Once()
		scenarioService := label.NewScenarioService(labelRepo)

		//WHEN
		actual, err := scenarioService.GetApplicationScenarioLabelsForAnyMatchingScenario(ctx, scenarios)
		//THEN
		assert.NoError(t, err)
		assert.Equal(t, labels, actual)
		labelRepo.AssertExpectations(t)
	})

	t.Run("error when cannot load tenant", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		scenarioService := label.NewScenarioService(labelRepo)
		//WHEN
		actual, err := scenarioService.GetApplicationScenarioLabelsForAnyMatchingScenario(ctx, []string{})
		//THEN
		assert.Error(t, err)
		assert.Nil(t, actual)
		labelRepo.AssertExpectations(t)
	})
}

func TestScenarioService_GetBundleInstanceAuthsScenarioLabels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		appID := "appId"
		runtimeID := "runtimeId"
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

		labelRepo := &automock.LabelRepository{}
		labels := []model.Label{
			{
				ID:         "1",
				Key:        model.ScenariosKey,
				Value:      []string{"scenario1", "scenario2"},
				ObjectType: model.ApplicationLabelableObject,
			},
			{
				ID:         "2",
				Key:        model.ScenariosKey,
				Value:      []string{"scenario1", "scenario3"},
				ObjectType: model.ApplicationLabelableObject,
			}}
		labelRepo.On("GetBundleInstanceAuthsScenarioLabels", ctx, tenantID, appID, runtimeID).Return(labels, nil).Once()
		scenarioService := label.NewScenarioService(labelRepo)

		//WHEN
		actual, err := scenarioService.GetBundleInstanceAuthsScenarioLabels(ctx, appID, runtimeID)
		//THEN
		assert.NoError(t, err)
		assert.Equal(t, labels, actual)
		labelRepo.AssertExpectations(t)
	})

	t.Run("error when cannot load tenant", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		scenarioService := label.NewScenarioService(labelRepo)
		//WHEN
		actual, err := scenarioService.GetBundleInstanceAuthsScenarioLabels(ctx, "", "")
		//THEN
		assert.Error(t, err)
		assert.Nil(t, actual)
		labelRepo.AssertExpectations(t)
	})
}

func TestMergeScenarios(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		scenarios := []interface{}{"scenario-1"}
		newScenario := "scenario-2"

		lbl := model.Label{
			Key:   model.ScenariosKey,
			Value: scenarios,
		}

		actual, err := label.MergeScenarios(lbl, []string{newScenario}, label.UniqueScenarios)
		require.NoError(t, err)
		assert.ElementsMatch(t, []interface{}{"scenario-1", "scenario-2"}, actual.Value)
	})

	t.Run("Fails when non-scenario label is provided", func(t *testing.T) {
		_, err := label.MergeScenarios(model.Label{Key: "something"}, []string{"scenario-1"}, label.UniqueScenarios)
		require.Error(t, err)
	})

	t.Run("Fails when getting scenarios fails", func(t *testing.T) {
		_, err := label.MergeScenarios(model.Label{Key: model.ScenariosKey, Value: 42}, []string{"scenario-1"}, label.UniqueScenarios)
		require.Error(t, err)
	})

	t.Run("Returns nil when no scenarios are left after applying merge function", func(t *testing.T) {
		scenarios := []interface{}{"scenario-1"}
		newScenario := "scenario-2"

		lbl := model.Label{
			Key:   model.ScenariosKey,
			Value: scenarios,
		}

		actual, err := label.MergeScenarios(lbl, []string{newScenario}, func(scenarios, diffScenario []string) []string {
			return make([]string, 0)
		})

		require.NoError(t, err)
		assert.Nil(t, actual)
	})
}
