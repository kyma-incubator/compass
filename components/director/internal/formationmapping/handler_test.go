package formationmapping_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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

	//faWithSourceAppAndTargetRuntime := fixFormationAssignmentModel(testFormationID, internalTntID, faSourceID, faTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime)
	//reverseFAWithSourceRuntimeAndTargetApp := fixFormationAssignmentModel(testFormationID, internalTntID, faTargetID, faSourceID, model.FormationAssignmentTypeRuntime, model.FormationAssignmentTypeApplication)

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
		faServiceFn         func() *automock.FormationAssignmentService
		faConverterFn       func() *automock.FormationAssignmentConverter
		faNotificationSvcFn func() *automock.FormationAssignmentNotificationService
		formationRepoFn     func() *automock.FormationRepository
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
			formationRepoFn:     fixUnusedFormationRepo,
			headers:             map[string][]string{httputils.HeaderContentTypeKey: {"invalidContentType"}},
			expectedStatusCode:  http.StatusUnsupportedMediaType,
			expectedErrOutput:   "Content-Type header is not application/json",
		},
		{
			name:                "Error when one or more of the required path parameters are missing",
			faServiceFn:         fixUnusedFormationAssignmentSvc,
			faConverterFn:       fixUnusedFormationAssignmentConverter,
			faNotificationSvcFn: fixUnusedFormationAssignmentNotificationSvc,
			formationRepoFn:     fixUnusedFormationRepo,
			reqBody: fm.RequestBody{
				State:         fm.ConfigurationStateReady,
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
			formationRepoFn:     fixUnusedFormationRepo,
			reqBody: fm.RequestBody{
				State:         fm.ConfigurationStateReady,
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
			formationRepoFn:     fixUnusedFormationRepo,
			reqBody: fm.RequestBody{
				State:         fm.ConfigurationStateCreateError,
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
			formationRepoFn:     fixUnusedFormationRepo,
			reqBody: fm.RequestBody{
				State: fm.ConfigurationStateReady,
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
				faSvc.On("UpdateFormationAssignment", mock.Anything, testAssignmentPair).Return(nil).Once()
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
			formationRepoFn: fixUnusedFormationRepo,
			reqBody: fm.RequestBody{
				State:         fm.ConfigurationStateReady,
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

			faSvc := tCase.faServiceFn()
			faConv := tCase.faConverterFn()
			faNotificationSvc := tCase.faNotificationSvcFn()
			formationRepo := tCase.formationRepoFn()
			defer mock.AssertExpectationsForObjects(t, faSvc, faConv, faNotificationSvc, formationRepo)

			handler := fm.NewFormationMappingHandler(faSvc, faConv, faNotificationSvc)

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
