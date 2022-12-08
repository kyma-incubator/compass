package formationtemplate_test

import (
	"context"
	"testing"

	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)

	testErr := errors.New("test error")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}

	testCases := []struct {
		Name                        string
		Input                       *model.FormationTemplateInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		FormationTemplateConverter  func() *automock.FormationTemplateConverter
		TenantSvc                   func() *automock.TenantService
		ExpectedOutput              string
		ExpectedError               error
	}{
		{
			Name:  "Success when tenant in ctx is GA",
			Input: &formationTemplateModelInput,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &formationTemplateModelInput, testID, testTenantID).Return(&formationTemplateModel).Once()
				return converter
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Create", ctx, &formationTemplateModel).Return(nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(newModelBusinessTenantMappingWithType(tenant.Account), nil).Once()
				return svc
			},
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name:  "Success when tenant in ctx is SA",
			Input: &formationTemplateModelInput,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &formationTemplateModelInput, testID, testTenantID).Return(&formationTemplateModel).Once()
				return converter
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Create", ctx, &formationTemplateModel).Return(nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				saTenant := newModelBusinessTenantMappingWithType(tenant.Subaccount)
				svc.On("GetTenantByID", ctx, testTenantID).Return(saTenant, nil)
				svc.On("GetTenantByID", ctx, saTenant.Parent).Return(newModelBusinessTenantMappingWithType(tenant.Account), nil)
				return svc
			},
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when getting tenant object",
			Input: &formationTemplateModelInput,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				return &automock.FormationTemplateConverter{}
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				return &automock.FormationTemplateRepository{}
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(nil, testErr)
				return svc
			},
			ExpectedOutput: "",
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when tenant object is not of type SA or GA",
			Input: &formationTemplateModelInput,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				return &automock.FormationTemplateConverter{}
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				return &automock.FormationTemplateRepository{}
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(newModelBusinessTenantMappingWithType(tenant.Customer), nil)
				return svc
			},
			ExpectedOutput: "",
			ExpectedError:  errors.New("tenant used for tenant scoped Formation Templates must be of type account or subaccount"),
		},
		{
			Name:  "Error when getting GA tenant object",
			Input: &formationTemplateModelInput,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				return &automock.FormationTemplateConverter{}
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				return &automock.FormationTemplateRepository{}
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				saTenant := newModelBusinessTenantMappingWithType(tenant.Subaccount)
				svc.On("GetTenantByID", ctx, testTenantID).Return(saTenant, nil)
				svc.On("GetTenantByID", ctx, saTenant.Parent).Return(nil, testErr)
				return svc
			},
			ExpectedOutput: "",
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when creating formation template",
			Input: &formationTemplateModelInput,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &formationTemplateModelInput, testID, testTenantID).Return(&formationTemplateModel).Once()
				return converter
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Create", ctx, &formationTemplateModel).Return(testErr).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(newModelBusinessTenantMappingWithType(tenant.Account), nil).Once()
				return svc
			},
			ExpectedOutput: "",
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()
			formationTemplateConv := testCase.FormationTemplateConverter()
			tenantSvc := testCase.TenantSvc()
			idSvc := uidSvcFn()

			svc := formationtemplate.NewService(formationTemplateRepo, idSvc, formationTemplateConv, tenantSvc)

			// WHEN
			result, err := svc.Create(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationTemplateRepo, idSvc, formationTemplateConv, tenantSvc)
		})
	}
	t.Run("Error when tenant is not in context", func(t *testing.T) {
		idSvc := uidSvcFn()
		svc := formationtemplate.NewService(nil, idSvc, nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), &formationTemplateModelInput)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
	t.Run("Error when there is only internalID in context", func(t *testing.T) {
		ctxWithoutExternalID := tnt.SaveToContext(context.TODO(), testTenantID, "")
		idSvc := uidSvcFn()
		svc := formationtemplate.NewService(nil, idSvc, nil, nil)
		// WHEN
		_, err := svc.Create(ctxWithoutExternalID, &formationTemplateModelInput)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewTenantNotFoundError("").Error())
	})
	t.Run("Error when there is only externalID in context", func(t *testing.T) {
		ctxWithoutInternalID := tnt.SaveToContext(context.TODO(), "", testTenantID)
		idSvc := uidSvcFn()
		svc := formationtemplate.NewService(nil, idSvc, nil, nil)
		// WHEN
		_, err := svc.Create(ctxWithoutInternalID, &formationTemplateModelInput)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewTenantNotFoundError(testTenantID).Error())
	})
}

func TestService_Get(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                        string
		Input                       string
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		ExpectedOutput              *model.FormationTemplate
		ExpectedError               error
	}{
		{
			Name:  "Success",
			Input: testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, testID).Return(&formationTemplateModel, nil).Once()
				return repo
			},
			ExpectedOutput: &formationTemplateModel,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when getting formation template",
			Input: testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, testID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, nil)

			// WHEN
			result, err := svc.Get(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationTemplateRepo)
		})
	}
}

