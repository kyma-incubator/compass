package tombstone_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	modelTombstone := fixTombstoneModel()
	modelInput := *fixTombstoneModelInput()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.TombstoneRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.TombstoneInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("Create", ctx, modelTombstone).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(tombstoneID)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error - Tombstone creation",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("Create", ctx, modelTombstone).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(tombstoneID)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			uidSvc := testCase.UIDServiceFn()

			svc := tombstone.NewService(repo, uidSvc)

			// when
			result, err := svc.Create(ctx, appID, testCase.Input)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := tombstone.NewService(nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), "", model.TombstoneInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelTombstone := fixTombstoneModel()
	modelInput := *fixTombstoneModelInput()

	inputTombstoneModel := mock.MatchedBy(func(ts *model.Tombstone) bool {
		return ts.OrdID == modelInput.OrdID
	})

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.TombstoneRepository
		Input        model.TombstoneInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("GetByID", ctx, tenantID, tombstoneID).Return(modelTombstone, nil).Once()
				repo.On("Update", ctx, inputTombstoneModel).Return(nil).Once()
				return repo
			},
			InputID:     tombstoneID,
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("GetByID", ctx, tenantID, tombstoneID).Return(modelTombstone, nil).Once()
				repo.On("Update", ctx, inputTombstoneModel).Return(testErr).Once()
				return repo
			},
			InputID:     tombstoneID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("GetByID", ctx, tenantID, tombstoneID).Return(nil, testErr).Once()
				return repo
			},
			InputID:     tombstoneID,
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := tombstone.NewService(repo, nil)

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

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
		svc := tombstone.NewService(nil, nil)
		// WHEN
		err := svc.Update(context.TODO(), "", model.TombstoneInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.TombstoneRepository
		Input        model.TombstoneInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("Delete", ctx, tenantID, tombstoneID).Return(nil).Once()
				return repo
			},
			InputID:     tombstoneID,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("Delete", ctx, tenantID, tombstoneID).Return(testErr).Once()
				return repo
			},
			InputID:     tombstoneID,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := tombstone.NewService(repo, nil)

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
		svc := tombstone.NewService(nil, nil)
		// WHEN
		err := svc.Delete(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Exist(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)

	testCases := []struct {
		Name           string
		RepoFn         func() *automock.TombstoneRepository
		ExpectedError  error
		ExpectedOutput bool
	}{
		{
			Name: "Success",
			RepoFn: func() *automock.TombstoneRepository {
				pkgRepo := &automock.TombstoneRepository{}
				pkgRepo.On("Exists", ctx, tenantID, tombstoneID).Return(true, nil).Once()
				return pkgRepo
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting Tombstone",
			RepoFn: func() *automock.TombstoneRepository {
				pkgRepo := &automock.TombstoneRepository{}
				pkgRepo.On("Exists", ctx, tenantID, tombstoneID).Return(false, testErr).Once()
				return pkgRepo
			},
			ExpectedError:  testErr,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tombstoneRepo := testCase.RepoFn()
			svc := tombstone.NewService(tombstoneRepo, nil)

			// WHEN
			result, err := svc.Exist(ctx, tombstoneID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			tombstoneRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := tombstone.NewService(nil, nil)
		// WHEN
		_, err := svc.Exist(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	tsModel := fixTombstoneModel()

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.TombstoneRepository
		Input              model.TombstoneInput
		InputID            string
		ExpectedTombstone  *model.Tombstone
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("GetByID", ctx, tenantID, tombstoneID).Return(tsModel, nil).Once()
				return repo
			},
			InputID:            tombstoneID,
			ExpectedTombstone:  tsModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Tombstone retrieval failed",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("GetByID", ctx, tenantID, tombstoneID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            tombstoneID,
			ExpectedTombstone:  tsModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := tombstone.NewService(repo, nil)

			// when
			ts, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedTombstone, ts)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := tombstone.NewService(nil, nil)
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationID(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	tombstones := []*model.Tombstone{
		fixTombstoneModel(),
		fixTombstoneModel(),
		fixTombstoneModel(),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.TombstoneRepository
		ExpectedResult     []*model.Tombstone
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, appID).Return(tombstones, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     tombstones,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Tombstone listing failed",
			RepositoryFn: func() *automock.TombstoneRepository {
				repo := &automock.TombstoneRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, appID).Return(nil, testErr).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := tombstone.NewService(repo, nil)

			// when
			docs, err := svc.ListByApplicationID(ctx, appID)

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
		svc := tombstone.NewService(nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
