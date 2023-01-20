package certsubjectmapping_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping"
	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_Create(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *model.CertSubjectMapping
		Repo           func() *automock.CertMappingRepository
		ExpectedOutput string
		ExpectedError  error
	}{
		{
			Name:  "Success",
			Input: CertSubjectMappingModel,
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Create", emptyCtx, CertSubjectMappingModel).Return(nil).Once()
				return repo
			},
			ExpectedOutput: TestID,
		},
		{
			Name:  "Error when creating certificate subject mapping",
			Input: CertSubjectMappingModel,
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Create", emptyCtx, CertSubjectMappingModel).Return(testErr).Once()
				return repo
			},
			ExpectedOutput: "",
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.Repo()

			svc := certsubjectmapping.NewService(repo)

			// WHEN
			result, err := svc.Create(emptyCtx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_Get(t *testing.T) {
	testCases := []struct {
		Name           string
		Repo           func() *automock.CertMappingRepository
		ExpectedOutput *model.CertSubjectMapping
		ExpectedError  error
	}{
		{
			Name: "Success",
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Get", emptyCtx, TestID).Return(CertSubjectMappingModel, nil).Once()
				return repo
			},
			ExpectedOutput: CertSubjectMappingModel,
		},
		{
			Name: "Error when getting certificate subject mapping",
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Get", emptyCtx, TestID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.Repo()

			svc := certsubjectmapping.NewService(repo)

			// WHEN
			result, err := svc.Get(emptyCtx, TestID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_Update(t *testing.T) {
	notFoundErr := apperrors.NewNotFoundError(resource.CertSubjectMapping, TestID)

	testCases := []struct {
		Name          string
		Input         *model.CertSubjectMapping
		Repo          func() *automock.CertMappingRepository
		ExpectedError error
	}{
		{
			Name:  "Success",
			Input: CertSubjectMappingModel,
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Exists", emptyCtx, TestID).Return(true, nil).Once()
				repo.On("Update", emptyCtx, CertSubjectMappingModel).Return(nil).Once()
				return repo
			},
		},
		{
			Name: "Error when checking for certificate subject mapping existence fails",
			Input: CertSubjectMappingModel,
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Exists", emptyCtx, TestID).Return(false, testErr).Once()
				return repo
			},
			ExpectedError: testErr,
		},
		{
			Name: "Error when the updated certificate subject mapping is not found",
			Input: CertSubjectMappingModel,
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Exists", emptyCtx, TestID).Return(false, nil).Once()
				return repo
			},
			ExpectedError: notFoundErr,
		},
		{
			Name:  "Error when updating certificate subject mapping fails",
			Input: CertSubjectMappingModel,
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Exists", emptyCtx, TestID).Return(true, nil).Once()
				repo.On("Update", emptyCtx, CertSubjectMappingModel).Return(testErr).Once()
				return repo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.Repo()

			svc := certsubjectmapping.NewService(repo)

			// WHEN
			err := svc.Update(emptyCtx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_Delete(t *testing.T) {
	testCases := []struct {
		Name           string
		Repo           func() *automock.CertMappingRepository
		ExpectedError  error
	}{
		{
			Name: "Success",
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Delete", emptyCtx, TestID).Return(nil).Once()
				return repo
			},
		},
		{
			Name: "Error when deleting certificate subject mapping",
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Delete", emptyCtx, TestID).Return(testErr).Once()
				return repo
			},
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.Repo()

			svc := certsubjectmapping.NewService(repo)

			// WHEN
			err := svc.Delete(emptyCtx, TestID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_Exists(t *testing.T) {
	testCases := []struct {
		Name           string
		Repo           func() *automock.CertMappingRepository
		ExpectedOutput bool
		ExpectedError  error
	}{
		{
			Name: "Success",
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Exists", emptyCtx, TestID).Return(true, nil).Once()
				return repo
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when checking for existence of certificate subject mapping fails",
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("Exists", emptyCtx, TestID).Return(false, testErr).Once()
				return repo
			},
			ExpectedOutput: false,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.Repo()

			svc := certsubjectmapping.NewService(repo)

			// WHEN
			result, err := svc.Exists(emptyCtx, TestID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_List(t *testing.T) {
	pageSize := 100
	invalidPageSize := -100

	certSubjectMappingPage := &model.CertSubjectMappingPage{
		Data:       []*model.CertSubjectMapping{CertSubjectMappingModel},
		PageInfo:   nil,
		TotalCount: 1,
	}

	invalidDataErr := apperrors.NewInvalidDataError("page size must be between 1 and 300")

	testCases := []struct {
		Name           string
		Repo           func() *automock.CertMappingRepository
		PageSize int
		ExpectedOutput *model.CertSubjectMappingPage
		ExpectedError  error
	}{
		{
			Name: "Success",
			Repo: func() *automock.CertMappingRepository {
				repo := &automock.CertMappingRepository{}
				repo.On("List", emptyCtx, pageSize, "").Return(certSubjectMappingPage, nil).Once()
				return repo
			},
			PageSize: pageSize,
			ExpectedOutput: certSubjectMappingPage,
		},
		{
			Name:           "Error when page size is incorrect",
			Repo:           fixUnusedCertSubjectMappingRepository,
			PageSize:       invalidPageSize,
			ExpectedOutput: nil,
			ExpectedError:  invalidDataErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.Repo()

			svc := certsubjectmapping.NewService(repo)

			// WHEN
			result, err := svc.List(emptyCtx, testCase.PageSize, "")

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}
