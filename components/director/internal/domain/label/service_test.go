package label_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabelService_UpsertMultipleLabels(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	id := "foo"
	emptyStringSchema := &model.LabelDefinition{
		Key:    "string",
		Tenant: tnt,
		ID:     "foo",
		Schema: nil,
	}
	var jsonSchema interface{} = map[string]interface{}{
		"$id":   "https://foo.com/bar.schema.json",
		"title": "foobarbaz",
		"type":  "object",
		"properties": map[string]interface{}{
			"foo": map[string]interface{}{
				"type":        "string",
				"description": "foo",
			},
		},
		"required": []interface{}{"foo"},
	}
	objectSchema := &model.LabelDefinition{
		Key:    "object",
		Tenant: tnt,
		ID:     "foo",
		Schema: &jsonSchema,
	}

	notFoundErr := errors.New("Label not found")
	testErr := errors.New("Test error")

	runtimeType := model.RuntimeLabelableObject
	runtimeID := "bar"

	stringValue := "lorem ipsum"
	arrayValue := []interface{}{"foo", "bar"}
	objectValue := map[string]interface{}{
		"foo": "bar",
	}

	testCases := []struct {
		Name           string
		LabelRepoFn    func() *automock.LabelRepository
		LabelDefRepoFn func() *automock.LabelDefinitionRepository
		UIDServiceFn   func() *automock.UIDService

		InputObjectType model.LabelableObject
		InputObjectID   string
		InputLabels     map[string]interface{}

		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			InputLabels: map[string]interface{}{
				"string": stringValue,
				"array":  arrayValue,
				"object": objectValue,
			},
			InputObjectID:   runtimeID,
			InputObjectType: runtimeType,
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Upsert", ctx, &model.Label{
					ID: id, Tenant: tnt, ObjectType: runtimeType, ObjectID: runtimeID, Key: "object", Value: objectValue,
				}).Return(nil).Once()
				repo.On("Upsert", ctx, &model.Label{
					ID: id, Tenant: tnt, ObjectType: runtimeType, ObjectID: runtimeID, Key: "string", Value: stringValue,
				}).Return(nil).Once()
				repo.On("Upsert", ctx, &model.Label{
					ID: id, Tenant: tnt, ObjectType: runtimeType, ObjectID: runtimeID, Key: "array", Value: arrayValue,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "object").Return(objectSchema, nil).Once()
				repo.On("GetByKey", ctx, tnt, "string").Return(emptyStringSchema, nil).Once()
				repo.On("GetByKey", ctx, tnt, "array").Return(nil, nil).Once()

				repo.On("Create", ctx, model.LabelDefinition{
					ID:     id,
					Tenant: tnt,
					Key:    "array",
					Schema: nil,
				}).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},

			ExpectedErrMessage: "",
		},
		{
			Name: "Error",
			InputLabels: map[string]interface{}{
				"object": objectValue,
				"string": stringValue,
			},
			InputObjectID:   runtimeID,
			InputObjectType: runtimeType,
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, runtimeType, runtimeID, "object").Return(nil, notFoundErr).Maybe()
				repo.On("GetByKey", ctx, tnt, runtimeType, runtimeID, "string").Return(nil, notFoundErr).Maybe()

				repo.On("Upsert", ctx, &model.Label{
					ID: id, Tenant: tnt, ObjectType: runtimeType, ObjectID: runtimeID, Key: "object", Value: objectValue,
				}).Return(testErr).Maybe()
				repo.On("Upsert", ctx, &model.Label{
					ID: id, Tenant: tnt, ObjectType: runtimeType, ObjectID: runtimeID, Key: "string", Value: stringValue,
				}).Return(nil).Maybe()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "object").Return(objectSchema, nil).Maybe()
				repo.On("GetByKey", ctx, tnt, "string").Return(emptyStringSchema, nil).Maybe()

				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},

			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelRepo := testCase.LabelRepoFn()
			labelDefRepo := testCase.LabelDefRepoFn()
			uidService := testCase.UIDServiceFn()

			svc := label.NewLabelService(labelRepo, labelDefRepo, uidService)

			// when
			err := svc.UpsertMultipleLabels(ctx, tnt, testCase.InputObjectType, testCase.InputObjectID, testCase.InputLabels)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			labelRepo.AssertExpectations(t)
			labelDefRepo.AssertExpectations(t)
			uidService.AssertExpectations(t)
		})
	}
}

