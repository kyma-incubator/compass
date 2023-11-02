package entitytypemapping_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytypemapping"
	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytypemapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(entityTypeMappingID)
		return uidSvc
	}

	testCases := []struct {
		Name                        string
		InputResourceType           resource.Type
		InputResourceID             string
		InputEntityTypeMappingInput model.EntityTypeMappingInput
		EntityTypeMappingRepoFn     func() *automock.EntityTypeMappingRepository
		UIDServiceFn                func() *automock.UIDService
		ExpectedError               error
		ExpectedOutput              string
	}{
		{
			Name:                        "Success with resource type API",
			InputResourceType:           resource.API,
			InputResourceID:             testAPIDefinitionID,
			InputEntityTypeMappingInput: fixEntityTypeMappingInputModel(),
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("Create", ctx, tenantID, mock.Anything).Return(nil).Once()
				return entityTypeRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: entityTypeMappingID,
		},
		{
			Name:                        "Success with resource type Event",
			InputResourceType:           resource.EventDefinition,
			InputResourceID:             testEventDefinitionID,
			InputEntityTypeMappingInput: fixEntityTypeMappingInputModel(),
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("Create", ctx, tenantID, mock.Anything).Return(nil).Once()
				return entityTypeRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: entityTypeMappingID,
		},
		{
			Name:                        "fail while creating entity type for Application",
			InputResourceType:           resource.Application,
			InputResourceID:             "application-id",
			InputEntityTypeMappingInput: fixEntityTypeMappingInputModel(),
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("Create", ctx, tenantID, mock.Anything).Return(errTest).Once()
				return entityTypeRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: errTest,
		},
		{
			Name:                        "fail while creating entity type for ApplicationTemplateVersion",
			InputResourceType:           resource.ApplicationTemplateVersion,
			InputResourceID:             "application-template-version-id",
			InputEntityTypeMappingInput: fixEntityTypeMappingInputModel(),
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("CreateGlobal", ctx, mock.Anything).Return(errTest).Once()
				return entityTypeRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeRepo := testCase.EntityTypeMappingRepoFn()
			idSvc := testCase.UIDServiceFn()
			svc := entitytypemapping.NewService(entityTypeRepo, idSvc)

			// WHEN
			result, err := svc.Create(ctx, testCase.InputResourceType, testCase.InputResourceID, testCase.InputEntityTypeMappingInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, result)
			}

			entityTypeRepo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name                    string
		InputResourceType       resource.Type
		InputID                 string
		EntityTypeMappingRepoFn func() *automock.EntityTypeMappingRepository
		ExpectedError           error
	}{
		{
			Name:              "Success with resource type Application",
			InputResourceType: resource.API,
			InputID:           entityTypeMappingID,
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("Delete", ctx, tenantID, entityTypeMappingID).Return(nil).Once()
				return entityTypeRepo
			},
		},
		{
			Name:              "Success with resource type Event",
			InputResourceType: resource.EventDefinition,
			InputID:           entityTypeMappingID,
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("Delete", ctx, tenantID, entityTypeMappingID).Return(nil).Once()
				return entityTypeRepo
			},
		},
		{
			Name:              "fail while deleting entity type for API",
			InputResourceType: resource.API,
			InputID:           entityTypeMappingID,
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("Delete", ctx, tenantID, entityTypeMappingID).Return(errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
		{
			Name:              "fail while deleting entity type for Event",
			InputResourceType: resource.EventDefinition,
			InputID:           entityTypeMappingID,
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("Delete", ctx, tenantID, entityTypeMappingID).Return(errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeRepo := testCase.EntityTypeMappingRepoFn()
			svc := entitytypemapping.NewService(entityTypeRepo, nil)

			// WHEN
			err := svc.Delete(ctx, testCase.InputResourceType, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			entityTypeRepo.AssertExpectations(t)
		})
	}
}

func TestService_Get(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name                    string
		InputID                 string
		EntityTypeMappingRepoFn func() *automock.EntityTypeMappingRepository
		ExpectedOutput          *model.EntityTypeMapping
		ExpectedError           error
	}{
		{
			Name:    "Success",
			InputID: entityTypeMappingID,
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("GetByID", ctx, tenantID, entityTypeMappingID).Return(fixEntityTypeMappingModel(entityTypeMappingID), nil).Once()
				return entityTypeRepo
			},
			ExpectedOutput: fixEntityTypeMappingModel(entityTypeMappingID),
		},
		{
			Name:    "fail while getting entity type",
			InputID: entityTypeMappingID,
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("GetByID", ctx, tenantID, entityTypeMappingID).Return(nil, errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeRepo := testCase.EntityTypeMappingRepoFn()
			svc := entitytypemapping.NewService(entityTypeRepo, nil)

			// WHEN
			entityType, err := svc.Get(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, entityType)
			}

			entityTypeRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListByAPIDefinitionID(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	entityTypeMappings := []*model.EntityTypeMapping{fixEntityTypeMappingModel(entityTypeMappingID)}
	testCases := []struct {
		Name                    string
		InputID                 string
		EntityTypeMappingRepoFn func() *automock.EntityTypeMappingRepository
		ExpectedOutput          []*model.EntityTypeMapping
		ExpectedError           error
	}{
		{
			Name:    "Success",
			InputID: testAPIDefinitionID,
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("ListByResourceID", ctx, tenantID, testAPIDefinitionID, resource.API).Return(entityTypeMappings, nil).Once()
				return entityTypeRepo
			},
			ExpectedOutput: entityTypeMappings,
		},
		{
			Name:    "fail while listing by resource id",
			InputID: testAPIDefinitionID,
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeMappingRepo := &automock.EntityTypeMappingRepository{}
				entityTypeMappingRepo.On("ListByResourceID", ctx, tenantID, testAPIDefinitionID, resource.API).Return(nil, errTest).Once()
				return entityTypeMappingRepo
			},
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeMappingRepo := testCase.EntityTypeMappingRepoFn()
			svc := entitytypemapping.NewService(entityTypeMappingRepo, nil)

			// WHEN
			entityTypeMappings, err := svc.ListByAPIDefinitionID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, entityTypeMappings)
			}

			entityTypeMappingRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListByEventDefinitionID(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	entityTypeMappings := []*model.EntityTypeMapping{fixEntityTypeMappingModel(entityTypeMappingID)}
	testCases := []struct {
		Name                    string
		InputID                 string
		EntityTypeMappingRepoFn func() *automock.EntityTypeMappingRepository
		ExpectedOutput          []*model.EntityTypeMapping
		ExpectedError           error
	}{
		{
			Name:    "Success",
			InputID: testEventDefinitionID,
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeRepo := &automock.EntityTypeMappingRepository{}
				entityTypeRepo.On("ListByResourceID", ctx, tenantID, testEventDefinitionID, resource.EventDefinition).Return(entityTypeMappings, nil).Once()
				return entityTypeRepo
			},
			ExpectedOutput: entityTypeMappings,
		},
		{
			Name:    "fail while listing by resource id",
			InputID: testEventDefinitionID,
			EntityTypeMappingRepoFn: func() *automock.EntityTypeMappingRepository {
				entityTypeMappingRepo := &automock.EntityTypeMappingRepository{}
				entityTypeMappingRepo.On("ListByResourceID", ctx, tenantID, testEventDefinitionID, resource.EventDefinition).Return(nil, errTest).Once()
				return entityTypeMappingRepo
			},
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeMappingRepo := testCase.EntityTypeMappingRepoFn()
			svc := entitytypemapping.NewService(entityTypeMappingRepo, nil)

			// WHEN
			entityTypeMappings, err := svc.ListByEventDefinitionID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, entityTypeMappings)
			}

			entityTypeMappingRepo.AssertExpectations(t)
		})
	}
}
