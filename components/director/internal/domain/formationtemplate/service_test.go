package formationtemplate_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	tnt "github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var nilStr *string

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tnt.SaveToContext(context.TODO(), testTenantID, testTenantID)
	ctxWithEmptyTenants := tnt.SaveToContext(context.TODO(), "", "")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testFormationTemplateID)
		return uidSvc
	}

	testCases := []struct {
		Name                        string
		Context                     context.Context
		Input                       *model.FormationTemplateRegisterInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		FormationTemplateConverter  func() *automock.FormationTemplateConverter
		TenantSvc                   func() *automock.TenantService
		LabelSvc                    func() *automock.LabelService
		WebhookRepo                 func() *automock.WebhookRepository
		ExpectedOutput              string
		ExpectedError               error
	}{
		{
			Name:    "Success",
			Context: ctx,
			Input:   &formationTemplateRegisterInputModelWithLabels,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelRegisterInputToModel", &formationTemplateRegisterInputModelWithLabels, testFormationTemplateID, testTenantID).Return(&formationTemplateModel).Once()
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
			LabelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("UpsertMultipleLabels", ctx, testTenantID, model.FormationTemplateLabelableObject, testFormationTemplateID, registerInputLabels).Return(nil).Once()
				return lblSvc
			},
			WebhookRepo: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, testTenantID, formationTemplateModel.Webhooks).Return(nil)
				return repo
			},
			ExpectedOutput: testFormationTemplateID,
		},
		{
			Name:    "Success when tenant in ctx is empty",
			Context: ctxWithEmptyTenants,
			Input:   &formationTemplateRegisterInputModel,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelRegisterInputToModel", &formationTemplateRegisterInputModel, testFormationTemplateID, "").Return(&formationTemplateModelNullTenant).Once()
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
			ExpectedOutput: testFormationTemplateID,
		},
		{
			Name:    "Success for application only template",
			Context: ctx,
			Input:   &formationTemplateModelInputAppOnly,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelRegisterInputToModel", &formationTemplateModelInputAppOnly, testFormationTemplateID, testTenantID).Return(&formationTemplateModelAppOnly).Once()
				return converter
			},
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Create", ctx, &formationTemplateModelAppOnly).Return(nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			WebhookRepo: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", ctx, testTenantID, formationTemplateModelAppOnly.Webhooks).Return(nil)
				return repo
			},
			ExpectedOutput: testFormationTemplateID,
		},
		{
			Name:    "Error when getting tenant object",
			Context: ctx,
			Input:   &formationTemplateRegisterInputModel,
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
			Name:    "Error when upserting input labels fail",
			Context: ctx,
			Input:   &formationTemplateRegisterInputModelWithLabels,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelRegisterInputToModel", &formationTemplateRegisterInputModelWithLabels, testFormationTemplateID, testTenantID).Return(&formationTemplateModel).Once()
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
			LabelSvc: func() *automock.LabelService {
				lblSvc := &automock.LabelService{}
				lblSvc.On("UpsertMultipleLabels", ctx, testTenantID, model.FormationTemplateLabelableObject, testFormationTemplateID, registerInputLabels).Return(testErr).Once()
				return lblSvc
			},
			ExpectedError: testErr,
		},
		{
			Name:    "Error when creating formation template",
			Context: ctx,
			Input:   &formationTemplateRegisterInputModel,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelRegisterInputToModel", &formationTemplateRegisterInputModel, testFormationTemplateID, testTenantID).Return(&formationTemplateModel).Once()
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
			Input:   &formationTemplateRegisterInputModel,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelRegisterInputToModel", &formationTemplateRegisterInputModel, testFormationTemplateID, testTenantID).Return(&formationTemplateModel).Once()
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
			ExpectedError:  errors.New("while creating webhooks for formation template with ID:"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()
			formationTemplateConv := testCase.FormationTemplateConverter()
			tenantSvc := testCase.TenantSvc()

			lblSvc := UnusedLabelService()
			if testCase.LabelSvc != nil {
				lblSvc = testCase.LabelSvc()
			}

			whRepo := UnusedWebhookRepo()
			if testCase.WebhookRepo != nil {
				whRepo = testCase.WebhookRepo()
			}
			idSvc := uidSvcFn()

			svc := formationtemplate.NewService(formationTemplateRepo, idSvc, formationTemplateConv, tenantSvc, whRepo, nil, lblSvc)

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

			mock.AssertExpectationsForObjects(t, formationTemplateRepo, idSvc, formationTemplateConv, tenantSvc, lblSvc, whRepo)
		})
	}
}

