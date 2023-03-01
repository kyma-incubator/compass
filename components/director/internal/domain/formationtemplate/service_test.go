package formationtemplate_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)
	ctxWithEmptyTenants := tnt.SaveToContext(context.TODO(), "", "")

	testErr := errors.New("test error")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}

	testCases := []struct {
		Name                        string
		Context                     context.Context
		Input                       *model.FormationTemplateInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		FormationTemplateConverter  func() *automock.FormationTemplateConverter
		TenantSvc                   func() *automock.TenantService
		WebhookRepo                 func() *automock.WebhookRepository
		ExpectedOutput              string
		ExpectedError               error
	}{
		{
			Name:    "Success",
			Context: ctx,
			Input:   &formationTemplateModelInput,
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
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			WebhookRepo: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, testTenantID, formationTemplateModel.Webhooks).Return(nil)
				return repo
			},
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name:    "Success when tenant in ctx is empty",
			Context: ctxWithEmptyTenants,
			Input:   &formationTemplateModelInput,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &formationTemplateModelInput, testID, "").Return(&formationTemplateModelNullTenant).Once()
				return converter
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Create", ctxWithEmptyTenants, &formationTemplateModelNullTenant).Return(nil).Once()
				return repo
			},
			WebhookRepo: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctxWithEmptyTenants, "", formationTemplateModelNullTenant.Webhooks).Return(nil)
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctxWithEmptyTenants).Return("", nil).Once()
				return svc
			},
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name:    "Success for application only template",
			Context: ctx,
			Input:   &formationTemplateModelInputAppOnly,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &formationTemplateModelInputAppOnly, testID, testTenantID).Return(&formationTemplateModelAppOnly).Once()
				return converter
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Create", ctx, &formationTemplateModelAppOnly).Return(nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByID", ctx, testTenantID).Return(newModelBusinessTenantMappingWithType(tenant.Account), nil).Once()
				return svc
			},
			WebhookRepo: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, testTenantID, formationTemplateModelAppOnly.Webhooks).Return(nil)
				return repo
			},
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name:    "Error when getting tenant object",
			Context: ctx,
			Input:   &formationTemplateModelInput,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				return &automock.FormationTemplateConverter{}
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				return &automock.FormationTemplateRepository{}
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return("", testErr)
				return svc
			},
			WebhookRepo: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedOutput: "",
			ExpectedError:  testErr,
		},
		{
			Name:    "Error when creating formation template",
			Context: ctx,
			Input:   &formationTemplateModelInput,
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
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			WebhookRepo: func() *automock.WebhookRepository {
				return &automock.WebhookRepository{}
			},
			ExpectedOutput: "",
			ExpectedError:  testErr,
		},
		{
			Name:    "Error when creating webhooks",
			Context: ctx,
			Input:   &formationTemplateModelInput,
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
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			WebhookRepo: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, testTenantID, formationTemplateModel.Webhooks).Return(testErr)
				return repo
			},
			ExpectedOutput: "",
			ExpectedError:  errors.New("while creating Webhooks for Formation Template"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()
			formationTemplateConv := testCase.FormationTemplateConverter()
			tenantSvc := testCase.TenantSvc()
			whRepo := testCase.WebhookRepo()
			idSvc := uidSvcFn()

			svc := formationtemplate.NewService(formationTemplateRepo, idSvc, formationTemplateConv, tenantSvc, whRepo, nil)

			// WHEN
			result, err := svc.Create(testCase.Context, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, formationTemplateRepo, idSvc, formationTemplateConv, tenantSvc, whRepo)
		})
	}
}

func TestService_Exist(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	testCases := []struct {
		Name                        string
		Input                       string
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		ExpectedOutput              bool
		ExpectedError               error
	}{
		{
			Name:  "Success",
			Input: testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(true, nil).Once()
				return repo
			},
			ExpectedOutput: true,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when checking if formation template exists",
			Input: testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctx, testID).Return(false, testErr).Once()
				return repo
			},
			ExpectedOutput: false,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, nil, nil, nil)

			// WHEN
			result, err := svc.Exist(ctx, testCase.Input)

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

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, nil, nil, nil)

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
	ctxWithEmptyTenants := tnt.SaveToContext(context.TODO(), "", "")

	testErr := errors.New("test error")
	pageSize := 20
	invalidPageSize := -100

	testCases := []struct {
		Name                        string
		Context                     context.Context
		PageSize                    int
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		TenantSvc                   func() *automock.TenantService
		ExpectedOutput              *model.FormationTemplatePage
		ExpectedError               error
	}{
		{
			Name:     "Success",
			Context:  ctx,
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("List", ctx, testTenantID, pageSize, mock.Anything).Return(&formationTemplateModelPage, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedOutput: &formationTemplateModelPage,
			ExpectedError:  nil,
		},
		{
			Name:     "Success when tenant in ctx is empty",
			Context:  ctxWithEmptyTenants,
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("List", ctxWithEmptyTenants, "", pageSize, mock.Anything).Return(&formationTemplateModelNullTenantPage, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctxWithEmptyTenants).Return("", nil).Once()
				return svc
			},
			ExpectedOutput: &formationTemplateModelNullTenantPage,
			ExpectedError:  nil,
		},
		{
			Name:     "Error when getting tenant object",
			Context:  ctx,
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				return &automock.FormationTemplateRepository{}
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return("", testErr)
				return svc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:     "Error when listing formation templates",
			Context:  ctx,
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("List", ctx, testTenantID, pageSize, mock.Anything).Return(nil, testErr).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                        "Error when invalid page size is given",
			Context:                     ctx,
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

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, tenantSvc, nil, nil)

			// WHEN
			result, err := svc.List(testCase.Context, testCase.PageSize, "")

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
}

