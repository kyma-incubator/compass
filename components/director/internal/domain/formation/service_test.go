package formation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tnt         = "tenant"
	externalTnt = "external-tnt"
)

func TestServiceCreateFormation(t *testing.T) {
	t.Run("success when no labeldef exists", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		expected := &model.Formation{
			Name: testFormation,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
		mockLabelDefService.On("CreateWithFormations", ctx, tnt, []string{testFormation}).Return(nil)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.CreateFormation(ctx, tnt, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("success when labeldef exists", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}

		schema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
		assert.NoError(t, err)
		lblDef := fixDefaultScenariosLabelDefinition(tnt, schema)

		expectedSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
		assert.NoError(t, err)
		expectedLblDef := fixDefaultScenariosLabelDefinition(tnt, expectedSchema)

		expected := &model.Formation{
			Name: testFormation,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&lblDef, nil)
		mockLabelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, expectedSchema, tnt, model.ScenariosKey).Return(nil)
		mockLabelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, expectedSchema, tnt, model.ScenariosKey).Return(nil)
		mockLabelDefRepository.On("UpdateWithVersion", ctx, expectedLblDef).Return(nil)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.CreateFormation(ctx, tnt, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("error when labeldef is missing and can not create it", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		testErr := errors.New("Test error")
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(nil, apperrors.NewNotFoundError(resource.LabelDefinition, ""))
		mockLabelDefService.On("CreateWithFormations", ctx, tnt, []string{testFormation}).Return(testErr)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.CreateFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
	})
	t.Run("error when can not get labeldef", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		testErr := errors.New("Test error")
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(nil, testErr)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.CreateFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
	})
	t.Run("error when labeldef's schema is missing", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}

		schema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
		assert.NoError(t, err)
		lblDef := fixDefaultScenariosLabelDefinition(tnt, schema)
		lblDef.Schema = nil
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&lblDef, nil)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.CreateFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), "missing schema")
	})
	t.Run("error when validating existing labels against the schema", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		testErr := errors.New("test error")
		schema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
		assert.NoError(t, err)
		lblDef := fixDefaultScenariosLabelDefinition(tnt, schema)
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
		assert.NoError(t, err)
		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&lblDef, nil)
		mockLabelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, lblDef.Key).Return(testErr)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.CreateFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		assert.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("error when validating automatic scenario assignment against the schema", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		testErr := errors.New("test error")
		schema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
		assert.NoError(t, err)
		lblDef := fixDefaultScenariosLabelDefinition(tnt, schema)
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
		assert.NoError(t, err)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&lblDef, nil)
		mockLabelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, lblDef.Key).Return(nil)
		mockLabelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, lblDef.Key).Return(testErr)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.CreateFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		assert.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("error when update with version fails", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		testErr := errors.New("test error")
		schema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
		assert.NoError(t, err)
		lblDef := fixDefaultScenariosLabelDefinition(tnt, schema)
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
		assert.NoError(t, err)
		expectedLblDef := fixDefaultScenariosLabelDefinition(tnt, newSchema)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&lblDef, nil)
		mockLabelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, lblDef.Key).Return(nil)
		mockLabelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, lblDef.Key).Return(nil)
		mockLabelDefRepository.On("UpdateWithVersion", ctx, expectedLblDef).Return(testErr)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.CreateFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		assert.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})
}