func TestService_Exist(t *testing.T) {
	testCases := []struct {
		Name                        string
		Input                       string
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		ExpectedOutput              bool
		ExpectedError               error
	}{
		{
			Name:  "Success",
			Input: testFormationTemplateID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				return repo
			},
			ExpectedOutput: true,
		},
		{
			Name:  "Error when checking if formation template exists",
			Input: testFormationTemplateID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(false, testErr).Once()
				return repo
			},
			ExpectedOutput: false,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, nil, nil, nil, nil)

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
	testCases := []struct {
		Name                        string
		Input                       string
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		ExpectedOutput              *model.FormationTemplate
		ExpectedError               error
	}{
		{
			Name:  "Success",
			Input: testFormationTemplateID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, testFormationTemplateID).Return(&formationTemplateModel, nil).Once()
				return repo
			},
			ExpectedOutput: &formationTemplateModel,
		},
		{
			Name:  "Error when getting formation template",
			Input: testFormationTemplateID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, testFormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationTemplateRepo := testCase.FormationTemplateRepository()

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, nil, nil, nil, nil)

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
				repo.On("List", ctx, nilLabelFilters, nilStr, testTenantID, pageSize, mock.Anything).Return(&formationTemplateModelPage, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedOutput: &formationTemplateModelPage,
		},
		{
			Name:     "Success when tenant in ctx is empty",
			Context:  ctxWithEmptyTenants,
			PageSize: pageSize,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("List", ctxWithEmptyTenants, nilLabelFilters, nilStr, "", pageSize, mock.Anything).Return(&formationTemplateModelNullTenantPage, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctxWithEmptyTenants).Return("", nil).Once()
				return svc
			},
			ExpectedOutput: &formationTemplateModelNullTenantPage,
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
				repo.On("List", ctx, nilLabelFilters, nilStr, testTenantID, pageSize, mock.Anything).Return(nil, testErr).Once()
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

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, tenantSvc, nil, nil, nil)

			// WHEN
			result, err := svc.List(testCase.Context, nil, nil, testCase.PageSize, "")

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

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testFormationTemplateID)
		return uidSvc
	}

	testCases := []struct {
		Name                         string
		Context                      context.Context
		Input                        string
		FormationTemplateUpdateInput *model.FormationTemplateUpdateInput
		FormationTemplateRepository  func() *automock.FormationTemplateRepository
		FormationTemplateConverter   func() *automock.FormationTemplateConverter
		TenantSvc                    func() *automock.TenantService
		ExpectedError                error
	}{
		{
			Name:                         "Success",
			Context:                      ctx,
			Input:                        testFormationTemplateID,
			FormationTemplateUpdateInput: &formationTemplateUpdateInputModel,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				repo.On("Update", ctx, &formationTemplateModel).Return(nil).Once()
				return repo
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelUpdateInputToModel", &formationTemplateUpdateInputModel, testFormationTemplateID, testTenantID).Return(&formationTemplateModel).Once()

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
			Name:                         "Success when tenant in context is empty",
			Context:                      ctxWithEmptyTenants,
			Input:                        testFormationTemplateID,
			FormationTemplateUpdateInput: &formationTemplateUpdateInputModel,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctxWithEmptyTenants, testFormationTemplateID).Return(true, nil).Once()
				repo.On("Update", ctxWithEmptyTenants, &formationTemplateModelNullTenant).Return(nil).Once()
				return repo
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelUpdateInputToModel", &formationTemplateUpdateInputModel, testFormationTemplateID, "").Return(&formationTemplateModelNullTenant).Once()

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
			Name:                         "Error when formation template does not exist",
			Context:                      ctx,
			Input:                        testFormationTemplateID,
			FormationTemplateUpdateInput: &formationTemplateUpdateInputModel,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(false, nil).Once()
				return repo
			},
			TenantSvc:                  UnusedTenantService,
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedError:              formationTemplateNotFoundErr,
		},
		{
			Name:                         "Error when formation existence check failed",
			Context:                      ctx,
			Input:                        testFormationTemplateID,
			FormationTemplateUpdateInput: &formationTemplateUpdateInputModel,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(false, testErr).Once()
				return repo
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			TenantSvc:                  UnusedTenantService,
			ExpectedError:              testErr,
		},
		{
			Name:                         "Error when getting tenant object",
			Context:                      ctx,
			Input:                        testFormationTemplateID,
			FormationTemplateUpdateInput: &formationTemplateUpdateInputModel,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
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
			Name:                         "Error when updating formation template fails",
			Context:                      ctx,
			Input:                        testFormationTemplateID,
			FormationTemplateUpdateInput: &formationTemplateUpdateInputModel,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				repo.On("Update", ctx, &formationTemplateModel).Return(testErr).Once()
				return repo
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromModelUpdateInputToModel", &formationTemplateUpdateInputModel, testFormationTemplateID, testTenantID).Return(&formationTemplateModel).Once()

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

			svc := formationtemplate.NewService(formationTemplateRepo, uidSvcFn(), formationTemplateConverter, tenantSvc, nil, nil, nil)

			// WHEN
			err := svc.Update(testCase.Context, testCase.Input, testCase.FormationTemplateUpdateInput)

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
			Input:   testFormationTemplateID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Delete", ctx, testFormationTemplateID, testTenantID).Return(nil).Once()
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
			Input:   testFormationTemplateID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Delete", ctxWithEmptyTenants, testFormationTemplateID, "").Return(nil).Once()
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
			Input:                       testFormationTemplateID,
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
			Input:   testFormationTemplateID,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Delete", ctx, testFormationTemplateID, testTenantID).Return(testErr).Once()
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

			svc := formationtemplate.NewService(formationTemplateRepo, nil, nil, tenantSvc, nil, nil, nil)

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
	testWebhook := &model.Webhook{
		ID:       testWebhookID,
		ObjectID: testFormationTemplateID,
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
			Input:   testFormationTemplateID,
			WebhookSvc: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListForFormationTemplate", ctx, testTenantID, testFormationTemplateID).Return([]*model.Webhook{testWebhook}, nil)
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
				webhookSvc.On("ListForFormationTemplate", ctxWithEmptyTenants, "", testFormationTemplateID).Return([]*model.Webhook{testWebhook}, nil)
				return webhookSvc
			},
			Context: ctxWithEmptyTenants,
			Input:   testFormationTemplateID,
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
			Input:      testFormationTemplateID,
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
			Input:   testFormationTemplateID,
			WebhookSvc: func() *automock.WebhookService {
				webhookSvc := &automock.WebhookService{}
				webhookSvc.On("ListForFormationTemplate", ctx, testTenantID, testFormationTemplateID).Return(nil, testErr)
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

			svc := formationtemplate.NewService(nil, nil, nil, tenantSvc, nil, webhookSvc, nil)

			// WHEN
			webhooks, err := svc.ListWebhooksForFormationTemplate(testCase.Context, testFormationTemplateID)

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

func TestService_SetLabel(t *testing.T) {
	testCases := []struct {
		Name                        string
		Context                     context.Context
		LabelInput                  *model.LabelInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		TenantSvc                   func() *automock.TenantService
		LabelSvc                    func() *automock.LabelService
		ExpectedError               error
	}{
		{
			Name:       "Success",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, testTenantID, formationTemplateLabelInput).Return(nil).Once()
				return svc
			},
		},
		{
			Name:       "Error when extracting tenant fails",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return("", testErr).Once()
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:       "Error when formation existence check failed",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(false, testErr).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:       "Error when upserting labels fail",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertLabel", ctx, testTenantID, formationTemplateLabelInput).Return(testErr).Once()
				return svc
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ftRepo := UnusedFormationTemplateRepository()
			if testCase.FormationTemplateRepository != nil {
				ftRepo = testCase.FormationTemplateRepository()
			}

			tenantSvc := UnusedTenantService()
			if testCase.TenantSvc != nil {
				tenantSvc = testCase.TenantSvc()
			}

			labelSvc := UnusedLabelService()
			if testCase.LabelSvc != nil {
				labelSvc = testCase.LabelSvc()
			}

			svc := formationtemplate.NewService(ftRepo, nil, nil, tenantSvc, nil, nil, labelSvc)

			// WHEN
			err := svc.SetLabel(testCase.Context, testCase.LabelInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, ftRepo, tenantSvc, labelSvc)
		})
	}
}

