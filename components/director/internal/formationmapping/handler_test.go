package formationmapping_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
	testFormationAssignmentID := "testFormationAsisgnmentID"
	url := fmt.Sprintf("/v1/businessIntegrations/{%s}/assignments/{%s}/status", fm.FormationIDParam, fm.FormationAssignmentIDParam)
	testValidConfig := `{"testK":"testV"}`
	urlVars := map[string]string{
		fm.FormationIDParam:           testFormationID,
		fm.FormationAssignmentIDParam: testFormationAssignmentID,
	}

	faSourceID := "testSourceID"
	faTargetID := "testTargetID"
	internalTntID := "testInternalID"

	faWithSourceAppAndTargetRuntime := fixFormationAssignmentModelWithStateAndConfig(testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.ReadyAssignmentState, testValidConfig)
	reverseFAWithSourceRuntimeAndTargetApp := fixFormationAssignmentModelWithStateAndConfig(testFormationID, internalTntID, faTargetID, faSourceID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication, model.ReadyAssignmentState, testValidConfig)

	faModelInput := fixFormationAssignmentInput(testFormationID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, model.ReadyAssignmentState, testValidConfig)

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

	testCases := []struct {
		name                string
		reqBody             fm.RequestBody
		transactFn          func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		faServiceFn         func() *automock.FormationAssignmentService
		faConverterFn       func() *automock.FormationAssignmentConverter
		faNotificationSvcFn func() *automock.FormationAssignmentNotificationService
		formationSvcFn      func() *automock.FormationService
		hasURLVars          bool
		headers             map[string][]string
		expectedStatusCode  int
		expectedErrOutput   string
	}{
		// Request(+metadata) validation checks
		{
			name:                "Decode Error: Content-Type header is not application/json",
			faServiceFn:         fixUnusedFormationAssignmentSvc,
			faConverterFn:       fixUnusedFormationAssignmentConverter,
			faNotificationSvcFn: fixUnusedFormationAssignmentNotificationSvc,
			headers:             map[string][]string{httputils.HeaderContentTypeKey: {"invalidContentType"}},
			expectedStatusCode:  http.StatusUnsupportedMediaType,
			expectedErrOutput:   "Content-Type header is not application/json",
		},
		{
			name:                "Error when one or more of the required path parameters are missing",
			faServiceFn:         fixUnusedFormationAssignmentSvc,
			faConverterFn:       fixUnusedFormationAssignmentConverter,
			faNotificationSvcFn: fixUnusedFormationAssignmentNotificationSvc,
			reqBody: fm.RequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Not all of the required parameters are provided",
		},
		// Request body validation checks
		{
			name:                "Validate Error: error when we have ready state with config but also an error provided",
			faServiceFn:         fixUnusedFormationAssignmentSvc,
			faConverterFn:       fixUnusedFormationAssignmentConverter,
			faNotificationSvcFn: fixUnusedFormationAssignmentNotificationSvc,
			reqBody: fm.RequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
				Error:         "testErrMsg",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		{
			name:                "Validate Error: error when configuration is provided but the state is incorrect",
			faServiceFn:         fixUnusedFormationAssignmentSvc,
			faConverterFn:       fixUnusedFormationAssignmentConverter,
			faNotificationSvcFn: fixUnusedFormationAssignmentNotificationSvc,
			reqBody: fm.RequestBody{
				State:         model.CreateErrorAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		{
			name:                "Validate Error: error when request body contains only state",
			faServiceFn:         fixUnusedFormationAssignmentSvc,
			faConverterFn:       fixUnusedFormationAssignmentConverter,
			faNotificationSvcFn: fixUnusedFormationAssignmentNotificationSvc,
			reqBody: fm.RequestBody{
				State: model.ReadyAssignmentState,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		// Business logic unit tests
		{
			name: "Success",
			faServiceFn: func() *automock.FormationAssignmentService {
				faSvc := &automock.FormationAssignmentService{}
				faSvc.On("GetGlobalByIDAndFormationID", mock.Anything, testFormationAssignmentID, testFormationID).Return(faWithSourceAppAndTargetRuntime, nil).Once()
				faSvc.On("Update", mock.Anything, testFormationAssignmentID, faModelInput).Return(nil).Once()
				faSvc.On("ProcessFormationAssignmentPair", mock.Anything, testAssignmentPair).Return(nil).Once()
				return faSvc
			},
			faConverterFn: func() *automock.FormationAssignmentConverter {
				faConv := &automock.FormationAssignmentConverter{}
				faConv.On("ToInput", faWithSourceAppAndTargetRuntime).Return(faModelInput).Once()
				return faConv
			},
			faNotificationSvcFn: func() *automock.FormationAssignmentNotificationService {
				faNotificationSvc := &automock.FormationAssignmentNotificationService{}
				faNotificationSvc.On("GenerateNotification", mock.Anything, faWithSourceAppAndTargetRuntime).Return(fixEmptyNotificationRequest(), nil).Once()
				faNotificationSvc.On("GenerateNotification", mock.Anything, reverseFAWithSourceRuntimeAndTargetApp).Return(fixEmptyNotificationRequest(), nil).Once()
				return faNotificationSvc
			},
			reqBody: fm.RequestBody{
				State:         model.ReadyAssignmentState,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
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

			persist, transact := tCase.transactFn()
			faSvc := tCase.faServiceFn()
			faConv := tCase.faConverterFn()
			faNotificationSvc := tCase.faNotificationSvcFn()
			formationSvc := tCase.formationSvcFn()
			defer mock.AssertExpectationsForObjects(t, persist, transact, faConv, faSvc, faNotificationSvc)

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