func TestServiceDeleteFormation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}

		schema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
		assert.NoError(t, err)
		lblDef := fixDefaultScenariosLabelDefinition(tnt, schema)

		expected := &model.Formation{
			Name: testFormation,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
		assert.NoError(t, err)
		expectedLblDef := fixDefaultScenariosLabelDefinition(tnt, newSchema)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&lblDef, nil)
		mockLabelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(nil)
		mockLabelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(nil)
		mockLabelDefRepository.On("UpdateWithVersion", ctx, expectedLblDef).Return(nil)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.DeleteFormation(ctx, tnt, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("error when can not get labeldef", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		testErr := errors.New("Test error")
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(nil, testErr)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.DeleteFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		require.Nil(t, actual)
	})
	t.Run("error when labeldef's schema is missing", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}

		schema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
		assert.NoError(t, err)
		lblDef := fixDefaultScenariosLabelDefinition(tnt, schema)
		lblDef.Schema = nil
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&lblDef, nil)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.DeleteFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), "missing schema")
	})
	t.Run("error when validating existing labels against the schema", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		testErr := errors.New("test error")
		schema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
		assert.NoError(t, err)
		lblDef := fixDefaultScenariosLabelDefinition(tnt, schema)
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
		assert.NoError(t, err)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&lblDef, nil)
		mockLabelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(testErr)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.DeleteFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		assert.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("error when validating automatic scenario assignment against the schema", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		testErr := errors.New("test error")
		schema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
		assert.NoError(t, err)
		lblDef := fixDefaultScenariosLabelDefinition(tnt, schema)
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
		assert.NoError(t, err)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&lblDef, nil)
		mockLabelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(nil)
		mockLabelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, model.ScenariosKey).Return(testErr)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.DeleteFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		assert.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})
	t.Run("error when update with version fails", func(t *testing.T) {
		//GIVEN
		mockLabelDefRepository := &automock.LabelDefRepository{}
		mockLabelDefService := &automock.LabelDefService{}
		defer mockLabelDefRepository.AssertExpectations(t)
		defer mockLabelDefService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		testErr := errors.New("test error")
		schema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT", testFormation})
		assert.NoError(t, err)
		lblDef := fixDefaultScenariosLabelDefinition(tnt, schema)
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		newSchema, err := labeldef.NewSchemaForFormations([]string{"DEFAULT"})
		assert.NoError(t, err)
		expectedLblDef := fixDefaultScenariosLabelDefinition(tnt, newSchema)

		mockLabelDefRepository.On("GetByKey", ctx, tnt, model.ScenariosKey).Return(&lblDef, nil)
		mockLabelDefService.On("ValidateExistingLabelsAgainstSchema", ctx, newSchema, tnt, lblDef.Key).Return(nil)
		mockLabelDefService.On("ValidateAutomaticScenarioAssignmentAgainstSchema", ctx, newSchema, tnt, lblDef.Key).Return(nil)
		mockLabelDefRepository.On("UpdateWithVersion", ctx, expectedLblDef).Return(testErr)

		sut := formation.NewService(mockLabelDefRepository, nil, nil, mockLabelDefService, nil)
		// WHEN
		actual, err := sut.DeleteFormation(ctx, tnt, in)
		// THEN
		require.Error(t, err)
		assert.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})
}