func TestService_DeleteLabel(t *testing.T) {
	testCases := []struct {
		Name                        string
		Context                     context.Context
		LabelInput                  *model.LabelInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		TenantSvc                   func() *automock.TenantService
		LabelSvc                    func() *automock.LabelService
		ExpectedError               error
	}{
		{
			Name:       "Success",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("Delete", ctx, testTenantID, model.FormationTemplateLabelableObject, testFormationTemplateID, testLabelKey).Return(nil).Once()
				return svc
			},
		},
		{
			Name:       "Error when extracting tenant fails",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return("", testErr).Once()
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:       "Error when formation existence check failed",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(false, testErr).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:       "Error when deleting labels fail",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("Delete", ctx, testTenantID, model.FormationTemplateLabelableObject, testFormationTemplateID, testLabelKey).Return(testErr).Once()
				return svc
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ftRepo := UnusedFormationTemplateRepository()
			if testCase.FormationTemplateRepository != nil {
				ftRepo = testCase.FormationTemplateRepository()
			}

			tenantSvc := UnusedTenantService()
			if testCase.TenantSvc != nil {
				tenantSvc = testCase.TenantSvc()
			}

			labelSvc := UnusedLabelService()
			if testCase.LabelSvc != nil {
				labelSvc = testCase.LabelSvc()
			}

			svc := formationtemplate.NewService(ftRepo, nil, nil, tenantSvc, nil, nil, labelSvc)

			// WHEN
			err := svc.DeleteLabel(testCase.Context, testFormationTemplateID, testLabelKey)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, ftRepo, tenantSvc, labelSvc)
		})
	}
}