func TestService_List(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)

	testErr := errors.New("test error")
	pageSize := 20
	invalidPageSize := -100

	testCases := []struct {
		Name                        string
		PageSize                    int
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		TenantSvc                   func() *automock.TenantService
		ExpectedOutput              *model.FormationTemplatePage
		ExpectedError               error
	}{
		{
			Name:     "Success when tenant in ctx is GA",
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("List", ctx, testTenantID, pageSize, mock.Anything).Return(&formationTemplateModelPage, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(newModelBusinessTenantMappingWithType(tenant.Account), nil).Once()
				return svc
			},
			ExpectedOutput: &formationTemplateModelPage,
			ExpectedError:  nil,
		},
		{
			Name:     "Success when tenant in ctx is SA",
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("List", ctx, testTenantID, pageSize, mock.Anything).Return(&formationTemplateModelPage, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				saTenant := newModelBusinessTenantMappingWithType(tenant.Subaccount)
				svc.On("GetTenantByID", ctx, testTenantID).Return(saTenant, nil)
				svc.On("GetTenantByID", ctx, saTenant.Parent).Return(newModelBusinessTenantMappingWithType(tenant.Account), nil)
				return svc
			},
			ExpectedOutput: &formationTemplateModelPage,
			ExpectedError:  nil,
		},
		{
			Name:     "Error when getting tenant object",
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				return &automock.FormationTemplateRepository{}
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(nil, testErr)
				return svc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:     "Error when tenant object is not of type SA or GA",
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				return &automock.FormationTemplateRepository{}
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(newModelBusinessTenantMappingWithType(tenant.Customer), nil)
				return svc
			},
			ExpectedOutput: nil,
			ExpectedError:  errors.New("tenant used for tenant scoped Formation Templates must be of type account or subaccount"),
		},
		{
			Name:     "Error when getting GA tenant object",
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				return &automock.FormationTemplateRepository{}
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				saTenant := newModelBusinessTenantMappingWithType(tenant.Subaccount)
				svc.On("GetTenantByID", ctx, testTenantID).Return(saTenant, nil)
				svc.On("GetTenantByID", ctx, saTenant.Parent).Return(nil, testErr)
				return svc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:     "Error when listing formation templates",
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("List", ctx, testTenantID, pageSize, mock.Anything).Return(nil, testErr).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(newModelBusinessTenantMappingWithType(tenant.Account), nil).Once()
				return svc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                        "Error when invalid page size is given",
			PageSize:                    invalidPageSize,
			FormationTemplateRepository: UnusedFormationTemplateRepository,
			TenantSvc:                   UnusedTenantService,
			ExpectedOutput:              nil,
			ExpectedError:               errors.New("page size must be between 1 and 200"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()
			tenantSvc := testCase.TenantSvc()

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, tenantSvc)

			// WHEN
			result, err := svc.List(ctx, testCase.PageSize, "")

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationTemplateRepo, tenantSvc)
		})
	}
	t.Run("Error when tenant is not in context", func(t *testing.T) {
		svc := formationtemplate.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.List(context.TODO(), 1, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
	t.Run("Error when there is only internalID in context", func(t *testing.T) {
		ctxWithoutExternalID := tnt.SaveToContext(context.TODO(), testTenantID, "")
		svc := formationtemplate.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.List(ctxWithoutExternalID, 1, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewTenantNotFoundError("").Error())
	})
	t.Run("Error when there is only externalID in context", func(t *testing.T) {
		ctxWithoutInternalID := tnt.SaveToContext(context.TODO(), "", testTenantID)
		svc := formationtemplate.NewService(nil, nil, nil, nil)
		// WHEN
		_, err := svc.List(ctxWithoutInternalID, 1, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewTenantNotFoundError(testTenantID).Error())
	})
}

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)

	testErr := errors.New("test error")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}

	testCases := []struct {
		Name                        string
		Input                       string
		InputFormationTemplate      *model.FormationTemplateInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		FormationTemplateConverter  func() *automock.FormationTemplateConverter
		TenantSvc                   func() *automock.TenantService
		ExpectedError               error
	}{
		{
			Name:                   "Success when tenant in context is GA",
			Input:                  testID,
			InputFormationTemplate: &formationTemplateModelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(true, nil).Once()
				repo.On("Update", ctx, &formationTemplateModel).Return(nil).Once()
				return repo
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &formationTemplateModelInput, testID, testTenantID).Return(&formationTemplateModel).Once()

				return converter
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(newModelBusinessTenantMappingWithType(tenant.Account), nil).Once()
				return svc
			},
			ExpectedError: nil,
		},
		{
			Name:                   "Success when tenant in context is SA",
			Input:                  testID,
			InputFormationTemplate: &formationTemplateModelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(true, nil).Once()
				repo.On("Update", ctx, &formationTemplateModel).Return(nil).Once()
				return repo
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &formationTemplateModelInput, testID, testTenantID).Return(&formationTemplateModel).Once()

				return converter
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				saTenant := newModelBusinessTenantMappingWithType(tenant.Subaccount)
				svc.On("GetTenantByID", ctx, testTenantID).Return(saTenant, nil)
				svc.On("GetTenantByID", ctx, saTenant.Parent).Return(newModelBusinessTenantMappingWithType(tenant.Account), nil)
				return svc
			},
			ExpectedError: nil,
		},
		{
			Name:                   "Error when formation template does not exist",
			Input:                  testID,
			InputFormationTemplate: &formationTemplateModelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(false, nil).Once()
				return repo
			},
			TenantSvc:                  UnusedTenantService,
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedError:              apperrors.NewNotFoundError(resource.FormationTemplate, testID),
		},
		{
			Name:                   "Error when formation existence check failed",
			Input:                  testID,
			InputFormationTemplate: &formationTemplateModelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(false, testErr).Once()
				return repo
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			TenantSvc:                  UnusedTenantService,
			ExpectedError:              testErr,
		},
		{
			Name:                   "Error when getting tenant object",
			Input:                  testID,
			InputFormationTemplate: &formationTemplateModelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(true, nil).Once()
				return repo
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(nil, testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:                   "Error when tenant object is not of type SA or GA",
			Input:                  testID,
			InputFormationTemplate: &formationTemplateModelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(true, nil).Once()
				return repo
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(newModelBusinessTenantMappingWithType(tenant.Customer), nil)
				return svc
			},
			ExpectedError: errors.New("tenant used for tenant scoped Formation Templates must be of type account or subaccount"),
		},
		{
			Name:                   "Error when getting GA tenant object",
			Input:                  testID,
			InputFormationTemplate: &formationTemplateModelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(true, nil).Once()
				return repo
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				saTenant := newModelBusinessTenantMappingWithType(tenant.Subaccount)
				svc.On("GetTenantByID", ctx, testTenantID).Return(saTenant, nil)
				svc.On("GetTenantByID", ctx, saTenant.Parent).Return(nil, testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:                   "Error when updating formation template fails",
			Input:                  testID,
			InputFormationTemplate: &formationTemplateModelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(true, nil).Once()
				repo.On("Update", ctx, &formationTemplateModel).Return(testErr).Once()
				return repo
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &formationTemplateModelInput, testID, testTenantID).Return(&formationTemplateModel).Once()

				return converter
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(newModelBusinessTenantMappingWithType(tenant.Account), nil).Once()
				return svc
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()
			formationTemplateConverter := testCase.FormationTemplateConverter()
			tenantSvc := testCase.TenantSvc()

			svc := formationtemplate.NewService(formationTemplateRepo, uidSvcFn(), formationTemplateConverter, tenantSvc)

			// WHEN
			err := svc.Update(ctx, testCase.Input, testCase.InputFormationTemplate)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, formationTemplateRepo, tenantSvc)
		})
	}

	t.Run("Error when tenant is not in context", func(t *testing.T) {
		repo := func() *automock.FormationTemplateRepository {
			repo := &automock.FormationTemplateRepository{}
			repo.On("Exists", context.TODO(), "").Return(true, nil).Once()
			return repo
		}
		svc := formationtemplate.NewService(repo(), nil, nil, nil)
		// WHEN
		err := svc.Update(context.TODO(), "", nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
	t.Run("Error when there is only internalID in context", func(t *testing.T) {
		ctxWithoutExternalID := tnt.SaveToContext(context.TODO(), testTenantID, "")
		repo := func() *automock.FormationTemplateRepository {
			repo := &automock.FormationTemplateRepository{}
			repo.On("Exists", ctxWithoutExternalID, "").Return(true, nil).Once()
			return repo
		}
		svc := formationtemplate.NewService(repo(), nil, nil, nil)
		// WHEN
		err := svc.Update(ctxWithoutExternalID, "", nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewTenantNotFoundError("").Error())
	})
	t.Run("Error when there is only externalID in context", func(t *testing.T) {
		ctxWithoutInternalID := tnt.SaveToContext(context.TODO(), "", testTenantID)
		repo := func() *automock.FormationTemplateRepository {
			repo := &automock.FormationTemplateRepository{}
			repo.On("Exists", ctxWithoutInternalID, "").Return(true, nil).Once()
			return repo
		}
		svc := formationtemplate.NewService(repo(), nil, nil, nil)
		// WHEN
		err := svc.Update(ctxWithoutInternalID, "", nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), apperrors.NewTenantNotFoundError(testTenantID).Error())
	})
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                        string
		Input                       string
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		ExpectedError               error
	}{
		{
			Name:  "Success",
			Input: testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Delete", ctx, testID).Return(nil).Once()
				return repo
			},
			ExpectedError: nil,
		},
		{
			Name:  "Error when deleting formation template",
			Input: testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Delete", ctx, testID).Return(testErr).Once()
				return repo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, nil)

			// WHEN
			err := svc.Delete(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, formationTemplateRepo)
		})
	}
}
