package bundlereferences_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetForBundle(t *testing.T) {
	testErr := errors.New("test err")

	objectID := "id"
	bundleID := "bundleID"
	targetURL := "http://test.com"

	bundleReferenceModel := &model.BundleReference{
		Tenant:              tenantID,
		BundleID:            &bundleID,
		ObjectType:          model.BundleAPIReference,
		ObjectID:            &objectID,
		APIDefaultTargetURL: &targetURL,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.BundleReferenceRepository
		Expected     *model.BundleReference
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetByID", ctx, model.BundleAPIReference, tenantID, &objectID, &bundleID).Return(bundleReferenceModel, nil).Once()
				return repo
			},
			Expected: bundleReferenceModel,
		},
		{
			Name: "Error on getting by id",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetByID", ctx, model.BundleAPIReference, tenantID, &objectID, &bundleID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			svc := bundlereferences.NewService(repo)

			// when
			bndlRef, err := svc.GetForBundle(ctx, model.BundleAPIReference, &objectID, &bundleID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Equal(t, testCase.Expected, bndlRef)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_GetBundleIDsForObject(t *testing.T) {
	testErr := errors.New("test err")

	objectID := "id"
	firstBundleID := "bundleID"
	secondBundleID := "bundleID2"

	bundleIDs := []string{firstBundleID, secondBundleID}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.BundleReferenceRepository
		Expected     []string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetBundleIDsForObject", ctx, tenantID, model.BundleAPIReference, &objectID).Return(bundleIDs, nil).Once()
				return repo
			},
			Expected: bundleIDs,
		},
		{
			Name: "Error on getting bundle ids",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetBundleIDsForObject", ctx, tenantID, model.BundleAPIReference, &objectID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			svc := bundlereferences.NewService(repo)

			// when
			bndlIDs, err := svc.GetBundleIDsForObject(ctx, model.BundleAPIReference, &objectID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Equal(t, testCase.Expected, bndlIDs)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_CreateByReferenceObjectID(t *testing.T) {
	testErr := errors.New("test err")

	objectID := "id"
	bundleID := "bundleID"
	targetURL := "http://test.com"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	bundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: &targetURL,
	}

	bundleReferenceModel := &model.BundleReference{
		Tenant:              tenantID,
		BundleID:            &bundleID,
		ObjectType:          model.BundleAPIReference,
		ObjectID:            &objectID,
		APIDefaultTargetURL: &targetURL,
	}

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.BundleReferenceRepository
		Input        model.BundleReferenceInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("Create", ctx, bundleReferenceModel).Return(nil).Once()
				return repo
			},
			Input: *bundleReferenceInput,
		},
		{
			Name: "Error on creation",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("Create", ctx, bundleReferenceModel).Return(testErr).Once()
				return repo
			},
			Input:       *bundleReferenceInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			svc := bundlereferences.NewService(repo)

			// when
			err := svc.CreateByReferenceObjectID(ctx, testCase.Input, model.BundleAPIReference, &objectID, &bundleID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_UpdateByReferenceObjectID(t *testing.T) {
	testErr := errors.New("test err")

	objectID := "id"
	bundleID := "bundleID"
	targetURL := "http://test.com"
	updatedTargetURL := "http://test-updated.com"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	bundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: &updatedTargetURL,
	}

	bundleReferenceModelBefore := &model.BundleReference{
		Tenant:              tenantID,
		BundleID:            &bundleID,
		ObjectType:          model.BundleAPIReference,
		ObjectID:            &objectID,
		APIDefaultTargetURL: &targetURL,
	}

	bundleReferenceModelAfter := &model.BundleReference{
		Tenant:              tenantID,
		BundleID:            &bundleID,
		ObjectType:          model.BundleAPIReference,
		ObjectID:            &objectID,
		APIDefaultTargetURL: &updatedTargetURL,
	}

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.BundleReferenceRepository
		Input        model.BundleReferenceInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetByID", ctx, model.BundleAPIReference, tenantID, &objectID, &bundleID).Return(bundleReferenceModelBefore, nil).Once()
				repo.On("Update", ctx, bundleReferenceModelAfter).Return(nil).Once()
				return repo
			},
			Input: *bundleReferenceInput,
		},
		{
			Name: "Error on getting by id",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetByID", ctx, model.BundleAPIReference, tenantID, &objectID, &bundleID).Return(nil, testErr).Once()
				return repo
			},
			Input:       *bundleReferenceInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error on update",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetByID", ctx, model.BundleAPIReference, tenantID, &objectID, &bundleID).Return(bundleReferenceModelBefore, nil).Once()
				repo.On("Update", ctx, bundleReferenceModelAfter).Return(testErr).Once()
				return repo
			},
			Input:       *bundleReferenceInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			svc := bundlereferences.NewService(repo)

			// when
			err := svc.UpdateByReferenceObjectID(ctx, testCase.Input, model.BundleAPIReference, &objectID, &bundleID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteByReferenceObjectID(t *testing.T) {
	testErr := errors.New("test err")

	objectID := "id"
	bundleID := "bundleID"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.BundleReferenceRepository
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenantID, bundleID, model.BundleAPIReference, objectID).Return(nil).Once()
				return repo
			},
		},
		{
			Name: "Error on deletion",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenantID, bundleID, model.BundleAPIReference, objectID).Return(testErr).Once()
				return repo
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			svc := bundlereferences.NewService(repo)

			// when
			err := svc.DeleteByReferenceObjectID(ctx, model.BundleAPIReference, &objectID, &bundleID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
}