func TestLabelService_UpsertLabel(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	id := "foo"
	version := 0
	emptyStringSchema := &model.LabelDefinition{
		Key:     "string",
		Tenant:  tnt,
		ID:      "foo",
		Schema:  nil,
		Version: version,
	}
	var jsonSchema interface{} = map[string]interface{}{
		"$id":   "https://foo.com/bar.schema.json",
		"title": "foobarbaz",
		"type":  "object",
		"properties": map[string]interface{}{
			"foo": map[string]interface{}{
				"type":        "string",
				"description": "foo",
			},
		},
		"required": []interface{}{"foo"},
	}
	objectSchema := &model.LabelDefinition{
		Key:     "object",
		Tenant:  tnt,
		ID:      "foo",
		Schema:  &jsonSchema,
		Version: version,
	}

	testErr := errors.New("Test error")

	objectValue := map[string]interface{}{
		"foo": "bar",
	}

	testCases := []struct {
		Name           string
		LabelRepoFn    func() *automock.LabelRepository
		LabelDefRepoFn func() *automock.LabelDefinitionRepository
		UIDServiceFn   func() *automock.UIDService

		LabelInput *model.LabelInput

		ExpectedErrMessage string
	}{
		{
			Name: "Success - No LabelDefinition",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      "string",
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Upsert", ctx, &model.Label{
					Key:        "test",
					Value:      "string",
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, nil).Once()

				repo.On("Create", ctx, model.LabelDefinition{
					ID:      id,
					Tenant:  tnt,
					Key:     "test",
					Schema:  nil,
					Version: version,
				}).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Success - LabelDefinition exists",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      "string",
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Upsert", ctx, &model.Label{
					Key:        "test",
					Value:      "string",
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(emptyStringSchema, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Success - Overwrite value",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      "string",
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Upsert", ctx, &model.Label{
					Key:        "test",
					Value:      "string",
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(emptyStringSchema, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Success - Validate value",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      objectValue,
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Upsert", ctx, &model.Label{
					Key:        "test",
					Value:      objectValue,
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(objectSchema, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Error - Validate value",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(objectSchema, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErrMessage: "Invalid type",
		},
		{
			Name: "Error - Creating new LabelDefinition",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, nil).Once()
				repo.On("Create", ctx, model.LabelDefinition{
					ID:      id,
					Tenant:  tnt,
					Key:     "test",
					Schema:  nil,
					Version: version,
				}).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "Test error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelRepo := testCase.LabelRepoFn()
			labelDefRepo := testCase.LabelDefRepoFn()
			uidService := testCase.UIDServiceFn()

			svc := label.NewLabelService(labelRepo, labelDefRepo, uidService)

			// when
			err := svc.UpsertLabel(ctx, tnt, testCase.LabelInput)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			labelRepo.AssertExpectations(t)
			labelDefRepo.AssertExpectations(t)
			uidService.AssertExpectations(t)
		})
	}
}

