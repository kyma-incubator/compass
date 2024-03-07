package formation_test

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"

	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceFinalizeDraftFormation(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, TntInternalID, TntExternalID)

	transactionError := errors.New("transaction error")
	txGen := txtest.NewTransactionContextGenerator(transactionError)

	allStates := model.ResynchronizableFormationAssignmentStates

	fa1 := fixFormationAssignmentModelWithParameters("id1", FormationID, RuntimeID, ApplicationID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication, model.InitialFormationState)
	fa2 := fixFormationAssignmentModelWithParameters("id2", FormationID, RuntimeContextID, ApplicationID, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeApplication, model.CreateErrorFormationState)
	fa3 := fixFormationAssignmentModelWithParameters("id3", FormationID, RuntimeID, RuntimeContextID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeRuntimeContext, model.DeletingFormationState)
	fa4 := fixFormationAssignmentModelWithParameters("id4", FormationID, RuntimeContextID, RuntimeContextID, model.FormationAssignmentTypeRuntimeContext, model.FormationAssignmentTypeRuntimeContext, model.DeleteErrorFormationState)
	formationAssignments := []*model.FormationAssignment{fa1, fa2, fa3, fa4}

	formationAssignmentsInDeletingState := cloneFormationAssignments(formationAssignments)
	setAssignmentsToState(model.DeletingAssignmentState, formationAssignmentsInDeletingState...)

	formationAssignmentsInInitialState := cloneFormationAssignments(formationAssignments)
	setAssignmentsToState(model.InitialAssignmentState, formationAssignmentsInInitialState...)

	webhookModeAsyncCallback := graphql.WebhookModeAsyncCallback
	notificationsForAssignments := []*webhookclient.FormationAssignmentNotificationRequest{
		{
			Webhook: &graphql.Webhook{
				ID: WebhookID,
			},
		},
		{
			Webhook: &graphql.Webhook{
				ID: Webhook2ID,
			},
		},
		{
			Webhook: &graphql.Webhook{
				ID:   Webhook3ID,
				Mode: &webhookModeAsyncCallback,
			},
		},
		{
			Webhook: &graphql.Webhook{
				ID: Webhook4ID,
			},
		},
	}

	var formationAssignmentPairs = make([]*formationassignment.AssignmentMappingPairWithOperation, 0, len(formationAssignments))
	for i := range formationAssignments {
		formationAssignmentPairs = append(formationAssignmentPairs, fixFormationAssignmentPairWithNoReverseAssignment(notificationsForAssignments[i], formationAssignments[i]))
	}

	testCases := []struct {
		Name                                     string
		FormationAssignments                     []*model.FormationAssignment
		ShouldReset                              bool
		TxFn                                     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		LabelServiceFn                           func() *automock.LabelService
		LabelRepoFn                              func() *automock.LabelRepository
		FormationRepositoryFn                    func() *automock.FormationRepository
		FormationTemplateRepositoryFn            func() *automock.FormationTemplateRepository
		FormationAssignmentNotificationServiceFN func() *automock.FormationAssignmentNotificationsService
		NotificationServiceFN                    func() *automock.NotificationsService
		FormationAssignmentServiceFn             func() *automock.FormationAssignmentService
		WebhookRepoFn                            func() *automock.WebhookRepository
		RuntimeContextRepoFn                     func() *automock.RuntimeContextRepository
		LabelDefRepositoryFn                     func() *automock.LabelDefRepository
		LabelDefServiceFn                        func() *automock.LabelDefService
		StatusServiceFn                          func() *automock.StatusService
		ExpectedErrMessage                       string
	}{
		{
			Name:                 "success when resynchronization is successful and there are leftover formation assignments",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(4)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()

				svc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[3].Source).Return([]*model.FormationAssignment{{ID: "id6"}}, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()
				return svc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(fixFormationModelWithState(model.DraftFormationState), nil).Once()
				repo.On("Update", txtest.CtxWithDBMatcher(), fixFormationModelWithState(model.ReadyFormationState)).Return(nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectIDGlobal", ctx, formationTemplate.ID, model.FormationTemplateWebhookReference).Return(nil, nil).Once()
				return repo
			},
		},
		{
			Name:                 "success when both formation and formation assignment resynchronization are successful and there no left formation assignments should unassign",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimes(5)
			},
			FormationAssignmentServiceFn: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("GetAssignmentsForFormationWithStates", txtest.CtxWithDBMatcher(), TntInternalID, FormationID, allStates).Return(formationAssignments, nil).Once()

				for _, fa := range formationAssignments {
					svc.On("GetReverseBySourceAndTarget", txtest.CtxWithDBMatcher(), FormationID, fa.Source, fa.Target).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, "")).Once()
				}

				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[0]).Return(false, nil).Once()
				svc.On("ProcessFormationAssignmentPair", txtest.CtxWithDBMatcher(), formationAssignmentPairs[1]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[2]).Return(false, nil).Once()
				svc.On("CleanupFormationAssignment", txtest.CtxWithDBMatcher(), formationAssignmentPairs[3]).Return(false, nil).Once()

				svc.On("ListFormationAssignmentsForObjectID", txtest.CtxWithDBMatcher(), FormationID, formationAssignments[3].Source).Return([]*model.FormationAssignment{{ID: "id6"}}, nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[0].ID, formationAssignmentsInInitialState[0]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[1].ID, formationAssignmentsInInitialState[1]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[2].ID, formationAssignmentsInDeletingState[2]).Return(nil).Once()
				svc.On("Update", txtest.CtxWithDBMatcher(), formationAssignments[3].ID, formationAssignmentsInDeletingState[3]).Return(nil).Once()
				return svc
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationTemplateID).Return(&formationTemplate, nil).Once() //todo~~ twice
				return repo
			},
			NotificationServiceFN: func() *automock.NotificationsService {
				notificationSvc := &automock.NotificationsService{}
				notificationSvc.On("GenerateFormationNotifications", txtest.CtxWithDBMatcher(), formationLifecycleSyncWebhooks, TntInternalID, fixFormationModelWithState(model.InitialFormationState), testFormationTemplateName, FormationTemplateID, model.CreateFormation).Return(formationNotificationSyncCreateRequests, nil).Once()
				notificationSvc.On("SendNotification", txtest.CtxWithDBMatcher(), formationNotificationSyncCreateRequest).Return(formationNotificationWebhookSuccessResponse, nil).Once()
				return notificationSvc
			},
			FormationAssignmentNotificationServiceFN: func() *automock.FormationAssignmentNotificationsService {
				svc := &automock.FormationAssignmentNotificationsService{}

				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[0], model.AssignFormation).Return(notificationsForAssignments[0], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[1], model.AssignFormation).Return(notificationsForAssignments[1], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[2], model.UnassignFormation).Return(notificationsForAssignments[2], nil).Once()
				svc.On("GenerateFormationAssignmentNotification", txtest.CtxWithDBMatcher(), formationAssignments[3], model.UnassignFormation).Return(notificationsForAssignments[3], nil).Once()
				return svc
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				webhookRepo := &automock.WebhookRepository{}
				webhookRepo.On("ListByReferenceObjectIDGlobal", ctx, FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				webhookRepo.On("ListByReferenceObjectIDGlobal", txtest.CtxWithDBMatcher(), FormationTemplateID, model.FormationTemplateWebhookReference).Return(formationLifecycleSyncWebhooks, nil).Once()
				return webhookRepo
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(fixFormationModelWithState(model.DraftFormationState), nil).Once()
				repo.On("Get", txtest.CtxWithDBMatcher(), FormationID, TntInternalID).Return(fixFormationModelWithState(model.ReadyFormationState), nil).Once()
				repo.On("Update", txtest.CtxWithDBMatcher(), fixFormationModelWithState(model.InitialFormationState)).Return(nil).Once()
				return repo
			},
			StatusServiceFn: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), fixFormationModelWithState(model.ReadyFormationState), model.CreateFormation).Return(nil).Once()
				return svc
			},
		},
		{
			Name:                 "error when fails to commit",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnCommit()
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(fixFormationModelWithState(model.DraftFormationState), nil).Once()
				repo.On("Update", txtest.CtxWithDBMatcher(), fixFormationModelWithState(model.ReadyFormationState)).Return(nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectIDGlobal", ctx, formationTemplate.ID, model.FormationTemplateWebhookReference).Return(nil, nil).Once()
				return repo
			},
			ExpectedErrMessage: "transaction error",
		},
		{
			Name:                 "error when updating formation state",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(fixFormationModelWithState(model.DraftFormationState), nil).Once()
				repo.On("Update", txtest.CtxWithDBMatcher(), fixFormationModelWithState(model.ReadyFormationState)).Return(testErr).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectIDGlobal", ctx, formationTemplate.ID, model.FormationTemplateWebhookReference).Return(nil, nil).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name:                 "error when opening transaction",
			FormationAssignments: formationAssignments,
			TxFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatFailsOnBegin()
			},
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(fixFormationModelWithState(model.DraftFormationState), nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectIDGlobal", ctx, formationTemplate.ID, model.FormationTemplateWebhookReference).Return(nil, nil).Once()
				return repo
			},
			ExpectedErrMessage: "transaction error",
		},
		{
			Name:                 "error when listing webhooks",
			FormationAssignments: formationAssignments,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(fixFormationModelWithState(model.DraftFormationState), nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(&formationTemplate, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByReferenceObjectIDGlobal", ctx, formationTemplate.ID, model.FormationTemplateWebhookReference).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrMessage: "when listing formation lifecycle webhooks for formation template with ID",
		},
		{
			Name:                 "error when getting formation template",
			FormationAssignments: formationAssignments,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(fixFormationModelWithState(model.DraftFormationState), nil).Once()
				return repo
			},
			FormationTemplateRepositoryFn: func() *automock.FormationTemplateRepository {
				repo := &automock.FormationTemplateRepository{}
				repo.On("Get", ctx, FormationTemplateID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrMessage: "An error occurred while getting formation template with ID:",
		},
		{
			Name:                 "error when formation is not in DRAFT state",
			FormationAssignments: formationAssignments,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(fixFormationModelWithState(model.ReadyFormationState), nil).Once()
				return repo
			},
			ExpectedErrMessage: "is not in DRAFT state",
		},
		{
			Name:                 "error when getting formation",
			FormationAssignments: formationAssignments,
			FormationRepositoryFn: func() *automock.FormationRepository {
				repo := &automock.FormationRepository{}
				repo.On("Get", ctx, FormationID, TntInternalID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrMessage: "while getting formation with ID",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := txGen.ThatDoesntStartTransaction()
			if testCase.TxFn != nil {
				persist, transact = testCase.TxFn()
			}

			labelService := unusedLabelService()
			if testCase.LabelServiceFn != nil {
				labelService = testCase.LabelServiceFn()
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

			formationAssignmentNotificationService := unusedFormationAssignmentNotificationService()
			if testCase.FormationAssignmentNotificationServiceFN != nil {
				formationAssignmentNotificationService = testCase.FormationAssignmentNotificationServiceFN()
			}

			webhookRepo := unusedWebhookRepository()
			if testCase.WebhookRepoFn != nil {
				webhookRepo = testCase.WebhookRepoFn()
			}

			labelDefRepo := unusedLabelDefRepository()
			if testCase.LabelDefRepositoryFn != nil {
				labelDefRepo = testCase.LabelDefRepositoryFn()
			}

			labelDefSvc := unusedLabelDefService()
			if testCase.LabelDefServiceFn != nil {
				labelDefSvc = testCase.LabelDefServiceFn()
			}

			statusService := &automock.StatusService{}
			if testCase.StatusServiceFn != nil {
				statusService = testCase.StatusServiceFn()
			}

			assignmentsBeforeModifications := make(map[string]*model.FormationAssignment)
			for _, a := range testCase.FormationAssignments {
				assignmentsBeforeModifications[a.ID] = a.Clone()
			}
			defer func() {
				for i, a := range testCase.FormationAssignments {
					testCase.FormationAssignments[i] = assignmentsBeforeModifications[a.ID]
				}
			}()

			svc := formation.NewServiceWithAsaEngine(transact, nil, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelService, nil, labelDefSvc, nil, nil, nil, nil, runtimeContextRepo, formationAssignmentSvc, webhookRepo, formationAssignmentNotificationService, notificationsSvc, nil, runtimeType, applicationType, nil, statusService)

			// WHEN
			_, err := svc.FinalizeDraftFormation(ctx, FormationID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			//todo assert persist and transact
			mock.AssertExpectationsForObjects(t, labelService, runtimeContextRepo, formationRepo, labelRepo, notificationsSvc, formationAssignmentSvc, formationAssignmentNotificationService, formationTemplateRepo, webhookRepo, statusService, persist, transact)
		})
	}

	t.Run("returns error when empty tenant", func(t *testing.T) {
		svc := formation.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
		_, err := svc.ResynchronizeFormationNotifications(context.TODO(), FormationID, false)
		require.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
