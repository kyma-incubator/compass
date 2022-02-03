package bundlereferences_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences/automock"
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
		BundleID:            &bundleID,
		ObjectType:          model.BundleAPIReference,
		ObjectID:            &objectID,
		APIDefaultTargetURL: &targetURL,
		Visibility:          visibility,
	}

	ctx := context.TODO()

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
				repo.On("GetByID", ctx, model.BundleAPIReference, &objectID, &bundleID).Return(bundleReferenceModel, nil).Once()
				return repo
			},
			Expected: bundleReferenceModel,
		},
		{
			Name: "Error on getting by id",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetByID", ctx, model.BundleAPIReference, &objectID, &bundleID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			svc := bundlereferences.NewService(repo, nil)

			// WHEN
			bndlRef, err := svc.GetForBundle(ctx, model.BundleAPIReference, &objectID, &bundleID)

			// THEN
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
				repo.On("GetBundleIDsForObject", ctx, model.BundleAPIReference, &objectID).Return(bundleIDs, nil).Once()
				return repo
			},
			Expected: bundleIDs,
		},
		{
			Name: "Error on getting bundle ids",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetBundleIDsForObject", ctx, model.BundleAPIReference, &objectID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			svc := bundlereferences.NewService(repo, nil)

			// WHEN
			bndlIDs, err := svc.GetBundleIDsForObject(ctx, model.BundleAPIReference, &objectID)

			// THEN
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

	bundleRefID := "foo"
	objectID := "id"
	bundleID := "bundleID"
	targetURL := "http://test.com"

	ctx := context.TODO()

	bundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: &targetURL,
	}

	bundleReferenceModel := &model.BundleReference{
		ID:                  bundleRefID,
		BundleID:            &bundleID,
		ObjectType:          model.BundleAPIReference,
		ObjectID:            &objectID,
		APIDefaultTargetURL: &targetURL,
		Visibility:          visibility,
	}

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.BundleReferenceRepository
		UIDServiceFn func() *automock.UIDService
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
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(bundleRefID).Once()
				return svc
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
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(bundleRefID).Once()
				return svc
			},
			Input:       *bundleReferenceInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			uidSvc := testCase.UIDServiceFn()
			svc := bundlereferences.NewService(repo, uidSvc)

			// WHEN
			err := svc.CreateByReferenceObjectID(ctx, testCase.Input, model.BundleAPIReference, &objectID, &bundleID)

			// THEN
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

	bundleReferenceInput := &model.BundleReferenceInput{
		APIDefaultTargetURL: &updatedTargetURL,
	}

	bundleReferenceModelBefore := &model.BundleReference{
		BundleID:            &bundleID,
		ObjectType:          model.BundleAPIReference,
		ObjectID:            &objectID,
		APIDefaultTargetURL: &targetURL,
		Visibility:          visibility,
	}

	bundleReferenceModelAfter := &model.BundleReference{
		BundleID:            &bundleID,
		ObjectType:          model.BundleAPIReference,
		ObjectID:            &objectID,
		APIDefaultTargetURL: &updatedTargetURL,
		Visibility:          visibility,
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
				repo.On("GetByID", ctx, model.BundleAPIReference, &objectID, &bundleID).Return(bundleReferenceModelBefore, nil).Once()
				repo.On("Update", ctx, bundleReferenceModelAfter).Return(nil).Once()
				return repo
			},
			Input: *bundleReferenceInput,
		},
		{
			Name: "Error on getting by id",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetByID", ctx, model.BundleAPIReference, &objectID, &bundleID).Return(nil, testErr).Once()
				return repo
			},
			Input:       *bundleReferenceInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error on update",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("GetByID", ctx, model.BundleAPIReference, &objectID, &bundleID).Return(bundleReferenceModelBefore, nil).Once()
				repo.On("Update", ctx, bundleReferenceModelAfter).Return(testErr).Once()
				return repo
			},
			Input:       *bundleReferenceInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			svc := bundlereferences.NewService(repo, nil)

			// WHEN
			err := svc.UpdateByReferenceObjectID(ctx, testCase.Input, model.BundleAPIReference, &objectID, &bundleID)

			// THEN
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

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.BundleReferenceRepository
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, bundleID, model.BundleAPIReference, objectID).Return(nil).Once()
				return repo
			},
		},
		{
			Name: "Error on deletion",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, bundleID, model.BundleAPIReference, objectID).Return(testErr).Once()
				return repo
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			svc := bundlereferences.NewService(repo, nil)

			// WHEN
			err := svc.DeleteByReferenceObjectID(ctx, model.BundleAPIReference, &objectID, &bundleID)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_ListByBundleIDs(t *testing.T) {
	testErr := errors.New("test err")

	firstAPIDefID := "apiID"
	secondAPIDefID := "apiID2"
	firstBundleID := "bundleID"
	secondBundleID := "bundleID2"
	bundleIDs := []string{firstBundleID, secondBundleID}

	firstAPIDefBundleRef := fixAPIBundleReferenceModel()
	firstAPIDefBundleRef.BundleID = str.Ptr(firstBundleID)
	firstAPIDefBundleRef.ObjectID = str.Ptr(firstAPIDefID)

	secondAPIDefBundleRef := fixAPIBundleReferenceModel()
	secondAPIDefBundleRef.BundleID = str.Ptr(secondBundleID)
	secondAPIDefBundleRef.ObjectID = str.Ptr(secondAPIDefID)
	bundleRefs := []*model.BundleReference{&firstAPIDefBundleRef, &secondAPIDefBundleRef}

	numberOfAPIsInFirstBundle := 1
	numberOfAPIsInSecondBundle := 1
	totalCounts := map[string]int{firstBundleID: numberOfAPIsInFirstBundle, secondBundleID: numberOfAPIsInSecondBundle}

	after := "test"

	ctx := context.TODO()

	testCases := []struct {
		Name                string
		PageSize            int
		RepositoryFn        func() *automock.BundleReferenceRepository
		ExpectedBundleRefs  []*model.BundleReference
		ExpectedTotalCounts map[string]int
		ExpectedErrMessage  string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("ListByBundleIDs", ctx, model.BundleAPIReference, bundleIDs, 2, after).Return(bundleRefs, totalCounts, nil).Once()
				return repo
			},
			PageSize:            2,
			ExpectedBundleRefs:  bundleRefs,
			ExpectedTotalCounts: totalCounts,
		},
		{
			Name: "Return error when page size is less than 1",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				return repo
			},
			PageSize:           0,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Return error when page size is more than 200",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				return repo
			},
			PageSize:           201,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Error on listing bundle references",
			RepositoryFn: func() *automock.BundleReferenceRepository {
				repo := &automock.BundleReferenceRepository{}
				repo.On("ListByBundleIDs", ctx, model.BundleAPIReference, bundleIDs, 2, after).Return(nil, nil, testErr).Once()
				return repo
			},
			PageSize:            2,
			ExpectedBundleRefs:  nil,
			ExpectedTotalCounts: nil,
			ExpectedErrMessage:  testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			svc := bundlereferences.NewService(repo, nil)

			// WHEN
			bndlRefs, counts, err := svc.ListByBundleIDs(ctx, model.BundleAPIReference, bundleIDs, testCase.PageSize, after)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedBundleRefs, bndlRefs)
				assert.Equal(t, testCase.ExpectedTotalCounts[firstBundleID], counts[firstBundleID])
				assert.Equal(t, testCase.ExpectedTotalCounts[secondBundleID], counts[secondBundleID])
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}