func TestLabelService_CreateLabel(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	id := "foo"
	version := 0
	emptyStringSchema := &model.LabelDefinition{
		Key:     "string",
		Tenant:  tnt,
		ID:      "foo",
		Schema:  nil,
		Version: version,
	}
	var jsonSchema interface{} = map[string]interface{}{
		"$id":   "https://foo.com/bar.schema.json",
		"title": "foobarbaz",
		"type":  "object",
		"properties": map[string]interface{}{
			"foo": map[string]interface{}{
				"type":        "string",
				"description": "foo",
			},
		},
		"required": []interface{}{"foo"},
	}
	objectSchema := &model.LabelDefinition{
		Key:     "object",
		Tenant:  tnt,
		ID:      "foo",
		Schema:  &jsonSchema,
		Version: version,
	}

	testErr := errors.New("Test error")

	objectValue := map[string]interface{}{
		"foo": "bar",
	}

	testCases := []struct {
		Name           string
		LabelRepoFn    func() *automock.LabelRepository
		LabelDefRepoFn func() *automock.LabelDefinitionRepository
		UIDServiceFn   func() *automock.UIDService

		LabelInput *model.LabelInput

		ExpectedErrMessage string
	}{
		{
			Name: "Success - No LabelDefinition exists",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      "string",
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Create", ctx, &model.Label{
					Key:        "test",
					Value:      "string",
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, nil).Once()

				repo.On("Create", ctx, model.LabelDefinition{
					ID:      id,
					Tenant:  tnt,
					Key:     "test",
					Schema:  nil,
					Version: version,
				}).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Success - LabelDefinition exists and is with empty schema",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      "string",
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Create", ctx, &model.Label{
					Key:        "test",
					Value:      "string",
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(emptyStringSchema, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Success - LabelDefinition exists and schema is valid",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      objectValue,
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Create", ctx, &model.Label{
					Key:        "test",
					Value:      objectValue,
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(objectSchema, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Error - labelInput is not valid against schema",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(objectSchema, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErrMessage: "Invalid type",
		},
		{
			Name: "Error - Creating new LabelDefinition",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, nil).Once()
				repo.On("Create", ctx, model.LabelDefinition{
					ID:      id,
					Tenant:  tnt,
					Key:     "test",
					Schema:  nil,
					Version: version,
				}).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "Test error",
		},
		{
			Name: "Error - Creating new Label",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Create", ctx, &model.Label{
					Key:        "test",
					Value:      []interface{}{"test"},
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(testErr)
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, nil).Once()
				repo.On("Create", ctx, model.LabelDefinition{
					ID:      id,
					Tenant:  tnt,
					Key:     "test",
					Schema:  nil,
					Version: version,
				}).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "Test error",
		},
		{
			Name: "Error while reading LabelDefinition value",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErrMessage: "while reading LabelDefinition",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelRepo := testCase.LabelRepoFn()
			labelDefRepo := testCase.LabelDefRepoFn()
			uidService := testCase.UIDServiceFn()

			svc := label.NewLabelService(labelRepo, labelDefRepo, uidService)

			// when
			err := svc.CreateLabel(ctx, tnt, id, testCase.LabelInput)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			labelRepo.AssertExpectations(t)
			labelDefRepo.AssertExpectations(t)
			uidService.AssertExpectations(t)
		})
	}
}

