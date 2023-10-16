package formation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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
		{ID: FormationID},
		{ID: FormationID2},
	}

	assignment1Clone := formationAssignments[0].Clone()
	assignment1Clone.State = string(model.DeletingAssignmentState)
	assignment2Clone := formationAssignments[1].Clone()
	assignment2Clone.State = string(model.DeletingAssignmentState)

	formationAssignmentsInDeletingState := []*model.FormationAssignment{
		assignment1Clone,
		assignment2Clone,
	}

	pendingAsyncAssignments := []*model.FormationAssignment{
		{ID: FormationID},
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

	expectedFormationTemplate := &model.FormationTemplate{
		ID:               FormationTemplateID,
		Name:             testFormationTemplateName,
		RuntimeTypes:     []string{runtimeType},
		ApplicationTypes: []string{applicationType},
	}

	mockData := []struct {
		Name                 string
		ObjectID             string
		ObjectType           graphql.FormationObjectType
		LabelType            model.LabelableObject
		ScenarioLabel        *model.Label
		ScenarioLabelInput   *model.LabelInput
		TypeLabel            *model.Label
		TypeLabelInput       *model.LabelInput
		SingleFormationLabel *model.Label
		UnassignDetails      *formationconstraint.UnassignFormationOperationDetails
	}{
		{
			ObjectID:   RuntimeID,
			ObjectType: graphql.FormationObjectTypeRuntime,
			LabelType:  model.RuntimeLabelableObject,
			ScenarioLabel: &model.Label{
				ID:         "123",
				Tenant:     str.Ptr(TntInternalID),
				Key:        model.ScenariosKey,
				Value:      []interface{}{testFormationName, secondTestFormationName},
				ObjectID:   RuntimeID,
				ObjectType: model.RuntimeLabelableObject,
				Version:    0,
			},
			ScenarioLabelInput: &model.LabelInput{
				Key:        model.ScenariosKey,
				Value:      []string{testFormationName},
				ObjectID:   RuntimeID,
				ObjectType: model.RuntimeLabelableObject,
				Version:    0,
			},
			TypeLabel: &model.Label{
				ID:         "123",
				Key:        runtimeType,
				Value:      runtimeType,
				Tenant:     str.Ptr(TntInternalID),
				ObjectID:   RuntimeID,
				ObjectType: model.RuntimeLabelableObject,
				Version:    0,
			},
			TypeLabelInput: &model.LabelInput{
				Key:        runtimeType,
				ObjectID:   RuntimeID,
				ObjectType: model.RuntimeLabelableObject,
				Version:    0,
			},
			SingleFormationLabel: &model.Label{
				ID:         "123",
				Tenant:     str.Ptr(TntInternalID),
				Key:        model.ScenariosKey,
				Value:      []interface{}{testFormationName},
				ObjectID:   RuntimeID,
				ObjectType: model.RuntimeLabelableObject,
				Version:    0,
			},
			UnassignDetails: unassignRuntimeDetails,
		},
		{
			ObjectID:   ApplicationID,
			ObjectType: graphql.FormationObjectTypeApplication,
			LabelType:  model.ApplicationLabelableObject,
			ScenarioLabel: &model.Label{
				ID:         "123",
				Tenant:     str.Ptr(TntInternalID),
				Key:        model.ScenariosKey,
				Value:      []interface{}{testFormationName, secondTestFormationName},
				ObjectID:   ApplicationID,
				ObjectType: model.ApplicationLabelableObject,
				Version:    0,
			},
			ScenarioLabelInput: &model.LabelInput{
				Key:        model.ScenariosKey,
				Value:      []string{testFormationName},
				ObjectID:   ApplicationID,
				ObjectType: model.ApplicationLabelableObject,
				Version:    0,
			},
			TypeLabel: &model.Label{
				ID:         "123",
				Key:        applicationType,
				Value:      applicationType,
				Tenant:     str.Ptr(TntInternalID),
				ObjectID:   ApplicationID,
				ObjectType: model.ApplicationLabelableObject,
				Version:    0,
			},
			TypeLabelInput: &model.LabelInput{
				Key:        applicationType,
				ObjectID:   ApplicationID,
				ObjectType: model.ApplicationLabelableObject,
				Version:    0,
			},
			SingleFormationLabel: &model.Label{
				ID:         "123",
				Tenant:     str.Ptr(TntInternalID),
				Key:        model.ScenariosKey,
				Value:      []interface{}{testFormationName},
				ObjectID:   ApplicationID,
				ObjectType: model.ApplicationLabelableObject,
				Version:    0,
			},
			UnassignDetails: unassignAppDetails,
		},
		{
			ObjectID:   RuntimeContextID,
			ObjectType: graphql.FormationObjectTypeRuntimeContext,
			LabelType:  model.RuntimeContextLabelableObject,
			ScenarioLabel: &model.Label{
				ID:         "123",
				Tenant:     str.Ptr(TntInternalID),
				Key:        model.ScenariosKey,
				Value:      []interface{}{testFormationName, secondTestFormationName},
				ObjectID:   RuntimeContextID,
				ObjectType: model.RuntimeContextLabelableObject,
				Version:    0,
			},
			ScenarioLabelInput: &model.LabelInput{
				Key:        model.ScenariosKey,
				Value:      []string{testFormationName},
				ObjectID:   RuntimeContextID,
				ObjectType: model.RuntimeContextLabelableObject,
				Version:    0,
			},
			TypeLabel: &model.Label{
				ID:         "123",
				Key:        runtimeType,
				Value:      runtimeType,
				Tenant:     str.Ptr(TntInternalID),
				ObjectID:   RuntimeContextRuntimeID,
				ObjectType: model.RuntimeLabelableObject,
				Version:    0,
			},
			TypeLabelInput: &model.LabelInput{
				Key:        runtimeType,
				ObjectID:   RuntimeContextRuntimeID,
				ObjectType: model.RuntimeLabelableObject,
				Version:    0,
			},
			SingleFormationLabel: &model.Label{
				ID:         "123",
				Tenant:     str.Ptr(TntInternalID),
				Key:        model.ScenariosKey,
				Value:      []interface{}{testFormationName},
				ObjectID:   RuntimeContextID,
				ObjectType: model.RuntimeContextLabelableObject,
				Version:    0,
			},
			UnassignDetails: unassignRuntimeContextDetails,
		},
	}

	for _, objectTypeData := range mockData {
		testCases := []struct {
			Name                          string
			TxFn                          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
			ApplicationRepoFn             func() *automock.ApplicationRepository
			LabelServiceFn                func() *automock.LabelService
			LabelRepoFn                   func() *automock.LabelRepository
			RuntimeContextRepoFn          func() *automock.RuntimeContextRepository
			FormationRepositoryFn         func() *automock.FormationRepository
			FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
			NotificationServiceFN         func() *automock.NotificationsService
			FormationAssignmentServiceFn  func() *automock.FormationAssignmentService
			ConstraintEngineFn            func() *automock.ConstraintEngine
			ASAEngineFn                   func() *automock.AsaEngine
			ObjectType                    graphql.FormationObjectType
			ObjectID                      string
			InputFormation                model.Formation
			ExpectedFormation             *model.Formation
			ExpectedErrMessage            string
		}{
			{
				Name: "success",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimes(3)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(objectTypeData.ScenarioLabel, nil).Once()
					labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabel.ID, &model.LabelInput{
						Key:        model.ScenariosKey,
						Value:      []string{secondTestFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return(nil, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					engine.On("EnforceConstraints", ctx, postUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:        objectTypeData.ObjectType,
				ObjectID:          objectTypeData.ObjectID,
				InputFormation:    in,
				ExpectedFormation: expected,
			},
			{
				Name: "success when formation has state different that ready should delete FA without sending notifications",
				TxFn: txGen.ThatDoesntStartTransaction,
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.ScenarioLabelInput).Return(objectTypeData.ScenarioLabel, nil).Once()
					labelService.On("UpdateLabel", ctx, TntInternalID, objectTypeData.ScenarioLabel.ID, &model.LabelInput{
						Key:        model.ScenariosKey,
						Value:      []string{secondTestFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
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
					formationAssignmentSvc.On("DeleteAssignmentsForObjectID", ctx, formationInInitialState.ID, objectTypeData.ObjectID).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:        objectTypeData.ObjectType,
				ObjectID:          objectTypeData.ObjectID,
				InputFormation:    in,
				ExpectedFormation: formationInInitialState,
			},
			{
				Name: "error when formation has state different that ready should delete FA without sending notifications but fails while updating scenario label",
				TxFn: txGen.ThatDoesntStartTransaction,
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.ScenarioLabelInput).Return(objectTypeData.ScenarioLabel, nil).Once()
					labelService.On("UpdateLabel", ctx, TntInternalID, objectTypeData.ScenarioLabel.ID, &model.LabelInput{
						Key:        model.ScenariosKey,
						Value:      []string{secondTestFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
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
					formationAssignmentSvc.On("DeleteAssignmentsForObjectID", ctx, formationInInitialState.ID, objectTypeData.ObjectID).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedFormation:  formationInInitialState,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "success if async unassignments exist",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimes(3)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return(pendingAsyncAssignments, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					engine.On("EnforceConstraints", ctx, postUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:        objectTypeData.ObjectType,
				ObjectID:          objectTypeData.ObjectID,
				InputFormation:    in,
				ExpectedFormation: expected,
			},
			{
				Name: "success if formation do not exist",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimes(3)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &model.LabelInput{
						Key:        model.ScenariosKey,
						Value:      []string{secondTestFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
						Version:    0,
					}).Return(objectTypeData.SingleFormationLabel, nil).Once()
					labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabel.ID, &model.LabelInput{
						Key:        model.ScenariosKey,
						Value:      []string{testFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, &secondFormation, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return(nil, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					engine.On("EnforceConstraints", ctx, postUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, secondTestFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:        objectTypeData.ObjectType,
				ObjectID:          objectTypeData.ObjectID,
				InputFormation:    secondIn,
				ExpectedFormation: &secondFormation,
			},
			{
				Name: "success when formation is last",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimes(3)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(objectTypeData.SingleFormationLabel, nil).Once()
					return labelService
				},
				FormationRepositoryFn: func() *automock.FormationRepository {
					formationRepo := &automock.FormationRepository{}
					formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
					return formationRepo
				},
				LabelRepoFn: func() *automock.LabelRepository {
					repo := &automock.LabelRepository{}
					repo.On("Delete", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.LabelType, objectTypeData.ObjectID, model.ScenariosKey).Return(nil).Once()
					return repo
				},
				ApplicationRepoFn:    expectEmptySliceApplicationRepo,
				RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
				NotificationServiceFN: func() *automock.NotificationsService {
					svc := &automock.NotificationsService{}
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return(nil, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					engine.On("EnforceConstraints", ctx, postUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:        objectTypeData.ObjectType,
				ObjectID:          objectTypeData.ObjectID,
				InputFormation:    in,
				ExpectedFormation: expected,
			},
			{
				Name: "error when formation has state different than ready and fails to delete the formation assignments",
				TxFn: txGen.ThatDoesntStartTransaction,
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					return labelService
				},
				FormationRepositoryFn: func() *automock.FormationRepository {
					formationRepo := &automock.FormationRepository{}
					formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(formationInInitialState, nil).Once()
					return formationRepo
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("DeleteAssignmentsForObjectID", ctx, formationInInitialState.ID, objectTypeData.ObjectID).Return(testErr).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error when fails to check if formations are coming from ASAs",
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
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, testErr)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedFormation:  expected,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error while getting label",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(nil, testErr).Once()
					return labelService
				},
				ApplicationRepoFn:    expectEmptySliceApplicationRepo,
				RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
				NotificationServiceFN: func() *automock.NotificationsService {
					svc := &automock.NotificationsService{}
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return([]*model.FormationAssignment{}, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()

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
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error while converting label values to string slice",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, &model.LabelInput{
						Key:        model.ScenariosKey,
						Value:      []string{testFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
						Version:    0,
					}).Return(&model.Label{
						ID:         "123",
						Tenant:     str.Ptr(TntInternalID),
						Key:        model.ScenariosKey,
						Value:      []string{testFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
						Version:    0,
					}, nil).Once()
					return labelService
				},
				ApplicationRepoFn:    expectEmptySliceApplicationRepo,
				RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
				NotificationServiceFN: func() *automock.NotificationsService {
					svc := &automock.NotificationsService{}
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return([]*model.FormationAssignment{}, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
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
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: "cannot convert label value to slice of strings",
			},
			{
				Name: "error while converting label value to string",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(&model.Label{
						ID:         "123",
						Tenant:     str.Ptr(TntInternalID),
						Key:        model.ScenariosKey,
						Value:      []interface{}{5},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
						Version:    0,
					}, nil).Once()
					return labelService
				},
				ApplicationRepoFn:    expectEmptySliceApplicationRepo,
				RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
				NotificationServiceFN: func() *automock.NotificationsService {
					svc := &automock.NotificationsService{}
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return([]*model.FormationAssignment{}, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()

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
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: "cannot cast label value as a string",
			},
			{
				Name: "error when formation is last and delete fails",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(objectTypeData.SingleFormationLabel, nil).Once()
					return labelService
				},
				LabelRepoFn: func() *automock.LabelRepository {
					labelRepo := &automock.LabelRepository{}
					labelRepo.On("Delete", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.LabelType, objectTypeData.ObjectID, model.ScenariosKey).Return(testErr).Once()
					return labelRepo
				},
				ApplicationRepoFn:    expectEmptySliceApplicationRepo,
				RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
				NotificationServiceFN: func() *automock.NotificationsService {
					svc := &automock.NotificationsService{}
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return([]*model.FormationAssignment{}, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
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
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error when updating label fails",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(objectTypeData.ScenarioLabel, nil).Once()
					labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabel.ID, &model.LabelInput{
						Key:        model.ScenariosKey,
						Value:      []string{secondTestFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
						Version:    0,
					}).Return(testErr).Once()
					return labelService
				},
				ApplicationRepoFn:    expectEmptySliceApplicationRepo,
				RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
				NotificationServiceFN: func() *automock.NotificationsService {
					svc := &automock.NotificationsService{}
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return([]*model.FormationAssignment{}, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
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
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
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
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedFormation:  expected,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "success if generating notifications fails with not found",
				TxFn: txGen.ThatSucceedsTwice,
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					return labelService
				},
				FormationRepositoryFn: func() *automock.FormationRepository {
					formationRepo := &automock.FormationRepository{}
					formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
					return formationRepo
				},
				NotificationServiceFN: func() *automock.NotificationsService {
					notificationSvc := &automock.NotificationsService{}
					notificationSvc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(nil, apperrors.NewNotFoundError(resource.Runtime, objectTypeData.ObjectID)).Once()
					return notificationSvc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ObjectType:        objectTypeData.ObjectType,
				ObjectID:          objectTypeData.ObjectID,
				InputFormation:    in,
				ExpectedFormation: expected,
			},
			{
				Name: "Error when there is an error invoking the defer cleanup and it fails when updating assignment",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					persistTx := &persistenceautomock.PersistenceTx{}
					transact := &persistenceautomock.Transactioner{}

					transact.On("Begin").Return(persistTx, nil).Times(3)
					transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(3)
					persistTx.On("Commit").Return(nil).Once()
					persistTx.On("Commit").Return(transactionError).Once()

					return persistTx, transact
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					return labelService
				},
				FormationRepositoryFn: func() *automock.FormationRepository {
					formationRepo := &automock.FormationRepository{}
					formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
					return formationRepo
				},
				NotificationServiceFN: func() *automock.NotificationsService {
					notificationSvc := &automock.NotificationsService{}
					notificationSvc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(nil, apperrors.NewNotFoundError(resource.Runtime, objectTypeData.ObjectID)).Once()
					return notificationSvc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignments[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignments[1]).Return(testErr).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: fmt.Sprintf("while committing transaction: %s", transactionError.Error()),
			},
			{
				Name: "error if generating notifications fails with not found and fails during commit",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					persistTx := &persistenceautomock.PersistenceTx{}
					transact := &persistenceautomock.Transactioner{}

					transact.On("Begin").Return(persistTx, nil).Times(3)
					transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(3)
					persistTx.On("Commit").Return(nil).Once()
					persistTx.On("Commit").Return(transactionError).Once()
					persistTx.On("Commit").Return(nil).Once()

					return persistTx, transact
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					return labelService
				},
				FormationRepositoryFn: func() *automock.FormationRepository {
					formationRepo := &automock.FormationRepository{}
					formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
					return formationRepo
				},
				NotificationServiceFN: func() *automock.NotificationsService {
					notificationSvc := &automock.NotificationsService{}
					notificationSvc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(nil, apperrors.NewNotFoundError(resource.Runtime, objectTypeData.ObjectID)).Once()
					return notificationSvc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignments[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignments[1]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: fmt.Sprintf("while committing transaction: %s", transactionError.Error()),
			},
			{
				Name: "Error while listing formation assignments when persisting them with DELETING state fails",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					persistTx := &persistenceautomock.PersistenceTx{}

					transact := &persistenceautomock.Transactioner{}
					transact.On("Begin").Return(persistTx, nil).Twice()
					transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Twice()

					return persistTx, transact
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					return labelService
				},
				FormationRepositoryFn: func() *automock.FormationRepository {
					formationRepo := &automock.FormationRepository{}
					formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
					return formationRepo
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(nil, testErr).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "Error while updating formation assignments with DELETING state when persisting them in the DB fails",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					persistTx := &persistenceautomock.PersistenceTx{}

					transact := &persistenceautomock.Transactioner{}
					transact.On("Begin").Return(persistTx, nil).Twice()
					transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Twice()

					return persistTx, transact
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					return labelService
				},
				FormationRepositoryFn: func() *automock.FormationRepository {
					formationRepo := &automock.FormationRepository{}
					formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
					return formationRepo
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(testErr).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "Error when committing DB transaction fails while persisting assignments in the DB",
				TxFn: txGen.ThatFailsOnCommit,
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					return labelService
				},
				FormationRepositoryFn: func() *automock.FormationRepository {
					formationRepo := &automock.FormationRepository{}
					formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
					return formationRepo
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: transactionError.Error(),
			},
			{
				Name: "error if generating notifications fails",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					return labelService
				},
				FormationRepositoryFn: func() *automock.FormationRepository {
					formationRepo := &automock.FormationRepository{}
					formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
					return formationRepo
				},
				NotificationServiceFN: func() *automock.NotificationsService {
					notificationSvc := &automock.NotificationsService{}
					notificationSvc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(nil, testErr).Once()
					return notificationSvc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignments[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignments[1]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error if listing applications fails",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignments[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignments[1]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error if listing runtime contexts fails",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignments[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignments[1]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error if transaction commit fails",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					persistTx := &persistenceautomock.PersistenceTx{}
					transact := &persistenceautomock.Transactioner{}

					transact.On("Begin").Return(persistTx, nil).Times(3)
					transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(3)
					persistTx.On("Commit").Return(nil).Once()
					persistTx.On("Commit").Return(transactionError).Once()
					persistTx.On("Commit").Return(nil).Once()

					return persistTx, transact
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignments[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignments[1]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedFormation:  expected,
				ExpectedErrMessage: transactionError.Error(),
			},
			{
				Name: "error if processing formation assignments fails",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimes(2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(testErr).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error while beginning scenario transaction",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndThenFailsOnBegin(2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: transactionError.Error(),
			},
			{
				Name: "error if list pending formation assignments fail",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return(nil, testErr).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error if assign due to async pending assignments fail - get application type label fail",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(nil, testErr).Once()
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return([]*model.FormationAssignment{}, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error if assign due to async pending assignments fail - get scenario label fail",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(nil, testErr).Once()
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return([]*model.FormationAssignment{}, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error if assign due to async pending assignments fail - update scenario label fail",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(objectTypeData.ScenarioLabel, nil).Once()
					labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabel.ID, &model.LabelInput{
						Key:        model.ScenariosKey,
						Value:      []string{secondTestFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return([]*model.FormationAssignment{}, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: testErr.Error(),
			},
			{
				Name: "error if transaction commit fails after processing formation assignments fails",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					persistTx := &persistenceautomock.PersistenceTx{}
					transact := &persistenceautomock.Transactioner{}

					transact.On("Begin").Return(persistTx, nil).Times(3)
					transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Times(3)
					persistTx.On("Commit").Return(nil).Once()
					persistTx.On("Commit").Return(transactionError).Once()
					persistTx.On("Commit").Return(nil).Once()

					return persistTx, transact
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
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
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(testErr).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignments[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignments[1]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedErrMessage: fmt.Sprintf("while committing transaction: %s", transactionError.Error()),
			},
			{
				Name: "error if transaction fails on begin in the first transaction",
				TxFn: txGen.ThatFailsOnBegin,
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
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
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					return formationAssignmentSvc
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedFormation:  expected,
				ExpectedErrMessage: transactionError.Error(),
			},
			{
				Name: "error if transaction fails on begin in the second transaction",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					persistTx := &persistenceautomock.PersistenceTx{}
					transact := &persistenceautomock.Transactioner{}

					transact.On("Begin").Return(persistTx, nil).Once()
					transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
					persistTx.On("Commit").Return(nil).Once()

					transact.On("Begin").Return(persistTx, transactionError).Once()

					transact.On("Begin").Return(persistTx, nil).Once()
					transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()
					persistTx.On("Commit").Return(nil).Once()

					return persistTx, transact
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil).Once()
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
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignments[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignments[1]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
					return formationAssignmentSvc
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedFormation:  expected,
				ExpectedErrMessage: transactionError.Error(),
			},
			{
				Name: "error while committing scenario label transaction",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndThenFailsOnCommit(2)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(objectTypeData.ScenarioLabel, nil).Once()
					labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabel.ID, &model.LabelInput{
						Key:        model.ScenariosKey,
						Value:      []string{secondTestFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
						Version:    0,
					}).Return(nil).Once()
					return labelService
				},
				ApplicationRepoFn:    expectEmptySliceApplicationRepo,
				RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
				NotificationServiceFN: func() *automock.NotificationsService {
					svc := &automock.NotificationsService{}
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return(nil, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
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
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedFormation:  expected,
				ExpectedErrMessage: transactionError.Error(),
			},
			{
				Name: "error while enforcing post constraints",
				TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
					return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(3, 3)
				},
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
					labelService.On("GetLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabelInput).Return(objectTypeData.ScenarioLabel, nil).Once()
					labelService.On("UpdateLabel", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ScenarioLabel.ID, &model.LabelInput{
						Key:        model.ScenariosKey,
						Value:      []string{secondTestFormationName},
						ObjectID:   objectTypeData.ObjectID,
						ObjectType: objectTypeData.LabelType,
						Version:    0,
					}).Return(nil).Once()
					return labelService
				},
				ApplicationRepoFn:    expectEmptySliceApplicationRepo,
				RuntimeContextRepoFn: expectEmptySliceRuntimeContextRepo,
				NotificationServiceFN: func() *automock.NotificationsService {
					svc := &automock.NotificationsService{}
					svc.On("GenerateFormationAssignmentNotifications", txtest.CtxWithDBMatcher(), TntInternalID, objectTypeData.ObjectID, expected, model.UnassignFormation, objectTypeData.ObjectType).Return(requests, nil).Once()
					return svc
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
					formationAssignmentSvc := &automock.FormationAssignmentService{}
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", ctx, expected.ID, objectTypeData.ObjectID).Return(formationAssignments, nil).Once()
					formationAssignmentSvc.On("ProcessFormationAssignments", txtest.CtxWithDBMatcher(), formationAssignments, make(map[string]string, 0), make(map[string]string, 0), requests, mock.Anything, model.UnassignFormation).Return(nil).Once()
					formationAssignmentSvc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), expected.ID, objectTypeData.ObjectID).Return(nil, nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInDeletingState[0]).Return(nil).Once()
					formationAssignmentSvc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInDeletingState[1]).Return(nil).Once()
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
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(nil).Once()
					engine.On("EnforceConstraints", ctx, postUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(testErr).Once()
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedFormation:  expected,
				ExpectedErrMessage: "while enforcing constraints for target operation",
			},
			{
				Name: "error while enforcing pre constraints",
				TxFn: txGen.ThatDoesntExpectCommit,
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(objectTypeData.TypeLabel, nil)
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
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				ConstraintEngineFn: func() *automock.ConstraintEngine {
					engine := &automock.ConstraintEngine{}
					engine.On("EnforceConstraints", ctx, preUnassignLocation, objectTypeData.UnassignDetails, FormationTemplateID).Return(testErr).Once()
					return engine
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedFormation:  expected,
				ExpectedErrMessage: "while enforcing constraints for target operation",
			},
			{
				Name: "error while preparing details",
				TxFn: txGen.ThatDoesntExpectCommit,
				LabelServiceFn: func() *automock.LabelService {
					labelService := &automock.LabelService{}
					labelService.On("GetLabel", ctx, TntInternalID, objectTypeData.TypeLabelInput).Return(nil, testErr)
					return labelService
				},
				FormationRepositoryFn: func() *automock.FormationRepository {
					formationRepo := &automock.FormationRepository{}
					formationRepo.On("GetByName", ctx, testFormationName, TntInternalID).Return(expected, nil).Once()
					return formationRepo
				},
				ASAEngineFn: func() *automock.AsaEngine {
					engine := &automock.AsaEngine{}
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(false, nil)
					return engine
				},
				FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
					repo := &automock.FormationTemplateRepository{}
					repo.On("Get", ctx, FormationTemplateID).Return(expectedFormationTemplate, nil).Once()
					return repo
				},
				ObjectType:         objectTypeData.ObjectType,
				ObjectID:           objectTypeData.ObjectID,
				InputFormation:     in,
				ExpectedFormation:  expected,
				ExpectedErrMessage: "while preparing joinpoint details for target operation",
			},
			{
				Name: "success when formation is coming from ASA",
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
					engine.On("IsFormationComingFromASA", ctx, objectTypeData.ObjectID, testFormationName, objectTypeData.ObjectType).Return(true, nil)
					return engine
				},
				ObjectType:        objectTypeData.ObjectType,
				ObjectID:          objectTypeData.ObjectID,
				InputFormation:    in,
				ExpectedFormation: expected,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				// GIVEN
				persist, transact := txGen.ThatDoesntStartTransaction()
				if testCase.TxFn != nil {
					persist, transact = testCase.TxFn()
				}
				applicationRepository := unusedApplicationRepository()
				if testCase.ApplicationRepoFn != nil {
					applicationRepository = testCase.ApplicationRepoFn()
				}
				labelService := unusedLabelService()
				if testCase.LabelServiceFn != nil {
					labelService = testCase.LabelServiceFn()
				}
				runtimeContextRepo := unusedRuntimeContextRepo()
				if testCase.RuntimeContextRepoFn != nil {
					runtimeContextRepo = testCase.RuntimeContextRepoFn()
				}
				if objectTypeData.ObjectType == graphql.FormationObjectTypeRuntimeContext {
					runtimeContextRepo.On("GetByID", ctx, TntInternalID, RuntimeContextID).Return(fixRuntimeContextModel(), nil).Maybe()
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
				asaEngine := unusedASAEngine()
				if testCase.ASAEngineFn != nil {
					asaEngine = testCase.ASAEngineFn()
				}

				svc := formation.NewServiceWithAsaEngine(transact, applicationRepository, nil, labelRepo, formationRepo, formationTemplateRepo, labelService, nil, nil, nil, nil, nil, nil, runtimeContextRepo, formationAssignmentSvc, nil, nil, notificationsSvc, constraintEngine, runtimeType, applicationType, asaEngine, nil)

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
				mock.AssertExpectationsForObjects(t, persist, labelService, applicationRepository, runtimeContextRepo, formationRepo, formationTemplateRepo, labelRepo, notificationsSvc, formationAssignmentSvc, constraintEngine, asaEngine)
			})
		}
	}
}

func TestServiceUnassignFormation_Tenant(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	transactionError := errors.New("transaction error")
	txGen := txtest.NewTransactionContextGenerator(transactionError)

	in := model.Formation{
		Name: testFormationName,
	}

	expected := &model.Formation{
		ID:                  fixUUID(),
		Name:                testFormationName,
		FormationTemplateID: FormationTemplateID,
		TenantID:            TntInternalID,
		State:               model.ReadyFormationState,
	}
	asa := model.AutomaticScenarioAssignment{
		ScenarioName:   testFormationName,
		Tenant:         TntInternalID,
		TargetTenantID: TargetTenant,
	}

	testCases := []struct {
		Name                          string
		TxFn                          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		AsaServiceFN                  func() *automock.AutomaticFormationAssignmentService
		AsaRepoFN                     func() *automock.AutomaticFormationAssignmentRepository
		FormationRepositoryFn         func() *automock.FormationRepository
		FormationTemplateRepositoryFn func() *automock.FormationTemplateRepository
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
			ObjectType:        graphql.FormationObjectTypeTenant,
			ObjectID:          TargetTenant,
			InputFormation:    in,
			ExpectedFormation: expected,
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
			Name:               "error when object type is unknown",
			ObjectType:         "UNKNOWN",
			InputFormation:     in,
			ExpectedErrMessage: "unknown formation type",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := txGen.ThatDoesntStartTransaction()
			if testCase.TxFn != nil {
				persist, transact = testCase.TxFn()
			}
			asaRepo := unusedASARepo()
			if testCase.AsaRepoFN != nil {
				asaRepo = testCase.AsaRepoFN()
			}
			asaService := unusedASAService()
			if testCase.AsaServiceFN != nil {
				asaService = testCase.AsaServiceFN()
			}
			formationRepo := unusedFormationRepo()
			if testCase.FormationRepositoryFn != nil {
				formationRepo = testCase.FormationRepositoryFn()
			}
			formationTemplateRepo := unusedFormationTemplateRepo()
			if testCase.FormationTemplateRepositoryFn != nil {
				formationTemplateRepo = testCase.FormationTemplateRepositoryFn()
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

			svc := formation.NewServiceWithAsaEngine(transact, nil, nil, nil, formationRepo, formationTemplateRepo, nil, nil, nil, asaRepo, asaService, tenantSvc, nil, nil, nil, nil, nil, nil, constraintEngine, runtimeType, applicationType, asaEngine, nil)

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
			mock.AssertExpectationsForObjects(t, persist, asaRepo, asaService, formationRepo, formationTemplateRepo, constraintEngine, tenantSvc, asaEngine)
		})
	}
}
