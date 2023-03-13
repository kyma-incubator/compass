package formationmapping_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/formationmapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"

	"github.com/gorilla/mux"
	fm "github.com/kyma-incubator/compass/components/director/internal/formationmapping"
	"github.com/stretchr/testify/require"
)

func TestHandler_UpdateFormationAssignmentStatus(t *testing.T) {
	url := fmt.Sprintf("/v1/businessIntegrations/{%s}/assignments/{%s}/status", fm.FormationIDParam, fm.FormationAssignmentIDParam)
	testValidConfig := `{"testK":"testV"}`
	urlVars := map[string]string{
		fm.FormationIDParam:           testFormationID,
		fm.FormationAssignmentIDParam: testFormationAssignmentID,
	}
	configurationErr := errors.New("formation assignment configuration error")

	// formation assignment fixtures with ASSIGN operation
	faWithSourceAppAndTargetRuntime := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.ReadyAssignmentState, testValidConfig)
	reverseFAWithSourceRuntimeAndTargetApp := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faTargetID, faSourceID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication, model.ReadyAssignmentState, testValidConfig)
	faModelInput := fixFormationAssignmentInput(testFormationID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.ReadyAssignmentState, testValidConfig)

	faWithSourceAppAndTargetRuntimeWithCreateErrorState := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.CreateErrorAssignmentState, "")
	faModelInputWithCreateErrorState := fixFormationAssignmentInput(testFormationID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.CreateErrorAssignmentState, "")

	testFAReqMapping := formationassignment.FormationAssignmentRequestMapping{
		Request:             fixEmptyNotificationRequest(),
		FormationAssignment: faWithSourceAppAndTargetRuntime,
	}

	testReverseFAReqMapping := formationassignment.FormationAssignmentRequestMapping{
		Request:             fixEmptyNotificationRequest(),
		FormationAssignment: reverseFAWithSourceRuntimeAndTargetApp,
	}

	testAssignmentPair := &formationassignment.AssignmentMappingPair{
		Assignment:        &testReverseFAReqMapping,
		ReverseAssignment: &testFAReqMapping,
	}

	// formation assignment fixtures with UNASSIGN operation
	faWithSourceAppAndTargetRuntimeForUnassingOp := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.DeletingAssignmentState, testValidConfig)
	faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.DeletingAssignmentState, "")

	faWithSameSourceAppAndTarget := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faSourceID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, model.ReadyAssignmentState, "")

	testFormationAssignmentsForObject := []*model.FormationAssignment{
		{
			ID: "ID1",
		},
		{
			ID: "ID2",
		},
	}

	emptyFormationAssignmentsForObject := []*model.FormationAssignment{}

	testFormation := &model.Formation{
		ID:                  testFormationID,
		TenantID:            internalTntID,
		FormationTemplateID: "testFormationTemplateID",
		Name:                "testFormationName",
		State:               model.ReadyFormationState,
	}

	testFormationInitialState := &model.Formation{
		ID:                  testFormationID,
		TenantID:            internalTntID,
		FormationTemplateID: "testFormationTemplateID",
		Name:                "testFormationName",
		State:               model.InitialFormationState,
	}

	testCases := []struct {
		name                string
		transactFn          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		faServiceFn         func() *automock.FormationAssignmentService
		faConverterFn       func() *automock.FormationAssignmentConverter
		faNotificationSvcFn func() *automock.FormationAssignmentNotificationService
		formationSvcFn      func() *automock.FormationService
		reqBody             fm.FormationAssignmentRequestBody
		hasURLVars          bool
		headers             map[string][]string
		expectedStatusCode  int
		expectedErrOutput   string
	}{
		// Request(+metadata) validation checks
		{
			name:               "Decode Error: Content-Type header is not application/json",
			headers:            map[string][]string{httputils.HeaderContentTypeKey: {"invalidContentType"}},
			expectedStatusCode: http.StatusUnsupportedMediaType,
			expectedErrOutput:  "Content-Type header is not application/json",
		},
		{
			name: "Error when one or more of the required path parameters are missing",
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Not all of the required parameters are provided",
		},
		// Request body validation checks
		{
			name: "Validate Error: error when we have ready state with config but also an error provided",
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
				Error:         configurationErr.Error(),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		{
			name: "Validate Error: error when configuration is provided but the state is incorrect",
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.CreateErrorAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		// Business logic unit tests for assign operation
		{
			name:       "Success when operation is assign",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("Update", contextThatHasTenant(internalTntID), testFormationAssignmentID, faModelInput).Return(nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, nil).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Success when state is not changed - only configuration is provided",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("Update", contextThatHasTenant(internalTntID), testFormationAssignmentID, faModelInput).Return(nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, nil).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Error when transaction fails to begin",
			transactFn: txGen.ThatFailsOnBegin,
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when getting formation assignment globally",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(nil, testErr).Once()
				return faSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name: "Error when getting formation by formation ID fail",
			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(nil, testErr).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name: "Error when the retrieved formation is not in READY state",
			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatDoesntExpectCommit()
			},
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationInitialState, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when request body state is not correct",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State: model.DeleteErrorAssignmentState,
				Error: configurationErr.Error(),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  fmt.Sprintf("An invalid state: %s is provided for %s operation", model.DeleteErrorAssignmentState, model.AssignFormation),
		},
		{
			name:       "Error when fail to set the formation assignment to error state",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorAssignmentState).Return(testErr).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State: model.CreateErrorAssignmentState,
				Error: configurationErr.Error(),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Successfully update formation assignment when input state is create error",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeWithCreateErrorState, nil).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorAssignmentState).Return(nil).Once()
				faSvc.On("Update", contextThatHasTenant(internalTntID), testFormationAssignmentID, faModelInputWithCreateErrorState).Return(nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntimeWithCreateErrorState).Return(faModelInputWithCreateErrorState).Once()
				return faConv
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State: model.CreateErrorAssignmentState,
				Error: configurationErr.Error(),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Error when transaction fail to commit after successful formation assignment update",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeWithCreateErrorState, nil).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorAssignmentState).Return(nil).Once()
				faSvc.On("Update", contextThatHasTenant(internalTntID), testFormationAssignmentID, faModelInputWithCreateErrorState).Return(nil).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntimeWithCreateErrorState).Return(faModelInputWithCreateErrorState).Once()
				return faConv
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State: model.CreateErrorAssignmentState,
				Error: configurationErr.Error(),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when setting formation assignment state to error fail",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeWithCreateErrorState, nil).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorAssignmentState).Return(nil).Once()
				faSvc.On("Update", contextThatHasTenant(internalTntID), testFormationAssignmentID, faModelInputWithCreateErrorState).Return(testErr).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntimeWithCreateErrorState).Return(faModelInputWithCreateErrorState).Once()
				return faConv
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State: model.CreateErrorAssignmentState,
				Error: configurationErr.Error(),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when generating notifications for assignment fail",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("Update", contextThatHasTenant(internalTntID), testFormationAssignmentID, faModelInput).Return(nil).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(nil, testErr).Once()
				return faNotificationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when getting reverse formation assignment fail",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("Update", contextThatHasTenant(internalTntID), testFormationAssignmentID, faModelInput).Return(nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(nil, testErr).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when generating reverse notifications for assignment fail",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("Update", contextThatHasTenant(internalTntID), testFormationAssignmentID, faModelInput).Return(nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(nil, testErr).Once()
				return faNotificationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when processing formation assignment pairs",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("Update", contextThatHasTenant(internalTntID), testFormationAssignmentID, faModelInput).Return(nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, testErr).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when committing transaction fail",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("Update", contextThatHasTenant(internalTntID), testFormationAssignmentID, faModelInput).Return(nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, nil).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		// Business logic unit tests for unassign operation
		{
			name:       "Success when operation is unassign",
			transactFn: txGen.ThatSucceedsTwice,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("Delete", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(testFormationAssignmentsForObject, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(testFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Error when request body state is not correct",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				return faSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ConfigPendingAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Successfully set the formation assignment to error state",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, nil).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorAssignmentState).Return(nil).Once()
				return faSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State: model.DeleteErrorAssignmentState,
				Error: configurationErr.Error(),
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Error when setting formation assignment state to error fail",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, nil).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorAssignmentState).Return(testErr).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State: model.DeleteErrorAssignmentState,
				Error: configurationErr.Error(),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when deleting formation assignment fail",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("Delete", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(testErr).Once()
				return faSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name: "Error when listing formation assignments for object fail",
			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("Delete", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(nil, testErr).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Successfully unassign formation when there are no formation assignment left",
			transactFn: txGen.ThatSucceedsTwice,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("Delete", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(emptyFormationAssignmentsForObject, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(emptyFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(faWithSourceAppAndTargetRuntimeForUnassingOp.SourceType), *testFormation).Return(testFormation, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faTargetID, graphql.FormationObjectType(faWithSourceAppAndTargetRuntimeForUnassingOp.TargetType), *testFormation).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name: "Error when unassigning source from formation fail when there are no formation assignment left",
			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("Delete", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(emptyFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(graphql.FormationAssignmentTypeApplication), *testFormation).Return(nil, testErr).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when trying to update formation assignment with same source and target",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSameSourceAppAndTarget, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Cannot update formation assignment with source ",
		},
		{
			name: "Error when unassigning target from formation fail when there are no formation assignment left",
			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("Delete", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(emptyFormationAssignmentsForObject, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(emptyFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(graphql.FormationAssignmentTypeApplication), *testFormation).Return(testFormation, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faTargetID, graphql.FormationObjectType(graphql.FormationAssignmentTypeRuntime), *testFormation).Return(nil, testErr).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when transaction fail to commit",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("Delete", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name: "Error when second transaction fail on begin",
			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("Begin").Return(nil, testErr).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("Delete", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name: "Error when second transaction fail to commit",
			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				persistTx.On("Commit").Return(testErr).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Twice()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Twice()

				return persistTx, transact
			},
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("Delete", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(testFormationAssignmentsForObject, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(testFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			marshalBody, err := json.Marshal(tCase.reqBody)
			require.NoError(t, err)

			httpReq := httptest.NewRequest(http.MethodPatch, url, bytes.NewBuffer(marshalBody))
			if tCase.hasURLVars {
				httpReq = mux.SetURLVars(httpReq, urlVars)
			}

			if tCase.headers != nil {
				httpReq.Header = tCase.headers
			}
			w := httptest.NewRecorder()

			persist, transact := fixUnusedTransactioner()
			if tCase.transactFn != nil {
				persist, transact = tCase.transactFn()
			}

			faSvc := fixUnusedFormationAssignmentSvc()
			if tCase.faServiceFn != nil {
				faSvc = tCase.faServiceFn()
			}

			faConv := fixUnusedFormationAssignmentConverter()
			if tCase.faConverterFn != nil {
				faConv = tCase.faConverterFn()
			}

			faNotificationSvc := fixUnusedFormationAssignmentNotificationSvc()
			if tCase.faNotificationSvcFn != nil {
				faNotificationSvc = tCase.faNotificationSvcFn()
			}

			formationSvc := fixUnusedFormationSvc()
			if tCase.formationSvcFn != nil {
				formationSvc = tCase.formationSvcFn()
			}

			defer mock.AssertExpectationsForObjects(t, persist, transact, faConv, faSvc, faNotificationSvc, formationSvc)

			handler := fm.NewFormationMappingHandler(transact, faConv, faSvc, faNotificationSvc, formationSvc)

			// WHEN
			handler.UpdateFormationAssignmentStatus(w, httpReq)

			// THEN
			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tCase.expectedErrOutput == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, string(body), tCase.expectedErrOutput)
			}
			require.Equal(t, tCase.expectedStatusCode, resp.StatusCode)
		})
	}
}

func TestHandler_UpdateFormationStatus(t *testing.T) {
	url := fmt.Sprintf("/v1/businessIntegrations/{%s}/status", fm.FormationIDParam)
	urlVars := map[string]string{
		fm.FormationIDParam: testFormationID,
	}

	formationWithInitialState := fixFormationWithState(model.InitialFormationState)
	formationWithReadyState := fixFormationWithState(model.ReadyFormationState)
	formationWithDeletingState := fixFormationWithState(model.DeletingFormationState)

	testCases := []struct {
		name               string
		transactFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		formationSvcFn     func() *automock.FormationService
		reqBody            fm.FormationRequestBody
		hasURLVars         bool
		headers            map[string][]string
		expectedStatusCode int
		expectedErrOutput  string
	}{
		// Request(+metadata) validation checks
		{
			name:               "Decode Error: Content-Type header is not application/json",
			headers:            map[string][]string{httputils.HeaderContentTypeKey: {"invalidContentType"}},
			expectedStatusCode: http.StatusUnsupportedMediaType,
			expectedErrOutput:  "Content-Type header is not application/json",
		},
		{
			name: "Error when the required path parameter is missing",
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Not all of the required parameters are provided",
		},
		// Request body validation checks
		{
			name: "Validate Error: error when the state has unsupported value",
			reqBody: fm.FormationRequestBody{
				State: "unsupported",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		{
			name: "Validate Error: error when we have an error with incorrect state",
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
				Error: testErr.Error(),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		// Business logic unit tests
		{
			name:       "Error when transaction fails to begin",
			transactFn: txGen.ThatFailsOnBegin,
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when getting formation globally fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(nil, testErr).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		// Business logic unit tests for delete formation operation
		{
			name:       "Successfully update formation status when operation is delete formation",
			transactFn: txGen.ThatSucceeds,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
				formationSvc.On("DeleteFormationEntityAndScenarios", contextThatHasTenant(internalTntID), internalTntID, testFormationName).Return(nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Error when request body state is not correct for delete formation operation",
			transactFn: txGen.ThatDoesntExpectCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.CreateErrorFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when updating the formation to delete error state fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
				formationSvc.On("SetFormationToErrorState", contextThatHasTenant(internalTntID), formationWithDeletingState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorFormationState).Return(testErr).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.DeleteErrorFormationState,
				Error: testErr.Error(),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when deleting formation fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
				formationSvc.On("DeleteFormationEntityAndScenarios", contextThatHasTenant(internalTntID), internalTntID, testFormationName).Return(testErr).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when transaction fails to commit after successful formation status update for delete operation",
			transactFn: txGen.ThatFailsOnCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
				formationSvc.On("DeleteFormationEntityAndScenarios", contextThatHasTenant(internalTntID), internalTntID, testFormationName).Return(nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		// Business logic unit tests for create formation operation
		{
			name:       "Successfully update formation status when operation is create formation",
			transactFn: txGen.ThatSucceeds,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				formationSvc.On("Update", contextThatHasTenant(internalTntID), formationWithReadyState).Return(nil).Once()
				formationSvc.On("ResynchronizeFormationNotifications", contextThatHasTenant(internalTntID), testFormationID).Return(nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Error when request body state is not correct for create formation operation",
			transactFn: txGen.ThatDoesntExpectCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.DeleteErrorFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when updating the formation to create error state fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				formationSvc.On("SetFormationToErrorState", contextThatHasTenant(internalTntID), formationWithInitialState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorFormationState).Return(testErr).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.CreateErrorFormationState,
				Error: testErr.Error(),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when updating the formation to ready state fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				formationSvc.On("Update", contextThatHasTenant(internalTntID), formationWithReadyState).Return(testErr).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when resynchronize formation notifications fails",
			transactFn: txGen.ThatDoesntExpectCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				formationSvc.On("Update", contextThatHasTenant(internalTntID), formationWithReadyState).Return(nil).Once()
				formationSvc.On("ResynchronizeFormationNotifications", contextThatHasTenant(internalTntID), testFormationID).Return(testErr).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when transaction fails to commit after successful formation status update for create operation",
			transactFn: txGen.ThatFailsOnCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				formationSvc.On("Update", contextThatHasTenant(internalTntID), formationWithReadyState).Return(nil).Once()
				formationSvc.On("ResynchronizeFormationNotifications", contextThatHasTenant(internalTntID), testFormationID).Return(nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			// GIVEN
			marshalBody, err := json.Marshal(tCase.reqBody)
			require.NoError(t, err)

			httpReq := httptest.NewRequest(http.MethodPatch, url, bytes.NewBuffer(marshalBody))
			if tCase.hasURLVars {
				httpReq = mux.SetURLVars(httpReq, urlVars)
			}

			if tCase.headers != nil {
				httpReq.Header = tCase.headers
			}
			w := httptest.NewRecorder()

			persist, transact := fixUnusedTransactioner()
			if tCase.transactFn != nil {
				persist, transact = tCase.transactFn()
			}

			formationSvc := fixUnusedFormationSvc()
			if tCase.formationSvcFn != nil {
				formationSvc = tCase.formationSvcFn()
			}

			defer mock.AssertExpectationsForObjects(t, persist, transact)

			handler := fm.NewFormationMappingHandler(transact, nil, nil, nil, formationSvc)

			// WHEN
			handler.UpdateFormationStatus(w, httpReq)

			// THEN
			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tCase.expectedErrOutput == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, string(body), tCase.expectedErrOutput)
			}
			require.Equal(t, tCase.expectedStatusCode, resp.StatusCode)
		})
	}
}
