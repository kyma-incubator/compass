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

func Test_StatusUpdate(t *testing.T) {
	testFormationID := "testFormationID"
	testFormationAssignmentID := "testFormationAssignmentID"
	url := fmt.Sprintf("/v1/businessIntegrations/{%s}/assignments/{%s}/status", fm.FormationIDParam, fm.FormationAssignmentIDParam)
	testValidConfig := `{"testK":"testV"}`
	urlVars := map[string]string{
		fm.FormationIDParam:           testFormationID,
		fm.FormationAssignmentIDParam: testFormationAssignmentID,
	}

	configurationErr := errors.New("formation assignment configuration error")
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	faSourceID := "testSourceID"
	faTargetID := "testTargetID"
	internalTntID := "testInternalID"

	// formation assignment fixtures with ASSIGN operation
	faWithSourceAppAndTargetRuntime := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.AssignFormation, model.ReadyAssignmentState, testValidConfig)
	reverseFAWithSourceRuntimeAndTargetApp := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faTargetID, faSourceID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication, model.AssignFormation, model.ReadyAssignmentState, testValidConfig)
	faModelInput := fixFormationAssignmentInput(testFormationID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.AssignFormation, model.ReadyAssignmentState, testValidConfig)

	faWithSourceAppAndTargetRuntimeWithCreateErrorState := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.AssignFormation, model.CreateErrorAssignmentState, "")
	faModelInputWithCreateErrorState := fixFormationAssignmentInput(testFormationID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.AssignFormation, model.CreateErrorAssignmentState, "")

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
	faWithSourceAppAndTargetRuntimeForUnassingOp := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.UnassignFormation, model.ReadyAssignmentState, testValidConfig)
	faWithSourceAppAndTargetRuntimeForUnassingOpWithDeleteErrorState := fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.UnassignFormation, model.DeleteErrorAssignmentState, "")

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
		ID:                  "testFormationID",
		TenantID:            internalTntID,
		FormationTemplateID: "testFormationTemplateID",
		Name:                "testFormationName",
	}

	testCases := []struct {
		name                string
		transactFn          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		faServiceFn         func() *automock.FormationAssignmentService
		faConverterFn       func() *automock.FormationAssignmentConverter
		faNotificationSvcFn func() *automock.FormationAssignmentNotificationService
		formationSvcFn      func() *automock.FormationService
		reqBody             fm.RequestBody
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
			reqBody: fm.RequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Not all of the required parameters are provided",
		},
		// Request body validation checks
		{
			name: "Validate Error: error when we have ready state with config but also an error provided",
			reqBody: fm.RequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
				Error:         configurationErr.Error(),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		{
			name: "Validate Error: error when configuration is provided but the state is incorrect",
			reqBody: fm.RequestBody{
				State:         model.CreateErrorAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		{
			name: "Validate Error: error when request body contains only state",
			reqBody: fm.RequestBody{
				State: model.ReadyAssignmentState,
			},
			hasURLVars:         true,
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
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(nil).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			formationSvcFn: fixUnusedFormationSvc,
			reqBody: fm.RequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name:       "Error when transaction fails to begin",
			transactFn: txGen.ThatFailsOnBegin,
			reqBody: fm.RequestBody{
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
			reqBody: fm.RequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusInternalServerError,
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
			reqBody: fm.RequestBody{
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
			reqBody: fm.RequestBody{
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
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntimeWithCreateErrorState).Return(faModelInputWithCreateErrorState).Once()
				return faConv
			},
			reqBody: fm.RequestBody{
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
			reqBody: fm.RequestBody{
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
			reqBody: fm.RequestBody{
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
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(nil, testErr).Once()
				return faNotificationSvc
			},
			reqBody: fm.RequestBody{
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
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			reqBody: fm.RequestBody{
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
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(nil, testErr).Once()
				return faNotificationSvc
			},
			reqBody: fm.RequestBody{
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
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(testErr).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			reqBody: fm.RequestBody{
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
				faSvc.On("ProcessFormationAssignmentPair", contextThatHasTenant(internalTntID), testAssignmentPair).Return(nil).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateNotification", contextThatHasTenant(internalTntID), faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateNotification", contextThatHasTenant(internalTntID), reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			reqBody: fm.RequestBody{
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
				return faSvc
			},
			reqBody: fm.RequestBody{
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
			reqBody: fm.RequestBody{
				State:         model.ConfigPendingAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
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
			reqBody: fm.RequestBody{
				State: model.DeleteErrorAssignmentState,
				Error: configurationErr.Error(),
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
			reqBody: fm.RequestBody{
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
			reqBody: fm.RequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
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
			reqBody: fm.RequestBody{
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
				formationSvc.On("Get", contextThatHasTenant(internalTntID), testFormationID).Return(nil, testErr).Once()
				return formationSvc
			},
			reqBody: fm.RequestBody{
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
				return faSvc
			},
			formationSvcFn: func() *automock.FormationService {
				formationSvc := &automock.FormationService{}
				formationSvc.On("Get", contextThatHasTenant(internalTntID), testFormationID).Return(testFormation, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(faWithSourceAppAndTargetRuntimeForUnassingOp.LastOperationInitiatorType), *testFormation).Return(testFormation, nil).Once()
				return formationSvc
			},
			reqBody: fm.RequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name: "Error when unassigning formation fail when there are no formation assignment left",
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
				formationSvc.On("Get", contextThatHasTenant(internalTntID), testFormationID).Return(testFormation, nil).Once()
				formationSvc.On("UnassignFormation", contextThatHasTenant(internalTntID), internalTntID, faSourceID, graphql.FormationObjectType(faWithSourceAppAndTargetRuntimeForUnassingOp.LastOperationInitiatorType), *testFormation).Return(nil, testErr).Once()
				return formationSvc
			},
			reqBody: fm.RequestBody{
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
			reqBody: fm.RequestBody{
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
			reqBody: fm.RequestBody{
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
				return faSvc
			},
			reqBody: fm.RequestBody{
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
			handler.UpdateStatus(w, httpReq)

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
