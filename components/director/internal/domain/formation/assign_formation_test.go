package formation_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

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
	ctxWithTenant := tenant.SaveToContext(context.TODO(), TntInternalID, TntExternalID)

	transactionError := errors.New("transaction error")
	txGen := txtest.NewTransactionContextGenerator(transactionError)

	inputFormation := model.Formation{
		ID:   testFormationID,
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

	notifications := []*webhookclient.FormationAssignmentNotificationRequestTargetMapping{{
		FormationAssignmentNotificationRequest: &webhookclient.FormationAssignmentNotificationRequest{
			Webhook: &graphql.Webhook{
				ID: "wid1",
			},
		},
	}}

	formationAssignmentInputs := []*model.FormationAssignmentInput{{
		FormationID: FormationID,
	}}

	formationAssignments := []*model.FormationAssignment{{
		ID:          FormationAssignmentID,
		FormationID: FormationID,
	}}

	formationAssignments2 := []*model.FormationAssignment{
		{
			ID:          FormationAssignmentID,
			FormationID: FormationID,
			TargetType:  model.FormationAssignmentTypeApplication,
			Target:      ApplicationID,
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
	formationInDraftState := fixFormationModelWithState(model.DraftFormationState)
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
	existingApp := &model.Application{
		BaseEntity: &model.BaseEntity{
			DeletedAt: nil,
			Ready:     true,
		},
	}

	assignments := []*model.FormationAssignment{
		{
			ID:          FormationAssignmentID,
			FormationID: FormationID,
			Target:      RuntimeContextID,
			TargetType:  model.FormationAssignmentTypeRuntimeContext,
		},
	}
	asa := &model.AutomaticScenarioAssignment{
		ScenarioName:   testFormationName,
		Tenant:         TntInternalID,
		TargetTenantID: TargetTenant,
	}

	runtime := &model.Runtime{ID: RuntimeID}

	assignmentOperation := mock.MatchedBy(func(op *model.AssignmentOperationInput) bool {
		return op.Type == model.Assign && op.FormationAssignmentID == FormationAssignmentID && op.FormationID == FormationID && op.TriggeredBy == model.AssignObject
	})

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
		RuntimeRepoFn                 func() *automock.RuntimeRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		NotificationServiceFN         func() *automock.NotificationsService
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		FormationAssignmentServiceFn  func() *automock.FormationAssignmentService
		ConstraintEngineFn            func() *automock.ConstraintEngine
		ASAEngineFn                   func() *automock.AsaEngine
		AssignmentOperationServiceFn  func() *automock.AssignmentOperationService
		ObjectID                      string
		ObjectType                    graphql.FormationObjectType
		InputFormation                model.Formation
		ExpectedFormation             *model.Formation
		ExpectedErrMessage            string
	}{
		{
			Name: "success for application if label does not exist",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: expectEmptySliceApplicationAndReadyApplicationRepo,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeApplication,
			ObjectID:          ApplicationID,
			InputFormation:    inputFormation,
			ExpectedFormation: expectedFormation,
		},
		{
			Name: "success for application if formation is already added",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(applicationLblNoFormations, nil).Once()
				labelService.On("UpdateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, applicationLbl.ID, &applicationLblInput).Return(nil).Once()
				return labelService
			},
			ApplicationRepoFn: expectEmptySliceApplicationAndReadyApplicationRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeApplication,
			ObjectID:          ApplicationID,
			InputFormation:    inputFormation,
			ExpectedFormation: expectedFormation,
		},
		{
			Name: "success for application with new formation",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, applicationLbl.ID, &model.LabelInput{
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

				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), secondTestFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, expectedSecondFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedSecondFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments2, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments2, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeApplication,
			ObjectID:          ApplicationID,
			InputFormation:    inputSecondFormation,
			ExpectedFormation: expectedSecondFormation,
		},
		{
			Name: "success for application when formation does not support runtime, runtime context and tenant",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, applicationLbl.ID, &model.LabelInput{
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

				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), secondTestFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplateNotSupportingRuntime, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, expectedSecondFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedSecondFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments2, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments2, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeApplication,
			ObjectID:          ApplicationID,
			InputFormation:    inputSecondFormation,
			ExpectedFormation: expectedSecondFormation,
		},
		{
			Name: "success for runtime if label does not exist",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &runtimeLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return runtimeRepo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID, graphql.FormationObjectTypeRuntime, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeRuntime,
			ObjectID:          RuntimeID,
			InputFormation:    inputFormation,
			ExpectedFormation: expectedFormation,
		},
		{
			Name: "success for runtime if formation is already added",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeLblInput).Return(runtimeLblNoFormations, nil).Once()
				labelService.On("UpdateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, runtimeLbl.ID, &runtimeLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return runtimeRepo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID, graphql.FormationObjectTypeRuntime, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeRuntime,
			ObjectID:          RuntimeID,
			InputFormation:    inputFormation,
			ExpectedFormation: expectedFormation,
		},
		{
			Name: "success for runtime with new formation",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
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
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID, expectedSecondFormation, model.AssignFormation, graphql.FormationObjectTypeRuntime).Return(notifications, nil).Once()
				return notificationSvc
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return runtimeRepo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID, graphql.FormationObjectTypeRuntime, expectedSecondFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), secondTestFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignRuntimeDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeRuntime,
			ObjectID:          RuntimeID,
			InputFormation:    inputSecondFormation,
			ExpectedFormation: expectedSecondFormation,
		},
		{
			Name: "success for runtime context if label does not exist",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeContextLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Twice()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(assignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), assignments, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ObjectID:          RuntimeContextID,
			InputFormation:    inputFormation,
			ExpectedFormation: expectedFormation,
		},
		{
			Name: "success for runtime context if formation is already added",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeContextLblInput).Return(runtimeContextLblNoFormations, nil).Once()
				labelService.On("UpdateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, runtimeContextLblNoFormations.ID, &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(assignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), assignments, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ObjectID:          RuntimeContextID,
			InputFormation:    inputFormation,
			ExpectedFormation: expectedFormation,
		},
		{
			Name: "success for runtime context with new formation",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeContextID,
					ObjectType: model.RuntimeContextLabelableObject,
					Version:    0,
				}).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
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
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), secondTestFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, expectedSecondFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedSecondFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(assignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), assignments, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeCtxDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignRuntimeCtxDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeRuntimeContext,
			ObjectID:          RuntimeContextID,
			InputFormation:    inputSecondFormation,
			ExpectedFormation: expectedSecondFormation,
		},
		{
			Name: "success for tenant",
			TxFn: txGen.ThatDoesntStartTransaction,
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByExternalID", ctxWithTenantAndLoggerMatcher(), TargetTenant).Return(&model.BusinessTenantMapping{Type: "account"}, nil).Once()
				svc.On("GetInternalTenant", ctxWithTenantAndLoggerMatcher(), TargetTenant).Return(TargetTenant, nil).Once()
				return svc
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("GetAvailableScenarios", ctxWithTenantAndLoggerMatcher(), TntInternalID).Return([]string{testFormationName}, nil).Once()

				return labelDefSvc
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctxWithTenantAndLoggerMatcher(), asa).Return(nil).Once()

				return asaRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignTenantDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignTenantDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("EnsureScenarioAssigned", ctxWithTenantAndLoggerMatcher(), asa, mock.Anything).Return(nil).Once()
				return engine
			},
			ObjectType:        graphql.FormationObjectTypeTenant,
			ObjectID:          TargetTenant,
			InputFormation:    inputFormation,
			ExpectedFormation: expectedFormation,
		},
		{
			Name: "error when creating operation",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenDoesntExpectCommit(1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, applicationLbl.ID, &model.LabelInput{
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

				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), secondTestFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedSecondFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments2, nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", testErr).Once()

				return svc
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputSecondFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when label does not exist and can't create it",
			TxFn: txGen.ThatDoesntExpectCommit,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLbl2Input).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLbl2Input).Return(testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLbl2Input).Return(nil, testErr).Once()
				return labelService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &model.LabelInput{
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
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error when the application is not in ready state",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				return labelService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID:    ApplicationID,
						Ready: false,
					},
				}, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: fmt.Sprintf("application with ID %q is not ready", ApplicationID),
		},
		{
			Name: "error for application if application is not being deleted",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				return labelService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(&model.Application{
					BaseEntity: &model.BaseEntity{
						ID:        ApplicationID,
						DeletedAt: &time.Time{},
						Ready:     true,
					},
				}, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: fmt.Sprintf("application with ID %q is currently being deleted", ApplicationID),
		},
		{
			Name: "error for application while converting label value to string",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLbl2Input).Return(&model.Label{
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
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLbl2Input).Return(applicationLblNoFormations, nil).Once()
				labelService.On("UpdateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, applicationLbl.ID, &applicationLbl2Input).Return(testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(emptyApplicationType, nil).Once()
				return labelService
			},
			ApplicationRepoFn: expectEmptySliceApplicationAndReadyApplicationRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(emptyApplicationType, nil).Once()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(emptyApplicationType, nil).Once()
				return labelService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, assignAppInvalidTypeDetails, FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(secondTestFormationName), FormationTemplateID).Return(nil).Once()
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
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(nil, testErr).Once()
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
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(nil, testErr).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime when runtime does not exist",
			TxFn: txGen.ThatDoesntExpectCommit,
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(nil, testErr).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
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
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &runtimeLblInput).Return(testErr).Once()
				return labelService
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &model.LabelInput{
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
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeLblInput).Return(&model.Label{
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
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeLblInput).Return(runtimeLblNoFormations, nil).Once()
				labelService.On("UpdateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, runtimeLbl.ID, &runtimeLblInput).Return(testErr).Once()
				return labelService
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				svc.On("GetTenantByExternalID", ctxWithTenantAndLoggerMatcher(), TargetTenant).Return(&model.BusinessTenantMapping{Type: "account"}, nil).Once()
				svc.On("GetInternalTenant", ctxWithTenantAndLoggerMatcher(), TargetTenant).Return("", testErr).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignTenantDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				svc.On("GetTenantByExternalID", ctxWithTenantAndLoggerMatcher(), TargetTenant).Return(&model.BusinessTenantMapping{Type: "account"}, nil).Once()
				svc.On("GetInternalTenant", ctxWithTenantAndLoggerMatcher(), TargetTenant).Return(TargetTenant, nil).Once()
				return svc
			},
			AsaRepoFn: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("Create", ctxWithTenantAndLoggerMatcher(), asa).Return(testErr).Once()

				return asaRepo
			},
			LabelDefServiceFn: func() *automock.LabelDefService {
				labelDefSvc := &automock.LabelDefService{}

				labelDefSvc.On("GetAvailableScenarios", ctxWithTenantAndLoggerMatcher(), TntInternalID).Return([]string{testFormationName}, nil).Once()

				return labelDefSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignTenantDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplateNotSupportingRuntime, nil).Once()
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
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), secondTestFormationName, TntInternalID).Return(nil, testErr).Once()
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
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(&formationTemplate, nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(&model.FormationTemplate{
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
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, assignRuntimeOtherTemplateDetails, FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(emptyRuntimeType, nil).Once()
				return labelService
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeID).Return(runtime, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplateNotSupportingRuntime, nil).Once()
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
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(nil, testErr).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeWithRuntimeContextRuntimeTypeLbl, nil).Twice()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Twice()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(&model.FormationTemplate{
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
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, assignRuntimeContextOtherTemplateDetails, FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Twice()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplateNotSupportingRuntime, nil).Once()
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
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(nil, testErr).Once()
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
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Once()
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(nil, testErr).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeWithRuntimeContextRuntimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for application when formation is in initial state",
			TxFn: txGen.ThatSucceedsTwice,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(formationInInitialState, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher, TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, formationInInitialState).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(assignments, nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeApplication,
			ObjectID:          ApplicationID,
			InputFormation:    inputFormation,
			ExpectedFormation: formationInInitialState,
		},
		{
			Name: "success for application when formation is in draft state",
			TxFn: txGen.ThatSucceedsTwice,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(formationInDraftState, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher, TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, formationInDraftState).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(assignments, nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:        graphql.FormationObjectTypeApplication,
			ObjectID:          ApplicationID,
			InputFormation:    inputFormation,
			ExpectedFormation: formationInDraftState,
		},
		{
			Name: "error for application when formation is in deleting state",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(formationInDeletingState, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: "cannot assign to formation with ID",
		},
		{
			Name: "error for application if generating notifications fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			ApplicationRepoFn: expectEmptySliceApplicationAndReadyApplicationRepo,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher, TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(assignments, nil).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(nil, testErr).Once()
				return notificationSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
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
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: unusedNotificationsService,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(nil, testErr).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if persisting formation assignments fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", mock.Anything, TntInternalID, ApplicationID).Return(existingApp, nil).Once()
				return repo
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: unusedNotificationsService,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(nil, testErr).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if processing formation assignments fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenDoesntExpectCommit(3)
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: expectEmptySliceApplicationAndReadyApplicationRepo,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, notifications, mock.Anything, model.AssignFormation).Return(testErr).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if generating notifications fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(3)
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeContextLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Twice()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(nil, nil).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), RuntimeContextID).Return(nil).Once()
				return formationAssignmentSvc
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return(nil, testErr).Once()
				return notificationSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeContextLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Twice()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: unusedNotificationsService,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(nil, testErr).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime context if processing formation assignments fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenDoesntExpectCommit(3)
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeContextLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &runtimeContextLblInput).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Twice()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeRuntimeContext).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID, graphql.FormationObjectTypeRuntimeContext, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(assignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), assignments, notifications, mock.Anything, model.AssignFormation).Return(testErr).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), RuntimeContextID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignRuntimeCtxDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:         graphql.FormationObjectTypeRuntimeContext,
			ObjectID:           RuntimeContextID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if generating formation assignments fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenDoesntExpectCommit(3)
			},
			ApplicationRepoFn: expectEmptySliceApplicationAndReadyApplicationRepo,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, notifications, mock.Anything, model.AssignFormation).Return(testErr).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     inputFormation,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if processing formation assignments fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenDoesntExpectCommit(3)
			},
			ApplicationRepoFn: expectEmptySliceApplicationAndReadyApplicationRepo,
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, notifications, mock.Anything, model.AssignFormation).Return(testErr).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(testErr).Once()
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
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(4)
			},
			UIDServiceFn: func() *automock.UuidService {
				uidService := &automock.UuidService{}
				uidService.On("Generate").Return(fixUUID()).Once()
				return uidService
			},
			ApplicationRepoFn: expectEmptySliceApplicationAndReadyApplicationRepo,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationLblInput).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				labelService.On("CreateLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, fixUUID(), &applicationLblInput).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, expectedFormation, model.AssignFormation, graphql.FormationObjectTypeApplication).Return(notifications, nil).Once()
				return notificationSvc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("GenerateAssignments", ctxWithTenantAndLoggerMatcher(), TntInternalID, ApplicationID, graphql.FormationObjectTypeApplication, expectedFormation).Return(formationAssignmentInputs, nil).Once()
				formationAssignmentSvc.On("PersistAssignments", txtest.CtxWithDBMatcher(), TntInternalID, formationAssignmentInputs).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, notifications, mock.Anything, model.AssignFormation).Return(nil).Once()
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", txtest.CtxWithDBMatcher(), fixUUID(), ApplicationID).Return(nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), preAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctxWithTenantAndLoggerMatcher(), postAssignLocation, fixAssignAppDetails(testFormationName), FormationTemplateID).Return(testErr).Once()
				return engine
			},
			AssignmentOperationServiceFn: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), assignmentOperation).Return("", nil).Once()

				return svc
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &applicationTypeLblInput).Return(emptyApplicationType, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedSecondFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(emptyRuntimeType, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
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
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(nil, testErr).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
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
				labelService.On("GetLabel", ctxWithTenantAndLoggerMatcher(), TntInternalID, &runtimeTypeLblInput).Return(emptyRuntimeType, nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenantAndLoggerMatcher(), TntInternalID, RuntimeContextID).Return(runtimeContext, nil).Once()
				return repo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
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
				svc.On("GetTenantByExternalID", ctxWithTenantAndLoggerMatcher(), TargetTenant).Return(nil, testErr).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctxWithTenantAndLoggerMatcher(), testFormationName, TntInternalID).Return(expectedFormation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctxWithTenantAndLoggerMatcher(), FormationTemplateID).Return(&formationTemplate, nil).Once()
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
			runtimeRepo := unusedRuntimeRepo()
			if testCase.RuntimeRepoFn != nil {
				runtimeRepo = testCase.RuntimeRepoFn()
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
			assignmentOperationService := &automock.AssignmentOperationService{}
			if testCase.AssignmentOperationServiceFn != nil {
				assignmentOperationService = testCase.AssignmentOperationServiceFn()
			}

			svc := formation.NewServiceWithAsaEngine(transact, applicationRepository, nil, nil, formationRepo, formationTemplateRepo, labelService, uidService, labelDefService, asaRepo, asaService, tenantSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, nil, nil, notificationSvc, constraintEngine, runtimeType, applicationType, asaEngine, nil, assignmentOperationService)

			// WHEN
			actual, err := svc.AssignFormation(ctxWithTenant, TntInternalID, testCase.ObjectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}

			mock.AssertExpectationsForObjects(t, persist, uidService, applicationRepository, labelService, asaRepo, asaService, tenantSvc, labelDefService, runtimeContextRepo, runtimeRepo, formationRepo, formationTemplateRepo, webhookClient, notificationSvc, formationAssignmentSvc, constraintEngine, asaEngine, assignmentOperationService)
		})
	}
}
