package formationmapping_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	fm "github.com/kyma-incubator/compass/components/director/internal/formationmapping"
	"github.com/kyma-incubator/compass/components/director/internal/formationmapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
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

	faWithSourceAppAndTargetRuntimeWithCreateErrorState := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.CreateErrorAssignmentState, "")

	testFAReqMapping := formationassignment.FormationAssignmentRequestMapping{
		Request:             fixEmptyNotificationRequest(),
		FormationAssignment: faWithSourceAppAndTargetRuntime,
	}

	testReverseFAReqMapping := formationassignment.FormationAssignmentRequestMapping{
		Request:             fixEmptyNotificationRequest(),
		FormationAssignment: reverseFAWithSourceRuntimeAndTargetApp,
	}

	testAssignmentPair := &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			Assignment:        &testReverseFAReqMapping,
			ReverseAssignment: &testFAReqMapping,
		},
		Operation: model.AssignFormation,
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

	testFormationWithReadyState := fixFormationWithState(model.ReadyFormationState)
	testFormationWithInitialState := fixFormationWithState(model.InitialFormationState)

	testCases := []struct {
		name                string
		transactFn          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		faServiceFn         func() *automock.FormationAssignmentService
		faNotificationSvcFn func() *automock.FormationAssignmentNotificationService
		formationSvcFn      func() *automock.FormationService
		faStatusSvcFn       func() *automock.FormationAssignmentStatusService
		reqBody             fm.FormationAssignmentRequestBody
		hasURLVars          bool
		headers             map[string][]string
		expectedStatusCode  int
		expectedErrOutput   string
		shouldSleep         bool // set to true for the test cases where we enter the go routine so that we give it some extra time to complete and then be able to assert the mocks
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
			transactFn: txGen.ThatSucceedsTwice,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, nil).Once()
				return faSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Success when state is not changed - only configuration is provided",
			transactFn: txGen.ThatSucceedsTwice,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, nil).Once()
				return faSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithInitialState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
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
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithInitialState, nil).Once()
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
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
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
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", txtest.CtxWithDBMatcher(), faWithSourceAppAndTargetRuntime, configurationErr.Error(), formationassignment.AssignmentErrorCode(2), model.CreateErrorAssignmentState, model.AssignFormation).Return(testErr).Once()
				return updater
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
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorAssignmentState, model.AssignFormation).Return(nil).Once()
				return updater
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
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeWithCreateErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorAssignmentState, model.AssignFormation).Return(nil).Once()
				return updater
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
			name:       "Error when update with constraints fail",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(testErr).Once()
				return updater
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
			name:       "Error when committing transaction fail before go routine",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		// Unit tests for formation assignment notifications processing in the go routine
		{
			name:       "Error when transaction fails to begin in go routine when operation is assign",
			transactFn: ThatFailsOnBeginInGoRoutine,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				return faSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				return &automock.FormationAssignmentNotificationService{}
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Error in go routine when generating notifications for assignment fail",
			transactFn: ThatDoesNotCommitInGoRoutine,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil, testErr).Once()
				return faNotificationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Returning when no formation assignment notification is generated in the go routine",
			transactFn: ThatDoesNotCommitInGoRoutine,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				return faSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil, nil).Once()
				return faNotificationSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithInitialState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Error in go routine when getting reverse formation assignment fail",
			transactFn: ThatDoesNotCommitInGoRoutine,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(nil, testErr).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Error in go routine when generating notifications for reverse assignment fail",
			transactFn: ThatDoesNotCommitInGoRoutine,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp, model.AssignFormation).Return(nil, testErr).Once()
				return faNotificationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Updating assignment to error state when processing formation assignment pairs in the go routine fails",
			transactFn: txGen.ThatSucceedsTwice,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, testErr).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.TechnicalError), model.CreateErrorAssignmentState).Return(nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Error in go routine when processing formation assignment pairs and setting assignment to error state fails",
			transactFn: ThatDoesNotCommitInGoRoutine,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, testErr).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.TechnicalError), model.CreateErrorAssignmentState).Return(testErr).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Error in go routine when processing formation assignment pairs fails and committing transaction fails",
			transactFn: ThatFailsOnCommitInGoRoutine,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, testErr).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.TechnicalError), model.CreateErrorAssignmentState).Return(nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Processing formation assignment pair succeeds but committing transaction fail in the go routine",
			transactFn: ThatFailsOnCommitInGoRoutine,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp, nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp, model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime, model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		// Business logic unit tests for unassign operation
		{
			name:       "Success when operation is unassign",
			transactFn: txGen.ThatSucceedsTwice,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(testFormationAssignmentsForObject, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(testFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				return faStatusSvc
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
			name:       "Error when request body state is not correct and commit fails",
			transactFn: txGen.ThatFailsOnCommit,
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
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when request body state is not correct",
			transactFn: txGen.ThatSucceeds,
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
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
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
				return faSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State: model.DeleteErrorAssignmentState,
				Error: configurationErr.Error(),
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorAssignmentState, model.UnassignFormation).Return(nil).Once()
				return updater
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Error when setting formation assignment state to error fail and then commit fails",
			transactFn: txGen.ThatFailsOnCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorAssignmentState, model.UnassignFormation).Return(testErr).Once()
				return updater
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
			name:       "Error when setting formation assignment state to error fail",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState, configurationErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorAssignmentState, model.UnassignFormation).Return(testErr).Once()
				return updater
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
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeForUnassingOp, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.TechnicalError), model.DeleteErrorAssignmentState).Return(nil).Once()
				return faSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(testErr).Once()
				return faStatusSvc
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when deleting formation assignment fails and then setting assignment to error state fails",
			transactFn: txGen.ThatSucceeds,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("SetAssignmentToErrorState", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntimeForUnassingOp, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.TechnicalError), model.DeleteErrorAssignmentState).Return(testErr).Once()
				return faSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(testErr).Once()
				return faStatusSvc
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Success when operation is unassign and deletion with constraints returns not found",
			transactFn: txGen.ThatSucceedsTwice,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(testFormationAssignmentsForObject, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(testFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(apperrors.NewNotFoundErrorWithType(resource.FormationAssignment)).Once()
				return faStatusSvc
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
			name: "Error when listing formation assignments for object fail",
			transactFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				return txGen.ThatSucceedsMultipleTimesAndCommitsMultipleTimes(2, 1)
			},
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntimeForUnassingOp, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(nil, testErr).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				return faStatusSvc
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
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(emptyFormationAssignmentsForObject, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(emptyFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(faWithSourceAppAndTargetRuntimeForUnassingOp.SourceType), *testFormationWithReadyState).Return(testFormationWithReadyState, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faTargetID, graphql.FormationObjectType(faWithSourceAppAndTargetRuntimeForUnassingOp.TargetType), *testFormationWithReadyState).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				return faStatusSvc
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
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(emptyFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(graphql.FormationAssignmentTypeApplication), *testFormationWithReadyState).Return(nil, testErr).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				return faStatusSvc
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
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
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
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(emptyFormationAssignmentsForObject, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(emptyFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(graphql.FormationAssignmentTypeApplication), *testFormationWithReadyState).Return(testFormationWithReadyState, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faTargetID, graphql.FormationObjectType(graphql.FormationAssignmentTypeRuntime), *testFormationWithReadyState).Return(nil, testErr).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				return faStatusSvc
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
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				return faStatusSvc
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
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				return faStatusSvc
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
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faSourceID).Return(testFormationAssignmentsForObject, nil).Once()
				faSvc.On("ListFormationAssignmentsForObjectID", contextThatHasTenant(internalTntID), testFormationID, faTargetID).Return(testFormationAssignmentsForObject, nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				faStatusSvc := &automock.FormationAssignmentStatusService{}
				faStatusSvc.On("DeleteWithConstraints", contextThatHasTenant(internalTntID), testFormationAssignmentID).Return(nil).Once()
				return faStatusSvc
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

			faNotificationSvc := fixUnusedFormationAssignmentNotificationSvc()
			if tCase.faNotificationSvcFn != nil {
				faNotificationSvc = tCase.faNotificationSvcFn()
			}

			formationSvc := fixUnusedFormationSvc()
			if tCase.formationSvcFn != nil {
				formationSvc = tCase.formationSvcFn()
			}

			faStatusSvcFn := &automock.FormationAssignmentStatusService{}
			if tCase.faStatusSvcFn != nil {
				faStatusSvcFn = tCase.faStatusSvcFn()
			}

			defer mock.AssertExpectationsForObjects(t, persist, transact, faSvc, faNotificationSvc, formationSvc, faStatusSvcFn)

			handler := fm.NewFormationMappingHandler(transact, faSvc, faStatusSvcFn, faNotificationSvc, formationSvc, nil)

			// WHEN
			handler.UpdateFormationAssignmentStatus(w, httpReq)

			if tCase.shouldSleep {
				time.Sleep(500 * time.Millisecond)
			}

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

func TestHandler_ResetFormationAssignmentStatus(t *testing.T) {
	url := fmt.Sprintf("/v1/businessIntegrations/{%s}/assignments/{%s}/status", fm.FormationIDParam, fm.FormationAssignmentIDParam)
	testValidConfig := `{"testK":"testV"}`
	urlVars := map[string]string{
		fm.FormationIDParam:           testFormationID,
		fm.FormationAssignmentIDParam: testFormationAssignmentID,
	}

	// formation assignment fixtures with ASSIGN operation
	faWithSourceAppAndTargetRuntime := func(state model.FormationAssignmentState) *model.FormationAssignment {
		return fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, state, testValidConfig)
	}
	reverseFAWithSourceRuntimeAndTargetApp := func(state model.FormationAssignmentState) *model.FormationAssignment {
		return fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faTargetID, faSourceID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication, state, testValidConfig)
	}

	testFAReqMapping := formationassignment.FormationAssignmentRequestMapping{
		Request:             fixEmptyNotificationRequest(),
		FormationAssignment: faWithSourceAppAndTargetRuntime(model.InitialAssignmentState),
	}

	testReverseFAReqMapping := formationassignment.FormationAssignmentRequestMapping{
		Request:             fixEmptyNotificationRequest(),
		FormationAssignment: reverseFAWithSourceRuntimeAndTargetApp(model.InitialAssignmentState),
	}

	testAssignmentPair := &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			Assignment:        &testReverseFAReqMapping,
			ReverseAssignment: &testFAReqMapping,
		},
		Operation: model.AssignFormation,
	}

	testFormationWithReadyState := fixFormationWithState(model.ReadyFormationState)

	testCases := []struct {
		name                string
		transactFn          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		faServiceFn         func() *automock.FormationAssignmentService
		faNotificationSvcFn func() *automock.FormationAssignmentNotificationService
		formationSvcFn      func() *automock.FormationService
		faStatusSvcFn       func() *automock.FormationAssignmentStatusService
		reqBody             fm.FormationAssignmentRequestBody
		hasURLVars          bool
		headers             map[string][]string
		expectedStatusCode  int
		expectedErrOutput   string
		shouldSleep         bool // set to true for the test cases where we enter the go routine so that we give it some extra time to complete and then be able to assert the mocks
	}{
		{
			name:       "Success when both assignments are in READY state",
			transactFn: txGen.ThatSucceedsTwice,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime(model.ReadyAssignmentState), nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp(model.ReadyAssignmentState), nil).Twice()
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(false, nil).Once()
				return faSvc
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime(model.InitialAssignmentState), model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateFormationAssignmentNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp(model.InitialAssignmentState), model.AssignFormation).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			faStatusSvcFn: func() *automock.FormationAssignmentStatusService {
				updater := &automock.FormationAssignmentStatusService{}
				updater.On("UpdateWithConstraints", txtest.CtxWithDBMatcher(), faWithSourceAppAndTargetRuntime(model.ReadyAssignmentState), model.AssignFormation).Return(nil).Once()
				return updater
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Fail when reverse assignment is not in READY state",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime(model.ReadyAssignmentState), nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(reverseFAWithSourceRuntimeAndTargetApp(model.InitialAssignmentState), nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Fail when assignment is not in READY state",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime(model.InitialAssignmentState), nil).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Fail when failing to get reverse assignment",
			transactFn: txGen.ThatDoesntExpectCommit,
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", txtest.CtxWithDBMatcher(), testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime(model.ReadyAssignmentState), nil).Once()
				faSvc.On("GetReverseBySourceAndTarget", contextThatHasTenant(internalTntID), testFormationID, faSourceID, faTargetID).Return(nil, testErr).Once()
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", txtest.CtxWithDBMatcher(), testFormationID).Return(testFormationWithReadyState, nil).Once()
				return formationSvc
			},
			reqBody: fm.FormationAssignmentRequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "",
			shouldSleep:        true,
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

			faNotificationSvc := fixUnusedFormationAssignmentNotificationSvc()
			if tCase.faNotificationSvcFn != nil {
				faNotificationSvc = tCase.faNotificationSvcFn()
			}

			formationSvc := fixUnusedFormationSvc()
			if tCase.formationSvcFn != nil {
				formationSvc = tCase.formationSvcFn()
			}

			faStatusSvcFn := &automock.FormationAssignmentStatusService{}
			if tCase.faStatusSvcFn != nil {
				faStatusSvcFn = tCase.faStatusSvcFn()
			}

			defer mock.AssertExpectationsForObjects(t, persist, transact, faSvc, faNotificationSvc, formationSvc, faStatusSvcFn)

			handler := fm.NewFormationMappingHandler(transact, faSvc, faStatusSvcFn, faNotificationSvc, formationSvc, nil)

			// WHEN
			handler.ResetFormationAssignmentStatus(w, httpReq)

			if tCase.shouldSleep {
				time.Sleep(500 * time.Millisecond)
			}

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
		name                 string
		transactFn           func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		formationSvcFn       func() *automock.FormationService
		faStatusSvcFn        func() *automock.FormationAssignmentStatusService
		formationStatusSvcFn func() *automock.FormationStatusService
		reqBody              fm.FormationRequestBody
		hasURLVars           bool
		headers              map[string][]string
		expectedStatusCode   int
		expectedErrOutput    string
		shouldSleep          bool // set to true for the test cases where we enter the go routine so that we give it some extra time to complete and then be able to assert the mocks
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
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("DeleteFormationEntityAndScenariosWithConstraints", contextThatHasTenant(internalTntID), internalTntID, formationWithDeletingState).Return(nil).Once()
				return formationStatusSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:       "Successfully update formation status when operation is delete formation and state is DELETE_ERROR",
			transactFn: txGen.ThatSucceeds,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithDeletingState, nil).Once()
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("SetFormationToErrorStateWithConstraints", contextThatHasTenant(internalTntID), formationWithDeletingState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorFormationState, model.DeleteFormation).Return(nil).Once()
				return formationStatusSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.DeleteErrorFormationState,
				Error: testErr.Error(),
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
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("SetFormationToErrorStateWithConstraints", contextThatHasTenant(internalTntID), formationWithDeletingState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.DeleteErrorFormationState, model.DeleteFormation).Return(testErr).Once()
				return formationStatusSvc
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
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("DeleteFormationEntityAndScenariosWithConstraints", contextThatHasTenant(internalTntID), internalTntID, formationWithDeletingState).Return(testErr).Once()
				return formationStatusSvc
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
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("DeleteFormationEntityAndScenariosWithConstraints", contextThatHasTenant(internalTntID), internalTntID, formationWithDeletingState).Return(nil).Once()
				return formationStatusSvc
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
			transactFn: txGen.ThatSucceedsTwice,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				formationSvc.On("ResynchronizeFormationNotifications", contextThatHasTenant(internalTntID), testFormationID, false).Return(nil, nil).Once()
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), formationWithReadyState, model.CreateFormation).Return(nil).Once()
				return formationStatusSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			shouldSleep:        true,
		},
		{
			name:       "Successfully update formation status when operation is create formation and state is CREATE_ERROR",
			transactFn: txGen.ThatSucceeds,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("SetFormationToErrorStateWithConstraints", contextThatHasTenant(internalTntID), formationWithInitialState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorFormationState, model.CreateFormation).Return(nil).Once()
				return formationStatusSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.CreateErrorFormationState,
				Error: testErr.Error(),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			shouldSleep:        true,
		},
		{
			name:       "Error on begin transaction in go routine",
			transactFn: ThatFailsOnBeginInGoRoutine,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), formationWithReadyState, model.CreateFormation).Return(nil).Once()
				return formationStatusSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			shouldSleep:        true,
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
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("SetFormationToErrorStateWithConstraints", contextThatHasTenant(internalTntID), formationWithInitialState, testErr.Error(), formationassignment.AssignmentErrorCode(formationassignment.ClientError), model.CreateErrorFormationState, model.CreateFormation).Return(testErr).Once()
				return formationStatusSvc
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
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), formationWithReadyState, model.CreateFormation).Return(testErr).Once()
				return formationStatusSvc
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
			transactFn: ThatDoesNotCommitInGoRoutine,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				formationSvc.On("ResynchronizeFormationNotifications", contextThatHasTenant(internalTntID), testFormationID, false).Return(nil, testErr).Once()
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), formationWithReadyState, model.CreateFormation).Return(nil).Once()
				return formationStatusSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
		},
		{
			name:       "Error when transaction fails to commit after successful formation status update for create operation",
			transactFn: txGen.ThatFailsOnCommit,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), formationWithReadyState, model.CreateFormation).Return(nil).Once()
				return formationStatusSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedErrOutput:  "An unexpected error occurred while processing the request. X-Request-Id:",
		},
		{
			name:       "Error when transaction fails to commit in go routine after successful formation status update for create operation",
			transactFn: ThatFailsOnCommitInGoRoutine,
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("GetGlobalByID", txtest.CtxWithDBMatcher(), testFormationID).Return(formationWithInitialState, nil).Once()
				formationSvc.On("ResynchronizeFormationNotifications", contextThatHasTenant(internalTntID), testFormationID, false).Return(nil, nil).Once()
				return formationSvc
			},
			formationStatusSvcFn: func() *automock.FormationStatusService {
				formationStatusSvc := &automock.FormationStatusService{}
				formationStatusSvc.On("UpdateWithConstraints", contextThatHasTenant(internalTntID), formationWithReadyState, model.CreateFormation).Return(nil).Once()
				return formationStatusSvc
			},
			reqBody: fm.FormationRequestBody{
				State: model.ReadyFormationState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
			shouldSleep:        true,
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

			faUpdater := &automock.FormationAssignmentStatusService{}
			if tCase.faStatusSvcFn != nil {
				faUpdater = tCase.faStatusSvcFn()
			}

			formationStatusSvc := &automock.FormationStatusService{}
			if tCase.formationStatusSvcFn != nil {
				formationStatusSvc = tCase.formationStatusSvcFn()
			}

			defer mock.AssertExpectationsForObjects(t, persist, transact, formationSvc, faUpdater, formationStatusSvc)

			handler := fm.NewFormationMappingHandler(transact, nil, faUpdater, nil, formationSvc, formationStatusSvc)

			// WHEN
			handler.UpdateFormationStatus(w, httpReq)

			if tCase.shouldSleep {
				time.Sleep(500 * time.Millisecond)
			}

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