func TestServiceAssignFormation(t *testing.T) {
	t.Run("success for application if label does not exist", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		mockUIDService := &automock.UIDService{}
		defer mockUIDService.AssertExpectations(t)
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lblInput := model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		expected := &model.Formation{
			Name: testFormation,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockUIDService.On("Generate").Return(fixUUID())
		mockLabelService.On("GetLabel", ctx, tnt, &lblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
		mockLabelService.On("CreateLabel", ctx, tnt, fixUUID(), &lblInput).Return(nil)

		sut := formation.NewService(nil, mockLabelService, mockUIDService, nil, nil)
		// WHEN
		actual, err := sut.AssignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("success for application if formation is already added", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lbl := &model.Label{
			ID:         "123",
			Tenant:     tnt,
			Key:        model.ScenariosKey,
			Value:      []interface{}{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		lblInput := &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		expected := &model.Formation{
			Name: testFormation,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelService.On("GetLabel", ctx, tnt, lblInput).Return(lbl, nil)
		mockLabelService.On("UpdateLabel", ctx, tnt, lbl.ID, lblInput).Return(nil)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.AssignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("success for application with new formation", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: "test-formation-2",
		}
		objectID := "123"
		lbl := &model.Label{
			ID:         "123",
			Tenant:     tnt,
			Key:        model.ScenariosKey,
			Value:      []interface{}{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    1,
		}
		lblInput := &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{"test-formation-2"},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		expected := &model.Formation{
			Name: "test-formation-2",
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelService.On("GetLabel", ctx, tnt, lblInput).Return(lbl, nil)
		mockLabelService.On("UpdateLabel", ctx, tnt, lbl.ID, &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation, "test-formation-2"},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    1,
		}).Return(nil)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.AssignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("success for tenant", func(t *testing.T) {
		//GIVEN
		mockAsaService := &automock.AutomaticFormationAssignmentService{}
		defer mockAsaService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		expected := &model.Formation{
			Name: testFormation,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		asa := model.AutomaticScenarioAssignment{
			ScenarioName: testFormation,
			Tenant:       tnt,
			Selector: model.LabelSelector{
				Key:   "global_subaccount_id",
				Value: objectID,
			},
		}

		mockAsaService.On("Create", ctx, asa).Return(asa, nil)

		sut := formation.NewService(nil, nil, nil, nil, mockAsaService)
		// WHEN
		actual, err := sut.AssignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeTenant, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("error for application when label does not exist and can't create it", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		mockUIDService := &automock.UIDService{}
		defer mockUIDService.AssertExpectations(t)
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lblInput := model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		testErr := errors.New("test error")

		mockUIDService.On("Generate").Return(fixUUID())
		mockLabelService.On("GetLabel", ctx, tnt, &lblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
		mockLabelService.On("CreateLabel", ctx, tnt, fixUUID(), &lblInput).Return(testErr)

		sut := formation.NewService(nil, mockLabelService, mockUIDService, nil, nil)
		// WHEN
		actual, err := sut.AssignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("error for application while getting label", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lblInput := model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		testErr := errors.New("test error")

		mockLabelService.On("GetLabel", ctx, tnt, &lblInput).Return(nil, testErr)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.AssignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("error for application while converting label values to string slice", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lblInput := model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		lbl := &model.Label{
			ID:         "123",
			Tenant:     tnt,
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}

		mockLabelService.On("GetLabel", ctx, tnt, &lblInput).Return(lbl, nil)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.AssignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), "cannot convert label value to slice of strings")
	})

	t.Run("error for application while converting label value to string", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lblInput := model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		lbl := &model.Label{
			ID:         "123",
			Tenant:     tnt,
			Key:        model.ScenariosKey,
			Value:      []interface{}{5},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}

		mockLabelService.On("GetLabel", ctx, tnt, &lblInput).Return(lbl, nil)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.AssignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), "cannot cast label value as a string")
	})

	t.Run("error for application when updating label fails", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lbl := &model.Label{
			ID:         "123",
			Tenant:     tnt,
			Key:        model.ScenariosKey,
			Value:      []interface{}{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		lblInput := &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		testErr := errors.New("test error")

		mockLabelService.On("GetLabel", ctx, tnt, lblInput).Return(lbl, nil)
		mockLabelService.On("UpdateLabel", ctx, tnt, lbl.ID, lblInput).Return(testErr)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.AssignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("error for tenant when create fails", func(t *testing.T) {
		//GIVEN
		mockAsaService := &automock.AutomaticFormationAssignmentService{}
		defer mockAsaService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		asa := model.AutomaticScenarioAssignment{
			ScenarioName: testFormation,
			Tenant:       tnt,
			Selector: model.LabelSelector{
				Key:   "global_subaccount_id",
				Value: objectID,
			},
		}
		testErr := errors.New("test error")

		mockAsaService.On("Create", ctx, asa).Return(asa, testErr)

		sut := formation.NewService(nil, nil, nil, nil, mockAsaService)
		// WHEN
		actual, err := sut.AssignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeTenant, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("error when object type is unknown", func(t *testing.T) {
		//GIVEN
		sut := formation.NewService(nil, nil, nil, nil, nil)
		// WHEN
		actual, err := sut.AssignFormation(context.TODO(), "", "", "UNKNOWN", model.Formation{})
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), "unknown formation type")
	})
}

func TestServiceUnassignFormation(t *testing.T) {
	t.Run("success for application", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lbl := &model.Label{
			ID:         "123",
			Tenant:     tnt,
			Key:        model.ScenariosKey,
			Value:      []interface{}{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		lblInput := &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		expected := &model.Formation{
			Name: testFormation,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelService.On("GetLabel", ctx, tnt, lblInput).Return(lbl, nil)
		mockLabelService.On("UpdateLabel", ctx, tnt, lbl.ID, &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}).Return(nil)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.UnassignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("success for application if formation do not exist", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: "test-formation-2",
		}
		objectID := "123"
		lbl := &model.Label{
			ID:         "123",
			Tenant:     tnt,
			Key:        model.ScenariosKey,
			Value:      []interface{}{testFormation, "test-formation-2"},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		lblInput := &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{"test-formation-2"},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		expected := &model.Formation{
			Name: "test-formation-2",
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		mockLabelService.On("GetLabel", ctx, tnt, lblInput).Return(lbl, nil)
		mockLabelService.On("UpdateLabel", ctx, tnt, lbl.ID, &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}).Return(nil)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.UnassignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("success for tenant", func(t *testing.T) {
		//GIVEN
		mockAsaService := &automock.AutomaticFormationAssignmentService{}
		defer mockAsaService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		expected := &model.Formation{
			Name: testFormation,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		asa := model.AutomaticScenarioAssignment{
			ScenarioName: testFormation,
			Tenant:       tnt,
			Selector: model.LabelSelector{
				Key:   "global_subaccount_id",
				Value: objectID,
			},
		}

		mockAsaService.On("GetForScenarioName", ctx, testFormation).Return(asa, nil)
		mockAsaService.On("Delete", ctx, asa).Return(nil)

		sut := formation.NewService(nil, nil, nil, nil, mockAsaService)
		// WHEN
		actual, err := sut.UnassignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeTenant, in)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("error for application while getting label", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lblInput := model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		testErr := errors.New("test error")

		mockLabelService.On("GetLabel", ctx, tnt, &lblInput).Return(nil, testErr)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.UnassignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("error for application while converting label values to string slice", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lblInput := model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		lbl := &model.Label{
			ID:         "123",
			Tenant:     tnt,
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}

		mockLabelService.On("GetLabel", ctx, tnt, &lblInput).Return(lbl, nil)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.UnassignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), "cannot convert label value to slice of strings")
	})

	t.Run("error for application while converting label value to string", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		lblInput := model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		lbl := &model.Label{
			ID:         "123",
			Tenant:     tnt,
			Key:        model.ScenariosKey,
			Value:      []interface{}{5},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}

		mockLabelService.On("GetLabel", ctx, tnt, &lblInput).Return(lbl, nil)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.UnassignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), "cannot cast label value as a string")
	})

	t.Run("error for application when updating label fails", func(t *testing.T) {
		//GIVEN
		mockLabelService := &automock.LabelService{}
		defer mockLabelService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		expected := &model.Formation{
			Name: testFormation,
		}
		lbl := &model.Label{
			ID:         "123",
			Tenant:     tnt,
			Key:        model.ScenariosKey,
			Value:      []interface{}{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		lblInput := &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{testFormation},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		testErr := errors.New("test error")

		mockLabelService.On("GetLabel", ctx, tnt, lblInput).Return(lbl, nil)
		mockLabelService.On("UpdateLabel", ctx, tnt, lbl.ID, &model.LabelInput{
			Key:        model.ScenariosKey,
			Value:      []string{},
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
			Version:    0,
		}).Return(testErr)

		sut := formation.NewService(nil, mockLabelService, nil, nil, nil)
		// WHEN
		actual, err := sut.UnassignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeApplication, in)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		require.Equal(t, expected, actual)
	})

	t.Run("error for tenant when delete fails", func(t *testing.T) {
		//GIVEN
		mockAsaService := &automock.AutomaticFormationAssignmentService{}
		defer mockAsaService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		asa := model.AutomaticScenarioAssignment{
			ScenarioName: testFormation,
			Tenant:       tnt,
			Selector: model.LabelSelector{
				Key:   "global_subaccount_id",
				Value: objectID,
			},
		}
		testErr := errors.New("test error")

		mockAsaService.On("GetForScenarioName", ctx, testFormation).Return(asa, nil)
		mockAsaService.On("Delete", ctx, asa).Return(testErr)

		sut := formation.NewService(nil, nil, nil, nil, mockAsaService)
		// WHEN
		actual, err := sut.UnassignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeTenant, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("error for tenant when getting automatic assignment scenario fails", func(t *testing.T) {
		//GIVEN
		mockAsaService := &automock.AutomaticFormationAssignmentService{}
		defer mockAsaService.AssertExpectations(t)

		in := model.Formation{
			Name: testFormation,
		}
		objectID := "123"
		ctx := context.TODO()
		ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

		testErr := errors.New("test error")

		mockAsaService.On("GetForScenarioName", ctx, testFormation).Return(model.AutomaticScenarioAssignment{}, testErr)

		sut := formation.NewService(nil, nil, nil, nil, mockAsaService)
		// WHEN
		actual, err := sut.UnassignFormation(ctx, tnt, objectID, graphql.FormationObjectTypeTenant, in)
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("error when object type is unknown", func(t *testing.T) {
		//GIVEN
		sut := formation.NewService(nil, nil, nil, nil, nil)
		// WHEN
		actual, err := sut.AssignFormation(context.TODO(), "", "", "UNKNOWN", model.Formation{})
		// THEN
		require.Error(t, err)
		require.Nil(t, actual)
		require.Contains(t, err.Error(), "unknown formation type")
	})
}

func fixUUID() string {
	return "003a0855-4eb0-486d-8fc6-3ab2f2312ca0"
}

func fixDefaultScenariosLabelDefinition(tenantID string, schema interface{}) model.LabelDefinition {
	return model.LabelDefinition{
		Key:     model.ScenariosKey,
		Tenant:  tenantID,
		Schema:  &schema,
		Version: 1,
	}
}
