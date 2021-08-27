package tenant_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	tenant2 "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetExternalTenant(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")
	tenantMappingModel := newModelBusinessTenantMapping(testID, testName)

	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedError       error
		ExpectedOutput      string
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("Get", ctx, testID).Return(tenantMappingModel, nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testExternal,
		},
		{
			Name: "Error when getting the internal tenant",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("Get", ctx, testID).Return(nil, testError).Once()
				return tenantMappingRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepoFn := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepoFn, nil)

			// WHEN
			result, err := svc.GetExternalTenant(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			tenantMappingRepoFn.AssertExpectations(t)
		})
	}
}

func TestService_GetInternalTenant(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")
	tenantMappingModel := newModelBusinessTenantMapping(testID, testName)

	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedError       error
		ExpectedOutput      string
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("GetByExternalTenant", ctx, testExternal).Return(tenantMappingModel, nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testID,
		},
		{
			Name: "Error when getting the internal tenant",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("GetByExternalTenant", ctx, testExternal).Return(nil, testError).Once()
				return tenantMappingRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepoFn := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepoFn, nil)

			// WHEN
			result, err := svc.GetInternalTenant(ctx, testExternal)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			tenantMappingRepoFn.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")
	modelTenantMappings := []*model.BusinessTenantMapping{
		newModelBusinessTenantMapping("foo1", "bar1"),
		newModelBusinessTenantMapping("foo2", "bar2"),
	}

	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedError       error
		ExpectedOutput      []*model.BusinessTenantMapping
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx).Return(modelTenantMappings, nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: modelTenantMappings,
		},
		{
			Name: "Error when tenants",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("List", ctx).Return([]*model.BusinessTenantMapping{}, testError).Once()
				return tenantMappingRepo
			},
			ExpectedError:  testError,
			ExpectedOutput: []*model.BusinessTenantMapping{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepo := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepo, nil)

			// WHEN
			result, err := svc.List(ctx)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			tenantMappingRepo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteMany(t *testing.T) {
	//GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")
	tenantInput := newModelBusinessTenantMappingInput(testName, "")
	testErr := errors.New("test")
	testCases := []struct {
		Name                string
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		ExpectedOutput      error
	}{
		{
			Name: "Success",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("DeleteByExternalTenant", ctx, tenantInput.ExternalTenant).Return(nil).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: nil,
		},
		{
			Name: "Error while deleting the tenant mapping",
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("DeleteByExternalTenant", ctx, tenantInput.ExternalTenant).Return(testErr).Once()
				return tenantMappingRepo
			},
			ExpectedOutput: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantMappingRepo := testCase.TenantMappingRepoFn()
			svc := tenant.NewService(tenantMappingRepo, nil)

			// WHEN
			err := svc.DeleteMany(ctx, []model.BusinessTenantMappingInput{tenantInput})

			// THEN
			if testCase.ExpectedOutput != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedOutput.Error())
			} else {
				assert.NoError(t, err)
			}

			tenantMappingRepo.AssertExpectations(t)
		})
	}

}

