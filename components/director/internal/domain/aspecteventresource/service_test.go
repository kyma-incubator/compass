package aspecteventresource_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/aspecteventresource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/aspecteventresource/automock"
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
		uidSvc.On("Generate").Return(aspectEventResourceID)
		return uidSvc
	}

	testCases := []struct {
		Name                      string
		InputResourceType         resource.Type
		InputResourceID           string
		AspectEventResourceInput  model.AspectEventResourceInput
		AspectEventResourceRepoFn func() *automock.AspectEventResourceRepository
		UIDServiceFn              func() *automock.UIDService
		ExpectedError             error
		ExpectedOutput            string
	}{
		{
			Name:                     "Success with resource type Application",
			InputResourceType:        resource.Application,
			InputResourceID:          "application-id",
			AspectEventResourceInput: fixAspectEventResourceInputModel(),
			AspectEventResourceRepoFn: func() *automock.AspectEventResourceRepository {
				aspectEventResourceRepo := &automock.AspectEventResourceRepository{}
				aspectEventResourceRepo.On("Create", ctx, tenantID, mock.Anything).Return(nil).Once()
				return aspectEventResourceRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: aspectEventResourceID,
		},
		{
			Name:                     "Success with resource type ApplicationTemplateVersion",
			InputResourceType:        resource.ApplicationTemplateVersion,
			InputResourceID:          "application-template-version-id",
			AspectEventResourceInput: fixAspectEventResourceInputModel(),
			AspectEventResourceRepoFn: func() *automock.AspectEventResourceRepository {
				aspectEventResourceRepo := &automock.AspectEventResourceRepository{}
				aspectEventResourceRepo.On("Create", ctx, tenantID, mock.Anything).Return(nil).Once()
				return aspectEventResourceRepo
			},
			UIDServiceFn:   uidSvcFn,
			ExpectedOutput: aspectEventResourceID,
		},
		{
			Name:                     "Fail while creating Aspect Event Resources for Application",
			InputResourceType:        resource.Application,
			InputResourceID:          "application-id",
			AspectEventResourceInput: fixAspectEventResourceInputModel(),
			AspectEventResourceRepoFn: func() *automock.AspectEventResourceRepository {
				aspectEventResourceRepo := &automock.AspectEventResourceRepository{}
				aspectEventResourceRepo.On("Create", ctx, tenantID, mock.Anything).Return(testErr).Once()
				return aspectEventResourceRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: testErr,
		},
		{
			Name:                     "Fail while creating Aspect Event Resources for ApplicationTemplateVersion",
			InputResourceType:        resource.ApplicationTemplateVersion,
			InputResourceID:          "application-template-version-id",
			AspectEventResourceInput: fixAspectEventResourceInputModel(),
			AspectEventResourceRepoFn: func() *automock.AspectEventResourceRepository {
				aspectEventResourceRepo := &automock.AspectEventResourceRepository{}
				aspectEventResourceRepo.On("Create", ctx, tenantID, mock.Anything).Return(testErr).Once()
				return aspectEventResourceRepo
			},
			UIDServiceFn:  uidSvcFn,
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			aspectEventResourceRepo := testCase.AspectEventResourceRepoFn()
			idSvc := testCase.UIDServiceFn()
			svc := aspecteventresource.NewService(aspectEventResourceRepo, idSvc)

			// WHEN
			result, err := svc.Create(ctx, testCase.InputResourceType, testCase.InputResourceID, aspectID, testCase.AspectEventResourceInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, result)
			}

			aspectEventResourceRepo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := aspecteventresource.NewService(nil, uid.NewService())
		// WHEN
		_, err := svc.Create(context.TODO(), "", "", "", model.AspectEventResourceInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationIDs(t *testing.T) {
	//GIVEN
	firstAspectEventResourceID := "aspectEventResourceID1"
	secondAspectEventResourceID := "aspectEventResourceID2"
	firstAppID := "appID1"
	secondAppID := "appID2"
	appIDs := []string{firstAppID, secondAppID}

	aspectEventResourceFirstApp := fixAspectEventResourceModel(firstAspectEventResourceID)
	aspectEventResourceFirstApp.ApplicationID = &firstAppID
	aspectEventResourceSecondApp := fixAspectEventResourceModel(secondAspectEventResourceID)
	aspectEventResourceSecondApp.ApplicationID = &secondAppID

	aspectEventResources := []*model.AspectEventResource{aspectEventResourceFirstApp, aspectEventResourceSecondApp}

	numberOfAspectEventResourcesInFirstApp := 1
	numberOfAspectEventResourcesInSecondApp := 1
	totalCounts := map[string]int{firstAppID: numberOfAspectEventResourcesInFirstApp, secondAppID: numberOfAspectEventResourcesInSecondApp}

	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name                string
		PageSize            int
		RepositoryFn        func() *automock.AspectEventResourceRepository
		ExpectedResult      []*model.AspectEventResource
		ExpectedTotalCounts map[string]int
		ExpectedErrMessage  string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.AspectEventResourceRepository {
				repo := &automock.AspectEventResourceRepository{}
				repo.On("ListByApplicationIDs", ctx, tenantID, appIDs, 2, after).Return(aspectEventResources, totalCounts, nil).Once()
				return repo
			},
			PageSize:            2,
			ExpectedResult:      aspectEventResources,
			ExpectedTotalCounts: totalCounts,
			ExpectedErrMessage:  "",
		},
		{
			Name: "Return error when page size is less than 1",
			RepositoryFn: func() *automock.AspectEventResourceRepository {
				repo := &automock.AspectEventResourceRepository{}
				return repo
			},
			PageSize:           0,
			ExpectedResult:     aspectEventResources,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Return error when page size is bigger than 200",
			RepositoryFn: func() *automock.AspectEventResourceRepository {
				repo := &automock.AspectEventResourceRepository{}
				return repo
			},
			PageSize:           201,
			ExpectedResult:     aspectEventResources,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when Aspect Event Resources listing failed",
			RepositoryFn: func() *automock.AspectEventResourceRepository {
				repo := &automock.AspectEventResourceRepository{}
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

			svc := aspecteventresource.NewService(repo, nil)

			// WHEN
			aspectEventResources, counts, err := svc.ListByApplicationIDs(ctx, appIDs, testCase.PageSize, after)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, aspectEventResources)
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
		svc := aspecteventresource.NewService(nil, nil)
		// WHEN
		_, _, err := svc.ListByApplicationIDs(context.TODO(), nil, 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByAspectID(t *testing.T) {
	// GIVEN
	aspectEventResources := []*model.AspectEventResource{
		fixAspectEventResourceModel(aspectEventResourceID),
		fixAspectEventResourceModel(aspectEventResourceID),
		fixAspectEventResourceModel(aspectEventResourceID),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.AspectEventResourceRepository
		ExpectedResult     []*model.AspectEventResource
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.AspectEventResourceRepository {
				repo := &automock.AspectEventResourceRepository{}
				repo.On("ListByAspectID", ctx, tenantID, aspectID).Return(aspectEventResources, nil).Once()
				return repo
			},
			ExpectedResult:     aspectEventResources,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Aspect Event Resources listing failed",
			RepositoryFn: func() *automock.AspectEventResourceRepository {
				repo := &automock.AspectEventResourceRepository{}
				repo.On("ListByAspectID", ctx, tenantID, aspectID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := aspecteventresource.NewService(repo, nil)

			// WHEN
			aspectEventResources, err := svc.ListByAspectID(ctx, aspectID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, aspectEventResources)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := aspecteventresource.NewService(nil, nil)
		// WHEN
		_, err := svc.ListByAspectID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
