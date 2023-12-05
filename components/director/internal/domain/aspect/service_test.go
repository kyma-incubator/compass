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

func TestService_ListByApplicationIDs(t *testing.T) {
	//GIVEN
	firstAspectID := "aspectID1"
	secondAspectID := "aspectID2"
	firstAppID := "appID1"
	secondAppID := "appID2"
	appIDs := []string{firstAppID, secondAppID}

	aspectFirstApp := fixAspectModel(firstAspectID)
	aspectFirstApp.ApplicationID = &firstAppID
	aspectSecondApp := fixAspectModel(secondAspectID)
	aspectSecondApp.ApplicationID = &secondAppID

	aspects := []*model.Aspect{aspectFirstApp, aspectSecondApp}

	numberOfAspectsInFirstApp := 1
	numberOfAspectsInSecondApp := 1
	totalCounts := map[string]int{firstAppID: numberOfAspectsInFirstApp, secondAppID: numberOfAspectsInSecondApp}

	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name                string
		PageSize            int
		RepositoryFn        func() *automock.AspectRepository
		ExpectedResult      []*model.Aspect
		ExpectedTotalCounts map[string]int
		ExpectedErrMessage  string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.AspectRepository {
				repo := &automock.AspectRepository{}
				repo.On("ListByApplicationIDs", ctx, tenantID, appIDs, 2, after).Return(aspects, totalCounts, nil).Once()
				return repo
			},
			PageSize:            2,
			ExpectedResult:      aspects,
			ExpectedTotalCounts: totalCounts,
			ExpectedErrMessage:  "",
		},
		{
			Name: "Return error when page size is less than 1",
			RepositoryFn: func() *automock.AspectRepository {
				repo := &automock.AspectRepository{}
				return repo
			},
			PageSize:           0,
			ExpectedResult:     aspects,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Return error when page size is bigger than 200",
			RepositoryFn: func() *automock.AspectRepository {
				repo := &automock.AspectRepository{}
				return repo
			},
			PageSize:           201,
			ExpectedResult:     aspects,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when Aspects listing failed",
			RepositoryFn: func() *automock.AspectRepository {
				repo := &automock.AspectRepository{}
				repo.On("ListByApplicationIDs", ctx, tenantID, appIDs, 2, after).Return(nil, nil, testErr).Once()
				return repo
			},
			PageSize:            2,
			ExpectedResult:      nil,
			ExpectedTotalCounts: nil,
			ExpectedErrMessage:  testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := aspect.NewService(repo, nil)

			// WHEN
			aspects, counts, err := svc.ListByApplicationIDs(ctx, appIDs, testCase.PageSize, after)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, aspects)
				assert.Equal(t, testCase.ExpectedTotalCounts[firstAppID], counts[firstAppID])
				assert.Equal(t, testCase.ExpectedTotalCounts[secondAppID], counts[secondAppID])
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := aspect.NewService(nil, nil)
		// WHEN
		_, _, err := svc.ListByApplicationIDs(context.TODO(), nil, 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByIntegrationDependencyID(t *testing.T) {
	// GIVEN
	aspects := []*model.Aspect{
		fixAspectModel(aspectID),
		fixAspectModel(aspectID),
		fixAspectModel(aspectID),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.AspectRepository
		ExpectedResult     []*model.Aspect
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.AspectRepository {
				repo := &automock.AspectRepository{}
				repo.On("ListByIntegrationDependencyID", ctx, tenantID, integrationDependencyID).Return(aspects, nil).Once()
				return repo
			},
			ExpectedResult:     aspects,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Aspects listing failed",
			RepositoryFn: func() *automock.AspectRepository {
				repo := &automock.AspectRepository{}
				repo.On("ListByIntegrationDependencyID", ctx, tenantID, integrationDependencyID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := aspect.NewService(repo, nil)

			// WHEN
			aspects, err := svc.ListByIntegrationDependencyID(ctx, integrationDependencyID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, aspects)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := aspect.NewService(nil, nil)
		// WHEN
		_, err := svc.ListByIntegrationDependencyID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