func TestService_Update(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)
	ctxWithEmptyTenants := tnt.SaveToContext(context.TODO(), "", "")

	testErr := errors.New("test error")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID)
		return uidSvc
	}

	testCases := []struct {
		Name                        string
		Context                     context.Context
		Input                       string
		InputFormationTemplate      *model.FormationTemplateInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		FormationTemplateConverter  func() *automock.FormationTemplateConverter
		TenantSvc                   func() *automock.TenantService
		ExpectedError               error
	}{
		{
			Name:                   "Success",
			Context:                ctx,
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
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedError: nil,
		},
		{
			Name:                   "Success when tenant in context is empty",
			Context:                ctxWithEmptyTenants,
			Input:                  testID,
			InputFormationTemplate: &formationTemplateModelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Exists", ctxWithEmptyTenants, testID).Return(true, nil).Once()
				repo.On("Update", ctxWithEmptyTenants, &formationTemplateModelNullTenant).Return(nil).Once()
				return repo
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelInputToModel", &formationTemplateModelInput, testID, "").Return(&formationTemplateModelNullTenant).Once()

				return converter
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctxWithEmptyTenants).Return("", nil).Once()
				return svc
			},
			ExpectedError: nil,
		},
		{
			Name:                   "Error when formation template does not exist",
			Context:                ctx,
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
			Context:                ctx,
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
			Context:                ctx,
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
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return("", testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:                   "Error when updating formation template fails",
			Context:                ctx,
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
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
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

			svc := formationtemplate.NewService(formationTemplateRepo, uidSvcFn(), formationTemplateConverter, tenantSvc, nil, nil)

			// WHEN
			err := svc.Update(testCase.Context, testCase.Input, testCase.InputFormationTemplate)

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
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)
	ctxWithEmptyTenants := tnt.SaveToContext(context.TODO(), "", "")
	testErr := errors.New("test error")

	testCases := []struct {
		Name                        string
		Context                     context.Context
		Input                       string
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		TenantSvc                   func() *automock.TenantService
		ExpectedError               error
	}{
		{
			Name:    "Success",
			Context: ctx,
			Input:   testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Delete", ctx, testID, testTenantID).Return(nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedError: nil,
		},
		{
			Name:    "Success when tenant in ctx is empty",
			Context: ctxWithEmptyTenants,
			Input:   testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Delete", ctxWithEmptyTenants, testID, "").Return(nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctxWithEmptyTenants).Return("", nil).Once()
				return svc
			},
			ExpectedError: nil,
		},
		{
			Name:                        "Error when getting tenant object",
			Context:                     ctx,
			Input:                       testID,
			FormationTemplateRepository: UnusedFormationTemplateRepository,
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return("", testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:    "Error when deleting formation template",
			Context: ctx,
			Input:   testID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Delete", ctx, testID, testTenantID).Return(testErr).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()
			tenantSvc := testCase.TenantSvc()

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, tenantSvc, nil, nil)

			// WHEN
			err := svc.Delete(testCase.Context, testCase.Input)

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
}

func TestService_ListWebhooksForFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)
	ctxWithEmptyTenants := tnt.SaveToContext(context.TODO(), "", "")
	testErr := errors.New("test error")
	testWebhook := &model.Webhook{
		ID:       testWebhookID,
		ObjectID: testID,
	}

	testCases := []struct {
		Name             string
		Context          context.Context
		Input            string
		WebhookSvc       func() *automock.WebhookService
		TenantSvc        func() *automock.TenantService
		ExpectedWebhooks []*model.Webhook
		ExpectedError    error
	}{
		{
			Name:    "Success",
			Context: ctx,
			Input:   testID,
			WebhookSvc: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListForFormationTemplate", ctx, testTenantID, testID).Return([]*model.Webhook{testWebhook}, nil)
				return webhookSvc
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedError:    nil,
			ExpectedWebhooks: []*model.Webhook{testWebhook},
		},
		{
			Name: "Success when tenant in ctx is empty",
			WebhookSvc: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListForFormationTemplate", ctxWithEmptyTenants, "", testID).Return([]*model.Webhook{testWebhook}, nil)
				return webhookSvc
			},
			Context: ctxWithEmptyTenants,
			Input:   testID,
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctxWithEmptyTenants).Return("", nil).Once()
				return svc
			},
			ExpectedError:    nil,
			ExpectedWebhooks: []*model.Webhook{testWebhook},
		},
		{
			Name:       "Error when getting tenant object",
			Context:    ctx,
			Input:      testID,
			WebhookSvc: UnusedWebhookService,
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return("", testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:    "Error when listing formation template webhooks",
			Context: ctx,
			Input:   testID,
			WebhookSvc: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListForFormationTemplate", ctx, testTenantID, testID).Return(nil, testErr)
				return webhookSvc
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			tenantSvc := testCase.TenantSvc()
			webhookSvc := testCase.WebhookSvc()

			svc := formationtemplate.NewService(nil, nil, nil, tenantSvc, nil, webhookSvc)

			// WHEN
			webhooks, err := svc.ListWebhooksForFormationTemplate(testCase.Context, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.ElementsMatch(t, webhooks, testCase.ExpectedWebhooks)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, webhookSvc, tenantSvc)
		})
	}
}
