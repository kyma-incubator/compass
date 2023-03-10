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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthenticator_FormationAssignmentHandler(t *testing.T) {
	faSourceID := "testSourceID"
	faTargetID := "testTargetID"
	testFormationAssignmentID := "testFormationAssignmentID"
	consumerUUID := uuid.New().String()
	appTemplateID := "testAppTemplateID"
	intSystemID := "intSystemID"
	runtimeID := "testRuntimeID"

	urlVars := map[string]string{
		fm.FormationIDParam:           testFormationID,
		fm.FormationAssignmentIDParam: testFormationAssignmentID,
	}

	faWithSourceAppAndTargetRuntime := fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime)
	faWithSourceRuntimeAndTargetApp := fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication)
	faWithSourceAppAndTargetRuntimeContext := fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntimeContext)

	intSysApp := &model.Application{
		IntegrationSystemID: &intSystemID,
	}

	appWithAppTemplate := &model.Application{
		ApplicationTemplateID: &appTemplateID,
	}

	consumerSubaccountLabelKey := "consumerSubaccountLabelKey"

	appTemplateLbls := map[string]*model.Label{
		consumerSubaccountLabelKey: {Key: consumerSubaccountLabelKey, Value: externalTntID},
	}

	appTemplateLblsWithInvalidConsumerSubaccount := map[string]*model.Label{
		consumerSubaccountLabelKey: {Key: consumerSubaccountLabelKey, Value: "invalidConsumerSubaccountID"},
	}

	appTemplateLblsWithIncorrectType := map[string]*model.Label{
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
		appTemplateRepoFn          func() *automock.ApplicationTemplateRepository
		labelRepoFn                func() *automock.LabelRepository
		tenantRepoFn               func() *automock.TenantRepository
		consumerSubaccountLabelKey string
		hasURLVars                 bool
		contextFn                  func() context.Context
		httpMethod                 string
		expectedStatusCode         int
		expectedErrOutput          string
	}{
		// Common authorization checks
		{
			name:       "Error when the http request method is not PATCH",
			transactFn: fixUnusedTransactioner,
			contextFn: func() context.Context {
				return emptyCtx
			},
			hasURLVars:         true,
			httpMethod:         http.MethodGet,
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedErrOutput:  "",
		},
		{
			name: "Error when required parameters are missing",
			contextFn: func() context.Context {
				return emptyCtx
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  fixBuildExpectedErrResponse(t, "Not all of the required parameters are provided"),
		},
		{
			name:       "Unauthorized error when authorization check is unsuccessful but there is no error",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, "invalid", "invalid"), nil).Once()
				return faSvc
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "",
		},
		{
			name: "Authorization fail: error when consumer info is missing in the context",
			contextFn: func() context.Context {
				return emptyCtx
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when transaction begin fails",
			transactFn: txGen.ThatFailsOnBegin,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(nil, testErr)
				return faSvc
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", mock.Anything, testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil)
				return faSvc
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, "invalid", "invalid"), nil)
				return faSvc
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		// Application/ApplicationTemplate authorization checks
		{
			name:       "Authorization fail: error when getting tenant",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(nil, testErr)
				return tenantRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when getting application",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(nil, testErr)
				return appRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(intSysApp, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, testErr)
				return appRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(intSysApp, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return appRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return appRepo
			},
			appTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", contextThatHasTenant(internalTntID), appTemplateID).Return(false, testErr)
				return appTemplateRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return appRepo
			},
			appTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", contextThatHasTenant(internalTntID), appTemplateID).Return(false, nil)
				return appTemplateRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return appRepo
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
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when listing application template labels doesn't include subaccount label",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return appRepo
			},
			appTemplateRepoFn: func() *automock.ApplicationTemplateRepository {
				appTemplateRepo := &automock.ApplicationTemplateRepository{}
				appTemplateRepo.On("Exists", contextThatHasTenant(internalTntID), appTemplateID).Return(true, nil)
				return appTemplateRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				lblRepo := &automock.LabelRepository{}
				lblRepo.On("ListForGlobalObject", contextThatHasTenant(internalTntID), model.AppTemplateLabelableObject, appTemplateID).Return(map[string]*model.Label{}, nil)
				return lblRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return appRepo
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
			consumerSubaccountLabelKey: consumerSubaccountLabelKey,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: when caller has NOT owner access to the target FA with type app that is made through subscription",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return appRepo
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
			consumerSubaccountLabelKey: consumerSubaccountLabelKey,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization success: when the int system caller has owner access to the target formation assignment with type application",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasConsumer(intSystemID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasConsumer(intSystemID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasConsumer(intSystemID), internalTntID, faTargetID).Return(intSysApp, nil)
				return appRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(intSystemID, consumer.IntegrationSystem)
				return fixContextWithConsumer(c)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization fail: when the int system caller manages the target FA with type application but the transaction fail",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasConsumer(intSystemID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasConsumer(intSystemID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasConsumer(intSystemID), internalTntID, faTargetID).Return(intSysApp, nil)
				return appRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(intSystemID, consumer.IntegrationSystem)
				return fixContextWithConsumer(c)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization success: when caller is business integration and the formation is in a tenant of type resource group",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasConsumer(consumerUUID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasConsumer(consumerUUID), internalTntID).Return(fixResourceGroupBusinessTenantMapping(), nil)
				return tenantRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.BusinessIntegration)
				return fixContextWithConsumer(c)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization fail: when caller is business integration and the formation is in a tenant of type resource group but the transaction fail",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasConsumer(consumerUUID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasConsumer(consumerUUID), internalTntID).Return(fixResourceGroupBusinessTenantMapping(), nil)
				return tenantRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.BusinessIntegration)
				return fixContextWithConsumer(c)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: when the caller is the parent of the formation assignment target with type application but the transaction fail",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasConsumer(appTemplateID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasConsumer(appTemplateID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasConsumer(appTemplateID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				return appRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(appTemplateID, consumer.ExternalCertificate)
				return fixContextWithConsumer(c)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization success: when the caller is the parent of the formation assignment target with type application",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasConsumer(appTemplateID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasConsumer(appTemplateID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasConsumer(appTemplateID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				return appRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(appTemplateID, consumer.ExternalCertificate)
				return fixContextWithConsumer(c)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization fail: error when consumer info is missing in the context for formation assignment with target type application",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasConsumer(consumerUUID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasConsumer(consumerUUID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasConsumer(consumerUUID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				return appRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithConsumer(c)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization success: when the caller has owner access to the target of the FA with type application",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(intSysApp, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(true, nil)
				return appRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization fail: when the caller has owner access to the target of the FA with type application but the transaction fail",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(intSysApp, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(true, nil)
				return appRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization success: when the caller has owner access to the FA target's parent for type app that is made through subscription",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return appRepo
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
			consumerSubaccountLabelKey: consumerSubaccountLabelKey,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Authz fail: when caller has owner access to FA target's parent for type app that is made through subscription but transact fail",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceRuntimeAndTargetApp, nil)
				return faSvc
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("Get", contextThatHasTenant(internalTntID), internalTntID).Return(fixBusinessTenantMapping(), nil)
				return tenantRepo
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(appWithAppTemplate, nil)
				appRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return appRepo
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
			consumerSubaccountLabelKey: consumerSubaccountLabelKey,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		// Runtime authorization checks
		{
			name:       "Authorization fail: error when runtime owner existence check fail",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, testErr)
				return rtmRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(false, nil)
				return rtmRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization success: when the caller has owner access to the target of the formation assignment with type runtime",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(true, nil)
				return rtmRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "OK",
		},
		{
			name:       "Authorization fail: when the caller has owner access to the target of the FA with type runtime but transaction fail",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(true, nil)
				return rtmRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		// Runtime context authorization checks
		{
			name:       "Authorization fail: error when getting runtime context globally",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeContext, nil)
				return faSvc
			},
			runtimeContextRepoFn: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(nil, testErr)
				return rtmCtxRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeContext, nil).Once()
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, runtimeID).Return(false, testErr).Once()
				return runtimeRepo
			},
			runtimeContextRepoFn: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(rtmContext, nil).Once()
				return rtmCtxRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
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
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeContext, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, runtimeID).Return(false, nil)
				return rtmRepo
			},
			runtimeContextRepoFn: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(rtmContext, nil)
				return rtmCtxRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization success: when caller has owner access to the target of the formation assignment with type rtm context",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeContext, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, runtimeID).Return(true, nil)
				return rtmRepo
			},
			runtimeContextRepoFn: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(rtmContext, nil)
				return rtmCtxRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Authorization fail: when caller has owner access to the target of the FA with type rtm context but transaction fail",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", contextThatHasTenant(internalTntID), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeContext, nil)
				return faSvc
			},
			runtimeRepoFn: func() *automock.RuntimeRepository {
				rtmRepo := &automock.RuntimeRepository{}
				rtmRepo.On("OwnerExists", contextThatHasTenant(internalTntID), internalTntID, runtimeID).Return(true, nil)
				return rtmRepo
			},
			runtimeContextRepoFn: func() *automock.RuntimeContextRepository {
				rtmCtxRepo := &automock.RuntimeContextRepository{}
				rtmCtxRepo.On("GetByID", contextThatHasTenant(internalTntID), internalTntID, faTargetID).Return(rtmContext, nil)
				return rtmCtxRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerUUID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			persist, transact := fixUnusedTransactioner()
			if tCase.transactFn != nil {
				persist, transact = tCase.transactFn()
			}

			faSvc := fixUnusedFormationAssignmentSvc()
			if tCase.faServiceFn != nil {
				faSvc = tCase.faServiceFn()
			}

			rtmRepo := fixUnusedRuntimeRepo()
			if tCase.runtimeRepoFn != nil {
				rtmRepo = tCase.runtimeRepoFn()
			}

			rtmCtxRepo := fixUnusedRuntimeContextRepo()
			if tCase.runtimeContextRepoFn != nil {
				rtmCtxRepo = tCase.runtimeContextRepoFn()
			}

			appRepo := fixUnusedAppRepo()
			if tCase.appRepoFn != nil {
				appRepo = tCase.appRepoFn()
			}

			appTemplateRepo := fixUnusedAppTemplateRepo()
			if tCase.appTemplateRepoFn != nil {
				appTemplateRepo = tCase.appTemplateRepoFn()
			}

			labelRepo := fixUnusedLabelRepo()
			if tCase.labelRepoFn != nil {
				labelRepo = tCase.labelRepoFn()
			}

			tenantRepo := fixUnusedTenantRepo()
			if tCase.tenantRepoFn != nil {
				tenantRepo = tCase.tenantRepoFn()
			}

			defer mock.AssertExpectationsForObjects(t, persist, transact, faSvc, rtmRepo, rtmCtxRepo, appRepo, appTemplateRepo, labelRepo)

			// GIVEN
			fmAuthenticator := fm.NewFormationMappingAuthenticator(transact, faSvc, rtmRepo, rtmCtxRepo, appRepo, appTemplateRepo, labelRepo, nil, nil, tenantRepo, tCase.consumerSubaccountLabelKey)
			fmAuthMiddleware := fmAuthenticator.FormationAssignmentHandler()
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

func TestAuthenticator_FormationHandler(t *testing.T) {
	urlVars := map[string]string{
		fm.FormationIDParam: testFormationID,
	}

	consumerID := "2c755564-97ef-4499-8c88-7b8518edc171"
	leadingProductID := consumerID
	leadingProductID2 := "leading-product-id-2"
	leadingProductIDs := []string{leadingProductID, leadingProductID2}

	formation := &model.Formation{
		ID:                  testFormationID,
		TenantID:            internalTntID,
		FormationTemplateID: testFormationTemplateID,
		Name:                testFormationName,
	}

	formationTemplate := &model.FormationTemplate{
		ID:                testFormationTemplateID,
		Name:              "formationTemplateName",
		TenantID:          &internalTntID,
		LeadingProductIDs: leadingProductIDs,
	}

	formationTemplateWithNonMatchingProductIDs := &model.FormationTemplate{
		ID:                testFormationTemplateID,
		Name:              "formationTemplateName",
		TenantID:          &internalTntID,
		LeadingProductIDs: []string{leadingProductID2},
	}

	testCases := []struct {
		name                       string
		transactFn                 func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		formationRepoFn            func() *automock.FormationRepository
		formationTemplateRepoFn    func() *automock.FormationTemplateRepository
		consumerSubaccountLabelKey string
		hasURLVars                 bool
		contextFn                  func() context.Context
		httpMethod                 string
		expectedStatusCode         int
		expectedErrOutput          string
	}{
		// Common authorization checks
		{
			name:       "Error when the http request method is not PATCH",
			transactFn: fixUnusedTransactioner,
			contextFn: func() context.Context {
				return emptyCtx
			},
			hasURLVars:         true,
			httpMethod:         http.MethodGet,
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedErrOutput:  "",
		},
		{
			name: "Error when the required parameter is missing",
			contextFn: func() context.Context {
				return emptyCtx
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  fixBuildExpectedErrResponse(t, "Not all of the required parameters are provided"),
		},
		{
			name:       "Authorization success: when the consumer ID is one of the formation templates product IDs",
			transactFn: txGen.ThatSucceeds,
			formationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationID).Return(formation, nil).Once()
				return formationRepo
			},
			formationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				ftRepo := &automock.FormationTemplateRepository{}
				ftRepo.On("Get", contextThatHasTenant(internalTntID), testFormationTemplateID).Return(formationTemplate, nil).Once()
				return ftRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Unauthorized error when authorization check is unsuccessful but there is no error",
			transactFn: txGen.ThatSucceeds,
			formationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationID).Return(formation, nil).Once()
				return formationRepo
			},
			formationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				ftRepo := &automock.FormationTemplateRepository{}
				ftRepo.On("Get", contextThatHasTenant(internalTntID), testFormationTemplateID).Return(formationTemplateWithNonMatchingProductIDs, nil).Once()
				return ftRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Authorization fail: error when consumer info is missing in the context",
			contextFn: func() context.Context {
				return emptyCtx
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when transaction begin fails",
			transactFn: txGen.ThatFailsOnBegin,
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when getting formation fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			formationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationID).Return(nil, testErr).Once()
				return formationRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when getting formation template fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			formationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationID).Return(formation, nil).Once()
				return formationRepo
			},
			formationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				ftRepo := &automock.FormationTemplateRepository{}
				ftRepo.On("Get", contextThatHasTenant(internalTntID), testFormationTemplateID).Return(nil, testErr).Once()
				return ftRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
		{
			name:       "Authorization fail: error when committing transaction",
			transactFn: txGen.ThatFailsOnCommit,
			formationRepoFn: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("GetGlobalByID", contextThatHasTenant(internalTntID), testFormationID).Return(formation, nil).Once()
				return formationRepo
			},
			formationTemplateRepoFn: func() *automock.FormationTemplateRepository {
				ftRepo := &automock.FormationTemplateRepository{}
				ftRepo.On("Get", contextThatHasTenant(internalTntID), testFormationTemplateID).Return(formationTemplateWithNonMatchingProductIDs, nil).Once()
				return ftRepo
			},
			contextFn: func() context.Context {
				c := fixGetConsumer(consumerID, consumer.ExternalCertificate)
				return fixContextWithTenantAndConsumer(c, internalTntID, externalTntID)
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			persist, transact := fixUnusedTransactioner()
			if tCase.transactFn != nil {
				persist, transact = tCase.transactFn()
			}

			formationRepo := fixUnusedFormationRepo()
			if tCase.formationRepoFn != nil {
				formationRepo = tCase.formationRepoFn()
			}

			formationTemplateRepo := fixUnusedFormationTemplateRepo()
			if tCase.formationTemplateRepoFn != nil {
				formationTemplateRepo = tCase.formationTemplateRepoFn()
			}

			defer mock.AssertExpectationsForObjects(t, persist, transact, formationRepo, formationTemplateRepo)

			// GIVEN
			fmAuthenticator := fm.NewFormationMappingAuthenticator(transact, nil, nil, nil, nil, nil, nil, formationRepo, formationTemplateRepo, nil, tCase.consumerSubaccountLabelKey)
			formationAuthMiddleware := fmAuthenticator.FormationHandler()
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
			formationAuthMiddleware(fixTestHandler(t)).ServeHTTP(rw, httpReq)

			// THEN
			require.Equal(t, tCase.expectedStatusCode, rw.Code)
			require.Contains(t, rw.Body.String(), tCase.expectedErrOutput)
		})
	}
}