func TestService_CreateManyIfNotExists(t *testing.T) {
	//GIVEN
	ctx := tenant.SaveToContext(context.TODO(), "test", "external-test")

	tenantInputs := []model.BusinessTenantMappingInput{newModelBusinessTenantMappingInput("test1", ""),
		newModelBusinessTenantMappingInput("test2", "").WithExternalTenant("external2")}
	tenantInputsWithSubdomains := []model.BusinessTenantMappingInput{newModelBusinessTenantMappingInput("test1", testSubdomain),
		newModelBusinessTenantMappingInput("test2", "").WithExternalTenant("external2")}
	tenantModels := []model.BusinessTenantMapping{*newModelBusinessTenantMapping(testID, "test1"),
		newModelBusinessTenantMapping(testID, "test2").WithExternalTenant("external2")}

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}
	noopLabelRepo := func() *automock.LabelRepository {
		return &automock.LabelRepository{}
	}
	noopLabelUpsertSvc := func() *automock.LabelUpsertService {
		return &automock.LabelUpsertService{}
	}
	testErr := errors.New("test")
	testCases := []struct {
		Name                string
		tenantInputs        []model.BusinessTenantMappingInput
		TenantMappingRepoFn func() *automock.TenantMappingRepository
		LabelRepoFn         func() *automock.LabelRepository
		LabelUpsertSvcFn    func() *automock.LabelUpsertService
		ExpectedOutput      error
	}{
		{
			Name:         "Success",
			tenantInputs: tenantInputs,
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[0].ExternalTenant).Return(false, nil)
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[1].ExternalTenant).Return(true, nil)
				tenantMappingRepo.On("Create", ctx, tenantModels[0]).Return(nil).Once()
				return tenantMappingRepo
			},
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedOutput:   nil,
		},
		{
			Name:         "Success when subdomain should be added",
			tenantInputs: tenantInputsWithSubdomains,
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantInputsWithSubdomains[0].ExternalTenant).Return(false, nil)
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantInputsWithSubdomains[1].ExternalTenant).Return(true, nil)
				tenantMappingRepo.On("Create", ctx, tenantModels[0]).Return(nil).Once()
				return tenantMappingRepo
			},
			LabelRepoFn: noopLabelRepo,
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				label := &model.LabelInput{
					Key:        "subdomain",
					Value:      testSubdomain,
					ObjectID:   testID,
					ObjectType: model.TenantLabelableObject,
				}
				svc.On("UpsertLabel", ctx, testID, label).Return(nil).Once()
				return svc
			},
			ExpectedOutput: nil,
		},
		{
			Name:         "Error when checking the existence of tenant",
			tenantInputs: tenantInputs,
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[0].ExternalTenant).Return(false, testErr)
				return tenantMappingRepo
			},
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedOutput:   testErr,
		},
		{
			Name:         "Error when subdomain creation fails",
			tenantInputs: tenantInputsWithSubdomains,
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantInputsWithSubdomains[0].ExternalTenant).Return(false, nil)
				tenantMappingRepo.On("Create", ctx, tenantModels[0]).Return(nil).Once()
				return tenantMappingRepo
			},
			LabelRepoFn: noopLabelRepo,
			LabelUpsertSvcFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				label := &model.LabelInput{
					Key:        "subdomain",
					Value:      testSubdomain,
					ObjectID:   testID,
					ObjectType: model.TenantLabelableObject,
				}
				svc.On("UpsertLabel", ctx, testID, label).Return(testErr).Once()
				return svc
			},
			ExpectedOutput: testErr,
		},
		{
			Name:         "Error when creating the tenant",
			tenantInputs: tenantInputs,
			TenantMappingRepoFn: func() *automock.TenantMappingRepository {
				tenantMappingRepo := &automock.TenantMappingRepository{}
				tenantMappingRepo.On("ExistsByExternalTenant", ctx, tenantModels[0].ExternalTenant).Return(false, nil)
				tenantMappingRepo.On("Create", ctx, tenantModels[0]).Return(testErr).Once()
				return tenantMappingRepo
			},
			LabelRepoFn:      noopLabelRepo,
			LabelUpsertSvcFn: noopLabelUpsertSvc,
			ExpectedOutput:   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			uidSvc := uidSvcFn()
			tenantMappingRepo := testCase.TenantMappingRepoFn()
			labelRepo := testCase.LabelRepoFn()
			labelUpsertSvc := testCase.LabelUpsertSvcFn()

			svc := tenant.NewServiceWithLabels(tenantMappingRepo, uidSvc, labelRepo, labelUpsertSvc)

			// WHEN
			err := svc.CreateManyIfNotExists(ctx, testCase.tenantInputs...)

			// THEN
			if testCase.ExpectedOutput != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedOutput.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, tenantMappingRepo, uidSvc, labelRepo, labelUpsertSvc)
		})
	}
}

