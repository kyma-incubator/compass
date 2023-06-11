package apptemplateversion_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplateversion"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplateversion/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	modelApplicationTemplateVersion := fixModelApplicationTemplateVersion(appTemplateVersionID)

	testCases := []struct {
		Name                     string
		Input                    *model.ApplicationTemplateVersionInput
		AppTemplateVersionRepoFn func() *automock.ApplicationTemplateVersionRepository
		UIDSvcFn                 func() *automock.UIDService
		TimeSvcFn                func() *automock.TimeService
		ExpectedError            error
		ExpectedOutput           string
	}{
		{
			Name:  "Success",
			Input: fixModelApplicationTemplateVersionInput(),
			AppTemplateVersionRepoFn: func() *automock.ApplicationTemplateVersionRepository {
				appTemplateVersionRepo := &automock.ApplicationTemplateVersionRepository{}
				appTemplateVersionRepo.On("Create", ctx, *modelApplicationTemplateVersion).Return(nil).Once()
				return appTemplateVersionRepo
			},
			UIDSvcFn:       fixUIDService,
			TimeSvcFn:      fixTimeService,
			ExpectedOutput: appTemplateVersionID,
		},
		{
			Name:                     "Nothing happens when input is empty",
			Input:                    nil,
			AppTemplateVersionRepoFn: fixEmptyappTemplateVersionRepo,
			UIDSvcFn:                 fixEmptyUIDService,
			TimeSvcFn:                fixEmptyTimeService,
			ExpectedOutput:           "",
		},
		{
			Name:  "Returns error when repo layer cannot create Application Template Version",
			Input: fixModelApplicationTemplateVersionInput(),
			AppTemplateVersionRepoFn: func() *automock.ApplicationTemplateVersionRepository {
				appTemplateVersionRepo := &automock.ApplicationTemplateVersionRepository{}
				appTemplateVersionRepo.On("Create", ctx, *modelApplicationTemplateVersion).Return(testError).Once()
				return appTemplateVersionRepo
			},
			UIDSvcFn:      fixUIDService,
			TimeSvcFn:     fixTimeService,
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateVersionRepo := testCase.AppTemplateVersionRepoFn()
			idSvc := testCase.UIDSvcFn()
			timeSvc := testCase.TimeSvcFn()
			svc := apptemplateversion.NewService(appTemplateVersionRepo, idSvc, timeSvc)

			defer mock.AssertExpectationsForObjects(t, idSvc, appTemplateVersionRepo, timeSvc)

			// WHEN
			result, err := svc.Create(ctx, appTemplateID, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	modelApplicationTemplateVersion := fixModelApplicationTemplateVersion(appTemplateVersionID)
	modelApplicationTemplateVersion.CreatedAt = time.Time{}

	testCases := []struct {
		Name                     string
		Input                    model.ApplicationTemplateVersionInput
		AppTemplateVersionRepoFn func() *automock.ApplicationTemplateVersionRepository
		ExpectedError            error
		ExpectedOutput           string
	}{
		{
			Name:  "Success",
			Input: *fixModelApplicationTemplateVersionInput(),
			AppTemplateVersionRepoFn: func() *automock.ApplicationTemplateVersionRepository {
				appTemplateVersionRepo := &automock.ApplicationTemplateVersionRepository{}
				appTemplateVersionRepo.On("Exists", ctx, appTemplateVersionID).Return(true, nil).Once()
				appTemplateVersionRepo.On("Update", ctx, *modelApplicationTemplateVersion).Return(nil).Once()
				return appTemplateVersionRepo
			},
			ExpectedOutput: appTemplateVersionID,
		},
		{
			Name:  "Returns an error when the entity does not exist",
			Input: *fixModelApplicationTemplateVersionInput(),
			AppTemplateVersionRepoFn: func() *automock.ApplicationTemplateVersionRepository {
				appTemplateVersionRepo := &automock.ApplicationTemplateVersionRepository{}
				appTemplateVersionRepo.On("Exists", ctx, appTemplateVersionID).Return(false, nil).Once()
				return appTemplateVersionRepo
			},
			ExpectedError: errors.New("Application Template Version with ID 44444444-1111-2222-3333-51d5356e7e09 does not exist"),
		},
		{
			Name:  "Returns an error when cannot determine if the entity exists",
			Input: *fixModelApplicationTemplateVersionInput(),
			AppTemplateVersionRepoFn: func() *automock.ApplicationTemplateVersionRepository {
				appTemplateVersionRepo := &automock.ApplicationTemplateVersionRepository{}
				appTemplateVersionRepo.On("Exists", ctx, appTemplateVersionID).Return(false, testError).Once()
				return appTemplateVersionRepo
			},
			ExpectedError: testError,
		},
		{
			Name:  "Returns an error when repo layer cannot update Application Template Version",
			Input: *fixModelApplicationTemplateVersionInput(),
			AppTemplateVersionRepoFn: func() *automock.ApplicationTemplateVersionRepository {
				appTemplateVersionRepo := &automock.ApplicationTemplateVersionRepository{}
				appTemplateVersionRepo.On("Exists", ctx, appTemplateVersionID).Return(true, nil).Once()
				appTemplateVersionRepo.On("Update", ctx, *modelApplicationTemplateVersion).Return(testError).Once()
				return appTemplateVersionRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateVersionRepo := testCase.AppTemplateVersionRepoFn()
			idSvc := fixEmptyUIDService()
			timeSvc := fixEmptyTimeService()
			svc := apptemplateversion.NewService(appTemplateVersionRepo, idSvc, timeSvc)

			defer mock.AssertExpectationsForObjects(t, idSvc, appTemplateVersionRepo, timeSvc)

			// WHEN
			err := svc.Update(ctx, appTemplateVersionID, appTemplateID, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetByAppTemplateIDAndVersion(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	modelApplicationTemplateVersion := fixModelApplicationTemplateVersion(appTemplateVersionID)

	testCases := []struct {
		Name                     string
		AppTemplateVersionRepoFn func() *automock.ApplicationTemplateVersionRepository
		ExpectedError            error
		ExpectedOutput           *model.ApplicationTemplateVersion
	}{
		{
			Name: "Success",
			AppTemplateVersionRepoFn: func() *automock.ApplicationTemplateVersionRepository {
				appTemplateVersionRepo := &automock.ApplicationTemplateVersionRepository{}
				appTemplateVersionRepo.On("GetByAppTemplateIDAndVersion", ctx, appTemplateID, testVersion).Return(modelApplicationTemplateVersion, nil).Once()
				return appTemplateVersionRepo
			},
			ExpectedOutput: modelApplicationTemplateVersion,
		},
		{
			Name: "Returns an error when getting the Application Template Version",
			AppTemplateVersionRepoFn: func() *automock.ApplicationTemplateVersionRepository {
				appTemplateVersionRepo := &automock.ApplicationTemplateVersionRepository{}
				appTemplateVersionRepo.On("GetByAppTemplateIDAndVersion", ctx, appTemplateID, testVersion).Return(nil, testError).Once()
				return appTemplateVersionRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateVersionRepo := testCase.AppTemplateVersionRepoFn()
			idSvc := fixEmptyUIDService()
			timeSvc := fixEmptyTimeService()
			svc := apptemplateversion.NewService(appTemplateVersionRepo, idSvc, timeSvc)

			defer mock.AssertExpectationsForObjects(t, idSvc, appTemplateVersionRepo, timeSvc)

			// WHEN
			result, err := svc.GetByAppTemplateIDAndVersion(ctx, appTemplateID, testVersion)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}

func TestService_ListByAppTemplateID(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	modelApplicationTemplateVersion := fixModelApplicationTemplateVersion(appTemplateVersionID)

	testCases := []struct {
		Name                     string
		AppTemplateVersionRepoFn func() *automock.ApplicationTemplateVersionRepository
		ExpectedError            error
		ExpectedOutput           []*model.ApplicationTemplateVersion
	}{
		{
			Name: "Success",
			AppTemplateVersionRepoFn: func() *automock.ApplicationTemplateVersionRepository {
				appTemplateVersionRepo := &automock.ApplicationTemplateVersionRepository{}
				appTemplateVersionRepo.On("ListByAppTemplateID", ctx, appTemplateID).Return([]*model.ApplicationTemplateVersion{modelApplicationTemplateVersion}, nil).Once()
				return appTemplateVersionRepo
			},
			ExpectedOutput: []*model.ApplicationTemplateVersion{modelApplicationTemplateVersion},
		},
		{
			Name: "Returns an error when listing the Application Template Versions",
			AppTemplateVersionRepoFn: func() *automock.ApplicationTemplateVersionRepository {
				appTemplateVersionRepo := &automock.ApplicationTemplateVersionRepository{}
				appTemplateVersionRepo.On("ListByAppTemplateID", ctx, appTemplateID).Return(nil, testError).Once()
				return appTemplateVersionRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateVersionRepo := testCase.AppTemplateVersionRepoFn()
			idSvc := fixEmptyUIDService()
			timeSvc := fixEmptyTimeService()
			svc := apptemplateversion.NewService(appTemplateVersionRepo, idSvc, timeSvc)

			defer mock.AssertExpectationsForObjects(t, idSvc, appTemplateVersionRepo, timeSvc)

			// WHEN
			result, err := svc.ListByAppTemplateID(ctx, appTemplateID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}

func fixUIDService() *automock.UIDService {
	uidSvc := &automock.UIDService{}
	uidSvc.On("Generate").Return(appTemplateVersionID)
	return uidSvc
}

func fixTimeService() *automock.TimeService {
	timeSvc := &automock.TimeService{}
	timeSvc.On("Now").Return(mockedTimestamp)
	return timeSvc
}

func fixEmptyappTemplateVersionRepo() *automock.ApplicationTemplateVersionRepository {
	return &automock.ApplicationTemplateVersionRepository{}
}

func fixEmptyUIDService() *automock.UIDService {
	return &automock.UIDService{}
}

func fixEmptyTimeService() *automock.TimeService {
	return &automock.TimeService{}
}
