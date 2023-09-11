package formation_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"testing"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceUnassignFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	transactionError := errors.New("transaction error")
	txGen := txtest.NewTransactionContextGenerator(transactionError)

	formationAssignments := []*model.FormationAssignment{
		{ID: "id1"},
		{ID: "id2"},
	}

	pendingAsyncAssignments := []*model.FormationAssignment{
		{ID: "id1"},
	}

	requests := []*webhookclient.FormationAssignmentNotificationRequest{
		{
			Webhook:       graphql.Webhook{},
			Object:        nil,
			CorrelationID: "123",
		},
		{
			Webhook:       graphql.Webhook{},
			Object:        nil,
			CorrelationID: "456",
		},
	}

	in := model.Formation{
		Name: testFormationName,
	}
	secondIn := model.Formation{
		Name: secondTestFormationName,
	}

	expected := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
		State:               model.ReadyFormationState,
	}
	secondFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                secondTestFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
		State:               model.ReadyFormationState,
	}
	formationInInitialState := fixFormationModelWithState(model.InitialFormationState)

	//applicationLblSingleFormation := &model.Label{
	//	ID:         "123",
	//	Tenant:     str.Ptr(TntInternalID),
	//	Key:        model.ScenariosKey,
	//	Value:      []interface{}{testFormationName},
	//	ObjectID:   ApplicationID,
	//	ObjectType: model.ApplicationLabelableObject,
	//	Version:    0,
	//}
	//applicationLbl := &model.Label{
	//	ID:         "123",
	//	Tenant:     str.Ptr(TntInternalID),
	//	Key:        model.ScenariosKey,
	//	Value:      []interface{}{testFormationName, secondTestFormationName},
	//	ObjectID:   ApplicationID,
	//	ObjectType: model.ApplicationLabelableObject,
	//	Version:    0,
	//}
	//applicationLblInput := &model.LabelInput{
	//	Key:        model.ScenariosKey,
	//	Value:      []string{testFormationName},
	//	ObjectID:   ApplicationID,
	//	ObjectType: model.ApplicationLabelableObject,
	//	Version:    0,
	//}

	runtimeLblSingleFormation := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(TntInternalID),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName},
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(TntInternalID),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName, secondTestFormationName},
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeLblInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	asa := model.AutomaticScenarioAssignment{
		ScenarioName:   testFormationName,
		Tenant:         TntInternalID,
		TargetTenantID: TargetTenant,
	}
	expectedFormationTemplate := &model.FormationTemplate{
		ID:               FormationTemplateID,
		Name:             testFormationTemplateName,
		RuntimeTypes:     []string{runtimeType},
		ApplicationTypes: []string{applicationType},
	}
	//applicationTypeLblInput := model.LabelInput{
	//	Key:        applicationType,
	//	ObjectID:   ApplicationID,
	//	ObjectType: model.ApplicationLabelableObject,
	//	Version:    0,
	//}
	//applicationTypeLbl := &model.Label{
	//	ID:         "123",
	//	Key:        applicationType,
	//	Value:      applicationType,
	//	Tenant:     str.Ptr(TntInternalID),
	//	ObjectID:   ApplicationID,
	//	ObjectType: model.ApplicationLabelableObject,
	//	Version:    0,
	//}
	runtimeTypeLblInput := model.LabelInput{
		Key:        runtimeType,
		ObjectID:   RuntimeID,
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

	testCases := []struct {
		Name                          string
		TxFn                          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		UIDServiceFn                  func() *automock.UuidService
		ApplicationRepoFn             func() *automock.ApplicationRepository
		LabelServiceFn                func() *automock.LabelService
		LabelRepoFn                   func() *automock.LabelRepository
		AsaServiceFN                  func() *automock.AutomaticFormationAssignmentService
		AsaRepoFN                     func() *automock.AutomaticFormationAssignmentRepository
		RuntimeRepoFN                 func() *automock.RuntimeRepository
		RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
		NotificationServiceFN         func() *automock.NotificationsService
		FormationAssignmentServiceFn  func() *automock.FormationAssignmentService
		TenantServiceFn               func() *automock.TenantService
		ConstraintEngineFn            func() *automock.ConstraintEngine
		ASAEngineFn                   func() *automock.AsaEngine
		ObjectType                    graphql.FormationObjectType
		ObjectID                      string
		InputFormation                model.Formation
		ExpectedFormation             *model.Formation
		ExpectedErrMessage            string
	}{
		{
			Name: "success for runtime",
			TxFn: txGen.ThatSucceedsTwice,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime when formation has state different that ready should delete FA without sending notifications",
			TxFn: txGen.ThatDoesntStartTransaction,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, TntInternalID, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(formationInInitialState, nil).Once()
				return formationRepo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", ctx, formationInInitialState.ID, RuntimeID).Return(nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  formationInInitialState,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime when formation has state different that ready should delete FA without sending notifications but fails while updating scenario label",
			TxFn: txGen.ThatDoesntStartTransaction,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, TntInternalID, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(formationInInitialState, nil).Once()
				return formationRepo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", ctx, formationInInitialState.ID, RuntimeID).Return(nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  formationInInitialState,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "success for runtime if async unassignments exist",
			TxFn: txGen.ThatSucceedsTwice,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(pendingAsyncAssignments, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime if formation do not exist",
			TxFn: txGen.ThatSucceedsTwice,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLblSingleFormation, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, secondTestFormationName, TntInternalID).Return(&secondFormation, nil).Once()
				return formationRepo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, &secondFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				details := &formationconstraint.UnassignFormationOperationDetails{
					ResourceType:        model.RuntimeResourceType,
					ResourceSubtype:     runtimeType,
					ResourceID:          RuntimeID,
					FormationType:       testFormationTemplateName,
					FormationTemplateID: FormationTemplateID,
					FormationID:         FormationID,
					TenantID:            TntInternalID,
				}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, details, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, details, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, secondTestFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     secondIn,
			ExpectedFormation:  &secondFormation,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime when formation is last",
			TxFn: txGen.ThatSucceedsTwice,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(runtimeLblSingleFormation, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", txtest.CtxWithDBMatcher(), TntInternalID, model.RuntimeLabelableObject, RuntimeID, model.ScenariosKey).Return(nil).Once()
				return repo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				details := &formationconstraint.UnassignFormationOperationDetails{
					ResourceType:        model.RuntimeResourceType,
					ResourceSubtype:     runtimeType,
					ResourceID:          RuntimeID,
					FormationType:       testFormationTemplateName,
					FormationTemplateID: FormationTemplateID,
					FormationID:         FormationID,
					TenantID:            TntInternalID,
				}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, details, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, details, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for tenant",
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}

				asaRepo.On("DeleteForScenarioName", ctx, TntInternalID, testFormationName).Return(nil).Once()

				return asaRepo
			},
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormationName).Return(asa, nil).Once()
				return asaService
			},
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByExternalID", ctx, TargetTenant).Return(&model.BusinessTenantMapping{Type: "account"}, nil)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignTenantDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, unassignTenantDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, TargetTenant, testFormationName, graphql.FormationObjectTypeTenant).Return(false, nil)
				engine.On("RemoveAssignedScenario", ctx, asa, mock.Anything).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime when formation has state different than ready and fails to delete the formation assignments",
			TxFn: txGen.ThatDoesntStartTransaction,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(formationInInitialState, nil).Once()
				return formationRepo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("DeleteAssignmentsForObjectID", ctx, formationInInitialState.ID, RuntimeID).Return(testErr).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime when fails to check if formations are coming from ASAs",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, testErr)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while getting label",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return([]*model.FormationAssignment{}, nil).Once()
				return formationAssignmentSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while converting label values to string slice",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
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
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return([]*model.FormationAssignment{}, nil).Once()
				return formationAssignmentSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error for runtime while converting label value to string",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(&model.Label{
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
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return([]*model.FormationAssignment{}, nil).Once()
				return formationAssignmentSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name: "error for runtime when formation is last and delete fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(runtimeLblSingleFormation, nil).Once()
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", txtest.CtxWithDBMatcher(), TntInternalID, model.RuntimeLabelableObject, RuntimeID, model.ScenariosKey).Return(testErr).Once()
				return labelRepo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return([]*model.FormationAssignment{}, nil).Once()
				return formationAssignmentSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime when updating label fails",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(testErr).Once()
				return labelService
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return([]*model.FormationAssignment{}, nil).Once()
				return formationAssignmentSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when delete fails",
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}
				asaRepo.On("DeleteForScenarioName", ctx, TntInternalID, testFormationName).Return(testErr).Once()
				return asaRepo
			},
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormationName).Return(asa, nil).Once()
				return asaService
			},
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByExternalID", ctx, TargetTenant).Return(&model.BusinessTenantMapping{Type: "account"}, nil)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, TargetTenant, testFormationName, graphql.FormationObjectTypeTenant).Return(false, nil)
				return engine
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignTenantDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant while enforcing post constraints",
			AsaRepoFN: func() *automock.AutomaticFormationAssignmentRepository {
				asaRepo := &automock.AutomaticFormationAssignmentRepository{}

				asaRepo.On("DeleteForScenarioName", ctx, TntInternalID, testFormationName).Return(nil).Once()

				return asaRepo
			},
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormationName).Return(asa, nil).Once()
				return asaService
			},
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByExternalID", ctx, TargetTenant).Return(&model.BusinessTenantMapping{Type: "account"}, nil)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignTenantDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, unassignTenantDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, TargetTenant, testFormationName, graphql.FormationObjectTypeTenant).Return(false, nil)
				engine.On("RemoveAssignedScenario", ctx, asa, mock.Anything).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when getting asa fails",
			AsaServiceFN: func() *automock.AutomaticFormationAssignmentService {
				asaService := &automock.AutomaticFormationAssignmentService{}
				asaService.On("GetForScenarioName", ctx, testFormationName).Return(model.AutomaticScenarioAssignment{}, testErr).Once()
				return asaService
			},
			TenantServiceFn: func() *automock.TenantService {
				svc := &automock.TenantService{}
				svc.On("GetTenantByExternalID", ctx, TargetTenant).Return(&model.BusinessTenantMapping{Type: "account"}, nil)
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignTenantDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, TargetTenant, testFormationName, graphql.FormationObjectTypeTenant).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when fetching formation fails",
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(nil, testErr).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:               "error when object type is unknown",
			ObjectType:         "UNKNOWN",
			InputFormation:     in,
			ExpectedErrMessage: "unknown formation type",
		},
		{
			Name: "success for runtime if generating notifications fails with not found",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(nil, apperrors.NewNotFoundError(resource.Runtime, RuntimeID)).Once()
				return notificationSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:        graphql.FormationObjectTypeRuntime,
			ObjectID:          RuntimeID,
			InputFormation:    in,
			ExpectedFormation: expected,
		},
		{
			Name: "error for runtime if generating notifications fails with not found and fails during commit",
			TxFn: txGen.ThatFailsOnCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(nil, apperrors.NewNotFoundError(resource.Runtime, RuntimeID)).Once()
				return notificationSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: "while committing transaction with error",
		},
		{
			Name: "error for runtime if generating notifications fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if listing applications fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListAllByIDs", mock.Anything, TntInternalID, []string{}).Return(nil, testErr).Once()
				return repo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if listing runtime contexts fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByIDs", mock.Anything, TntInternalID, []string{}).Return(nil, testErr).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if listing formation assignments for formation fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(nil, testErr).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if transaction commit fails",
			TxFn: txGen.ThatFailsOnCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name: "error for runtime if processing formation assignments fails",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(testErr).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime while beginning scenario transaction",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenFailsOnBegin(1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name: "error for runtime if list pending formation assignments fail",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(nil, testErr).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if assign due to async pending assignments fail - get application type label fail",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return([]*model.FormationAssignment{}, nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if assign due to async pending assignments fail - get scenario label fail",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return([]*model.FormationAssignment{}, nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if assign due to async pending assignments fail - update scenario label fail",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return([]*model.FormationAssignment{}, nil).Once()
				return formationAssignmentSvc
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if transaction commit fails after processing formation assignments fails",
			TxFn: txGen.ThatFailsOnCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(testErr).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: "while committing transaction with error: Test error",
		},
		{
			Name: "error for runtime if transaction fails on begin",
			TxFn: txGen.ThatFailsOnBegin,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name: "error while committing scenario label transaction",
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndThenFailsOnCommit(1)
			},
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name: "error while enforcing post constraints",
			TxFn: txGen.ThatSucceedsTwice,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			ApplicationRepoFn:    expectEmptySliceApplicationRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "while enforcing constraints for target operation",
		},
		{
			Name: "error while enforcing pre constraints",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignRuntimeDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "while enforcing constraints for target operation",
		},
		{
			Name: "error while preparing details",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(nil, testErr)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "while preparing joinpoint details for target operation",
		},
		{
			Name: "success for runtime when formation is coming from ASA",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, TntInternalID, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				details := &formationconstraint.UnassignFormationOperationDetails{
					ResourceType:        model.RuntimeResourceType,
					ResourceSubtype:     runtimeType,
					ResourceID:          RuntimeID,
					FormationType:       testFormationTemplateName,
					FormationTemplateID: FormationTemplateID,
					FormationID:         FormationID,
					TenantID:            TntInternalID,
				}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, details, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(true, nil)
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
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
			if testCase.AsaRepoFN != nil {
				asaRepo = testCase.AsaRepoFN()
			}
			asaService := unusedASAService()
			if testCase.AsaServiceFN != nil {
				asaService = testCase.AsaServiceFN()
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
			labelRepo := unusedLabelRepo()
			if testCase.LabelRepoFn != nil {
				labelRepo = testCase.LabelRepoFn()
			}
			notificationsSvc := unusedNotificationsService()
			if testCase.NotificationServiceFN != nil {
				notificationsSvc = testCase.NotificationServiceFN()
			}
			formationAssignmentSvc := unusedFormationAssignmentService()
			if testCase.FormationAssignmentServiceFn != nil {
				formationAssignmentSvc = testCase.FormationAssignmentServiceFn()
			}
			constraintEngine := unusedConstraintEngine()
			if testCase.ConstraintEngineFn != nil {
				constraintEngine = testCase.ConstraintEngineFn()
			}
			tenantSvc := unusedTenantService()
			if testCase.TenantServiceFn != nil {
				tenantSvc = testCase.TenantServiceFn()
			}
			asaEngine := unusedASAEngine()
			if testCase.ASAEngineFn != nil {
				asaEngine = testCase.ASAEngineFn()
			}

			svc := formation.NewServiceWithAsaEngine(transact, applicationRepository, nil, labelRepo, formationRepo, formationTemplateRepo, labelService, uidService, nil, asaRepo, asaService, tenantSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, nil, nil, notificationsSvc, constraintEngine, runtimeType, applicationType, asaEngine, nil)

			// WHEN
			actual, err := svc.UnassignFormation(ctx, TntInternalID, testCase.ObjectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}
			mock.AssertExpectationsForObjects(t, persist, uidService, labelService, applicationRepository, asaRepo, asaService, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, labelRepo, notificationsSvc, formationAssignmentSvc, constraintEngine, tenantSvc, asaEngine)
		})
	}
}
