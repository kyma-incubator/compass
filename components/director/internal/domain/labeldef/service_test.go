package labeldef_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	tenant2 "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const defaultScenarioEnabled = true

func TestServiceCreateWithFormations(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		testFormations := []string{"test-formation-one", "test-formation-two"}
		expectedFormations := append(testFormations, "DEFAULT")
		ctx := context.TODO()

		mockUID := &automock.UIDService{}
		defer mockUID.AssertExpectations(t)
		mockUID.On("Generate").Return(fixUUID())

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			if schemaArgs, ok := args.Get(1).(model.LabelDefinition); ok {
				formations, err := labeldef.ParseFormationsFromSchema(schemaArgs.Schema)
				require.NoError(t, err)
				require.ElementsMatch(t, formations, expectedFormations)
				return
			}
			t.Fatal("schema should contain desired formations")
		})

		sut := labeldef.NewService(mockRepository, nil, nil, nil, mockUID, defaultScenarioEnabled)
		// WHEN
		err := sut.CreateWithFormations(ctx, "tenant", testFormations)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success when default scenario is disabled", func(t *testing.T) {
		// GIVEN
		testFormations := []string{"test-formation-one", "test-formation-two"}
		expectedFormations := testFormations
		ctx := context.TODO()

		mockUID := &automock.UIDService{}
		defer mockUID.AssertExpectations(t)
		mockUID.On("Generate").Return(fixUUID())

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			if schemaArgs, ok := args.Get(1).(model.LabelDefinition); ok {
				formations, err := labeldef.ParseFormationsFromSchema(schemaArgs.Schema)
				require.NoError(t, err)
				require.ElementsMatch(t, formations, expectedFormations)
				return
			}
			t.Fatal("schema should contain desired formations")
		})

		sut := labeldef.NewService(mockRepository, nil, nil, nil, mockUID, false)
		// WHEN
		err := sut.CreateWithFormations(ctx, "tenant", testFormations)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error if cannot create Label Definition", func(t *testing.T) {
		// GIVEN
		testFormations := []string{"test-formation-one", "test-formation-two"}
		expectedFormations := append(testFormations, "DEFAULT")
		ctx := context.TODO()
		testError := errors.New("test error")

		mockUID := &automock.UIDService{}
		defer mockUID.AssertExpectations(t)
		mockUID.On("Generate").Return(fixUUID())

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("Create", mock.Anything, mock.Anything).Return(testError).Run(func(args mock.Arguments) {
			if schemaArgs, ok := args.Get(1).(model.LabelDefinition); ok {
				formations, err := labeldef.ParseFormationsFromSchema(schemaArgs.Schema)
				require.NoError(t, err)
				require.ElementsMatch(t, formations, expectedFormations)
				return
			}
			t.Fatal("schema should contain desired formations")
		})

		sut := labeldef.NewService(mockRepository, nil, nil, nil, mockUID, defaultScenarioEnabled)
		// WHEN
		err := sut.CreateWithFormations(ctx, "tenant", testFormations)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
	})
}

