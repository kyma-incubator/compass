package formation_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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
	ctx = tenant.SaveToContext(ctx, Tnt, ExternalTnt)

	testErr := errors.New("test error")
	transactionError := errors.New("transaction error")
	txGen := txtest.NewTransactionContextGenerator(transactionError)

	formationAssignments := []*model.FormationAssignment{
		{ID: "id1"},
		{ID: "id2"},
	}

	pendingAsyncAssignments := []*model.FormationAssignment{
		{ID: "id1"},
	}

	requests := []*webhookclient.NotificationRequest{
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
		TenantID:            Tnt,
	}
	secondFormation := model.Formation{
		ID:                  fixUUID(),
		Name:                secondTestFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            Tnt,
	}

	applicationLblSingleFormation := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName},
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName, secondTestFormationName},
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}
	applicationLblInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{testFormationName},
		ObjectID:   ApplicationID,
		ObjectType: model.ApplicationLabelableObject,
		Version:    0,
	}

	runtimeLblSingleFormation := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
		Key:        model.ScenariosKey,
		Value:      []interface{}{testFormationName},
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}
	runtimeLbl := &model.Label{
		ID:         "123",
		Tenant:     str.Ptr(Tnt),
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
		Tenant:         Tnt,
		TargetTenantID: TargetTenant,
	}
	expectedFormationTemplate := &model.FormationTemplate{
		ID:               FormationTemplateID,
		Name:             testFormationTemplateName,
		RuntimeTypes:     []string{runtimeType},
		ApplicationTypes: []string{applicationType},
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
		Tenant:     str.Ptr(Tnt),
		ObjectID:   RuntimeID,
		ObjectType: model.RuntimeLabelableObject,
		Version:    0,
	}

	testCases := []struct {
		Name                          string
		TxFn                          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		UIDServiceFn                  func() *automock.UuidService
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
			Name: "success for application",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, ApplicationID).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application if async unassignments exist",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Twice()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, ApplicationID).Return(pendingAsyncAssignments, nil).Once()
				return formationAssignmentSvc
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application if formation do not exist",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			LabelRepoFn:          unusedLabelRepo,
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, ApplicationID).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for application when formation is last",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)

				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLblSingleFormation, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID, model.ScenariosKey).Return(nil).Once()
				return repo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, ApplicationID).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
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
			LabelRepoFn:        unusedLabelRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime if async unassignments exist",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Twice()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Twice()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
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
			LabelRepoFn:        unusedLabelRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "",
		},
		{
			Name: "success for runtime when formation is coming from ASA",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
					FormationType:       "formation-template",
					FormationTemplateID: FormationTemplateID,
					FormationID:         FormationID,
					TenantID:            Tnt,
				}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, details, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, details, FormationTemplateID).Return(nil).Once()
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
		{
			Name: "success for runtime if formation do not exist",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(runtimeLblSingleFormation, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, secondTestFormationName, Tnt).Return(&secondFormation, nil).Once()
				return formationRepo
			},
			LabelRepoFn: unusedLabelRepo,
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{}).Return(nil, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, &secondFormation, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
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
					FormationType:       "formation-template",
					FormationTemplateID: FormationTemplateID,
					FormationID:         FormationID,
					TenantID:            Tnt,
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
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLblSingleFormation, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			LabelRepoFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("Delete", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID, model.ScenariosKey).Return(nil).Once()
				return repo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByIDs", ctx, Tnt, []string{}).Return(nil, nil).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
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
					FormationType:       "formation-template",
					FormationTemplateID: FormationTemplateID,
					FormationID:         FormationID,
					TenantID:            Tnt,
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

				asaRepo.On("DeleteForScenarioName", ctx, Tnt, testFormationName).Return(nil).Once()

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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			Name: "error for application while getting label",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
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
				}, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: "cannot convert label value to slice of strings",
		},
		{
			Name: "error for application while converting label value to string",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: "cannot cast label value as a string",
		},
		{
			Name: "error for application when formation is last and delete fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLblSingleFormation, nil).Once()
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, Tnt, model.ApplicationLabelableObject, ApplicationID, model.ScenariosKey).Return(testErr).Once()
				return labelRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application when updating label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime when fails to check if formations are coming from ASAs",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(nil, testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
				}, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(&model.Label{
					ID:         "123",
					Tenant:     str.Ptr(Tnt),
					Key:        model.ScenariosKey,
					Value:      []interface{}{5},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}, nil).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLblSingleFormation, nil).Once()
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID, model.ScenariosKey).Return(testErr).Once()
				return labelRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			Name: "error for runtime when formation is last and delete fails with not found",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLblSingleFormation, nil).Once()
				return labelService
			},
			LabelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("Delete", ctx, Tnt, model.RuntimeLabelableObject, RuntimeID, model.ScenariosKey).Return(apperrors.NewNotFoundErrorWithType(resource.Formations)).Once()
				return labelRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			ExpectedErrMessage: "",
		},
		{
			Name: "error for runtime when updating label fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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

				asaRepo.On("DeleteForScenarioName", ctx, Tnt, testFormationName).Return(testErr).Once()

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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
				engine.On("RemoveAssignedScenario", ctx, asa, mock.Anything).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for tenant when delete fails",
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
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
			ObjectType:         graphql.FormationObjectTypeTenant,
			ObjectID:           TargetTenant,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when fetching formation fails",
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(nil, testErr).Once()
				return formationRepo
			},
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error when object type is unknown",
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			ObjectType:         "UNKNOWN",
			InputFormation:     in,
			ExpectedErrMessage: "unknown formation type",
		},
		{
			Name: "error for application if generating notifications fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(nil, testErr).Once()
				return notificationSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if generating notifications fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(nil, testErr).Once()
				return notificationSvc
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
			LabelRepoFn:        unusedLabelRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if listing formation assignments for formation fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil)
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil)
				return labelService
			},
			NotificationServiceFN: noActionNotificationsService,
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(nil, testErr)
				return formationAssignmentSvc
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if transaction commit fails after successful assignments processing",
			TxFn: txGen.ThatFailsOnCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, ApplicationID).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name: "error for application if processing formation assignments fails",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(testErr).Once()
				return formationAssignmentSvc
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if list pending formation assignments fail",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, ApplicationID).Return(nil, testErr).Once()
				return formationAssignmentSvc
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if assign due to async pending assignments fail - get application type label fail",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(nil, testErr).Once()
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, ApplicationID).Return(pendingAsyncAssignments, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if assign due to async pending assignments fail - get scenario label fail",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(nil, testErr).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, ApplicationID).Return(pendingAsyncAssignments, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if assign due to async pending assignments fail - update scenario label fail",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Twice()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(testErr).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, ApplicationID).Return(pendingAsyncAssignments, nil).Once()
				return formationAssignmentSvc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},

			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for application if transaction commit fails after processing formation assignments fails",
			TxFn: txGen.ThatFailsOnCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(testErr).Once()
				return formationAssignmentSvc
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: "while committing transaction with error: test error",
		},
		{
			Name: "error for application if listing runtime contexts fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByIDs", mock.Anything, Tnt, []string{}).Return(nil, testErr).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				return formationAssignmentSvc
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if listing runtime contexts fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByIDs", mock.Anything, Tnt, []string{}).Return(nil, testErr).Once()
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
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
			LabelRepoFn:        unusedLabelRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if listing formation assignments for formation fails",
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(nil, testErr).Once()
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
			LabelRepoFn:        unusedLabelRepo,
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
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
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
				return engine
			},
			ASAEngineFn: func() *automock.AsaEngine {
				engine := &automock.AsaEngine{}
				engine.On("IsFormationComingFromASA", ctx, RuntimeID, testFormationName, graphql.FormationObjectTypeRuntime).Return(false, nil)
				return engine
			},
			LabelRepoFn:        unusedLabelRepo,
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
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(testErr).Once()
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
			LabelRepoFn:        unusedLabelRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if list pending formation assignments fail",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
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
			LabelRepoFn:        unusedLabelRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if assign due to async pending assignments fail - get application type label fail",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(nil, testErr).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(pendingAsyncAssignments, nil).Once()
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
			LabelRepoFn:        unusedLabelRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if assign due to async pending assignments fail - get scenario label fail",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(nil, testErr).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(pendingAsyncAssignments, nil).Once()
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
			LabelRepoFn:        unusedLabelRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "error for runtime if assign due to async pending assignments fail - update scenario label fail",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Twice()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{testFormationName, secondTestFormationName},
					ObjectID:   RuntimeID,
					ObjectType: model.RuntimeLabelableObject,
					Version:    0,
				}).Return(testErr).Once()
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, RuntimeID).Return(pendingAsyncAssignments, nil).Once()
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
			LabelRepoFn:        unusedLabelRepo,
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
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(testErr).Once()
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
			LabelRepoFn:        unusedLabelRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedErrMessage: "while committing transaction with error: test error",
		},
		{
			Name: "error for application if transaction begin fails",
			TxFn: txGen.ThatFailsOnBegin,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				return formationAssignmentSvc
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name: "error for runtime if transaction fails on begin",
			TxFn: txGen.ThatFailsOnBegin,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &runtimeTypeLblInput).Return(runtimeTypeLbl, nil).Once()
				labelService.On("GetLabel", ctx, Tnt, runtimeLblInput).Return(runtimeLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, runtimeLbl.ID, &model.LabelInput{
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
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, RuntimeID, expected, model.UnassignFormation, graphql.FormationObjectTypeRuntime).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, RuntimeID).Return(formationAssignments, nil).Once()
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
			LabelRepoFn:        unusedLabelRepo,
			ObjectType:         graphql.FormationObjectTypeRuntime,
			ObjectID:           RuntimeID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: transactionError.Error(),
		},
		{
			Name: "error while enforcing post constraints",
			TxFn: txGen.ThatSucceeds,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				labelService.On("GetLabel", ctx, Tnt, applicationLblInput).Return(applicationLbl, nil).Once()
				labelService.On("UpdateLabel", ctx, Tnt, applicationLbl.ID, &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []string{secondTestFormationName},
					ObjectID:   ApplicationID,
					ObjectType: model.ApplicationLabelableObject,
					Version:    0,
				}).Return(nil).Once()
				return labelService
			},
			RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
			NotificationServiceFN: func() *automock.NotificationsService {
				svc := &automock.NotificationsService{}
				svc.On("GenerateNotifications", ctx, Tnt, ApplicationID, expected, model.UnassignFormation, graphql.FormationObjectTypeApplication).Return(requests, nil).Once()
				return svc
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				formationAssignmentSvc := &automock.FormationAssignmentService{}
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, ApplicationID).Return(formationAssignments, nil).Once()
				formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), requests, mock.Anything).Return(nil).Once()
				formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, ApplicationID).Return(nil, nil).Once()
				return formationAssignmentSvc
			},
			LabelRepoFn: unusedLabelRepo,
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(nil).Once()
				engine.On("EnforceConstraints", ctx, postUnassignLocation, unassignAppDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "While enforcing constraints for target operation",
		},
		{
			Name: "error while enforcing pre constraints",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(applicationTypeLbl, nil)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ConstraintEngineFn: func() *automock.ConstraintEngine {
				engine := &automock.ConstraintEngine{}
				engine.On("EnforceConstraints", ctx, preUnassignLocation, unassignAppDetails, FormationTemplateID).Return(testErr).Once()
				return engine
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "While enforcing constraints for target operation",
		},
		{
			Name: "error while preparing details",
			TxFn: txGen.ThatDoesntExpectCommit,
			LabelServiceFn: func() *automock.LabelService {
				labelService := &automock.LabelService{}
				labelService.On("GetLabel", ctx, Tnt, &applicationTypeLblInput).Return(nil, testErr)
				return labelService
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetByName", ctx, testFormationName, Tnt).Return(expected, nil).Once()
				return formationRepo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
				return repo
			},
			ObjectType:         graphql.FormationObjectTypeApplication,
			ObjectID:           ApplicationID,
			InputFormation:     in,
			ExpectedFormation:  expected,
			ExpectedErrMessage: "While preparing joinpoint details for target operation",
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

			svc := formation.NewServiceWithAsaEngine(transact, nil, labelRepo, formationRepo, formationTemplateRepo, labelService, uidService, nil, asaRepo, asaService, tenantSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, notificationsSvc, constraintEngine, runtimeType, applicationType, asaEngine)

			// WHEN
			actual, err := svc.UnassignFormation(ctx, Tnt, testCase.ObjectID, testCase.ObjectType, testCase.InputFormation)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedFormation, actual)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
				require.Nil(t, actual)
			}
			mock.AssertExpectationsForObjects(t, persist, uidService, labelService, asaRepo, asaService, runtimeRepo, runtimeContextRepo, formationRepo, formationTemplateRepo, labelRepo, notificationsSvc, formationAssignmentSvc, constraintEngine, tenantSvc)
		})
	}
}