func Test_MultipleToTenantMapping(t *testing.T) {
	testCases := []struct {
		Name          string
		InputSlice    []model.BusinessTenantMappingInput
		ExpectedSlice []model.BusinessTenantMapping
	}{
		{
			Name: "success with more than one parent chain",
			InputSlice: []model.BusinessTenantMappingInput{
				{
					Name:           "acc1",
					ExternalTenant: "0",
				},
				{
					Name:           "acc2",
					ExternalTenant: "1",
					Parent:         "2",
				},
				{
					Name:           "customer1",
					ExternalTenant: "2",
					Parent:         "4",
				},
				{
					Name:           "acc3",
					ExternalTenant: "3",
				},
				{
					Name:           "x1",
					ExternalTenant: "4",
				},
			},
			ExpectedSlice: []model.BusinessTenantMapping{
				{
					ID:             "0",
					Name:           "acc1",
					ExternalTenant: "0",
					Status:         tenant2.Active,
					Type:           tenant2.Unknown,
				},
				{
					ID:             "4",
					Name:           "x1",
					ExternalTenant: "4",
					Status:         tenant2.Active,
					Type:           tenant2.Unknown,
				},
				{
					ID:             "2",
					Name:           "customer1",
					ExternalTenant: "2",
					Parent:         "4",
					Status:         tenant2.Active,
					Type:           tenant2.Unknown,
				},
				{
					ID:             "1",
					Name:           "acc2",
					ExternalTenant: "1",
					Parent:         "2",
					Status:         tenant2.Active,
					Type:           tenant2.Unknown,
				},
				{
					ID:             "3",
					Name:           "acc3",
					ExternalTenant: "3",
					Status:         tenant2.Active,
					Type:           tenant2.Unknown,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := tenant.NewService(nil, &serialUUIDService{})
			require.Equal(t, testCase.ExpectedSlice, svc.MultipleToTenantMapping(testCase.InputSlice))
		})
	}

}

func Test_MoveBeforeIndex(t *testing.T) {
	testCases := []struct {
		Name           string
		InputSlice     []model.BusinessTenantMapping
		TargetTenantID string
		TargetIndex    int
		ExpectedSlice  []model.BusinessTenantMapping
		ShouldMove     bool
	}{
		{
			Name: "success",
			InputSlice: []model.BusinessTenantMapping{
				{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
			},
			TargetTenantID: "4",
			TargetIndex:    1,
			ExpectedSlice: []model.BusinessTenantMapping{
				{ID: "1"}, {ID: "4"}, {ID: "2"}, {ID: "3"}, {ID: "5"},
			},
			ShouldMove: true,
		},
		{
			Name: "move before first element",
			InputSlice: []model.BusinessTenantMapping{
				{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
			},
			TargetTenantID: "3",
			TargetIndex:    0,
			ExpectedSlice: []model.BusinessTenantMapping{
				{ID: "3"}, {ID: "1"}, {ID: "2"}, {ID: "4"}, {ID: "5"},
			},
			ShouldMove: true,
		},
		{
			Name: "move before last element",
			InputSlice: []model.BusinessTenantMapping{
				{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
			},
			TargetTenantID: "3",
			TargetIndex:    4,
			ExpectedSlice: []model.BusinessTenantMapping{
				{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result, moved := tenant.MoveBeforeIfShould(testCase.InputSlice, testCase.TargetTenantID, testCase.TargetIndex)
			require.Equal(t, testCase.ShouldMove, moved)
			require.Equal(t, testCase.ExpectedSlice, result)
		})
	}
}

func Test_ListLabels(t *testing.T) {
	const tenantID = "edc6857b-b0c7-49e6-9f0a-e87a9c2a4eb8"

	ctx := context.TODO()
	testErr := errors.New("failed to list labels")

	t.Run("Success", func(t *testing.T) {
		labels := map[string]*model.Label{
			"label-key": {
				ID:         "5ef5ebd0-987d-4cb6-a3c1-7d710de259a2",
				Tenant:     tenantID,
				Key:        "label-key",
				Value:      "value",
				ObjectID:   tenantID,
				ObjectType: model.TenantLabelableObject,
			},
		}

		uidSvc := &automock.UIDService{}
		labelUpsertSvc := &automock.LabelUpsertService{}

		tenantRepo := &automock.TenantMappingRepository{}
		tenantRepo.On("Exists", ctx, tenantID).Return(true, nil)

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("ListForObject", ctx, tenantID, model.TenantLabelableObject, tenantID).Return(labels, nil)

		defer mock.AssertExpectationsForObjects(t, tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		svc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		actualLabels, err := svc.ListLabels(ctx, tenantID)
		assert.NoError(t, err)
		assert.Equal(t, labels, actualLabels)
	})

	t.Run("Error when tenant existence cannot be ensured", func(t *testing.T) {
		uidSvc := &automock.UIDService{}
		labelRepo := &automock.LabelRepository{}
		labelUpsertSvc := &automock.LabelUpsertService{}

		tenantRepo := &automock.TenantMappingRepository{}
		tenantRepo.On("Exists", ctx, tenantID).Return(false, testErr)

		defer mock.AssertExpectationsForObjects(t, tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		svc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		_, err := svc.ListLabels(ctx, tenantID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("while checking if tenant with ID %s exists", tenantID))
	})

	t.Run("Error when tenant does not exist", func(t *testing.T) {
		uidSvc := &automock.UIDService{}
		labelRepo := &automock.LabelRepository{}
		labelUpsertSvc := &automock.LabelUpsertService{}

		tenantRepo := &automock.TenantMappingRepository{}
		tenantRepo.On("Exists", ctx, tenantID).Return(false, nil)

		defer mock.AssertExpectationsForObjects(t, tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		svc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		_, err := svc.ListLabels(ctx, tenantID)
		assert.Error(t, err)
		assert.True(t, apperrors.IsNotFoundError(err))
	})

	t.Run("Error when fails to list labels from repo", func(t *testing.T) {
		uidSvc := &automock.UIDService{}
		labelUpsertSvc := &automock.LabelUpsertService{}

		tenantRepo := &automock.TenantMappingRepository{}
		tenantRepo.On("Exists", ctx, tenantID).Return(true, nil)

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("ListForObject", ctx, tenantID, model.TenantLabelableObject, tenantID).Return(nil, testErr)

		defer mock.AssertExpectationsForObjects(t, tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		svc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelUpsertSvc)

		_, err := svc.ListLabels(ctx, tenantID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("whilie listing labels for tenant with ID %s", tenantID))
	})
}

type serialUUIDService struct {
	i int
}

func (s *serialUUIDService) Generate() string {
	result := s.i
	s.i++
	return fmt.Sprintf("%d", result)
}