func TestServiceGet(t *testing.T) {
	t.Run("success when key is not scenarios key", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    "key",
			Tenant: "tenant",
		}
		mockRepository.On("GetByKey", ctx, "tenant", "key").Return(&given, nil)
		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil, defaultScenarioEnabled)
		// WHEN
		actual, err := sut.Get(ctx, "tenant", "key")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, &given, actual)
	})

	t.Run("success when LD exists", func(t *testing.T) {
		// GIVEN
		testKey := model.ScenariosKey
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    testKey,
			Tenant: "tenant",
		}

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("GetByKey", ctx, "tenant", testKey).Return(&given, nil)
		mockRepository.On("Exists", ctx, "tenant", testKey).Return(true, nil)

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil, defaultScenarioEnabled)
		// WHEN
		actual, err := sut.Get(ctx, "tenant", testKey)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, &given, actual)
	})

	t.Run("success when getting scenarios LD if it doesn't exist", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		testTenant := "tenant"
		testKey := model.ScenariosKey
		var schema interface{} = model.ScenariosSchema
		given := model.LabelDefinition{
			ID:      fixUUID(),
			Key:     testKey,
			Tenant:  testTenant,
			Schema:  &schema,
			Version: 0,
		}

		mockUID := &automock.UIDService{}
		mockUID.On("Generate").Return(fixUUID())
		defer mockUID.AssertExpectations(t)

		mockRepository := &automock.Repository{}
		mockRepository.On("GetByKey", ctx, testTenant, testKey).Return(&given, nil).Once()
		mockRepository.On("Exists", ctx, testTenant, testKey).Return(false, nil)
		mockRepository.On("Create", ctx, given).Return(nil)
		defer mockRepository.AssertExpectations(t)

		sut := labeldef.NewService(mockRepository, nil, nil, nil, mockUID, defaultScenarioEnabled)

		// WHEN
		actual, err := sut.Get(ctx, testTenant, testKey)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &given, actual)
	})

	t.Run("on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("GetByKey", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("some error"))

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil, defaultScenarioEnabled)
		// WHEN
		_, err := sut.Get(context.TODO(), "tenant", "key")
		// THEN
		require.EqualError(t, err, "while fetching Label Definition: some error")
	})

	t.Run("error while checking if LD exists", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		testTenant := "tenant"
		testKey := model.ScenariosKey

		testError := errors.New("some error")

		mockRepository := &automock.Repository{}
		mockRepository.On("Exists", ctx, testTenant, testKey).Return(false, testError)
		defer mockRepository.AssertExpectations(t)

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil, defaultScenarioEnabled)

		// WHEN
		actual, err := sut.Get(ctx, testTenant, testKey)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		assert.Nil(t, actual)
	})
}

func TestServiceGetWithoutCreating(t *testing.T) {
	t.Run("success when repository returns label definition", func(t *testing.T) {
		testKey := model.ScenariosKey
		ctx := context.TODO()
		given := model.LabelDefinition{
			Key:    testKey,
			Tenant: "tenant",
		}

		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("GetByKey", ctx, "tenant", testKey).Return(&given, nil)

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil, defaultScenarioEnabled)
		// WHEN
		actual, err := sut.GetWithoutCreating(ctx, "tenant", testKey)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, &given, actual)
	})
	t.Run("on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		mockRepository.On("GetByKey", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("some error"))

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil, defaultScenarioEnabled)
		// WHEN
		_, err := sut.GetWithoutCreating(context.TODO(), "tenant", "key")
		// THEN
		require.EqualError(t, err, "while fetching Label Definition: some error")
	})
}

func TestServiceList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		ctx := context.TODO()
		givenDefs := []model.LabelDefinition{
			{
				Tenant: "tenant",
				Key:    "key1",
			},
			{
				Tenant: "tenant",
				Key:    "key2",
			},
		}
		mockRepository.On("List", ctx, "tenant").Return(givenDefs, nil)

		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil, defaultScenarioEnabled)
		// WHEN
		actual, err := sut.List(ctx, "tenant")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenDefs, actual)
	})

	t.Run("on error from repository", func(t *testing.T) {
		// GIVEN
		mockRepository := &automock.Repository{}
		defer mockRepository.AssertExpectations(t)
		ctx := context.TODO()
		mockRepository.On("List", ctx, "tenant").Return(nil, errors.New("some error"))
		sut := labeldef.NewService(mockRepository, nil, nil, nil, nil, defaultScenarioEnabled)
		// WHEN
		_, err := sut.List(ctx, "tenant")
		// THEN
		require.EqualError(t, err, "while fetching Label Definitions: some error")
	})
}

func TestService_EnsureScenariosLabelDefinitionExists(t *testing.T) {
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
			svc := labeldef.NewService(ldRepo, nil, nil, nil, uidSvc, true)

			// WHEN
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
		sut := labeldef.NewService(mockService, nil, nil, nil, nil, true)
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
		sut := labeldef.NewService(mockService, nil, nil, nil, nil, true)
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
		sut := labeldef.NewService(mockService, nil, nil, nil, nil, true)
		// WHEN
		_, err := sut.GetAvailableScenarios(context.TODO(), fixTenant())
		// THEN
		require.EqualError(t, err, "missing schema for `scenarios` label definition")
	})
}