func TestService_GetLabel(t *testing.T) {
	testCases := []struct {
		Name                        string
		Context                     context.Context
		LabelInput                  *model.LabelInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		TenantSvc                   func() *automock.TenantService
		LabelSvc                    func() *automock.LabelService
		ExpectedOutput              *model.Label
		ExpectedError               error
	}{
		{
			Name:       "Success",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.FormationTemplateLabelableObject, testFormationTemplateID, testLabelKey).Return(modelLabel, nil).Once()
				return svc
			},
			ExpectedOutput: modelLabel,
		},
		{
			Name:       "Error when extracting tenant fails",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return("", testErr).Once()
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:       "Error when formation existence check failed",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(false, testErr).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:       "Error when getting label by key fail",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.FormationTemplateLabelableObject, testFormationTemplateID, testLabelKey).Return(nil, testErr).Once()
				return svc
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ftRepo := UnusedFormationTemplateRepository()
			if testCase.FormationTemplateRepository != nil {
				ftRepo = testCase.FormationTemplateRepository()
			}

			tenantSvc := UnusedTenantService()
			if testCase.TenantSvc != nil {
				tenantSvc = testCase.TenantSvc()
			}

			labelSvc := UnusedLabelService()
			if testCase.LabelSvc != nil {
				labelSvc = testCase.LabelSvc()
			}

			svc := formationtemplate.NewService(ftRepo, nil, nil, tenantSvc, nil, nil, labelSvc)

			// WHEN
			lbl, err := svc.GetLabel(testCase.Context, testFormationTemplateID, testLabelKey)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedOutput, lbl)
			}

			mock.AssertExpectationsForObjects(t, ftRepo, tenantSvc, labelSvc)
		})
	}
}

func TestService_ListLabels(t *testing.T) {
	testCases := []struct {
		Name                        string
		Context                     context.Context
		LabelInput                  *model.LabelInput
		FormationTemplateRepository func() *automock.FormationTemplateRepository
		TenantSvc                   func() *automock.TenantService
		LabelSvc                    func() *automock.LabelService
		ExpectedOutput              map[string]*model.Label
		ExpectedError               error
	}{
		{
			Name:       "Success",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("ListForObject", ctx, testTenantID, model.FormationTemplateLabelableObject, testFormationTemplateID).Return(modelLabels, nil).Once()
				return svc
			},
			ExpectedOutput: modelLabels,
		},
		{
			Name:       "Error when extracting tenant fails",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return("", testErr).Once()
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:       "Error when formation existence check failed",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(false, testErr).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:       "Error when listing labels fail",
			Context:    ctx,
			LabelInput: formationTemplateLabelInput,
			FormationTemplateRepository: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("ExistsGlobal", ctx, testFormationTemplateID).Return(true, nil).Once()
				return repo
			},
			TenantSvc: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("ExtractTenantIDForTenantScopedFormationTemplates", ctx).Return(testTenantID, nil).Once()
				return svc
			},
			LabelSvc: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("ListForObject", ctx, testTenantID, model.FormationTemplateLabelableObject, testFormationTemplateID).Return(nil, testErr).Once()
				return svc
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ftRepo := UnusedFormationTemplateRepository()
			if testCase.FormationTemplateRepository != nil {
				ftRepo = testCase.FormationTemplateRepository()
			}

			tenantSvc := UnusedTenantService()
			if testCase.TenantSvc != nil {
				tenantSvc = testCase.TenantSvc()
			}

			labelSvc := UnusedLabelService()
			if testCase.LabelSvc != nil {
				labelSvc = testCase.LabelSvc()
			}

			svc := formationtemplate.NewService(ftRepo, nil, nil, tenantSvc, nil, nil, labelSvc)

			// WHEN
			lbl, err := svc.ListLabels(testCase.Context, testFormationTemplateID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedOutput, lbl)
			}

			mock.AssertExpectationsForObjects(t, ftRepo, tenantSvc, labelSvc)
		})
	}
}