func TestLabelService_UpdateLabel(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	id := "foo"
	version := 0
	emptyStringSchema := &model.LabelDefinition{
		Key:     "string",
		Tenant:  tnt,
		ID:      "foo",
		Schema:  nil,
		Version: version,
	}
	var jsonSchema interface{} = map[string]interface{}{
		"$id":   "https://foo.com/bar.schema.json",
		"title": "foobarbaz",
		"type":  "object",
		"properties": map[string]interface{}{
			"foo": map[string]interface{}{
				"type":        "string",
				"description": "foo",
			},
		},
		"required": []interface{}{"foo"},
	}
	objectSchema := &model.LabelDefinition{
		Key:     "object",
		Tenant:  tnt,
		ID:      "foo",
		Schema:  &jsonSchema,
		Version: version,
	}

	testErr := errors.New("Test error")

	objectValue := map[string]interface{}{
		"foo": "bar",
	}

	testCases := []struct {
		Name           string
		LabelRepoFn    func() *automock.LabelRepository
		LabelDefRepoFn func() *automock.LabelDefinitionRepository
		UIDServiceFn   func() *automock.UIDService

		LabelInput *model.LabelInput

		ExpectedErrMessage string
	}{
		{
			Name: "Success - No LabelDefinition exists",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      "string",
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("UpdateWithVersion", ctx, &model.Label{
					Key:        "test",
					Value:      "string",
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, nil).Once()

				repo.On("Create", ctx, model.LabelDefinition{
					ID:      id,
					Tenant:  tnt,
					Key:     "test",
					Schema:  nil,
					Version: version,
				}).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Success - LabelDefinition exists and is with empty schema",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      "string",
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("UpdateWithVersion", ctx, &model.Label{
					Key:        "test",
					Value:      "string",
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(emptyStringSchema, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Success - LabelDefinition exists and schema is valid",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      objectValue,
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("UpdateWithVersion", ctx, &model.Label{
					Key:        "test",
					Value:      objectValue,
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(objectSchema, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Error - labelInput is not valid against schema",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(objectSchema, nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErrMessage: "Invalid type",
		},
		{
			Name: "Error - Creating new LabelDefinition",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, nil).Once()
				repo.On("Create", ctx, model.LabelDefinition{
					ID:      id,
					Tenant:  tnt,
					Key:     "test",
					Schema:  nil,
					Version: version,
				}).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "Test error",
		},
		{
			Name: "Error - Updating Label",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("UpdateWithVersion", ctx, &model.Label{
					Key:        "test",
					Value:      []interface{}{"test"},
					ObjectType: model.ApplicationLabelableObject,
					ObjectID:   "appID",
					Tenant:     tnt,
					ID:         id,
					Version:    version,
				}).Return(testErr)
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, nil).Once()
				repo.On("Create", ctx, model.LabelDefinition{
					ID:      id,
					Tenant:  tnt,
					Key:     "test",
					Schema:  nil,
					Version: version,
				}).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			ExpectedErrMessage: "Test error",
		},
		{
			Name: "Error while reading LabelDefinition value",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			ExpectedErrMessage: "while reading LabelDefinition",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelRepo := testCase.LabelRepoFn()
			labelDefRepo := testCase.LabelDefRepoFn()
			uidService := testCase.UIDServiceFn()

			svc := label.NewLabelService(labelRepo, labelDefRepo, uidService)

			// when
			err := svc.UpdateLabel(ctx, tnt, id, testCase.LabelInput)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			labelRepo.AssertExpectations(t)
			labelDefRepo.AssertExpectations(t)
			uidService.AssertExpectations(t)
		})
	}
}

func TestLabelService_GetLabel(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tenant"
	id := "foo"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	version := 0

	testErr := errors.New("Test error")

	testCases := []struct {
		Name           string
		LabelRepoFn    func() *automock.LabelRepository
		LabelDefRepoFn func() *automock.LabelDefinitionRepository
		UIDServiceFn   func() *automock.UIDService

		LabelInput *model.LabelInput

		ExpectedLabel      *model.Label
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, "appID", "test").Return(&model.Label{
					ID:         id,
					Tenant:     tnt,
					Key:        "test",
					Value:      []interface{}{"test"},
					ObjectID:   "appID",
					ObjectType: model.ApplicationLabelableObject,
					Version:    version,
				}, nil)
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},

			ExpectedLabel: &model.Label{
				ID:         id,
				Tenant:     tnt,
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectID:   "appID",
				ObjectType: model.ApplicationLabelableObject,
				Version:    version,
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Error while getting Label",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      []interface{}{"test"},
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
				Version:    version,
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.ApplicationLabelableObject, "appID", "test").Return(nil, testErr)
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},

			ExpectedLabel:      nil,
			ExpectedErrMessage: "Test error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelRepo := testCase.LabelRepoFn()
			labelDefRepo := testCase.LabelDefRepoFn()
			uidService := testCase.UIDServiceFn()

			svc := label.NewLabelService(labelRepo, labelDefRepo, uidService)

			// when
			lbl, err := svc.GetLabel(ctx, tnt, testCase.LabelInput)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedLabel, lbl)
			} else {
				require.Error(t, err)
				require.Nil(t, lbl)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			labelRepo.AssertExpectations(t)
			labelDefRepo.AssertExpectations(t)
			uidService.AssertExpectations(t)
		})
	}
}