func TestScenariosService_AddDefaultScenarioIfEnabled(t *testing.T) {
	t.Run("Adds default scenario when enabled an tenant is account type and no scenario assigned", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"
		externalTnt := "external-tnt"
		ctx := context.TODO()
		expected := map[string]interface{}{
			"scenarios": model.ScenariosDefaultValue,
		}
		tenantRepo := &automock.TenantRepository{}
		tenantRepo.On("Get", ctx, tnt).Return(&model.BusinessTenantMapping{
			ID:             tnt,
			ExternalTenant: externalTnt,
			Type:           tenant2.Account,
		}, nil)
		sut := labeldef.NewService(nil, nil, nil, tenantRepo, nil, true)
		labels := map[string]interface{}{}

		// WHEN
		sut.AddDefaultScenarioIfEnabled(ctx, tnt, &labels)

		// THEN
		assert.Equal(t, expected, labels)
	})
	t.Run("Should not add default scenario when enabled an tenant is subaccount type and no scenario assigned", func(t *testing.T) {
		// GIVEN
		tnt := "sub-tenant"
		externalTnt := "sub-external-tnt"
		ctx := context.TODO()
		expected := map[string]interface{}{}
		tenantRepo := &automock.TenantRepository{}
		tenantRepo.On("Get", ctx, tnt).Return(&model.BusinessTenantMapping{
			ID:             tnt,
			ExternalTenant: externalTnt,
			Type:           tenant2.Subaccount,
		}, nil)
		sut := labeldef.NewService(nil, nil, nil, tenantRepo, nil, true)
		labels := map[string]interface{}{}

		// WHEN
		sut.AddDefaultScenarioIfEnabled(ctx, tnt, &labels)

		// THEN
		assert.Equal(t, expected, labels)
	})

	t.Run("Adds default scenario when enabled and labels is nil", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"
		externalTnt := "external-tnt"
		ctx := context.TODO()
		expected := map[string]interface{}{
			"scenarios": model.ScenariosDefaultValue,
		}
		tenantRepo := &automock.TenantRepository{}
		tenantRepo.On("Get", ctx, tnt).Return(&model.BusinessTenantMapping{
			ID:             tnt,
			ExternalTenant: externalTnt,
			Type:           tenant2.Account,
		}, nil)
		sut := labeldef.NewService(nil, nil, nil, tenantRepo, nil, true)
		var labels map[string]interface{}

		// WHEN
		sut.AddDefaultScenarioIfEnabled(ctx, tnt, &labels)

		// THEN
		assert.Equal(t, expected, labels)
	})

	t.Run("Doesn't add default scenario when enable and any scenario assigned", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"
		externalTnt := "external-tnt"
		ctx := context.TODO()
		tenantRepo := &automock.TenantRepository{}
		tenantRepo.On("Get", ctx, tnt).Return(&model.BusinessTenantMapping{
			ID:             tnt,
			ExternalTenant: externalTnt,
			Type:           tenant2.Account,
		}, nil)
		expected := map[string]interface{}{
			"scenarios": []string{"TEST"},
		}
		sut := labeldef.NewService(nil, nil, nil, tenantRepo, nil, true)
		labels := map[string]interface{}{
			"scenarios": []string{"TEST"},
		}

		// WHEN
		sut.AddDefaultScenarioIfEnabled(context.TODO(), tnt, &labels)

		// THEN
		assert.Equal(t, expected, labels)
	})

	t.Run("Doesn't add default scenario when disabled and no scenario assigned", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"
		externalTnt := "external-tnt"
		ctx := context.TODO()
		tenantRepo := &automock.TenantRepository{}
		tenantRepo.On("Get", ctx, tnt).Return(&model.BusinessTenantMapping{
			ID:             tnt,
			ExternalTenant: externalTnt,
			Type:           tenant2.Account,
		}, nil)
		expected := map[string]interface{}{
			"scenarios": []string{"TEST"},
		}
		sut := labeldef.NewService(nil, nil, nil, tenantRepo, nil, false)
		labels := map[string]interface{}{
			"scenarios": []string{"TEST"},
		}

		// WHEN
		sut.AddDefaultScenarioIfEnabled(context.TODO(), tnt, &labels)

		// THEN
		assert.Equal(t, expected, labels)
	})

	t.Run("Doesn't add default scenario when fails to retrieve tenant", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"
		err := errors.New("test-err")
		ctx := context.TODO()
		tenantRepo := &automock.TenantRepository{}
		tenantRepo.On("Get", ctx, tnt).Return(nil, err)
		expected := map[string]interface{}{}
		sut := labeldef.NewService(nil, nil, nil, tenantRepo, nil, true)
		labels := map[string]interface{}{}

		// WHEN
		sut.AddDefaultScenarioIfEnabled(context.TODO(), tnt, &labels)

		// THEN
		assert.Equal(t, expected, labels)
	})
}

