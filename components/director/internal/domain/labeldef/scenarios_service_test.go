package labeldef_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenariosService_EnsureScenariosLabelDefinitionExists(t *testing.T) {
	testErr := errors.New("Test error")
	id := "foo"

	tnt := "tenant"
	externalTnt := "external-tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	var scenariosSchema interface{} = model.ScenariosSchema
	scenariosLD := model.LabelDefinition{
		ID:     id,
		Tenant: tnt,
		Key:    model.ScenariosKey,
		Schema: &scenariosSchema,
	}

	testCases := []struct {
		Name           string
		LabelDefRepoFn func() *automock.Repository
		UIDServiceFn   func() *automock.UIDService
		ExpectedErr    error
	}{
		{
			Name: "Success",
			LabelDefRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Exists", contextThatHasTenant(tnt), tnt, model.ScenariosKey).Return(true, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErr: nil,
		},
		{
			Name: "Success when scenarios label definition does not exist",
			LabelDefRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Exists", contextThatHasTenant(tnt), tnt, model.ScenariosKey).Return(false, nil).Once()
				repo.On("Create", contextThatHasTenant(tnt), scenariosLD).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when checking if label definition exists failed",
			LabelDefRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Exists", contextThatHasTenant(tnt), tnt, model.ScenariosKey).Return(false, testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when creating scenarios label definition failed",
			LabelDefRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Exists", contextThatHasTenant(tnt), tnt, model.ScenariosKey).Return(false, nil).Once()
				repo.On("Create", contextThatHasTenant(tnt), scenariosLD).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ldRepo := testCase.LabelDefRepoFn()
			uidSvc := testCase.UIDServiceFn()
			svc := labeldef.NewScenariosService(ldRepo, uidSvc, true)

			// when
			err := svc.EnsureScenariosLabelDefinitionExists(ctx, tnt)

			// then
			if testCase.ExpectedErr != nil {
				require.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			ldRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}
}

func TestGetAvailableScenarios(t *testing.T) {
	t.Run("returns value from default schema", func(t *testing.T) {
		// GIVEN
		mockService := &automock.Repository{}
		defer mockService.AssertExpectations(t)
		var givenSchema interface{} = model.ScenariosSchema
		givenDef := model.LabelDefinition{
			Tenant: fixTenant(),
			Key:    model.ScenariosKey,
			Schema: &givenSchema,
		}
		mockService.On("GetByKey", mock.Anything, fixTenant(), model.ScenariosKey).Return(&givenDef, nil)
		sut := labeldef.NewScenariosService(mockService, nil, true)
		// WHEN
		actualScenarios, err := sut.GetAvailableScenarios(context.TODO(), fixTenant())
		// THEN
		require.NoError(t, err)
		assert.Equal(t, []string{"DEFAULT"}, actualScenarios)
	})

	t.Run("returns error from repository", func(t *testing.T) {
		// GIVEN
		mockService := &automock.Repository{}
		defer mockService.AssertExpectations(t)
		mockService.On("GetByKey", mock.Anything, fixTenant(), model.ScenariosKey).Return(nil, fixError())
		sut := labeldef.NewScenariosService(mockService, nil, true)
		// WHEN
		_, err := sut.GetAvailableScenarios(context.TODO(), fixTenant())
		// THEN
		require.EqualError(t, err, "while getting `scenarios` label definition: some error")
	})

	t.Run("returns error when missing schema in label def", func(t *testing.T) {
		// GIVEN
		mockService := &automock.Repository{}
		defer mockService.AssertExpectations(t)
		mockService.On("GetByKey", mock.Anything, fixTenant(), model.ScenariosKey).Return(&model.LabelDefinition{}, nil)
		sut := labeldef.NewScenariosService(mockService, nil, true)
		// WHEN
		_, err := sut.GetAvailableScenarios(context.TODO(), fixTenant())
		// THEN
		require.EqualError(t, err, "missing schema for `scenarios` label definition")
	})

}

func TestScenariosService_AddDefaultScenarioIfEnabled(t *testing.T) {
	t.Run("Adds default scenario when enabled and no scenario assigned", func(t *testing.T) {
		// GIVEN
		expected := map[string]interface{}{
			"scenarios": model.ScenariosDefaultValue,
		}
		sut := labeldef.NewScenariosService(nil, nil, true)
		labels := map[string]interface{}{}

		// WHEN
		sut.AddDefaultScenarioIfEnabled(context.TODO(), &labels)

		// THEN
		assert.Equal(t, expected, labels)
	})

	t.Run("Adds default scenario when enabled and labels is nil", func(t *testing.T) {
		// GIVEN
		expected := map[string]interface{}{
			"scenarios": model.ScenariosDefaultValue,
		}
		sut := labeldef.NewScenariosService(nil, nil, true)
		var labels map[string]interface{}

		// WHEN
		sut.AddDefaultScenarioIfEnabled(context.TODO(), &labels)

		// THEN
		assert.Equal(t, expected, labels)
	})

	t.Run("Doesn't add default scenario when enable and any scenario assigned", func(t *testing.T) {
		// GIVEN
		expected := map[string]interface{}{
			"scenarios": []string{"TEST"},
		}
		sut := labeldef.NewScenariosService(nil, nil, false)
		labels := map[string]interface{}{
			"scenarios": []string{"TEST"},
		}

		// WHEN
		sut.AddDefaultScenarioIfEnabled(context.TODO(), &labels)

		// THEN
		assert.Equal(t, expected, labels)
	})

	t.Run("Doesn't add default scenario when disabled and no scenario assigned", func(t *testing.T) {
		// GIVEN
		expected := map[string]interface{}{}
		sut := labeldef.NewScenariosService(nil, nil, false)
		labels := map[string]interface{}{}

		// WHEN
		sut.AddDefaultScenarioIfEnabled(context.TODO(), &labels)

		// THEN
		assert.Equal(t, expected, labels)
	})
}
