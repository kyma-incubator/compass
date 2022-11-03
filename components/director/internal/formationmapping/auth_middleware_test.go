package formationmapping_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	fm "github.com/kyma-incubator/compass/components/director/internal/formationmapping"
	"github.com/kyma-incubator/compass/components/director/internal/formationmapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthenticator_Handler(t *testing.T) {
	emptyCtx := context.Background()
	internalTntID := "testInternalID"
	externalTntID := "testExternalID"
	parentTntID := "parentTenantID"
	faSourceID := "testSourceID"
	faTargetID := "testTargetID"
	testFormationID := "testFormationID"
	testFormationAssignmentID := "testFormationAssignmentID"
	consumerUUID := uuid.New().String()
	appTemplateID := "testAppTemplateID"
	runtimeID := "testRuntimeID"

	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	urlVars := map[string]string{
		fm.FormationIDParam:           testFormationID,
		fm.FormationAssignmentIDParam: testFormationAssignmentID,
	}

	consumerTenant := &model.BusinessTenantMapping{
		ID:             internalTntID,
		ExternalTenant: externalTntID,
		Parent:         parentTntID,
	}

	consumerTenantWithEmptyParent := &model.BusinessTenantMapping{
		ID:             internalTntID,
		ExternalTenant: externalTntID,
	}

	faWithSourceAppAndTargetRuntime := fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime)
	faWithSourceRuntimeAndTargetApp := fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication)
	faWithSourceAppAndTargetRuntimeContext := fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntimeContext)

	intSysApp := &model.Application{
		ApplicationTemplateID: &appTemplateID,
		IntegrationSystemID:   &faTargetID,
	}

	selfRegDistinguishLabelKey := "selfRegDistinguishLabelKey"
	consumerSubaccountLabelKey := "consumerSubaccountLabelKey"

	appTemplateLbls := map[string]*model.Label{
		selfRegDistinguishLabelKey: {Key: selfRegDistinguishLabelKey, Value: "selfRegDistinguishLabelValue"},
		consumerSubaccountLabelKey: {Key: consumerSubaccountLabelKey, Value: externalTntID},
	}

	appTemplateLblsWithInvalidConsumerSubaccount := map[string]*model.Label{
		selfRegDistinguishLabelKey: {Key: selfRegDistinguishLabelKey, Value: "selfRegDistinguishLabelValue"},
		consumerSubaccountLabelKey: {Key: consumerSubaccountLabelKey, Value: "invalidConsumerSubaccountID"},
	}

	appTemplateLblsWithIncorrectType := map[string]*model.Label{
		selfRegDistinguishLabelKey: {Key: selfRegDistinguishLabelKey, Value: "selfRegDistinguishLabelValue"},
		consumerSubaccountLabelKey: {Key: consumerSubaccountLabelKey, Value: model.FormationAssignmentTypeRuntime},
	}

	rtmContext := &model.RuntimeContext{
		RuntimeID: runtimeID,
	}

	testCases := []struct {
		name                       string
		transactFn                 func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		faServiceFn                func() *automock.FormationAssignmentService
		runtimeRepoFn              func() *automock.RuntimeRepository
		runtimeContextRepoFn       func() *automock.RuntimeContextRepository
		appRepoFn                  func() *automock.ApplicationRepository
		tenantRepoFn               func() *automock.TenantRepository
		appTemplateRepoFn          func() *automock.ApplicationTemplateRepository
		labelRepoFn                func() *automock.LabelRepository
		selfRegDistinguishLabelKey string
		consumerSubaccountLabelKey string
		hasURLVars                 bool
		contextFn                  func() context.Context
		httpMethod                 string
		expectedStatusCode         int
		expectedErrOutput          string
	}{
		// Common authorization checks
		{
			name:                 "Error when the http request method is not PATCH",
			transactFn:           fixUnusedTransactioner,
			faServiceFn:          fixUnusedFormationAssignmentSvc,
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			contextFn: func() context.Context {
				return emptyCtx
			},
			hasURLVars:         true,
			httpMethod:         http.MethodGet,
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedErrOutput:  "",
		},
		{
			name:                 "Error when required parameters are missing",
			transactFn:           fixUnusedTransactioner,
			faServiceFn:          fixUnusedFormationAssignmentSvc,
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			contextFn: func() context.Context {
				return emptyCtx
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  fixBuildExpectedErrResponse(t, "Not all of the required parameters are provided"),
		},
		{
			name:       "Unauthorized error when authorization check is unsuccessful",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, "invalid", "invalid"), nil).Once()
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "",
		},
		{
			name:                 "Authorization fail: error when consumer info is missing in the context",
			transactFn:           fixUnusedTransactioner,
			faServiceFn:          fixUnusedFormationAssignmentSvc,
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			contextFn: func() context.Context {
				return emptyCtx
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:                 "Authorization fail: error when transaction begin fails",
			transactFn:           txGen.ThatFailsOnBegin,
			faServiceFn:          fixUnusedFormationAssignmentSvc,
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when getting formation assignment globally fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil, testErr)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when tenant loading from context fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", mock.Anything, testFormationAssignmentID).Return(faWithSourceAppAndTargetRuntime, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				ctxOnlyWithConsumer := consumer.SaveToContext(emptyCtx, c)
				return ctxOnlyWithConsumer
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when committing transaction",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, "invalid", "invalid"), nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		// Application/ApplicationTemplate authorization checks
		{
			name:       "Authorization fail: error when getting application globally",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(nil, testErr)
				return appRepo
			},
			tenantRepoFn:      fixUnusedTenantRepo,
			appTemplateRepoFn: fixUnusedAppTemplateRepo,
			labelRepoFn:       fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Application)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when getting application parent tenant",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(&model.Application{}, nil)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(nil, testErr).Once()
				return tenantRepo
			},
			appTemplateRepoFn: fixUnusedAppTemplateRepo,
			labelRepoFn:       fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when application owner existence check fail",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(&model.Application{}, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), parentTntID, faTargetID).Return(false, testErr)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(consumerTenant, nil).Once()
				return tenantRepo
			},
			appTemplateRepoFn: fixUnusedAppTemplateRepo,
			labelRepoFn:       fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when application template is nil or empty",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(&model.Application{}, nil)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(consumerTenantWithEmptyParent, nil).Once()
				return tenantRepo
			},
			appTemplateRepoFn: fixUnusedAppTemplateRepo,
			labelRepoFn:       fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization fail: error when application template existence check fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(intSysApp, nil)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(consumerTenantWithEmptyParent, nil).Once()
				return tenantRepo
			},
			appTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", contextThatHasTenant(internalTntID), appTemplateID).Return(false, testErr)
				return appTemplateRepo
			},
			labelRepoFn: fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when application template does not exists",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(intSysApp, nil)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(consumerTenantWithEmptyParent, nil).Once()
				return tenantRepo
			},
			appTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", contextThatHasTenant(internalTntID), appTemplateID).Return(false, nil)
				return appTemplateRepo
			},
			labelRepoFn: fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization fail: error when listing application template labels",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(intSysApp, nil)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(consumerTenantWithEmptyParent, nil).Once()
				return tenantRepo
			},
			appTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", contextThatHasTenant(internalTntID), appTemplateID).Return(true, nil)
				return appTemplateRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForGlobalObject", contextThatHasTenant(internalTntID), model.AppTemplateLabelableObject, appTemplateID).Return(nil, testErr)
				return lblRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when either self register or consumer subaccount label is missing",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(intSysApp, nil)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(consumerTenantWithEmptyParent, nil).Once()
				return tenantRepo
			},
			appTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", contextThatHasTenant(internalTntID), appTemplateID).Return(true, nil)
				return appTemplateRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForGlobalObject", contextThatHasTenant(internalTntID), model.AppTemplateLabelableObject, appTemplateID).Return(appTemplateLbls, nil)
				return lblRepo
			},
			selfRegDistinguishLabelKey: "nonExistingLabel",
			consumerSubaccountLabelKey: "nonExistingLabel",
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when consumer subaccount label is not of type string",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(intSysApp, nil)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(consumerTenantWithEmptyParent, nil).Once()
				return tenantRepo
			},
			appTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", contextThatHasTenant(internalTntID), appTemplateID).Return(true, nil)
				return appTemplateRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForGlobalObject", contextThatHasTenant(internalTntID), model.AppTemplateLabelableObject, appTemplateID).Return(appTemplateLblsWithIncorrectType, nil)
				return lblRepo
			},
			selfRegDistinguishLabelKey: selfRegDistinguishLabelKey,
			consumerSubaccountLabelKey: consumerSubaccountLabelKey,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: when caller has NOT owner access to the FA with target type app that is made through subscription",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(intSysApp, nil)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(consumerTenantWithEmptyParent, nil).Once()
				return tenantRepo
			},
			appTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", contextThatHasTenant(internalTntID), appTemplateID).Return(true, nil)
				return appTemplateRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForGlobalObject", contextThatHasTenant(internalTntID), model.AppTemplateLabelableObject, appTemplateID).Return(appTemplateLblsWithInvalidConsumerSubaccount, nil)
				return lblRepo
			},
			selfRegDistinguishLabelKey: selfRegDistinguishLabelKey,
			consumerSubaccountLabelKey: consumerSubaccountLabelKey,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization success: when the int system caller has owner access to the formation assignment with target type application",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(intSysApp, nil)
				return appRepo
			},
			tenantRepoFn:      fixUnusedTenantRepo,
			appTemplateRepoFn: fixUnusedAppTemplateRepo,
			labelRepoFn:       fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(faTargetID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization success: when the caller has owner access to the formation assignment with target type application",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(&model.Application{}, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), parentTntID, faTargetID).Return(true, nil)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(consumerTenant, nil).Once()
				return tenantRepo
			},
			appTemplateRepoFn: fixUnusedAppTemplateRepo,
			labelRepoFn:       fixUnusedLabelRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization success: when the caller has owner access to the FA with target type app that is made through subscription",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			runtimeRepoFn:        fixUnusedRuntimeRepo,
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(intSysApp, nil)
				return appRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(consumerTenantWithEmptyParent, nil).Once()
				return tenantRepo
			},
			appTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", contextThatHasTenant(internalTntID), appTemplateID).Return(true, nil)
				return appTemplateRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForGlobalObject", contextThatHasTenant(internalTntID), model.AppTemplateLabelableObject, appTemplateID).Return(appTemplateLbls, nil)
				return lblRepo
			},
			selfRegDistinguishLabelKey: selfRegDistinguishLabelKey,
			consumerSubaccountLabelKey: consumerSubaccountLabelKey,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.IntegrationSystem)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		// Runtime authorization checks
		{
			name:       "Authorization fail: error when runtime owner existence check fail",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceAppAndTargetRuntime, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, testErr)
				return rtmRepo
			},
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when the caller has NOT owner access to the formation assignment with target type runtime",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceAppAndTargetRuntime, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return rtmRepo
			},
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization success: when the caller has owner access to the formation assignment with target type runtime",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceAppAndTargetRuntime, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(true, nil)
				return rtmRepo
			},
			runtimeContextRepoFn: fixUnusedRuntimeContextRepo,
			appRepoFn:            fixUnusedAppRepo,
			appTemplateRepoFn:    fixUnusedAppTemplateRepo,
			labelRepoFn:          fixUnusedLabelRepo,
			tenantRepoFn:         fixUnusedTenantRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "OK",
		},
		// Runtime context authorization checks
		{
			name:       "Authorization fail: error when getting runtime context globally",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceAppAndTargetRuntimeContext, nil)
				return faSvc
			},
			runtimeRepoFn: fixUnusedRuntimeRepo,
			runtimeContextRepoFn: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(nil, testErr)
				return rtmCtxRepo
			},
			appRepoFn:         fixUnusedAppRepo,
			appTemplateRepoFn: fixUnusedAppTemplateRepo,
			labelRepoFn:       fixUnusedLabelRepo,
			tenantRepoFn:      fixUnusedTenantRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when runtime context owner check for runtime fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceAppAndTargetRuntimeContext, nil).Once()
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, runtimeID).Return(false, testErr).Once()
				return runtimeRepo
			},
			runtimeContextRepoFn: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(rtmContext, nil).Once()
				return rtmCtxRepo
			},
			appRepoFn:         fixUnusedAppRepo,
			appTemplateRepoFn: fixUnusedAppTemplateRepo,
			labelRepoFn:       fixUnusedLabelRepo,
			tenantRepoFn:      fixUnusedTenantRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: when caller has NOT owner access to FA with target type rtm context made through subscription",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceAppAndTargetRuntimeContext, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, runtimeID).Return(false, nil)
				return rtmRepo
			},
			runtimeContextRepoFn: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(rtmContext, nil)
				return rtmCtxRepo
			},
			appRepoFn:         fixUnusedAppRepo,
			appTemplateRepoFn: fixUnusedAppTemplateRepo,
			labelRepoFn:       fixUnusedLabelRepo,
			tenantRepoFn:      fixUnusedTenantRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization success: when caller has owner access to the formation assignment with target type rtm context",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(faWithSourceAppAndTargetRuntimeContext, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, runtimeID).Return(true, nil)
				return rtmRepo
			},
			runtimeContextRepoFn: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), faTargetID).Return(rtmContext, nil)
				return rtmCtxRepo
			},
			appRepoFn:         fixUnusedAppRepo,
			appTemplateRepoFn: fixUnusedAppTemplateRepo,
			labelRepoFn:       fixUnusedLabelRepo,
			tenantRepoFn:      fixUnusedTenantRepo,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.Runtime)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			persist, transact := tCase.transactFn()
			faSvc := tCase.faServiceFn()
			rtmRepo := tCase.runtimeRepoFn()
			rtmCtxRepo := tCase.runtimeContextRepoFn()
			appRepo := tCase.appRepoFn()
			appTemplateRepo := tCase.appTemplateRepoFn()
			labelRepo := tCase.labelRepoFn()
			tenantRepo := tCase.tenantRepoFn()

			defer mock.AssertExpectationsForObjects(t, persist, transact, faSvc, rtmRepo, rtmCtxRepo, appRepo, appTemplateRepo, labelRepo, tenantRepo)

			// GIVEN
			fmAuthenticator := fm.NewFormationMappingAuthenticator(transact, faSvc, rtmRepo, rtmCtxRepo, appRepo, tenantRepo, appTemplateRepo, labelRepo, tCase.selfRegDistinguishLabelKey, tCase.consumerSubaccountLabelKey)
			fmAuthMiddleware := fmAuthenticator.Handler()
			rw := httptest.NewRecorder()

			httpMethod := http.MethodPatch
			if tCase.httpMethod != "" {
				httpMethod = tCase.httpMethod
			}

			httpReq := fixRequestWithContext(t, tCase.contextFn(), httpMethod)

			if tCase.hasURLVars {
				httpReq = mux.SetURLVars(httpReq, urlVars)
			}

			// WHEN
			fmAuthMiddleware(fixTestHandler(t)).ServeHTTP(rw, httpReq)

			// THEN
			require.Equal(t, tCase.expectedStatusCode, rw.Code)
			require.Contains(t, rw.Body.String(), tCase.expectedErrOutput)
		})
	}
}
