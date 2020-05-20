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

func TestLabelUpsertService_UpsertMultipleLabels(t *testing.T) {
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

			svc := label.NewLabelUpsertService(labelRepo, labelDefRepo, uidService)

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

func TestLabelUpsertService_UpsertLabel(t *testing.T) {
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
				}).Return(nil).Once()
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, nil).Once()

				repo.On("Create", ctx, model.LabelDefinition{
					ID:     id,
					Tenant: tnt,
					Key:    "test",
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
			Name: "Success - LabelDefinition exists",
			LabelInput: &model.LabelInput{
				Key:        "test",
				Value:      "string",
				ObjectType: model.ApplicationLabelableObject,
				ObjectID:   "appID",
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
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelDefRepoFn: func() *automock.LabelDefinitionRepository {
				repo := &automock.LabelDefinitionRepository{}
				repo.On("GetByKey", ctx, tnt, "test").Return(nil, nil).Once()
				repo.On("Create", ctx, model.LabelDefinition{
					ID:     id,
					Tenant: tnt,
					Key:    "test",
					Schema: nil,
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

			svc := label.NewLabelUpsertService(labelRepo, labelDefRepo, uidService)

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
