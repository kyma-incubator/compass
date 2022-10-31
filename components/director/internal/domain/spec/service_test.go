package spec_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec/automock"
	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const tenantID = "b91b59f7-2563-40b2-aba9-fef726037aa3"

var testErr = errors.New("Test error")

func TestService_GetByID(t *testing.T) {
	testSpec := &model.Spec{}

	testCases := []struct {
		Name           string
		Context        context.Context
		SpecRepoMock   *automock.SpecRepository
		ExpectedResult *model.Spec
		ExpectedError  error
	}{
		{
			Name:    "Success",
			Context: tnt.SaveToContext(context.TODO(), tenantID, tenantID),
			SpecRepoMock: func() *automock.SpecRepository {
				specRepositoryMock := automock.SpecRepository{}
				specRepositoryMock.On("GetByID", mock.Anything, tenantID, mock.Anything, mock.Anything).Return(testSpec, nil).Once()
				return &specRepositoryMock
			}(),
			ExpectedResult: testSpec,
			ExpectedError:  nil,
		},
		{
			Name:           "Fails when tenant is missing in context",
			Context:        context.TODO(),
			SpecRepoMock:   &automock.SpecRepository{},
			ExpectedResult: nil,
			ExpectedError:  apperrors.NewCannotReadTenantError(),
		},
		{
			Name:    "Fails when repo get by id fails",
			Context: tnt.SaveToContext(context.TODO(), tenantID, tenantID),
			SpecRepoMock: func() *automock.SpecRepository {
				specRepositoryMock := automock.SpecRepository{}
				specRepositoryMock.On("GetByID", mock.Anything, tenantID, mock.Anything, mock.Anything).Return(nil, testErr).Once()
				return &specRepositoryMock
			}(),
			ExpectedResult: testSpec,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.SpecRepoMock
			svc := spec.NewService(repo, nil, nil, nil)

			// WHEN
			spec, err := svc.GetByID(testCase.Context, "123", model.APISpecReference)

			// then
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, spec)
			} else {
				require.Error(t, err)
				assert.Equal(t, err, testCase.ExpectedError)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_ListByReferenceObjectID(t *testing.T) {
	// GIVEN
	specs := []*model.Spec{
		fixModelAPISpec(),
		fixModelAPISpec(),
		fixModelAPISpec(),
	}

	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.SpecRepository
		ExpectedResult     []*model.Spec
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("ListByReferenceObjectID", ctx, tenant, model.APISpecReference, apiID).Return(specs, nil).Once()
				return repo
			},
			ExpectedResult:     specs,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition listing failed",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("ListByReferenceObjectID", ctx, tenant, model.APISpecReference, apiID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := spec.NewService(repo, nil, nil, nil)

			// WHEN
			docs, err := svc.ListByReferenceObjectID(ctx, model.APISpecReference, apiID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, docs)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := spec.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByReferenceObjectID(context.TODO(), model.APISpecReference, apiID)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByReferenceObjectIDs(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	firstAPIID := "apiID"
	secondAPIID := "apiID2"
	apiIDs := []string{firstAPIID, secondAPIID}

	specForFirstAPI := fixModelAPISpecWithID(firstAPIID)
	specForSecondAPI := fixModelAPISpecWithID(secondAPIID)
	specs := []*model.Spec{specForFirstAPI, specForSecondAPI}

	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.SpecRepository
		ExpectedResult     []*model.Spec
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("ListByReferenceObjectIDs", ctx, tenant, model.APISpecReference, apiIDs).Return(specs, nil).Once()
				return repo
			},
			ExpectedResult: specs,
		},
		{
			Name: "Returns error when Specs listing failed",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("ListByReferenceObjectIDs", ctx, tenant, model.APISpecReference, apiIDs).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := spec.NewService(repo, nil, nil, nil)

			// WHEN
			specifications, err := svc.ListByReferenceObjectIDs(ctx, model.APISpecReference, apiIDs)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, specifications)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := spec.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByReferenceObjectIDs(context.TODO(), model.APISpecReference, apiIDs)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_DeleteByReferenceObjectID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	specs := []*model.Spec{
		fixModelAPISpec(),
		fixModelAPISpec(),
		fixModelAPISpec(),
	}

	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.SpecRepository
		ExpectedResult     []*model.Spec
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenant, model.APISpecReference, apiID).Return(nil).Once()
				return repo
			},
			ExpectedResult:     specs,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition listing failed",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenant, model.APISpecReference, apiID).Return(testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := spec.NewService(repo, nil, nil, nil)

			// WHEN
			err := svc.DeleteByReferenceObjectID(ctx, model.APISpecReference, apiID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := spec.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.DeleteByReferenceObjectID(context.TODO(), model.APISpecReference, apiID)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetByReferenceObjectID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	specs := []*model.Spec{
		fixModelAPISpec(),
		fixModelAPISpec(),
		fixModelAPISpec(),
	}

	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.SpecRepository
		ExpectedResult     *model.Spec
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("ListByReferenceObjectID", ctx, tenant, model.APISpecReference, apiID).Return(specs, nil).Once()
				return repo
			},
			ExpectedResult:     specs[0],
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition listing failed",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("ListByReferenceObjectID", ctx, tenant, model.APISpecReference, apiID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns nil when no specs are found",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("ListByReferenceObjectID", ctx, tenant, model.APISpecReference, apiID).Return([]*model.Spec{}, nil).Once()
				return repo
			},
			ExpectedResult: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := spec.NewService(repo, nil, nil, nil)

			// WHEN
			docs, err := svc.GetByReferenceObjectID(ctx, model.APISpecReference, apiID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, docs)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := spec.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.GetByReferenceObjectID(context.TODO(), model.APISpecReference, apiID)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_CreateByReferenceObjectID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	specData := "specData"

	specInputWithFR := fixModelAPISpecInputWithFetchRequest()
	specInputWithFR.Data = nil

	specModel := fixModelAPISpec()
	specModel.Data = nil

	timestamp := time.Now()

	fr := &model.FetchRequest{
		ID:   specID,
		URL:  "foo.bar",
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: model.APISpecFetchRequestReference,
		ObjectID:   specID,
	}

	testCases := []struct {
		Name                  string
		RepositoryFn          func() *automock.SpecRepository
		FetchRequestRepoFn    func() *automock.FetchRequestRepository
		UIDServiceFn          func() *automock.UIDService
		FetchRequestServiceFn func() *automock.FetchRequestService
		Input                 model.SpecInput
		ExpectedErr           error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Create", ctx, tenant, specModel).Return(nil).Once()
				repo.On("Update", ctx, tenant, specModel).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, tenant, fr).Return(nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Twice()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleSpec", ctx, fr).Return(nil)
				return svc
			},
			Input:       *specInputWithFR,
			ExpectedErr: nil,
		},
		{
			Name: "Success fetched Spec",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Create", ctx, tenant, specModel).Return(nil).Once()
				repo.On("Update", ctx, tenant, fixModelAPISpec()).Return(nil).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, tenant, fr).Return(nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Twice()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleSpec", ctx, fr).Return(&specData)
				return svc
			},
			Input:       *specInputWithFR,
			ExpectedErr: nil,
		},
		{
			Name: "Error - Spec Creation",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Create", ctx, tenant, specModel).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				return &automock.FetchRequestRepository{}
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				return &automock.FetchRequestService{}
			},
			Input:       *specInputWithFR,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Fetch Request Creation",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Create", ctx, tenant, specModel).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, tenant, fr).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Twice()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				return &automock.FetchRequestService{}
			},
			Input:       *specInputWithFR,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Spec Update",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Create", ctx, tenant, specModel).Return(nil).Once()
				repo.On("Update", ctx, tenant, specModel).Return(testErr).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, tenant, fr).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Twice()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleSpec", ctx, fr).Return(nil)
				return svc
			},
			Input:       *specInputWithFR,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			uidService := testCase.UIDServiceFn()
			fetchRequestService := testCase.FetchRequestServiceFn()

			svc := spec.NewService(repo, fetchRequestRepo, uidService, fetchRequestService)
			svc.SetTimestampGen(func() time.Time {
				return timestamp
			})

			// WHEN
			result, err := svc.CreateByReferenceObjectID(ctx, testCase.Input, model.APISpecReference, apiID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}

			repo.AssertExpectations(t)
			fetchRequestRepo.AssertExpectations(t)
			uidService.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := spec.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.CreateByReferenceObjectID(context.TODO(), model.SpecInput{}, model.APISpecReference, apiID)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_CreateByReferenceObjectIDWithDelayedFetchRequest(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	specInputWithFR := fixModelAPISpecInputWithFetchRequest()
	specInputWithFR.Data = nil

	specModel := fixModelAPISpec()
	specModel.Data = nil

	timestamp := time.Now()

	fr := &model.FetchRequest{
		ID:   specID,
		URL:  "foo.bar",
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: model.APISpecFetchRequestReference,
		ObjectID:   specID,
	}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.SpecRepository
		FetchRequestRepoFn func() *automock.FetchRequestRepository
		UIDServiceFn       func() *automock.UIDService
		Input              model.SpecInput
		ExpectedErr        error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Create", ctx, tenant, specModel).Return(nil).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, tenant, fr).Return(nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Twice()
				return svc
			},
			Input:       *specInputWithFR,
			ExpectedErr: nil,
		},
		{
			Name: "Error - Spec Creation",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Create", ctx, tenant, specModel).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				return &automock.FetchRequestRepository{}
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Once()
				return svc
			},
			Input:       *specInputWithFR,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Fetch Request Creation",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Create", ctx, tenant, specModel).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, tenant, fr).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Twice()
				return svc
			},
			Input:       *specInputWithFR,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			uidService := testCase.UIDServiceFn()

			svc := spec.NewService(repo, fetchRequestRepo, uidService, nil)
			svc.SetTimestampGen(func() time.Time {
				return timestamp
			})

			// WHEN
			result, fr, err := svc.CreateByReferenceObjectIDWithDelayedFetchRequest(ctx, testCase.Input, model.APISpecReference, apiID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
				assert.NotEmpty(t, fr)
			}

			repo.AssertExpectations(t)
			fetchRequestRepo.AssertExpectations(t)
			uidService.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := spec.NewService(nil, nil, nil, nil)
		// WHEN
		_, _, err := svc.CreateByReferenceObjectIDWithDelayedFetchRequest(context.TODO(), model.SpecInput{}, model.APISpecReference, apiID)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_UpdateByReferenceObjectID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	specData := "specData"

	specInputWithFR := fixModelAPISpecInputWithFetchRequest()
	specInputWithFR.Data = nil

	specModel := fixModelAPISpec()
	specModel.Data = nil

	timestamp := time.Now()

	fr := &model.FetchRequest{
		ID:   specID,
		URL:  "foo.bar",
		Mode: model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: model.APISpecFetchRequestReference,
		ObjectID:   specID,
	}

	testCases := []struct {
		Name                  string
		RepositoryFn          func() *automock.SpecRepository
		FetchRequestRepoFn    func() *automock.FetchRequestRepository
		UIDServiceFn          func() *automock.UIDService
		FetchRequestServiceFn func() *automock.FetchRequestService
		Input                 model.SpecInput
		InputID               string
		ExpectedErr           error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("GetByID", ctx, tenant, specID, model.APISpecReference).Return(specModel, nil).Once()
				repo.On("Update", ctx, tenant, fixModelAPISpec()).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(nil).Once()
				repo.On("Create", ctx, tenant, fr).Return(nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleSpec", ctx, fr).Return(&specData)
				return svc
			},
			InputID:     specID,
			Input:       *specInputWithFR,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("GetByID", ctx, tenant, specID, model.APISpecReference).Return(specModel, nil).Once()
				repo.On("Update", ctx, tenant, specModel).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(nil).Once()
				repo.On("Create", ctx, tenant, fr).Return(nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleSpec", ctx, fr).Return(nil)
				return svc
			},
			InputID:     specID,
			Input:       *specInputWithFR,
			ExpectedErr: testErr,
		},
		{
			Name: "Delete FetchRequest by reference Error",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("GetByID", ctx, tenant, specID, model.APISpecReference).Return(specModel, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			InputID:     specID,
			Input:       *specInputWithFR,
			ExpectedErr: testErr,
		},
		{
			Name: "Fetch Request Creation Error",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("GetByID", ctx, tenant, specID, model.APISpecReference).Return(specModel, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(nil).Once()
				repo.On("Create", ctx, tenant, fr).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(specID).Once()
				return svc
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			InputID:     specID,
			Input:       *specInputWithFR,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("GetByID", ctx, tenant, specID, model.APISpecReference).Return(nil, testErr).Once()
				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			InputID:     specID,
			Input:       *specInputWithFR,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			uidSvc := testCase.UIDServiceFn()
			fetchRequestSvc := testCase.FetchRequestServiceFn()

			svc := spec.NewService(repo, fetchRequestRepo, uidSvc, fetchRequestSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// WHEN
			err := svc.UpdateByReferenceObjectID(ctx, testCase.InputID, testCase.Input, model.APISpecReference, apiID)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			fetchRequestRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := spec.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.UpdateByReferenceObjectID(context.TODO(), "", model.SpecInput{}, model.APISpecReference, apiID)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_UpdateSpecOnly(t *testing.T) {
	testSpec := &model.Spec{}

	testCases := []struct {
		Name           string
		Context        context.Context
		SpecRepoMock   *automock.SpecRepository
		ExpectedResult *model.Spec
		ExpectedError  error
	}{
		{
			Name:    "Success",
			Context: tnt.SaveToContext(context.TODO(), tenantID, tenantID),
			SpecRepoMock: func() *automock.SpecRepository {
				specRepositoryMock := automock.SpecRepository{}
				specRepositoryMock.On("Update", mock.Anything, tenantID, testSpec).Return(nil).Once()
				return &specRepositoryMock
			}(),
			ExpectedError: nil,
		},
		{
			Name:          "Fails when tenant is missing in context",
			Context:       context.TODO(),
			SpecRepoMock:  &automock.SpecRepository{},
			ExpectedError: apperrors.NewCannotReadTenantError(),
		},
		{
			Name:    "Fails when repo update fails",
			Context: tnt.SaveToContext(context.TODO(), tenantID, tenantID),
			SpecRepoMock: func() *automock.SpecRepository {
				specRepositoryMock := automock.SpecRepository{}
				specRepositoryMock.On("Update", mock.Anything, tenantID, testSpec).Return(testErr).Once()
				return &specRepositoryMock
			}(),
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.SpecRepoMock
			svc := spec.NewService(repo, nil, nil, nil)

			err := svc.UpdateSpecOnly(testCase.Context, *testSpec)
			if testCase.ExpectedError != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	id := "foo"

	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	var testCases = []struct {
		Name         string
		RepositoryFn func() *automock.SpecRepository
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Delete", ctx, tenant, id, model.APISpecReference).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Delete", ctx, tenant, id, model.APISpecReference).Return(testErr).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := spec.NewService(repo, nil, nil, nil)

			// WHEN
			err := svc.Delete(ctx, testCase.InputID, model.APISpecReference)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := spec.NewService(nil, nil, nil, nil)
		// WHEN
		err := svc.Delete(context.TODO(), "", model.APISpecReference)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_RefetchSpec(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	dataBytes := "data"
	modelSpec := &model.Spec{
		Data: &dataBytes,
	}

	timestamp := time.Now()
	fr := &model.FetchRequest{
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
	}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.SpecRepository
		FetchRequestRepoFn func() *automock.FetchRequestRepository
		FetchRequestSvcFn  func() *automock.FetchRequestService
		ExpectedAPISpec    *model.Spec
		ExpectedErr        error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("GetByID", ctx, tenant, specID, model.APISpecReference).Return(modelSpec, nil).Once()
				repo.On("Update", ctx, tenant, modelSpec).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(nil, nil)
				return repo
			},
			FetchRequestSvcFn: func() *automock.FetchRequestService {
				return &automock.FetchRequestService{}
			},
			ExpectedAPISpec: modelSpec,
			ExpectedErr:     nil,
		},
		{
			Name: "Success - fetched API Spec",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("GetByID", ctx, tenant, specID, model.APISpecReference).Return(modelSpec, nil).Once()
				repo.On("Update", ctx, tenant, modelSpec).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(fr, nil)
				return repo
			},
			FetchRequestSvcFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleSpec", ctx, fr).Return(&dataBytes)
				return svc
			},
			ExpectedAPISpec: modelSpec,
			ExpectedErr:     nil,
		},
		{
			Name: "Get from repository error",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("GetByID", ctx, tenant, specID, model.APISpecReference).Return(nil, testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				return &automock.FetchRequestRepository{}
			},
			FetchRequestSvcFn: func() *automock.FetchRequestService {
				return &automock.FetchRequestService{}
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Get fetch request error",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("GetByID", ctx, tenant, specID, model.APISpecReference).Return(modelSpec, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(nil, testErr)
				return repo
			},
			FetchRequestSvcFn: func() *automock.FetchRequestService {
				return &automock.FetchRequestService{}
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     errors.Wrapf(testErr, "while getting FetchRequest for Specification with id %q", specID),
		},
		{
			Name: "Error when updating API Definition failed",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("GetByID", ctx, tenant, specID, model.APISpecReference).Return(modelSpec, nil).Once()
				repo.On("Update", ctx, tenant, modelSpec).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(nil, nil)
				return repo
			},
			FetchRequestSvcFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     errors.Wrapf(testErr, "while updating Specification with id %q", specID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			frRepo := testCase.FetchRequestRepoFn()
			frSvc := testCase.FetchRequestSvcFn()

			svc := spec.NewService(repo, frRepo, nil, frSvc)

			// WHEN
			result, err := svc.RefetchSpec(ctx, specID, model.APISpecReference)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)

			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := spec.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.RefetchSpec(context.TODO(), "", model.APISpecReference)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetFetchRequest(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	testErr := errors.New("Test error")

	frURL := "foo.bar"

	timestamp := time.Now()
	fr := &model.FetchRequest{
		URL: frURL,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
	}

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.SpecRepository
		FetchRequestRepoFn   func() *automock.FetchRequestRepository
		InputAPIDefID        string
		ExpectedFetchRequest *model.FetchRequest
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Exists", ctx, tenant, specID, model.APISpecReference).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(fr, nil).Once()
				return repo
			},
			InputAPIDefID:        specID,
			ExpectedFetchRequest: fr,
			ExpectedErrMessage:   "",
		},
		{
			Name: "Error - Spec Not Exist",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Exists", ctx, tenant, specID, model.APISpecReference).Return(false, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				return &automock.FetchRequestRepository{}
			},
			InputAPIDefID:        specID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   fmt.Sprintf("specification with id %q doesn't exist", specID),
		},
		{
			Name: "Success - Not Found",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Exists", ctx, tenant, specID, model.APISpecReference).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(nil, apperrors.NewNotFoundError(resource.API, "")).Once()
				return repo
			},
			InputAPIDefID:        specID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   fmt.Sprintf("while getting FetchRequest by Specification with id %q", specID),
		},
		{
			Name: "Error - Get FetchRequest",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Exists", ctx, tenant, specID, model.APISpecReference).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.APISpecFetchRequestReference, specID).Return(nil, testErr).Once()
				return repo
			},
			InputAPIDefID:        specID,
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name: "Error - Spec doesn't exist",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Exists", ctx, tenant, specID, model.APISpecReference).Return(false, testErr).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			InputAPIDefID:      specID,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			svc := spec.NewService(repo, fetchRequestRepo, nil, nil)

			// WHEN
			l, err := svc.GetFetchRequest(ctx, testCase.InputAPIDefID, model.APISpecReference)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedFetchRequest)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			fetchRequestRepo.AssertExpectations(t)
		})
	}
	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := spec.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.GetFetchRequest(context.TODO(), "dd", model.APISpecReference)
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_ListFetchRequestsByReferenceObjectIDs(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	frURL := "foo.bar"
	firstFRID := "frID"
	secondFRID := "frID2"
	firstSpecID := "specID"
	secondSpecID := "specID2"
	specIDs := []string{firstSpecID, secondSpecID}
	timestamp := time.Now()

	firstFetchRequest := &model.FetchRequest{
		ID:  firstFRID,
		URL: frURL,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
	}

	secondFetchRequest := &model.FetchRequest{
		ID:  secondFRID,
		URL: frURL,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
	}
	fetchRequests := []*model.FetchRequest{firstFetchRequest, secondFetchRequest}

	ctx := context.TODO()
	ctx = tnt.SaveToContext(ctx, tenant, externalTenant)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.FetchRequestRepository
		ExpectedResult     []*model.FetchRequest
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("ListByReferenceObjectIDs", ctx, tenant, model.APISpecFetchRequestReference, specIDs).Return(fetchRequests, nil).Once()
				return repo
			},
			ExpectedResult: fetchRequests,
		},
		{
			Name: "Returns error when Fetch Requests listing failed",
			RepositoryFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("ListByReferenceObjectIDs", ctx, tenant, model.APISpecFetchRequestReference, specIDs).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := spec.NewService(nil, repo, nil, nil)

			// WHEN
			frs, err := svc.ListFetchRequestsByReferenceObjectIDs(ctx, tenant, specIDs, model.APISpecReference)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, frs)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}
