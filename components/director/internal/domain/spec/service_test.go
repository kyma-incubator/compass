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
	"github.com/stretchr/testify/require"
)

func TestService_ListByReferenceObjectID(t *testing.T) {
	// given
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

			// when
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

func TestService_GetByReferenceObjectID(t *testing.T) {
	// given
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

			// when
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
	// given
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
		ID:     specID,
		Tenant: tenant,
		URL:    "foo.bar",
		Mode:   model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: model.SpecFetchRequestReference,
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
				repo.On("Create", ctx, specModel).Return(nil).Once()
				repo.On("Update", ctx, specModel).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, fr).Return(nil).Once()

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
				repo.On("Create", ctx, specModel).Return(nil).Once()
				repo.On("Update", ctx, fixModelAPISpec()).Return(nil).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, fr).Return(nil).Once()

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
				repo.On("Create", ctx, specModel).Return(testErr).Once()
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
				repo.On("Create", ctx, specModel).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, fr).Return(testErr).Once()
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
				repo.On("Create", ctx, specModel).Return(nil).Once()
				repo.On("Update", ctx, specModel).Return(testErr).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, fr).Return(nil).Once()
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
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			uidService := testCase.UIDServiceFn()
			fetchRequestService := testCase.FetchRequestServiceFn()

			svc := spec.NewService(repo, fetchRequestRepo, uidService, fetchRequestService)
			svc.SetTimestampGen(func() time.Time {
				return timestamp
			})

			// when
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

func TestService_UpdateByReferenceObjectID(t *testing.T) {
	// given
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
		ID:     specID,
		Tenant: tenant,
		URL:    "foo.bar",
		Mode:   model.FetchModeSingle,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: model.SpecFetchRequestReference,
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
				repo.On("GetByID", ctx, tenant, specID).Return(specModel, nil).Once()
				repo.On("Update", ctx, fixModelAPISpec()).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(nil).Once()
				repo.On("Create", ctx, fr).Return(nil).Once()

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
				repo.On("GetByID", ctx, tenant, specID).Return(specModel, nil).Once()
				repo.On("Update", ctx, specModel).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(nil).Once()
				repo.On("Create", ctx, fr).Return(nil).Once()

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
				repo.On("GetByID", ctx, tenant, specID).Return(specModel, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(testErr).Once()
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
				repo.On("GetByID", ctx, tenant, specID).Return(specModel, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(nil).Once()
				repo.On("Create", ctx, fr).Return(testErr).Once()
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
				repo.On("GetByID", ctx, tenant, specID).Return(nil, testErr).Once()
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
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			uidSvc := testCase.UIDServiceFn()
			fetchRequestSvc := testCase.FetchRequestServiceFn()

			svc := spec.NewService(repo, fetchRequestRepo, uidSvc, fetchRequestSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
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

func TestService_Delete(t *testing.T) {
	// given
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
				repo.On("Delete", ctx, tenant, id).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.SpecRepository {
				repo := &automock.SpecRepository{}
				repo.On("Delete", ctx, tenant, id).Return(testErr).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := spec.NewService(repo, nil, nil, nil)

			// when
			err := svc.Delete(ctx, testCase.InputID)

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
		err := svc.Delete(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_RefetchSpec(t *testing.T) {
	// given
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
				repo.On("GetByID", ctx, tenant, specID).Return(modelSpec, nil).Once()
				repo.On("Update", ctx, modelSpec).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(nil, nil)
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
				repo.On("GetByID", ctx, tenant, specID).Return(modelSpec, nil).Once()
				repo.On("Update", ctx, modelSpec).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(fr, nil)
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
				repo.On("GetByID", ctx, tenant, specID).Return(nil, testErr).Once()
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
				repo.On("GetByID", ctx, tenant, specID).Return(modelSpec, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(nil, testErr)
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
				repo.On("GetByID", ctx, tenant, specID).Return(modelSpec, nil).Once()
				repo.On("Update", ctx, modelSpec).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(nil, nil)
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
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			frRepo := testCase.FetchRequestRepoFn()
			frSvc := testCase.FetchRequestSvcFn()

			svc := spec.NewService(repo, frRepo, nil, frSvc)

			// when
			result, err := svc.RefetchSpec(ctx, specID)

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
		_, err := svc.RefetchSpec(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetFetchRequest(t *testing.T) {
	// given
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
				repo.On("Exists", ctx, tenant, specID).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(fr, nil).Once()
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
				repo.On("Exists", ctx, tenant, specID).Return(false, nil).Once()
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
				repo.On("Exists", ctx, tenant, specID).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(nil, apperrors.NewNotFoundError(resource.API, "")).Once()
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
				repo.On("Exists", ctx, tenant, specID).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenant, model.SpecFetchRequestReference, specID).Return(nil, testErr).Once()
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
				repo.On("Exists", ctx, tenant, specID).Return(false, testErr).Once()

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

			// when
			l, err := svc.GetFetchRequest(ctx, testCase.InputAPIDefID)

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
		// when
		_, err := svc.GetFetchRequest(context.TODO(), "dd")
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}
