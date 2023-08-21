package formation_test

import (
	"context"
	"errors"
	"fmt"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"testing"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceAssignFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	transactionError := errors.New("transaction error")
	txGen := txtest.NewTransactionContextGenerator(transactionError)

	inputFormation := model.Formation{
		Name: testFormationName,
	}
	expectedFormation := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
		State:               model.ReadyFormationState,
	}
	expectedFormationTemplate := &model.FormationTemplate{
		ID:                     FormationTemplateID,
		Name:                   testFormationTemplateName,
		RuntimeArtifactKind:    &subscriptionRuntimeArtifactKind,
		RuntimeTypeDisplayName: runtimeTypeDisplayName,
		RuntimeTypes:           []string{runtimeType},
		ApplicationTypes:       []string{applicationType},
	}
	expectedFormationTemplateNotSupportingRuntime := &model.FormationTemplate{
		ID:               FormationTemplateID,
		Name:             testFormationTemplateName,
		RuntimeTypes:     []string{},
		ApplicationTypes: []string{applicationType},
	}
	notifications := []*webhookclient.FormationAssignmentNotificationRequest{{
		Webhook: graphql.Webhook{
			ID: "wid1",
		},
	}}
	formationAssignments := []*model.FormationAssignment{{
		ID: "faid1",
	}}

	formationAssignments2 := []*model.FormationAssignment{
		{
			ID:         "faid1",
			TargetType: model.FormationAssignmentTypeApplication,
			Target:     ApplicationID,
		},
	}

	inputSecondFormation := model.Formation{
		Name: secondTestFormationName,
	}
	expectedSecondFormation := &model.Formation{
		ID:                  fixUUID(),
		Name:                secondTestFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
		State:               model.ReadyFormationState,
	}
	formationInInitialState := fixFormationModelWithState(model.InitialFormationState)
	formationInDeletingState := fixFormationModelWithState(model.DeletingFormationState)

	applicationLblNoFormations := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(TntInternalID),
		Key:        model.ScenariosKey,
		Value:      []interface{}{},
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	applicationLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(TntInternalID),
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
	applicationLbl2Input := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{secondTestFormationName},
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	runtimeLblNoFormations := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(TntInternalID),
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
		Tenant:     str.Ptr(TntInternalID),
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	runtimeLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(TntInternalID),
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
	runtimeWithRuntimeContextRuntimeTypeLbl := &model.Label{
		ID:         "123",
		Key:        runtimeType,
		Value:      runtimeType,
		Tenant:     str.Ptr(TntInternalID),
		ObjectID:   RuntimeContextRuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeTypeLbl := &model.Label{
		ID:         "123",
		Key:        runtimeType,
		Value:      runtimeType,
		Tenant:     str.Ptr(TntInternalID),
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	runtimeContext := &model.RuntimeContext{
		ID:        RuntimeContextRuntimeID,
		RuntimeID: RuntimeID,
	}
	runtimeContextLblInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   RuntimeContextID,
		ObjectType: model.RuntimeContextLabelableObject,
		Version:    0,
	}
	runtimeContextLblNoFormations := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(TntInternalID),
		Key:        model.ScenariosKey,
		Value:      []interface{}{},
		ObjectID:   RuntimeContextID,
		ObjectType: model.RuntimeContextLabelableObject,
		Version:    0,
	}

	assignments := []*model.FormationAssignment{
		{
			Target:     RuntimeContextID,
			TargetType: model.FormationAssignmentTypeRuntimeContext,
		},
	}

	asa := model.AutomaticScenarioAssignment{
		ScenarioName:   testFormationName,
		Tenant:         TntInternalID,
		TargetTenantID: TargetTenant,
	}

	testCases := []struct {
		Name                          string
		TxFn                          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		UIDServiceFn                  func() *automock.UuidService
		ApplicationRepoFn             func() *automock.ApplicationRepository
		LabelServiceFn                func() *automock.LabelService
		LabelDefServiceFn             func() *automock.LabelDefService
		TenantServiceFn               func() *automock.TenantService
		AsaRepoFn                     func() *automock.AutomaticFormationAssignmentRepository
		AsaServiceFN                  func() *automock.AutomaticFormationAssignmentService
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		NotificationServiceFN         func() *automock.NotificationsService
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		FormationAssignmentServiceFn  func() *automock.FormationAssignmentService
		ConstraintEngineFn            func() *automock.ConstraintEngine
		ASAEngineFn                   func() *automock.AsaEngine
		ObjectID                      string
		ObjectType                    graphql.FormationObjectType
		InputFormation                model.Formation
		ExpectedFormation             *model.Formation
		ExpectedErrMessage            string
	}{
		{
			Name: "success for application if label does not exist",
			TxFn: txGen.ThatSucceeds,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments, map[string]string{}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application if formation is already added",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(applicationLblNoFormations, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, applicationLbl.ID, &applicationLblInput).Return(nil).Once()
				return labelService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments, map[string]string{}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application with new formation",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{ApplicationID}).Return([]*model.Application{{BaseEntity: &model.BaseEntity{ID: ApplicationID}, ApplicationTemplateID: str.Ptr(ApplicationTemplateID)}}, nil).Once()

				return repo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, ApplicationID, expectedSecondFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedSecondFormation).Return(formationAssignments2, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments2, map[string]string{}, map[string]string{ApplicationID: ApplicationTemplateID}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application when formation does not support runtime, runtime context and tenant",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{ApplicationID}).Return([]*model.Application{{BaseEntity: &model.BaseEntity{ID: ApplicationID}, ApplicationTemplateID: str.Ptr(ApplicationTemplateID)}}, nil).Once()

				return repo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplateNotSupportingRuntime, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, ApplicationID, expectedSecondFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedSecondFormation).Return(formationAssignments2, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments2, map[string]string{}, map[string]string{ApplicationID: ApplicationTemplateID}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime if label does not exist",
			TxFn: txGen.ThatSucceeds,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &runtimeLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, RuntimeID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, graphql.FormationObjectTypeRuntime, expectedFormation).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments, map[string]string{}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime if formation is already added",
			TxFn: txGen.ThatSucceeds,
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeLblInput).Return(runtimeLblNoFormations, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &runtimeLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, RuntimeID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, graphql.FormationObjectTypeRuntime, expectedFormation).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments, map[string]string{}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime with new formation",
			TxFn: txGen.ThatSucceeds,
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, RuntimeID, expectedSecondFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, graphql.FormationObjectTypeRuntime, expectedSecondFormation).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments, map[string]string{}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignRuntimeDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime context if label does not exist",
			TxFn: txGen.ThatSucceeds,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeContextLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("GetByID", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("ListByIDs", ctx, TntInternalID, []string{RuntimeContextID}).Return([]*model.RuntimeContext{runtimeContext}, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, RuntimeContextID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(assignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, assignments, map[string]string{RuntimeContextRuntimeID: RuntimeID}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime context if formation is already added",
			TxFn: txGen.ThatSucceeds,
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeContextLblInput).Return(runtimeContextLblNoFormations, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeContextLblNoFormations.ID, &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("GetByID", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("ListByIDs", ctx, TntInternalID, []string{RuntimeContextID}).Return([]*model.RuntimeContext{runtimeContext}, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, RuntimeContextID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(assignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, assignments, map[string]string{RuntimeContextRuntimeID: RuntimeID}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime context with new formation",
			TxFn: txGen.ThatSucceeds,
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeContextID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   RuntimeContextID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("GetByID", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("ListByIDs", ctx, TntInternalID, []string{RuntimeContextID}).Return([]*model.RuntimeContext{runtimeContext}, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, RuntimeContextID, expectedSecondFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedSecondFormation).Return(assignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, assignments, map[string]string{RuntimeContextRuntimeID: RuntimeID}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeCtxDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignRuntimeCtxDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for tenant",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByExternalID", ctx, TargetTenant).Return(&model.BusinessTenantMapping{Type: "account"}, nil).Once()
				svc.On("GetInternalTenant", ctx, TargetTenant).Return(TargetTenant, nil).Once()
				return svc
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("GetAvailableScenarios", ctx, TntInternalID).Return([]string{testFormationName}, nil).Once()

				return labelDefSvc
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, asa).Return(nil).Once()

				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignTenantDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignTenantDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("EnsureScenarioAssigned", ctx, asa, mock.Anything).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for application when label does not exist and can't create it",
			TxFn: txGen.ThatDoesntExpectCommit,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLbl2Input).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLbl2Input).Return(testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application while getting label",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLbl2Input).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application while converting label values to string slice",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(TntInternalID),
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error for application while converting label value to string",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLbl2Input).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(TntInternalID),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name: "error for application when updating label fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLbl2Input).Return(applicationLblNoFormations, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, applicationLbl.ID, &applicationLbl2Input).Return(testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application type missing",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				emptyApplicationType := &model.Label{
					ID:         "123",
					Key:        applicationType,
					Tenant:     str.Ptr(TntInternalID),
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(emptyApplicationType, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "missing applicationType",
		},
		{
			Name: "error for application when updating label fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				emptyApplicationType := &model.Label{
					ID:         "123",
					Key:        applicationType,
					Value:      "invalidApplicationType",
					Tenant:     str.Ptr(TntInternalID),
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(emptyApplicationType, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(emptyApplicationType, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, assignAppInvalidTypeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "unsupported applicationType",
		},
		{
			Name: "error for application when getting application type label fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when getting formation template fails",
			TxFn: txGen.ThatDoesntStartTransaction,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
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
			TxFn: txGen.ThatDoesntStartTransaction,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(nil, testErr).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime when label does not exist and can't create it",
			TxFn: txGen.ThatDoesntExpectCommit,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &runtimeLblInput).Return(testErr).Once()
				return labelService
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while getting label",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while converting label values to string slice",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(TntInternalID),
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil).Once()
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error for runtime while converting label value to string",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(TntInternalID),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil).Once()
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name: "error for runtime when updating label fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeLblInput).Return(runtimeLblNoFormations, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &runtimeLblInput).Return(testErr).Once()
				return labelService
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when tenant conversion fails",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByExternalID", ctx, TargetTenant).Return(&model.BusinessTenantMapping{Type: "account"}, nil).Once()
				svc.On("GetInternalTenant", ctx, TargetTenant).Return("", testErr).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignTenantDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when create fails",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByExternalID", ctx, TargetTenant).Return(&model.BusinessTenantMapping{Type: "account"}, nil).Once()
				svc.On("GetInternalTenant", ctx, TargetTenant).Return(TargetTenant, nil).Once()
				return svc
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctx, model.AutomaticScenarioAssignment{ScenarioName: testFormationName, Tenant: TntInternalID, TargetTenantID: TargetTenant}).Return(testErr).Once()

				return asaRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("GetAvailableScenarios", ctx, TntInternalID).Return([]string{testFormationName}, nil).Once()

				return labelDefSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignTenantDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when assigning tenant when formation does not support runtime",
			TxFn: txGen.ThatDoesntStartTransaction,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplateNotSupportingRuntime, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     inputFormation,
			ExpectedErrMessage: fmt.Sprintf("Formation %q of type %q does not support resources of type %q", testFormationName, testFormationTemplateName, graphql.FormationObjectTypeTenant),
		},
		{
			Name: "error when can't get formation by name",
			TxFn: txGen.ThatDoesntStartTransaction,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, TntInternalID).Return(nil, testErr).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputSecondFormation,
			ExpectedFormation:  expectedSecondFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when object type is unknown",
			TxFn: txGen.ThatDoesntStartTransaction,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ObjectType:         "UNKNOWN",
			InputFormation:     inputFormation,
			ExpectedErrMessage: "unknown formation type",
		},
		{
			Name: "error when assigning runtime with runtime type that does not match formation template allowed type",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&model.FormationTemplate{
					ID:                     FormationTemplateID,
					Name:                   "some-other-template",
					RuntimeArtifactKind:    &subscriptionRuntimeArtifactKind,
					RuntimeTypeDisplayName: runtimeTypeDisplayName,
					RuntimeTypes:           []string{"not-the-expected-type"},
				}, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, assignRuntimeOtherTemplateDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: fmt.Sprintf("unsupported runtimeType %q for formation template %q, allowing only %q", runtimeTypeLbl.Value, "some-other-template", []string{"not-the-expected-type"}),
		},
		{
			Name: "error for runtime type label missing",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				emptyRuntimeType := &model.Label{
					ID:         "123",
					Key:        runtimeType,
					Tenant:     str.Ptr(TntInternalID),
					ObjectID:   RuntimeID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(emptyRuntimeType, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "missing runtimeType",
		},
		{
			Name: "error when assigning runtime fetching runtime type label fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when assigning runtime formation does not support runtime",
			TxFn: txGen.ThatDoesntStartTransaction,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplateNotSupportingRuntime, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: fmt.Sprintf("Formation %q of type %q does not support resources of type %q", testFormationName, testFormationTemplateName, graphql.FormationObjectTypeRuntime),
		},
		{
			Name: "error when assigning runtime fetching formation template",
			TxFn: txGen.ThatDoesntStartTransaction,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when assigning runtime context whose runtime is with runtime type that does not match formation template allowed type",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Once()
				repo.On("GetByID", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&model.FormationTemplate{
					ID:                     FormationTemplateID,
					RuntimeArtifactKind:    &subscriptionRuntimeArtifactKind,
					RuntimeTypeDisplayName: runtimeTypeDisplayName,
					Name:                   "some-other-template",
					RuntimeTypes:           []string{"not-the-expected-type"},
				}, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, assignRuntimeContextOtherTemplateDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: fmt.Sprintf("unsupported runtimeType %q for formation template %q, allowing only %q", runtimeTypeLbl.Value, "some-other-template", []string{"not-the-expected-type"}),
		},
		{
			Name: "error when assigning runtime context fetching runtime type label fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Once()
				repo.On("GetByID", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when assigning runtime context when formation does not support runtime context",
			TxFn: txGen.ThatDoesntStartTransaction,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplateNotSupportingRuntime, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: fmt.Sprintf("Formation %q of type %q does not support resources of type %q", testFormationName, testFormationTemplateName, graphql.FormationObjectTypeRuntimeContext),
		},
		{
			Name: "error when assigning runtime context fetching formation template",
			TxFn: txGen.ThatDoesntStartTransaction,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when assigning runtime context fetching runtime context fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Once()
				repo.On("GetByID", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID).Return(nil, testErr).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for application when formation is in initial state",
			TxFn: txGen.ThatSucceeds,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(formationInInitialState, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, formationInInitialState).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:        graphql.FormationObjectTypeApplication,
			ObjectID:          ApplicationID,
			InputFormation:    inputFormation,
			ExpectedFormation: formationInInitialState,
		},
		{
			Name: "error for application when formation is in deleting state",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(formationInDeletingState, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot assign to formation with ID",
		},
		{
			Name: "error for application if generating notifications fails",
			TxFn: txGen.ThatSucceedsTwice,
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once() // in defer
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(nil, nil).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(nil, testErr).Once()
				return notificationSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when listing applications by ID fails",
			TxFn: txGen.ThatSucceedsTwice,
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return(nil, testErr).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once() // in defer
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(nil, nil).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(testErr).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if generating formation assignments fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: unusedNotificationsService,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(nil, testErr).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if runtime context mapping fails",
			TxFn: txGen.ThatSucceedsTwice,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Twice() // second time in defer
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByIDs", ctx, TntInternalID, []string{}).Return(nil, testErr).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: unusedNotificationsService,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(nil, nil).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if processing formation assignments fails",
			TxFn: txGen.ThatSucceedsTwice,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Twice() // second time in defer
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments, map[string]string{}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(testErr).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if generating notifications fails",
			TxFn: txGen.ThatSucceedsTwice,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeContextLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("GetByID", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("ListByIDs", ctx, TntInternalID, []string{}).Return(nil, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(nil, nil).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), RuntimeContextID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", txtest.CtxWithDBMatcher(), RuntimeContextID, testFormationName, graphql.FormationObjectTypeRuntimeContext).Return(false, testErr).Once()
				return engine
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, RuntimeContextID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return(nil, testErr).Once()
				return notificationSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if generating formation assignments fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeContextLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("GetByID", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: unusedNotificationsService,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(nil, testErr).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if runtime context mapping fails",
			TxFn: txGen.ThatSucceedsTwice,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeContextLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("GetByID", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("ListByIDs", ctx, TntInternalID, []string{}).Return(nil, testErr).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", txtest.CtxWithDBMatcher(), RuntimeContextID, testFormationName, graphql.FormationObjectTypeRuntimeContext).Return(false, testErr).Once()
				return engine
			},
			NotificationServiceFN: unusedNotificationsService,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(nil, nil).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), RuntimeContextID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if processing formation assignments fails",
			TxFn: txGen.ThatSucceedsTwice,
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &runtimeContextLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("GetByID", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("ListByIDs", ctx, TntInternalID, []string{RuntimeContextID}).Return(nil, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, RuntimeContextID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(assignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, assignments, map[string]string{}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(testErr).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), RuntimeContextID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", txtest.CtxWithDBMatcher(), RuntimeContextID, testFormationName, graphql.FormationObjectTypeRuntimeContext).Return(false, testErr).Once()
				return engine
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if generating formation assignments fails",
			TxFn: txGen.ThatSucceedsTwice,
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Twice()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByIDs", ctx, TntInternalID, []string{}).Return(nil, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments, map[string]string{}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(testErr).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if processing formation assignments fails",
			TxFn: txGen.ThatSucceedsTwice,
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Twice()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByIDs", ctx, TntInternalID, []string{}).Return(nil, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments, map[string]string{}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(testErr).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error while enforcing constraints pre operation",
			TxFn: txGen.ThatDoesntStartTransaction,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(testErr).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "while enforcing constraints for target operation",
		},
		{
			Name: "error while enforcing constraints post operation",
			TxFn: txGen.ThatSucceedsTwice,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}

				repo.On("ListAllByIDs", ctx, TntInternalID, []string{}).Return([]*model.Application{}, nil).Once()

				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Twice()
				labelService.On("CreateLabel", txtest.CtxWithDBMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctx, TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", txtest.CtxWithDBMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", ctx, formationAssignments, map[string]string{}, map[string]string{}, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(testErr).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "while enforcing constraints for target operation",
		},
		{
			Name: "error while getting application subtype failed to get label",
			TxFn: txGen.ThatDoesntStartTransaction,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
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
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "while getting label",
		},
		{
			Name: "error while getting application subtype if application type missing",
			TxFn: txGen.ThatDoesntStartTransaction,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				emptyApplicationType := &model.Label{
					ID:         "123",
					Key:        applicationType,
					Tenant:     str.Ptr(TntInternalID),
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}
				labelService.On("GetLabel", ctx, TntInternalID, &applicationTypeLblInput).Return(emptyApplicationType, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
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
			ExpectedErrMessage: "Missing application type for application",
		},
		{
			Name: "error while getting runtime type failed to get label",
			TxFn: txGen.ThatDoesntStartTransaction,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "while getting label",
		},
		{
			Name: "error while getting runtime type if runtime type is missing",
			TxFn: txGen.ThatDoesntStartTransaction,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				emptyRuntimeType := &model.Label{
					ID:         "123",
					Key:        runtimeType,
					Tenant:     str.Ptr(TntInternalID),
					ObjectID:   RuntimeID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(emptyRuntimeType, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "Missing runtime type for runtime",
		},
		{
			Name: "error while getting runtime context",
			TxFn: txGen.ThatDoesntStartTransaction,
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(nil, testErr).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "while fetching runtime context with ID",
		},
		{
			Name: "error while getting runtime context type failed to get runtime label",
			TxFn: txGen.ThatDoesntStartTransaction,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "while getting label",
		},
		{
			Name: "error while getting runtime context type if runtime type is missing",
			TxFn: txGen.ThatDoesntStartTransaction,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				emptyRuntimeType := &model.Label{
					ID:         "123",
					Key:        runtimeType,
					Tenant:     str.Ptr(TntInternalID),
					ObjectID:   RuntimeID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(emptyRuntimeType, nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "Missing runtime type for runtime",
		},
		{
			Name: "error while fetching tenant by ID",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByExternalID", ctx, TargetTenant).Return(nil, testErr).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     inputFormation,
			ExpectedFormation:  expectedFormation,
			ExpectedErrMessage: "while getting tenant by external ID",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := txGen.ThatDoesntStartTransaction()
			if testCase.TxFn != nil {
				persist, transact = testCase.TxFn()
			}
			uidService := unusedUUIDService()
			if testCase.UIDServiceFn != nil {
				uidService = testCase.UIDServiceFn()
			}
			applicationRepository := unusedApplicationRepository()
			if testCase.ApplicationRepoFn != nil {
				applicationRepository = testCase.ApplicationRepoFn()
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
			webhookClient := unusedWebhookClient()

			notificationSvc := unusedNotificationsService()
			if testCase.NotificationServiceFN != nil {
				notificationSvc = testCase.NotificationServiceFN()
			}
			formationAssignmentSvc := unusedFormationAssignmentService()
			if testCase.FormationAssignmentServiceFn != nil {
				formationAssignmentSvc = testCase.FormationAssignmentServiceFn()
			}
			constraintEngine := unusedConstraintEngine()
			if testCase.ConstraintEngineFn != nil {
				constraintEngine = testCase.ConstraintEngineFn()
			}
			asaEngine := unusedASAEngine()
			if testCase.ASAEngineFn != nil {
				asaEngine = testCase.ASAEngineFn()
			}

			svc := formation.NewServiceWithAsaEngine(transact, applicationRepository, nil, nil, formationRepo, formationTemplateRepo, labelService, uidService, labelDefService, asaRepo, asaService, tenantSvc, nil, runtimeContextRepo, formationAssignmentSvc, nil, nil, notificationSvc, constraintEngine, runtimeType, applicationType, asaEngine, nil)

			// WHEN
			actual, err := svc.AssignFormation(ctx, TntInternalID, testCase.ObjectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, persist, uidService, applicationRepository, labelService, asaRepo, asaService, tenantSvc, labelDefService, runtimeContextRepo, formationRepo, formationTemplateRepo, webhookClient, notificationSvc, formationAssignmentSvc, constraintEngine, asaEngine)
		})
	}
}