func fixUUID() string {
	return "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
}

func fixBasicInputSchema() *graphql.JSONSchema {
	sch := `{
		"$id": "https://example.com/person.schema.json",
  		"$schema": "http://json-schema.org/draft-07/schema#",
  		"title": "Person",
  		"type": "object",
  		"properties": {
  		  "firstName": {
  		    "type": "string",
  		    "description": "The person's first name."
  		  },
  		  "lastName": {
  		    "type": "string",
  		    "description": "The person's last name."
  		  },
  		  "age": {
  		    "description": "Age in years which must be equal to or greater than zero.",
  		    "type": "integer",
  		    "minimum": 0
  		  }
  		}
	  }`
	jsonSchema := graphql.JSONSchema(sch)
	return &jsonSchema
}

func fixBasicSchema(t *testing.T) *interface{} {
	sch := fixBasicInputSchema()
	require.NotNil(t, sch)
	var obj map[string]interface{}

	err := json.Unmarshal([]byte(*sch), &obj)
	require.NoError(t, err)
	var objTemp interface{} = obj
	return &objTemp
}

func fixSchema(t *testing.T, propertyName, propertyType, propertyDescription, requiredProperty string) *interface{} {
	sch := fmt.Sprintf(`{
		"$id": "https://example.com/person.schema.json",
		"$schema": "http://json-schema.org/draft-07/schema#",
		"title": "Person",
		"type": "object",
		"properties": {
		  "%s": {
		    "type": "%s",
		    "description": "%s"
		  }
		},
		"required": ["%s"]
	  }`, propertyName, propertyType, propertyDescription, requiredProperty)
	var obj map[string]interface{}

	err := json.Unmarshal([]byte(sch), &obj)
	require.NoError(t, err)
	var objTemp interface{} = obj
	return &objTemp
}

func fixDefaultScenariosLabelDefinition(tenantID string) model.LabelDefinition {
	var schema interface{} = model.ScenariosSchema
	ld := model.LabelDefinition{
		Key:    model.ScenariosKey,
		Tenant: tenantID,
		Schema: &schema,
	}
	return ld
}

func fixModifiedScenariosLabelDefinition(tenantID string) model.LabelDefinition {
	m := map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type":      "string",
			"pattern":   "^[A-Za-z0-9]([-_A-Za-z0-9\\s]*[A-Za-z0-9])$",
			"enum":      []string{"DEFAULT", "scenario-A", "scenario-B"},
			"maxLength": 128,
		},
	}

	var schema interface{} = m
	ld := model.LabelDefinition{
		Key:    model.ScenariosKey,
		Tenant: tenantID,
		Schema: &schema,
	}
	return ld
}

func fixLabel(id, key string, value interface{}, objectID string, objectType model.LabelableObject) *model.Label {
	return &model.Label{
		ID:         id,
		Key:        key,
		Value:      value,
		ObjectID:   objectID,
		ObjectType: objectType,
	}
}
