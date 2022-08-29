package formation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceAssignFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("test error")

	inputFormation := model.Formation{
		Name: testFormationName,
	}
	expectedFormation := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            Tnt,
	}
	expectedFormationTemplate := &model.FormationTemplate{
		ID:               FormationTemplateID,
		Name:             testFormationTemplateName,
		RuntimeType:      runtimeType,
		ApplicationTypes: []string{applicationType},
	}

	inputSecondFormation := model.Formation{
		Name: secondTestFormationName,
	}
	expectedSecondFormation := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            Tnt,
	}

	applicationLblNoFormations := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{},
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	applicationLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName},
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationLblInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	applicationLblInputInDefaultScenario := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{model.DefaultScenario},
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	runtimeLblNoFormations := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{},
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	applicationTypeLblInput := model.LabelInput{
		Key:        applicationType,
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationTypeLbl := &model.Label{
		ID:         "123",
		Key:        applicationType,
		Value:      applicationType,
		Tenant:     str.Ptr(Tnt),
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	runtimeLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName},
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeLblInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeTypeLblInput := model.LabelInput{
		Key:        runtimeType,
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeWithRuntimeContextRuntimeTypeLblInput := model.LabelInput{
		Key:        runtimeType,
		ObjectID:   RuntimeContextRuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeTypeLbl := &model.Label{
		ID:         "123",
		Key:        runtimeType,
		Value:      runtimeType,
		Tenant:     str.Ptr(Tnt),
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeWithRuntimeContextRuntimeTypeLbl := &model.Label{
		ID:         "123",
		Key:        runtimeType,
		Value:      runtimeType,
		Tenant:     str.Ptr(Tnt),
		ObjectID:   RuntimeContextRuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeCtxLblInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   RuntimeContextID,
		ObjectType: model.RuntimeContextLabelableObject,
		Version:    0,
	}

	asa := model.AutomaticScenarioAssignment{
		ScenarioName:   testFormationName,
		Tenant:         Tnt,
		TargetTenantID: TargetTenant,
	}

	testCases := []struct {
		Name                          string
		UIDServiceFn                  func() *automock.UuidService
		LabelServiceFn                func() *automock.LabelService
		LabelDefServiceFn             func() *automock.LabelDefService
		TenantServiceFn               func() *automock.TenantService
		AsaRepoFn                     func() *automock.AutomaticFormationAssignmentRepository
		AsaServiceFN                  func() *automock.AutomaticFormationAssignmentService
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		ApplicationRepoFN             func() *automock.ApplicationRepository
		WebhookRepoFN                 func() *automock.WebhookRepository
		WebhookConverterFN            func() *automock.WebhookConverter
		WebhookClientFN               func() *automock.WebhookClient
		ApplicationTemplateRepoFN     func() *automock.ApplicationTemplateRepository
		LabelRepoFN                   func() *automock.LabelRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		ObjectID                      string
		ObjectType                    graphql.FormationObjectType
		InputFormation                model.Formation
		ExpectedFormation             *model.Formation
		ExpectedErrMessage            string
	}{
		{
			Name: "success for application if label does not exist",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModelWithoutTemplate(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(nil, nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application if formation is already added",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(applicationLblNoFormations, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(&model.Application{}, nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(nil, nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application with new formation",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(&model.Application{}, nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(nil, nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application with default formation",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInputInDefaultScenario).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInputInDefaultScenario).Return(nil)
				return labelService
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(&model.Application{}, nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(nil, nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     defaultFormation,
			ExpectedFormation:  &defaultFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime if label does not exist",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(nil, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, RuntimeID))
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime if formation is already added",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(runtimeLblNoFormations, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(nil, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(nil, nil)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, RuntimeID))
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime with new formation",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(nil, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, apperrors.NewNotFoundError(resource.Webhook, RuntimeID))
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for tenant",
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, TargetTenant).Return(TargetTenant, nil)
				return svc
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", ctx, Tnt).Return(nil)
				labelDefSvc.On("GetAvailableScenarios", ctx, Tnt).Return([]string{testFormationName}, nil)

				return labelDefSvc
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, asa).Return(nil)

				return asaRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("ListOwnedRuntimes", ctx, TargetTenant, runtimeLblFilters).Return(make([]*model.Runtime, 0), nil).Once()
				runtimeRepo.On("ListAll", ctx, TargetTenant, runtimeLblFilters).Return(make([]*model.Runtime, 0), nil).Once()
				return runtimeRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Twice()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for application when label does not exist and can't create it",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(testErr)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application while getting label",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, testErr)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application while converting label values to string slice",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error for application while converting label value to string",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name: "error for application when updating label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(applicationLblNoFormations, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &applicationLblInput).Return(testErr)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application type missing",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				emptyApplicationType := &model.Label{
					ID:         "123",
					Key:        applicationType,
					Tenant:     str.Ptr(Tnt),
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(emptyApplicationType, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "missing applicationType",
		},
		{
			Name: "error for application when updating label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				emptyApplicationType := &model.Label{
					ID:         "123",
					Key:        applicationType,
					Value:      "invalidApplicationType",
					Tenant:     str.Ptr(Tnt),
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(emptyApplicationType, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "unsupported applicationType",
		},
		{
			Name: "error for application when getting application type label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(nil, testErr)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when getting formation template fails",
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when getting formation fails",
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(nil, testErr).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime when label does not exist and can't create it",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(testErr)
				return labelService
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while getting label",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, testErr)
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while converting label values to string slice",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error for runtime while converting label value to string",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil)
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name: "error for runtime when updating label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(runtimeLblNoFormations, nil)
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &runtimeLblInput).Return(testErr)
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when tenant conversion fails",
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, TargetTenant).Return("", testErr)
				return svc
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when create fails",
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetInternalTenant", ctx, TargetTenant).Return(TargetTenant, nil)
				return svc
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, model.AutomaticScenarioAssignment{ScenarioName: testFormationName, Tenant: Tnt, TargetTenantID: TargetTenant}).Return(testErr)

				return asaRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("EnsureScenariosLabelDefinitionExists", ctx, Tnt).Return(nil)
				labelDefSvc.On("GetAvailableScenarios", ctx, Tnt).Return([]string{testFormationName}, nil)

				return labelDefSvc
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when can't get formation by name",
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, Tnt).Return(nil, testErr).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "error when object type is unknown",
			ObjectType:         "UNKNOWN",
			InputFormation:     inputFormation,
			ExpectedErrMessage: "unknown formation type",
		},
		{
			Name: "success for application if label does not exist with both runtime and app-to-app notifications",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeID, RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeID:               fixRuntimeLabelsMap(),
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{Application2ID}).Return(map[string]map[string]interface{}{
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{}).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookID, RuntimeID), fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID), nil)
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp2, Application2ID), nil)
				return repo
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, nil)
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil)
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID),
					Object: &webhook.ApplicationTenantMappingInput{
						Operation:                 model.AssignFormation,
						FormationID:               expectedFormation.ID,
						SourceApplicationTemplate: nil,
						SourceApplication: &webhook.ApplicationWithLabels{
							Application: fixApplicationModelWithoutTemplate(Application2ID),
							Labels:      fixApplicationLabelsMap(),
						},
						TargetApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						TargetApplication: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil)
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp2, Application2ID),
					Object: &webhook.ApplicationTenantMappingInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						SourceApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						SourceApplication: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						TargetApplicationTemplate: nil,
						TargetApplication: &webhook.ApplicationWithLabels{
							Application: fixApplicationModelWithoutTemplate(Application2ID),
							Labels:      fixApplicationLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil)

				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID), fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeID, RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for application webhook client request fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID), nil)
				return repo
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, testErr)

				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when webhook conversion fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(map[string]map[string]interface{}{
					RuntimeContextID: fixRuntimeContextLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtime context labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeContextLabelableObject, []string{RuntimeContextID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.RuntimeContext{fixRuntimeContextModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtime contexts in scenario fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenariosAndRuntimeIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching runtimes in scenario fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(map[string]map[string]interface{}{
					RuntimeContextRuntimeID: fixRuntimeLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching listening runtimes labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.RuntimeLabelableObject, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return([]*model.Runtime{fixRuntimeModel(RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching listening runtimes fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return([]*model.Webhook{fixWebhookModel(WebhookForRuntimeContextID, RuntimeContextRuntimeID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{RuntimeContextRuntimeID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching webhooks fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application template labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application template fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when fetching application fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime if label does not exist with notifications",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:           model.AssignFormation,
						FormationID:         expectedFormation.ID,
						ApplicationTemplate: nil,
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModelWithoutTemplate(Application2ID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime if webhook client call fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: FormationID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: nil,
					},
					CorrelationID: "",
				}).Return(nil, testErr).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if webhook conversion fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeID)).Return(nil, testErr)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application template labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application templates fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching application labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime if there are no applications to notify",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{}, nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime if fetching applications fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching webhooks fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching runtime labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(fixRuntimeModel(RuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if fetching runtime fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeLblInput).Return(nil)
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime context if label does not exist with notifications",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:           model.AssignFormation,
						FormationID:         expectedFormation.ID,
						ApplicationTemplate: nil,
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModelWithoutTemplate(Application2ID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, nil).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime context if webhook client call fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID), nil)
				return conv
			},
			WebhookClientFN: func() *automock.WebhookClient {
				client := &automock.WebhookClient{}
				client.On("Do", ctx, &webhookclient.Request{
					Webhook: *fixWebhookGQLModel(WebhookID, RuntimeContextRuntimeID),
					Object: &webhook.FormationConfigurationChangeInput{
						Operation:   model.AssignFormation,
						FormationID: expectedFormation.ID,
						ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
							ApplicationTemplate: fixApplicationTemplateModel(),
							Labels:              fixApplicationTemplateLabelsMap(),
						},
						Application: &webhook.ApplicationWithLabels{
							Application: fixApplicationModel(ApplicationID),
							Labels:      fixApplicationLabelsMap(),
						},
						Runtime: &webhook.RuntimeWithLabels{
							Runtime: fixRuntimeModel(RuntimeContextRuntimeID),
							Labels:  fixRuntimeLabelsMap(),
						},
						RuntimeContext: &webhook.RuntimeContextWithLabels{
							RuntimeContext: fixRuntimeContextModel(),
							Labels:         fixRuntimeContextLabelsMap(),
						},
					},
					CorrelationID: "",
				}).Return(nil, testErr).Once()
				return client
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if webhook conversion fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", fixWebhookModel(WebhookID, RuntimeContextRuntimeID)).Return(nil, testErr)
				return conv
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application template labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application templates fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(map[string]map[string]interface{}{
					ApplicationID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching application labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID)}, nil)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching applications fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr)
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(fixWebhookModel(WebhookID, RuntimeContextRuntimeID), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching webhook fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(fixRuntimeLabels(), nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByIDAndWebhookType", ctx, Tnt, RuntimeContextRuntimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(fixRuntimeModel(RuntimeContextRuntimeID), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				repo.On("ListForObject", ctx, Tnt, model.RuntimeLabelableObject, RuntimeContextRuntimeID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeRepoFN: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextRuntimeID).Return(nil, testErr)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(fixRuntimeContextLabels(), nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime context labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.RuntimeContextLabelableObject, RuntimeContextID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if fetching runtime context fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &runtimeCtxLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &runtimeCtxLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Once()
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(nil, testErr).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when assigning runtime with runtime type that does not match formation template allowed type",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&model.FormationTemplate{
					ID:          FormationTemplateID,
					Name:        "some-other-template",
					RuntimeType: "not-the-expected-type",
				}, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: fmt.Sprintf("unsupported runtimeType %q for formation template %q, allowing only %q", runtimeTypeLbl.Value, "some-other-template", "not-the-expected-type"),
		},
		{
			Name: "error when assigning runtime fetching runtime type label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(nil, testErr)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when assigning runtime fetching formation template",
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when assigning runtime context whose runtime is with runtime type that does not match formation template allowed type",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&model.FormationTemplate{
					ID:          FormationTemplateID,
					Name:        "some-other-template",
					RuntimeType: "not-the-expected-type",
				}, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: fmt.Sprintf("unsupported runtimeType %q for formation template %q, allowing only %q", runtimeTypeLbl.Value, "some-other-template", "not-the-expected-type"),
		},
		{
			Name: "error when assigning runtime context fetching runtime type label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(nil, testErr)
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when assigning runtime context fetching formation template",
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when assigning runtime context fetching runtime context fails",
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, Tnt, RuntimeContextID).Return(nil, testErr).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: webhook convertion fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{Application2ID}).Return(map[string]map[string]interface{}{
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{}).Return(nil, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: templates list labels for IDs fail",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{Application2ID}).Return(map[string]map[string]interface{}{
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: templates list by IDs fail",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{Application2ID}).Return(map[string]map[string]interface{}{
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				repo.On("ListByIDs", ctx, []string{}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: application labels list for IDs fail",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return([]*model.Application{fixApplicationModelWithoutTemplate(Application2ID)}, nil)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{Application2ID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: application by scenarios and IDs fail",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				repo.On("ListByScenariosAndIDs", ctx, Tnt, []string{inputFormation.Name}, []string{Application2ID}).Return(nil, testErr)
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(fixApplicationTenantMappingWebhookGQLModel(AppTenantMappingWebhookIDForApp1, ApplicationID), nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: self webhook convertion fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(map[string]map[string]interface{}{
					ApplicationTemplateID: fixApplicationTemplateLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			WebhookConverterFN: func() *automock.WebhookConverter {
				repo := &automock.WebhookConverter{}
				repo.On("ToGraphQL", fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID)).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: list labels for app templates of apps already in formation fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				repo.On("ListForObjectIDs", ctx, Tnt, model.AppTemplateLabelableObject, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return([]*model.ApplicationTemplate{fixApplicationTemplateModel()}, nil)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: list app templates of apps already in formation fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(map[string]map[string]interface{}{
					ApplicationID:  fixApplicationLabelsMap(),
					Application2ID: fixApplicationLabelsMap(),
				}, nil)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				repo.On("ListByIDs", ctx, []string{ApplicationTemplateID}).Return(nil, testErr)
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: list labels of apps already in formation fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return([]*model.Application{fixApplicationModel(ApplicationID), fixApplicationModelWithoutTemplate(Application2ID)}, nil).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				repo.On("ListForObjectIDs", ctx, Tnt, model.ApplicationLabelableObject, []string{ApplicationID, Application2ID}).Return(nil, testErr)
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: list apps already in formation fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				repo.On("ListByScenariosNoPaging", ctx, Tnt, []string{expectedFormation.Name}).Return(nil, testErr).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return([]*model.Webhook{fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp1, ApplicationID), fixApplicationTenantMappingWebhookModel(AppTenantMappingWebhookIDForApp2, Application2ID)}, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success when there are no listening apps",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				return repo
			},
			ObjectType:        graphql.FormationObjectTypeApplication,
			ObjectID:          ApplicationID,
			InputFormation:    inputFormation,
			ExpectedFormation: expectedFormation,
		},
		{
			Name: "error while generation app-to-app notifications: list listening apps' webhooks fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Twice()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference).Return(nil, testErr)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: get assigned app's template labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Once()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Twice()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: get assigned app's template fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Twice()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Once()
				repo.On("Get", ctx, ApplicationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: get assigned app's labels fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Twice()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Once()
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(nil, testErr).Once()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while generation app-to-app notifications: get assigned app fails",
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID())
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, ""))
				labelService.On("CreateLabel", ctx, Tnt, fixUUID(), &applicationLblInput).Return(nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFN: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(fixApplicationModel(ApplicationID), nil).Once()
				repo.On("GetByID", ctx, Tnt, ApplicationID).Return(nil, testErr).Once()
				return repo
			},
			LabelRepoFN: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID).Return(fixApplicationLabels(), nil).Once()
				repo.On("ListForObject", ctx, Tnt, model.AppTemplateLabelableObject, ApplicationTemplateID).Return(fixApplicationTemplateLabels(), nil).Once()
				return repo
			},
			WebhookRepoFN: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectTypeAndWebhookType", ctx, Tnt, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference).Return(nil, nil)
				return repo
			},
			ApplicationTemplateRepoFN: func() *automock.ApplicationTemplateRepository {
				repo := &automock.ApplicationTemplateRepository{}
				repo.On("Get", ctx, ApplicationTemplateID).Return(fixApplicationTemplateModel(), nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			uidService := unusedUUIDService()
			if testCase.UIDServiceFn != nil {
				uidService = testCase.UIDServiceFn()
			}
			labelService := unusedLabelService()
			if testCase.LabelServiceFn != nil {
				labelService = testCase.LabelServiceFn()
			}
			asaRepo := unusedASARepo()
			if testCase.AsaRepoFn != nil {
				asaRepo = testCase.AsaRepoFn()
			}
			asaService := unusedASAService()
			if testCase.AsaServiceFN != nil {
				asaService = testCase.AsaServiceFN()
			}
			tenantSvc := &automock.TenantService{}
			if testCase.TenantServiceFn != nil {
				tenantSvc = testCase.TenantServiceFn()
			}
			labelDefService := unusedLabelDefService()
			if testCase.LabelDefServiceFn != nil {
				labelDefService = testCase.LabelDefServiceFn()
			}
			runtimeRepo := unusedRuntimeRepo()
			if testCase.RuntimeRepoFN != nil {
				runtimeRepo = testCase.RuntimeRepoFN()
			}
			runtimeContextRepo := unusedRuntimeContextRepo()
			if testCase.RuntimeContextRepoFn != nil {
				runtimeContextRepo = testCase.RuntimeContextRepoFn()
			}
			formationRepo := unusedFormationRepo()
			if testCase.FormationRepositoryFn != nil {
				formationRepo = testCase.FormationRepositoryFn()
			}
			formationTemplateRepo := unusedFormationTemplateRepo()
			if testCase.FormationTemplateRepositoryFn != nil {
				formationTemplateRepo = testCase.FormationTemplateRepositoryFn()
			}
			applicationRepo := unusedApplicationRepo()
			if testCase.ApplicationRepoFN != nil {
				applicationRepo = testCase.ApplicationRepoFN()
			}
			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFN != nil {
				webhookRepo = testCase.WebhookRepoFN()
			}
			webhookConverter := unusedWebhookConverter()
			if testCase.WebhookConverterFN != nil {
				webhookConverter = testCase.WebhookConverterFN()
			}
			webhookClient := unusedWebhookClient()
			if testCase.WebhookClientFN != nil {
				webhookClient = testCase.WebhookClientFN()
			}
			appTemplateRepo := unusedAppTemplateRepository()
			if testCase.ApplicationTemplateRepoFN != nil {
				appTemplateRepo = testCase.ApplicationTemplateRepoFN()
			}
			labelRepo := unusedLabelRepo()
			if testCase.LabelRepoFN != nil {
				labelRepo = testCase.LabelRepoFN()
			}

			svc := formation.NewService(nil, labelRepo, formationRepo, formationTemplateRepo, labelService, uidService, labelDefService, asaRepo, asaService, tenantSvc, runtimeRepo, runtimeContextRepo, webhookRepo, webhookClient, applicationRepo, appTemplateRepo, webhookConverter, runtimeType, applicationType)

			// WHEN
			actual, err := svc.AssignFormation(ctx, Tnt, testCase.ObjectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, uidService, labelService, asaService, tenantSvc, asaRepo, labelDefService, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, applicationRepo, webhookRepo, webhookConverter, webhookClient, appTemplateRepo, labelRepo)
		})
	}
}
