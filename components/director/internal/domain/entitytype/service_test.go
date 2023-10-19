package entitytype_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytype"
	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytype/automock"
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
		uidSvc.On("Generate").Return(entityTypeID)
		return uidSvc
	}

	testCases := []struct {
		Name                 string
		InputResourceType    resource.Type
		InputResourceID      string
		InputEntityTypeInput model.EntityTypeInput
		EntityTypeRepoFn     func() *automock.EntityTypeRepository
		UIDServiceFn         func() *automock.UIDService
		ExpectedError        error
		ExpectedOutput       string
	}{
		{
			Name:                 "Success with resource type Application",
			InputResourceType:    resource.Application,
			InputResourceID:      "application-id",
			InputEntityTypeInput: fixEntityTypeInputModel(),
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("Create", ctx, tenantID, mock.Anything).Return(nil).Once()
				return entityTypeRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: entityTypeID,
		},
		{
			Name:                 "Success with resource type ApplicationTemplateVersion",
			InputResourceType:    resource.ApplicationTemplateVersion,
			InputResourceID:      "application-template-version-id",
			InputEntityTypeInput: fixEntityTypeInputModel(),
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("CreateGlobal", ctx, mock.Anything).Return(nil).Once()
				return entityTypeRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: entityTypeID,
		},
		{
			Name:                 "fail while creating entity type for Application",
			InputResourceType:    resource.Application,
			InputResourceID:      "application-id",
			InputEntityTypeInput: fixEntityTypeInputModel(),
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("Create", ctx, tenantID, mock.Anything).Return(errTest).Once()
				return entityTypeRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: errTest,
		},
		{
			Name:                 "fail while creating entity type for ApplicationTemplateVersion",
			InputResourceType:    resource.ApplicationTemplateVersion,
			InputResourceID:      "application-template-version-id",
			InputEntityTypeInput: fixEntityTypeInputModel(),
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("CreateGlobal", ctx, mock.Anything).Return(errTest).Once()
				return entityTypeRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeRepo := testCase.EntityTypeRepoFn()
			idSvc := testCase.UIDServiceFn()
			svc := entitytype.NewService(entityTypeRepo, idSvc)

			// WHEN
			result, err := svc.Create(ctx, testCase.InputResourceType, testCase.InputResourceID, testCase.InputEntityTypeInput, 123)

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

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	testCases := []struct {
		Name              string
		InputResourceType resource.Type
		InputID           string
		EntityTypeInput   model.EntityTypeInput
		EntityTypeRepoFn  func() *automock.EntityTypeRepository
		ExpectedError     error
		ExpectedOutput    string
	}{
		{
			Name:              "Success with resource type Application",
			InputResourceType: resource.Application,
			InputID:           entityTypeID,
			EntityTypeInput:   fixEntityTypeInputModel(),
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("GetByID", ctx, tenantID, entityTypeID).Return(fixEntityTypeModel(entityTypeID), nil).Once()
				entityTypeRepo.On("Update", ctx, tenantID, mock.Anything).Return(nil).Once()
				return entityTypeRepo
			},
			ExpectedOutput: entityTypeID,
		},
		{
			Name:              "Success with resource type ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           entityTypeID,
			EntityTypeInput:   fixEntityTypeInputModel(),
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("GetByIDGlobal", ctx, entityTypeID).Return(fixEntityTypeModel(entityTypeID), nil).Once()
				entityTypeRepo.On("UpdateGlobal", ctx, mock.Anything).Return(nil).Once()
				return entityTypeRepo
			},
			ExpectedOutput: entityTypeID,
		},
		{
			Name:              "fail while getting entity type by id for Application",
			InputResourceType: resource.Application,
			InputID:           entityTypeID,
			EntityTypeInput:   fixEntityTypeInputModel(),
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("GetByID", ctx, tenantID, entityTypeID).Return(nil, errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
		{
			Name:              "fail while getting entity type by id for ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           entityTypeID,
			EntityTypeInput:   fixEntityTypeInputModel(),
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("GetByIDGlobal", ctx, entityTypeID).Return(nil, errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
		{
			Name:              "fail while updating entity type for Application",
			InputResourceType: resource.Application,
			InputID:           entityTypeID,
			EntityTypeInput:   fixEntityTypeInputModel(),
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("GetByID", ctx, tenantID, entityTypeID).Return(fixEntityTypeModel(entityTypeID), nil).Once()
				entityTypeRepo.On("Update", ctx, tenantID, mock.Anything).Return(errTest).Once()

				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
		{
			Name:              "fail while updating entity type for ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           entityTypeID,
			EntityTypeInput:   fixEntityTypeInputModel(),
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("GetByIDGlobal", ctx, entityTypeID).Return(fixEntityTypeModel(entityTypeID), nil).Once()
				entityTypeRepo.On("UpdateGlobal", ctx, mock.Anything).Return(errTest).Once()

				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeRepo := testCase.EntityTypeRepoFn()
			svc := entitytype.NewService(entityTypeRepo, nil)

			// WHEN
			err := svc.Update(ctx, testCase.InputResourceType, testCase.InputID, testCase.EntityTypeInput, 123)

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

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name              string
		InputResourceType resource.Type
		InputID           string
		EntityTypeRepoFn  func() *automock.EntityTypeRepository
		ExpectedError     error
	}{
		{
			Name:              "Success with resource type Application",
			InputResourceType: resource.Application,
			InputID:           entityTypeID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("Delete", ctx, tenantID, entityTypeID).Return(nil).Once()
				return entityTypeRepo
			},
		},
		{
			Name:              "Success with resource type ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           entityTypeID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("DeleteGlobal", ctx, entityTypeID).Return(nil).Once()
				return entityTypeRepo
			},
		},
		{
			Name:              "fail while deleting entity type for Application",
			InputResourceType: resource.Application,
			InputID:           entityTypeID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("Delete", ctx, tenantID, entityTypeID).Return(errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
		{
			Name:              "fail while deleting entity type for ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputID:           entityTypeID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("DeleteGlobal", ctx, entityTypeID).Return(errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeRepo := testCase.EntityTypeRepoFn()
			svc := entitytype.NewService(entityTypeRepo, nil)

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
		Name             string
		InputID          string
		EntityTypeRepoFn func() *automock.EntityTypeRepository
		ExpectedOutput   *model.EntityType
		ExpectedError    error
	}{
		{
			Name:    "Success",
			InputID: entityTypeID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("GetByID", ctx, tenantID, entityTypeID).Return(fixEntityTypeModel(entityTypeID), nil).Once()
				return entityTypeRepo
			},
			ExpectedOutput: fixEntityTypeModel(entityTypeID),
		},
		{
			Name:    "fail while getting entity type",
			InputID: entityTypeID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("GetByID", ctx, tenantID, entityTypeID).Return(nil, errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeRepo := testCase.EntityTypeRepoFn()
			svc := entitytype.NewService(entityTypeRepo, nil)

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

func TestService_Exists(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name             string
		InputID          string
		EntityTypeRepoFn func() *automock.EntityTypeRepository
		ExpectedOutput   bool
		ExpectedError    error
	}{
		{
			Name:    "Success - the resource exists",
			InputID: entityTypeID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("Exists", ctx, tenantID, entityTypeID).Return(true, nil).Once()
				return entityTypeRepo
			},
			ExpectedOutput: true,
		},
		{
			Name:    "Success - the resource does not exist",
			InputID: entityTypeID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("Exists", ctx, tenantID, entityTypeID).Return(false, nil).Once()
				return entityTypeRepo
			},
			ExpectedOutput: false,
		},
		{
			Name:    "fail while checking if entity type exists",
			InputID: entityTypeID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("Exists", ctx, tenantID, entityTypeID).Return(false, errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeRepo := testCase.EntityTypeRepoFn()
			svc := entitytype.NewService(entityTypeRepo, nil)

			// WHEN
			exists, err := svc.Exist(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, exists)
			}

			entityTypeRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListByApplicationID(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	entityTypes := []*model.EntityType{fixEntityTypeModel(entityTypeID)}
	applicationID := "application-id"
	testCases := []struct {
		Name             string
		InputID          string
		EntityTypeRepoFn func() *automock.EntityTypeRepository
		ExpectedOutput   []*model.EntityType
		ExpectedError    error
	}{
		{
			Name:    "Success",
			InputID: applicationID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("ListByResourceID", ctx, tenantID, applicationID, resource.Application).Return(entityTypes, nil).Once()
				return entityTypeRepo
			},
			ExpectedOutput: entityTypes,
		},
		{
			Name:    "fail while listing by resource id",
			InputID: applicationID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("ListByResourceID", ctx, tenantID, applicationID, resource.Application).Return(nil, errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeRepo := testCase.EntityTypeRepoFn()
			svc := entitytype.NewService(entityTypeRepo, nil)

			// WHEN
			entityTypes, err := svc.ListByApplicationID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, entityTypes)
			}

			entityTypeRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListByApplicationTemplateVersionID(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	entityTypes := []*model.EntityType{fixEntityTypeModel(entityTypeID)}
	applicationTemplateVersionID := "application-template-version-id"
	testCases := []struct {
		Name             string
		InputID          string
		EntityTypeRepoFn func() *automock.EntityTypeRepository
		ExpectedOutput   []*model.EntityType
		ExpectedError    error
	}{
		{
			Name:    "Success",
			InputID: applicationTemplateVersionID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("ListByResourceID", ctx, "", applicationTemplateVersionID, resource.ApplicationTemplateVersion).Return(entityTypes, nil).Once()
				return entityTypeRepo
			},
			ExpectedOutput: entityTypes,
		},
		{
			Name:    "fail while listing by resource id",
			InputID: applicationTemplateVersionID,
			EntityTypeRepoFn: func() *automock.EntityTypeRepository {
				entityTypeRepo := &automock.EntityTypeRepository{}
				entityTypeRepo.On("ListByResourceID", ctx, "", applicationTemplateVersionID, resource.ApplicationTemplateVersion).Return(nil, errTest).Once()
				return entityTypeRepo
			},
			ExpectedError: errTest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			entityTypeRepo := testCase.EntityTypeRepoFn()
			svc := entitytype.NewService(entityTypeRepo, nil)

			// WHEN
			entityTypes, err := svc.ListByApplicationTemplateVersionID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, entityTypes)
			}

			entityTypeRepo.AssertExpectations(t)
		})
	}
}
