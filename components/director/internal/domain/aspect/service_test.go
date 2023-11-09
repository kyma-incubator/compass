package aspect_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect"
	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
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
		uidSvc.On("Generate").Return(aspectID)
		return uidSvc
	}

	testCases := []struct {
		Name              string
		InputResourceType resource.Type
		InputResourceID   string
		AspectInput       model.AspectInput
		AspectRepoFn      func() *automock.AspectRepository
		UIDServiceFn      func() *automock.UIDService
		ExpectedError     error
		ExpectedOutput    string
	}{
		{
			Name:              "Success with resource type Application",
			InputResourceType: resource.Application,
			InputResourceID:   "application-id",
			AspectInput:       fixAspectInputModel(),
			AspectRepoFn: func() *automock.AspectRepository {
				entityTypeRepo := &automock.AspectRepository{}
				entityTypeRepo.On("Create", ctx, tenantID, mock.Anything).Return(nil).Once()
				return entityTypeRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: aspectID,
		},
		{
			Name:              "Success with resource type ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputResourceID:   "application-template-version-id",
			AspectInput:       fixAspectInputModel(),
			AspectRepoFn: func() *automock.AspectRepository {
				entityTypeRepo := &automock.AspectRepository{}
				entityTypeRepo.On("Create", ctx, tenantID, mock.Anything).Return(nil).Once()
				return entityTypeRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: aspectID,
		},
		{
			Name:              "Fail while creating Aspect for Application",
			InputResourceType: resource.Application,
			InputResourceID:   "application-id",
			AspectInput:       fixAspectInputModel(),
			AspectRepoFn: func() *automock.AspectRepository {
				entityTypeRepo := &automock.AspectRepository{}
				entityTypeRepo.On("Create", ctx, tenantID, mock.Anything).Return(testErr).Once()
				return entityTypeRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: testErr,
		},
		{
			Name:              "Fail while creating Aspect for ApplicationTemplateVersion",
			InputResourceType: resource.ApplicationTemplateVersion,
			InputResourceID:   "application-template-version-id",
			AspectInput:       fixAspectInputModel(),
			AspectRepoFn: func() *automock.AspectRepository {
				entityTypeRepo := &automock.AspectRepository{}
				entityTypeRepo.On("Create", ctx, tenantID, mock.Anything).Return(testErr).Once()
				return entityTypeRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			aspectRepo := testCase.AspectRepoFn()
			idSvc := testCase.UIDServiceFn()
			svc := aspect.NewService(aspectRepo, idSvc)

			// WHEN
			result, err := svc.Create(ctx, testCase.InputResourceType, testCase.InputResourceID, integrationDependencyID, testCase.AspectInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, result)
			}

			aspectRepo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := aspect.NewService(nil, uid.NewService())
		// WHEN
		_, err := svc.Create(context.TODO(), "", "", "", model.AspectInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_DeleteByIntegrationDependencyID(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name          string
		AspectRepoFn  func() *automock.AspectRepository
		ExpectedError error
	}{
		{
			Name: "Success",
			AspectRepoFn: func() *automock.AspectRepository {
				entityTypeRepo := &automock.AspectRepository{}
				entityTypeRepo.On("DeleteByIntegrationDependencyID", ctx, tenantID, integrationDependencyID).Return(nil).Once()
				return entityTypeRepo
			},
			ExpectedError: nil,
		},
		{
			Name: "Error on deletion",
			AspectRepoFn: func() *automock.AspectRepository {
				entityTypeRepo := &automock.AspectRepository{}
				entityTypeRepo.On("DeleteByIntegrationDependencyID", ctx, tenantID, integrationDependencyID).Return(testErr).Once()
				return entityTypeRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			aspectRepo := testCase.AspectRepoFn()
			svc := aspect.NewService(aspectRepo, nil)

			// WHEN
			err := svc.DeleteByIntegrationDependencyID(ctx, integrationDependencyID)

			// THEN

			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			aspectRepo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := aspect.NewService(nil, uid.NewService())
		// WHEN
		err := svc.DeleteByIntegrationDependencyID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
