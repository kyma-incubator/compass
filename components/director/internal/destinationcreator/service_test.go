package destinationcreator_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	"github.com/kyma-incubator/compass/components/director/internal/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/internal/destinationcreator/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	faWithSourceAppAndTargetApp        = fixFormationAssignmentModelWithParameters(testAssignmentID, testFormationID, testTenantID, testSourceID, testTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)
	faWithSourceAppAndTargetRuntime    = fixFormationAssignmentModelWithParameters(testAssignmentID, testFormationID, testTenantID, testSourceID, testTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntime, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)
	faWithSourceAppAndTargetRuntimeCtx = fixFormationAssignmentModelWithParameters(testAssignmentID, testFormationID, testTenantID, testSourceID, testTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeRuntimeContext, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)
	faWithInvalidTargetType            = fixFormationAssignmentModelWithParameters(testAssignmentID, testFormationID, testTenantID, testSourceID, testTargetID, model.FormationAssignmentTypeApplication, invalidTargetType, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)

	destConfig = fixDestinationConfig()

	basicDestDetails          = fixBasicDestinationDetails()
	samlAssertionDestsDetails = fixSAMLAssertionDestinationsDetails()

	basicAuthCreds         = fixBasicAuthCreds(basicDestURL, basicDestUser, basicDestPassword)
	samlAssertionAuthCreds = fixSAMLAssertionAuthCreds(basicDestURL)

	createResp                   = fixHTTPResponse(http.StatusCreated, "")
	createRespWithConflict       = fixHTTPResponse(http.StatusConflict, "")
	deleteResp                   = fixHTTPResponse(http.StatusNoContent, "")
	respWithUnexpectedStatusCode = fixHTTPResponse(http.StatusOK, "test-body")

	correlationIDs = []string{"correlation-id-1", "correlation-id-2"}
)

func Test_CreateDesignTimeDestinations(t *testing.T) {
	designTimeDestDetails := fixDesignTimeDestinationDetails()

	designTimeDestDetailsWithoutName := fixDesignTimeDestinationDetails()
	designTimeDestDetailsWithoutName.Name = ""

	destConfigWithInvalidDestBaseURL := fixDestinationConfig()
	destConfigWithInvalidDestBaseURL.DestinationAPIConfig.BaseURL = ":wrong"

	testCases := []struct {
		name                string
		config              *destinationcreator.Config
		destinationDetails  operators.Destination
		formationAssignment *model.FormationAssignment
		httpClient          func() *automock.HttpClient
		labelRepoFn         func() *automock.LabelRepository
		tenantRepoFn        func() *automock.TenantRepository
		expectedErrMessage  string
	}{
		{
			name:                "Success when subaccount ID is provided in the destination details",
			config:              destConfig,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createResp, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
		},
		{
			name:                "Error while getting region and get external tenant fail",
			config:              destConfig,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting region label for tenant with ID: %s: while getting tenant by external ID: %q", destinationExternalSubaccountID, destinationExternalSubaccountID),
		},
		{
			name:                "Error while getting region and getting label by key fail",
			config:              destConfig,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(nil, testErr).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting region label for tenant with ID: %s: %s", destinationExternalSubaccountID, testErr.Error()),
		},
		{
			name:                "Error while getting region and label type is invalid",
			config:              destConfig,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLblWithInvalidType, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting region label for tenant with ID: %s: unexpected type of %q label, expect: string, got: %T", destinationExternalSubaccountID, destinationcreator.RegionLabelKey, invalidLblValue),
		},
		{
			name:                "Error while building url and region is empty",
			config:              destConfig,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(lblWithEmptyValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "The provided region and/or subaccount for the URL couldn't be empty",
		},
		{
			name:                "Error while building url and parse url fail",
			config:              destConfigWithInvalidDestBaseURL,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "missing protocol scheme",
		},
		{
			name:                "Error when validating design time destination request body",
			config:              destConfig,
			destinationDetails:  designTimeDestDetailsWithoutName,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "while validating no authentication destination request body",
		},
		{
			name:                "Error when executing remote design time destination request fail",
			config:              destConfig,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(nil, testErr).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: testErr.Error(),
		},
		{
			name:                "Error while executing remote design time destination request and the status code is not the expected one",
			config:              destConfig,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(respWithUnexpectedStatusCode, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("Failed to create entity with name: %q, status: %d, body: %s", designTimeDestDetails.Name, http.StatusOK, "test-body"),
		},
		{
			name:                "Success while executing remote design time destination request and the status code is conflict",
			config:              destConfig,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Once()
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Once()
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createResp, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Times(1)
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Times(3)
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Times(3)
				return tenantRepo
			},
		},
		{
			name:                "Error while executing remote design time destination request and maximum depth is reached",
			config:              destConfig,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Times(3)
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Twice()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Times(2)
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Times(5)
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Times(5)
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("Destination creator service retry limit: %d is exceeded", destinationcreator.DepthLimit),
		},
		{
			name:                "Error while executing remote design time destination request in case of conflict and delete destination fail",
			config:              destConfig,
			destinationDetails:  designTimeDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting tenant by external ID: %q: %s", destinationExternalSubaccountID, testErr.Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			httpClient := fixUnusedHTTPClient()
			if testCase.httpClient != nil {
				httpClient = testCase.httpClient()
			}

			labelRepo := fixUnusedLabelRepo()
			if testCase.labelRepoFn != nil {
				labelRepo = testCase.labelRepoFn()
			}

			tenantRepo := fixUnusedTenantRepo()
			if testCase.tenantRepoFn != nil {
				tenantRepo = testCase.tenantRepoFn()
			}
			defer mock.AssertExpectationsForObjects(t, httpClient, labelRepo, tenantRepo)

			svc := destinationcreator.NewService(httpClient, testCase.config, nil, nil, nil, labelRepo, tenantRepo)

			err := svc.CreateDesignTimeDestinations(emptyCtx, testCase.destinationDetails, testCase.formationAssignment, 0, false) //TODO:: propagate
			if testCase.expectedErrMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_CreateBasicDestinations(t *testing.T) {
	basicDestDetailsWithInvalidAuth := fixBasicDestinationDetails()
	basicDestDetailsWithInvalidAuth.Authentication = invalidDestAuthType

	testCases := []struct {
		name                string
		destinationDetails  operators.Destination
		formationAssignment *model.FormationAssignment
		httpClient          func() *automock.HttpClient
		labelRepoFn         func() *automock.LabelRepository
		tenantRepoFn        func() *automock.TenantRepository
		expectedErrMessage  string
	}{
		{
			name:                "Success",
			destinationDetails:  basicDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createResp, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
		},
		{
			name:                "Error while getting region and get external tenant fail",
			destinationDetails:  basicDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting region label for tenant with ID: %s: while getting tenant by external ID: %q", destinationExternalSubaccountID, destinationExternalSubaccountID),
		},
		{
			name:                "Error while building url and region is empty",
			destinationDetails:  basicDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(lblWithEmptyValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "The provided region and/or subaccount for the URL couldn't be empty",
		},
		{
			name:                "Error when preparing basic request body fail",
			destinationDetails:  basicDestDetailsWithInvalidAuth,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("The provided authentication type: %s in the destination details is invalid. It should be %s", invalidDestAuthType, destinationcreatorpkg.AuthTypeBasic),
		},
		{
			name:                "Error when executing remote basic destination request fail",
			destinationDetails:  basicDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(nil, testErr).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while creating inbound basic destination with name: %q in the destination service: %s", basicDestName, testErr.Error()),
		},
		{
			name:                "Success while executing remote basic destination request and the status code is conflict",
			destinationDetails:  basicDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Once()
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Once()
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createResp, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Times(1)
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Times(3)
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Times(3)
				return tenantRepo
			},
		},
		{
			name:                "Error while executing remote basic destination request and maximum depth is reached",
			destinationDetails:  basicDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Times(3)
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Twice()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Times(2)
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Times(5)
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Times(5)
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("Destination creator service retry limit: %d is exceeded", destinationcreator.DepthLimit),
		},
		{
			name:                "Error while executing remote basic destination request in case of conflict and delete destination fail",
			destinationDetails:  basicDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting tenant by external ID: %q: %s", destinationExternalSubaccountID, testErr.Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			httpClient := fixUnusedHTTPClient()
			if testCase.httpClient != nil {
				httpClient = testCase.httpClient()
			}

			labelRepo := fixUnusedLabelRepo()
			if testCase.labelRepoFn != nil {
				labelRepo = testCase.labelRepoFn()
			}

			tenantRepo := fixUnusedTenantRepo()
			if testCase.tenantRepoFn != nil {
				tenantRepo = testCase.tenantRepoFn()
			}
			defer mock.AssertExpectationsForObjects(t, httpClient, labelRepo, tenantRepo)

			svc := destinationcreator.NewService(httpClient, destConfig, nil, nil, nil, labelRepo, tenantRepo)

			err := svc.CreateBasicCredentialDestinations(emptyCtx, testCase.destinationDetails, basicAuthCreds, testCase.formationAssignment, correlationIDs, 0, false) //TODO:: propagate
			if testCase.expectedErrMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_CreateSAMLAssertionDestinations(t *testing.T) {
	samlAssertionDestDetails := fixSAMLAssertionDestinationDetails()

	samlAssertionDestDetailsWithoutName := fixSAMLAssertionDestinationDetails()
	samlAssertionDestDetailsWithoutName.Name = ""

	samlAssertionDestDetailsWithInvalidAuth := fixSAMLAssertionDestinationDetails()
	samlAssertionDestDetailsWithInvalidAuth.Authentication = invalidDestAuthType

	testCases := []struct {
		name                   string
		destinationDetails     operators.Destination
		samlAssertionAuthCreds operators.SAMLAssertionAuthentication
		formationAssignment    *model.FormationAssignment
		httpClient             func() *automock.HttpClient
		appRepoFn              func() *automock.ApplicationRepository
		labelRepoFn            func() *automock.LabelRepository
		tenantRepoFn           func() *automock.TenantRepository
		expectedErrMessage     string
	}{
		{
			name:                "Success",
			destinationDetails:  samlAssertionDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createResp, nil).Once()
				return client
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testSourceID).Return(testApp, nil).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
		},
		{
			name:                "Error while getting region and get external tenant fail",
			destinationDetails:  samlAssertionDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting region label for tenant with ID: %s: while getting tenant by external ID: %q", destinationExternalSubaccountID, destinationExternalSubaccountID),
		},
		{
			name:                "Error while building url and region is empty",
			destinationDetails:  samlAssertionDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(lblWithEmptyValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "while building destination URL: The provided region and/or subaccount for the URL couldn't be empty",
		},
		{
			name:                "Error when saml assertion authentication type is NOT correct",
			destinationDetails:  samlAssertionDestDetailsWithInvalidAuth,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("The provided authentication type: %s in the destination details is invalid. It should be %s", invalidDestAuthType, destinationcreatorpkg.AuthTypeSAMLAssertion),
		},
		{
			name:                "Error when getting application by ID fail",
			destinationDetails:  samlAssertionDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testSourceID).Return(nil, testErr).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting application with ID: %q: %s", testSourceID, testErr.Error()),
		},
		{
			name:                "Error when validating saml assertion request body",
			destinationDetails:  samlAssertionDestDetailsWithoutName,
			formationAssignment: faWithSourceAppAndTargetApp,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testSourceID).Return(testApp, nil).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "while validating SAML assertion destination request body",
		},
		{
			name:                "Error when executing remote saml assertion destination request fail",
			destinationDetails:  samlAssertionDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(nil, testErr).Once()
				return client
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testSourceID).Return(testApp, nil).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while creating SAML assertion destination with name: %q in the destination service: %s", samlAssertionDestName, testErr.Error()),
		},
		{
			name:                "Success while executing remote saml assertion destination request and the status code is conflict",
			destinationDetails:  samlAssertionDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Once()
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Once()
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createResp, nil).Once()
				return client
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testSourceID).Return(testApp, nil).Twice()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Times(1)
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Times(3)
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Times(3)
				return tenantRepo
			},
		},
		{
			name:                "Error while executing remote saml assertion destination request and maximum depth is reached",
			destinationDetails:  samlAssertionDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Times(3)
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Twice()
				return client
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testSourceID).Return(testApp, nil).Times(3)
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Times(2)
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Times(5)
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Times(5)
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("Destination creator service retry limit: %d is exceeded", destinationcreator.DepthLimit),
		},
		{
			name:                "Error while executing remote saml assertion destination request in case of conflict and delete destination fail",
			destinationDetails:  samlAssertionDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Once()
				return client
			},
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testSourceID).Return(testApp, nil).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting tenant by external ID: %q: %s", destinationExternalSubaccountID, testErr.Error()),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			httpClient := fixUnusedHTTPClient()
			if testCase.httpClient != nil {
				httpClient = testCase.httpClient()
			}

			appRepo := fixUnusedAppRepo()
			if testCase.appRepoFn != nil {
				appRepo = testCase.appRepoFn()
			}

			labelRepo := fixUnusedLabelRepo()
			if testCase.labelRepoFn != nil {
				labelRepo = testCase.labelRepoFn()
			}

			tenantRepo := fixUnusedTenantRepo()
			if testCase.tenantRepoFn != nil {
				tenantRepo = testCase.tenantRepoFn()
			}
			defer mock.AssertExpectationsForObjects(t, httpClient, appRepo, labelRepo, tenantRepo)

			svc := destinationcreator.NewService(httpClient, destConfig, appRepo, nil, nil, labelRepo, tenantRepo)

			err := svc.CreateSAMLAssertionDestination(emptyCtx, testCase.destinationDetails, samlAssertionAuthCreds, testCase.formationAssignment, correlationIDs, 0, false) //TODO:: propagate
			if testCase.expectedErrMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_CreateClientCertificateDestination(t *testing.T) {
	clientCertAuthDestDetails := fixClientCertAuthDestinationDetails()
	clientCertAuthTypeCreds := fixClientCertAuthTypeCreds()

	clientCertAuthDestDetailsWithoutName := fixClientCertAuthDestinationDetails()
	clientCertAuthDestDetailsWithoutName.Name = ""

	clientCertAuthDestDetailsWithInvalidAuth := fixClientCertAuthDestinationDetails()
	clientCertAuthDestDetailsWithInvalidAuth.Authentication = invalidDestAuthType

	testCases := []struct {
		name                   string
		destinationDetails     operators.Destination
		samlAssertionAuthCreds operators.SAMLAssertionAuthentication
		formationAssignment    *model.FormationAssignment
		httpClient             func() *automock.HttpClient
		labelRepoFn            func() *automock.LabelRepository
		tenantRepoFn           func() *automock.TenantRepository
		expectedErrMessage     string
	}{
		{
			name:                "Success",
			destinationDetails:  clientCertAuthDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createResp, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
		},
		{
			name:                "Error while getting region and get external tenant fail",
			destinationDetails:  clientCertAuthDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting region label for tenant with ID: %s: while getting tenant by external ID: %q", destinationExternalSubaccountID, destinationExternalSubaccountID),
		},
		{
			name:                "Error while building url and region is empty",
			destinationDetails:  clientCertAuthDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(lblWithEmptyValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "while building destination URL: The provided region and/or subaccount for the URL couldn't be empty",
		},
		{
			name:                "Error when saml assertion authentication type is NOT correct",
			destinationDetails:  clientCertAuthDestDetailsWithInvalidAuth,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("The provided authentication type: %s in the destination details is invalid. It should be %s", invalidDestAuthType, destinationcreatorpkg.AuthTypeClientCertificate),
		},
		{
			name:                "Error when validating saml assertion request body",
			destinationDetails:  clientCertAuthDestDetailsWithoutName,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "while validating client certificate destination request body",
		},
		{
			name:                "Error when executing remote saml assertion destination request fail",
			destinationDetails:  clientCertAuthDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(nil, testErr).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while creating client certificate authentication destination with name: %q in the destination service: %s", clientCertAuthDestName, testErr.Error()),
		},
		{
			name:                "Success while executing remote saml assertion destination request and the status code is conflict",
			destinationDetails:  clientCertAuthDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Once()
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Once()
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createResp, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Times(1)
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Times(3)
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Times(3)
				return tenantRepo
			},
		},
		{
			name:                "Error while executing remote saml assertion destination request and maximum depth is reached",
			destinationDetails:  clientCertAuthDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Times(3)
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Twice()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Times(2)
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Times(5)
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Times(5)
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("Destination creator service retry limit: %d is exceeded", destinationcreator.DepthLimit),
		},
		{
			name:                "Error while executing remote saml assertion destination request in case of conflict and delete destination fail",
			destinationDetails:  clientCertAuthDestDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while deleting destination with name: %q and subaccount ID: %q", clientCertAuthDestName, destinationExternalSubaccountID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			httpClient := fixUnusedHTTPClient()
			if testCase.httpClient != nil {
				httpClient = testCase.httpClient()
			}

			labelRepo := fixUnusedLabelRepo()
			if testCase.labelRepoFn != nil {
				labelRepo = testCase.labelRepoFn()
			}

			tenantRepo := fixUnusedTenantRepo()
			if testCase.tenantRepoFn != nil {
				tenantRepo = testCase.tenantRepoFn()
			}
			defer mock.AssertExpectationsForObjects(t, httpClient, labelRepo, tenantRepo)

			svc := destinationcreator.NewService(httpClient, destConfig, nil, nil, nil, labelRepo, tenantRepo)

			err := svc.CreateClientCertificateDestination(emptyCtx, testCase.destinationDetails, clientCertAuthTypeCreds, testCase.formationAssignment, correlationIDs, 0, false)//TODO:: propagate
			if testCase.expectedErrMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_DeleteDestination(t *testing.T) {
	testCases := []struct {
		name                    string
		destinationName         string
		destinationSubaccountID string
		destinationInstanceID   string
		formationAssignment     *model.FormationAssignment
		httpClient              func() *automock.HttpClient
		labelRepoFn             func() *automock.LabelRepository
		tenantRepoFn            func() *automock.TenantRepository
		expectedErrMessage      string
	}{
		{
			name:                    "Success",
			destinationName:         basicDestName,
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
		},
		{
			name:                "Error when validating destination subaccount fail",
			destinationName:     basicDestName,
			formationAssignment: faWithInvalidTargetType,
			expectedErrMessage:  fmt.Sprintf("Couldn't determine the label-able object type from assignment type: %q", invalidTargetType),
		},
		{
			name:                    "Error while getting region and get external tenant fail",
			destinationName:         basicDestName,
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting region label for tenant with ID: %s: while getting tenant by external ID: %q", destinationExternalSubaccountID, destinationExternalSubaccountID),
		},
		{
			name:                    "Error while building url and region is empty",
			destinationName:         basicDestName,
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(lblWithEmptyValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "while building destination URL: The provided region and/or subaccount for the URL couldn't be empty",
		},
		{
			name:                    "Error while building url and destination name is empty",
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("The entity name should not be empty in case of %s request", http.MethodDelete),
		},
		{
			name:                    "Error when executing remote delete destination request fail",
			destinationName:         basicDestName,
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(nil, testErr).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: testErr.Error(),
		},
		{
			name:                    "Error when executing remote delete destination request return unexpected status code",
			destinationName:         basicDestName,
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(respWithUnexpectedStatusCode, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("Failed to delete entity with name: %q from destination service, status: %d, body: %s", basicDestName, http.StatusOK, ""),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			httpClient := fixUnusedHTTPClient()
			if testCase.httpClient != nil {
				httpClient = testCase.httpClient()
			}

			labelRepo := fixUnusedLabelRepo()
			if testCase.labelRepoFn != nil {
				labelRepo = testCase.labelRepoFn()
			}

			tenantRepo := fixUnusedTenantRepo()
			if testCase.tenantRepoFn != nil {
				tenantRepo = testCase.tenantRepoFn()
			}
			defer mock.AssertExpectationsForObjects(t, httpClient, labelRepo, tenantRepo)

			svc := destinationcreator.NewService(httpClient, destConfig, nil, nil, nil, labelRepo, tenantRepo)

			err := svc.DeleteDestination(emptyCtx, testCase.destinationName, testCase.destinationSubaccountID, testCase.destinationInstanceID, testCase.formationAssignment, false) //TODO:: propagate
			if testCase.expectedErrMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_CreateCertificate(t *testing.T) {
	samlAssertionDestsDetailsWithoutName := fixSAMLAssertionDestinationsDetails()
	samlAssertionDestDetailsWithoutName := samlAssertionDestsDetailsWithoutName[0]
	samlAssertionDestDetailsWithoutName.Name = ""

	samlAssertionDestDetailsWithInvalidAuth := fixSAMLAssertionDestinationDetails()
	samlAssertionDestDetailsWithInvalidAuth.Authentication = invalidDestAuthType

	certName := destinationcreatorpkg.AuthTypeSAMLAssertion + "-" + testAssignmentID
	certResp := fixCertificateResponse(certificateFileNameValue, certificateCommonNameValue, certificateChainValue)
	certRespBytes, err := json.Marshal(certResp)
	require.NoError(t, err)

	certRespWithoutFileName := fixCertificateResponse("", certificateCommonNameValue, certificateChainValue)
	certRespWithoutFileNameBytes, err := json.Marshal(certRespWithoutFileName)
	require.NoError(t, err)

	certificateCreateRespWithoutName := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(bytes.NewBuffer(certRespWithoutFileNameBytes)),
	}

	invalidCertificateResp := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(strings.NewReader("{\"invalid")),
	}
	faWithSourceAppAndTargetAppAndLongID := fixFormationAssignmentModelWithParameters("some-long-id-0b2de27b-b500-47eb-bf7f-b0c105a65864", testFormationID, testTenantID, testSourceID, testTargetID, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), TestConfigValueRawJSON, TestEmptyErrorValueRawJSON)

	testCases := []struct {
		name                string
		destinationsDetails []operators.Destination
		destinationAuthType destinationcreatorpkg.AuthType
		formationAssignment *model.FormationAssignment
		httpClient          func() *automock.HttpClient
		labelRepoFn         func() *automock.LabelRepository
		tenantRepoFn        func() *automock.TenantRepository
		expectedResult      *operators.CertificateData
		expectedErrMessage  string
	}{
		{
			name:                "Success",
			destinationsDetails: samlAssertionDestsDetails,
			destinationAuthType: destinationcreatorpkg.AuthTypeSAMLAssertion,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(fixHTTPResponse(http.StatusCreated, string(certRespBytes)), nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedResult: &operators.CertificateData{
				FileName:         certificateFileNameValue,
				CommonName:       certificateCommonNameValue,
				CertificateChain: certificateChainValue,
			},
		},
		{
			name:                "Success when destination subaccount ID is NOT provided",
			destinationsDetails: fixDestinationsDetailsWithoutSubaccountID(),
			destinationAuthType: destinationcreatorpkg.AuthTypeSAMLAssertion,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(fixHTTPResponse(http.StatusCreated, string(certRespBytes)), nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedResult: &operators.CertificateData{
				FileName:         certificateFileNameValue,
				CommonName:       certificateCommonNameValue,
				CertificateChain: certificateChainValue,
			},
		},
		{
			name:                "Error when destination subaccount ID is NOT provided and validating destination subaccount fail",
			destinationsDetails: fixDestinationsDetailsWithoutSubaccountID(),
			formationAssignment: faWithInvalidTargetType,
			expectedErrMessage:  "while sanitizing destination subaccount IDs: An error occurred while getting consumer subaccount:",
		},
		{
			name:                "Error when the subaccount IDs in the destination details are different",
			destinationsDetails: fixDestinationsDetailsWitDifferentSubaccountIDs(),
			formationAssignment: faWithInvalidTargetType,
			expectedErrMessage:  "All subaccount in the destinations details should have one and the same subaccount ID. Currently there is/are: 2 subaacount ID(s)",
		},
		{
			name:                "Error when validating destination subaccount fail",
			destinationsDetails: fixDestinationsDetailsWithoutSubaccountID(),
			formationAssignment: faWithInvalidTargetType,
			expectedErrMessage:  fmt.Sprintf("Couldn't determine the label-able object type from assignment type: %q", invalidTargetType),
		},
		{
			name:                "Error while getting region and get external tenant fail",
			destinationsDetails: samlAssertionDestsDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting region label for tenant with ID: %s: while getting tenant by external ID: %q", destinationExternalSubaccountID, destinationExternalSubaccountID),
		},
		{
			name:                "Error while building certificate url and region is empty",
			destinationsDetails: samlAssertionDestsDetails,
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(lblWithEmptyValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "while building certificate URL: The provided region and/or subaccount for the URL couldn't be empty",
		},
		{
			name:                "Error when getting destination certificate name and destination auth type is invalid",
			destinationsDetails: samlAssertionDestsDetailsWithoutName,
			destinationAuthType: "",
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "Invalid destination authentication type: \"\" for certificate creation",
		},
		{
			name:                "Success when the certificate name length is more than accepted and it will be truncated",
			destinationsDetails: samlAssertionDestsDetailsWithoutName,
			destinationAuthType: destinationcreatorpkg.AuthTypeClientCertificate,
			formationAssignment: faWithSourceAppAndTargetAppAndLongID,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(fixHTTPResponse(http.StatusCreated, string(certRespBytes)), nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedResult: &operators.CertificateData{
				FileName:         certificateFileNameValue,
				CommonName:       certificateCommonNameValue,
				CertificateChain: certificateChainValue,
			},
		},
		{
			name:                "Error when executing remote create certificate request fail",
			destinationsDetails: samlAssertionDestsDetails,
			destinationAuthType: destinationcreatorpkg.AuthTypeSAMLAssertion,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(nil, testErr).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while creating certificate with name: %q for subaccount with ID: %q in the destination service: %s", certName, destinationExternalSubaccountID, testErr.Error()),
		},
		{
			name:                "Error while executing remote certificate request and maximum depth is reached",
			destinationsDetails: samlAssertionDestsDetails,
			destinationAuthType: destinationcreatorpkg.AuthTypeSAMLAssertion,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Times(3)
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Twice()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Times(2)
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Times(5)
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Times(5)
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("Destination creator service retry limit: %d is exceeded", destinationcreator.DepthLimit),
		},
		{
			name:                "Error while executing remote certificate request in case of conflict and delete destination fail",
			destinationsDetails: samlAssertionDestsDetails,
			destinationAuthType: destinationcreatorpkg.AuthTypeSAMLAssertion,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(createRespWithConflict, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting tenant by external ID: %q: %s", destinationExternalSubaccountID, testErr.Error()),
		},
		{
			name:                "Error when unmarshalling certificate response fail",
			destinationsDetails: samlAssertionDestsDetails,
			destinationAuthType: destinationcreatorpkg.AuthTypeSAMLAssertion,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(invalidCertificateResp, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "unexpected end of JSON input",
		},
		{
			name:                "Error when certificate response is not valid",
			destinationsDetails: samlAssertionDestsDetails,
			destinationAuthType: destinationcreatorpkg.AuthTypeSAMLAssertion,
			formationAssignment: faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodPost)).Return(certificateCreateRespWithoutName, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "while validation destination certificate data",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			httpClient := fixUnusedHTTPClient()
			if testCase.httpClient != nil {
				httpClient = testCase.httpClient()
			}

			labelRepo := fixUnusedLabelRepo()
			if testCase.labelRepoFn != nil {
				labelRepo = testCase.labelRepoFn()
			}

			tenantRepo := fixUnusedTenantRepo()
			if testCase.tenantRepoFn != nil {
				tenantRepo = testCase.tenantRepoFn()
			}
			defer mock.AssertExpectationsForObjects(t, httpClient, labelRepo, tenantRepo)

			svc := destinationcreator.NewService(httpClient, destConfig, nil, nil, nil, labelRepo, tenantRepo)

			result, err := svc.CreateCertificate(emptyCtx, testCase.destinationsDetails, testCase.destinationAuthType, testCase.formationAssignment, 0, false) //TODO:: propagate
			if testCase.expectedErrMessage != "" {
				require.Empty(t, result)
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NotEmpty(t, result)
				require.NoError(t, err)
				require.Equal(t, testCase.expectedResult, result)
			}
		})
	}
}

func Test_DeleteCertificate(t *testing.T) {
	testCases := []struct {
		name                    string
		certificateName         string
		destinationSubaccountID string
		destinationInstanceID   string
		formationAssignment     *model.FormationAssignment
		httpClient              func() *automock.HttpClient
		labelRepoFn             func() *automock.LabelRepository
		tenantRepoFn            func() *automock.TenantRepository
		expectedErrMessage      string
	}{
		{
			name:                    "Success",
			certificateName:         certificateName,
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(deleteResp, nil).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
		},
		{
			name:                "Error when validating destination subaccount fail",
			certificateName:     certificateName,
			formationAssignment: faWithInvalidTargetType,
			expectedErrMessage:  fmt.Sprintf("Couldn't determine the label-able object type from assignment type: %q", invalidTargetType),
		},
		{
			name:                    "Error while getting region and get external tenant fail",
			certificateName:         certificateName,
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting region label for tenant with ID: %s: while getting tenant by external ID: %q", destinationExternalSubaccountID, destinationExternalSubaccountID),
		},
		{
			name:                    "Error while building url and region is empty",
			certificateName:         certificateName,
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(lblWithEmptyValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: "while building certificate URL: The provided region and/or subaccount for the URL couldn't be empty",
		},
		{
			name:                    "Error while building url and certificate name is empty",
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("The entity name should not be empty in case of %s request", http.MethodDelete),
		},
		{
			name:                    "Error when executing remote delete certificate request fail",
			certificateName:         certificateName,
			destinationSubaccountID: destinationExternalSubaccountID,
			destinationInstanceID:   destinationInstanceID,
			formationAssignment:     faWithSourceAppAndTargetApp,
			httpClient: func() *automock.HttpClient {
				client := &automock.HttpClient{}
				client.On("Do", requestThatHasMethod(http.MethodDelete)).Return(nil, testErr).Once()
				return client
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				labelRepo.On("GetByKey", emptyCtx, destinationInternalSubaccountID, model.TenantLabelableObject, destinationExternalSubaccountID, destinationcreator.RegionLabelKey).Return(regionLbl, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			httpClient := fixUnusedHTTPClient()
			if testCase.httpClient != nil {
				httpClient = testCase.httpClient()
			}

			labelRepo := fixUnusedLabelRepo()
			if testCase.labelRepoFn != nil {
				labelRepo = testCase.labelRepoFn()
			}

			tenantRepo := fixUnusedTenantRepo()
			if testCase.tenantRepoFn != nil {
				tenantRepo = testCase.tenantRepoFn()
			}
			defer mock.AssertExpectationsForObjects(t, httpClient, labelRepo, tenantRepo)

			svc := destinationcreator.NewService(httpClient, destConfig, nil, nil, nil, labelRepo, tenantRepo)

			err := svc.DeleteCertificate(emptyCtx, testCase.certificateName, testCase.destinationSubaccountID, testCase.destinationInstanceID, testCase.formationAssignment, false) //TODO:: propagate
			if testCase.expectedErrMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_EnrichAssignmentConfigWithCertificateData(t *testing.T) {
	certData := &operators.CertificateData{
		FileName:         "certFileNameValue",
		CommonName:       "certCommonNameValue",
		CertificateChain: "certChainValue",
	}

	testCases := []struct {
		name               string
		assignmentConfig   json.RawMessage
		destinationConfig  *destinationcreator.Config
		certData           *operators.CertificateData
		expectedResult     json.RawMessage
		expectedErrMessage string
	}{
		{
			name:              "Success",
			assignmentConfig:  json.RawMessage(""),
			destinationConfig: destConfig,
			certData:          certData,
			expectedResult:    json.RawMessage("{\"credentials\":{\"inboundCommunication\":{\"clientCertificateAuthentication\":{\"certificate\":\"certChainValue\"}}}}"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			svc := destinationcreator.NewService(nil, testCase.destinationConfig, nil, nil, nil, nil, nil)

			result, err := svc.EnrichAssignmentConfigWithCertificateData(testCase.assignmentConfig, destinationcreatorpkg.ClientCertAuthDestPath, testCase.certData)
			if testCase.expectedErrMessage != "" {
				require.Empty(t, result)
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NotEmpty(t, result)
				require.NoError(t, err)
				require.Equal(t, testCase.expectedResult, result)
			}
		})
	}
}

func Test_EnrichAssignmentConfigWithSAMLCertificateData(t *testing.T) {
	certData := &operators.CertificateData{
		FileName:         "certFileNameValue",
		CommonName:       "certCommonNameValue",
		CertificateChain: "certChainValue",
	}

	testCases := []struct {
		name               string
		assignmentConfig   json.RawMessage
		destinationConfig  *destinationcreator.Config
		certData           *operators.CertificateData
		expectedResult     json.RawMessage
		expectedErrMessage string
	}{
		{
			name:              "Success",
			assignmentConfig:  json.RawMessage(""),
			destinationConfig: destConfig,
			certData:          certData,
			expectedResult:    json.RawMessage("{\"credentials\":{\"inboundCommunication\":{\"samlAssertion\":{\"certificate\":\"certChainValue\",\"assertionIssuer\":\"certCommonNameValue\"}}}}"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			svc := destinationcreator.NewService(nil, testCase.destinationConfig, nil, nil, nil, nil, nil)

			result, err := svc.EnrichAssignmentConfigWithSAMLCertificateData(testCase.assignmentConfig, destinationcreatorpkg.SAMLAssertionDestPath, testCase.certData)
			if testCase.expectedErrMessage != "" {
				require.Empty(t, result)
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NotEmpty(t, result)
				require.NoError(t, err)
				require.Equal(t, testCase.expectedResult, result)
			}
		})
	}
}

func Test_DetermineDestinationSubaccount(t *testing.T) {
	testCases := []struct {
		name                     string
		externalDestSubaccountID string
		formationAssignment      *model.FormationAssignment
		appRepoFn                func() *automock.ApplicationRepository
		runtimeRepoFn            func() *automock.RuntimeRepository
		runtimeCtxRepoFn         func() *automock.RuntimeCtxRepository
		labelRepoFn              func() *automock.LabelRepository
		tenantRepoFn             func() *automock.TenantRepository
		skipValidation           bool
		expectedErrMessage       string
	}{
		// unit tests WITHOUT provided subaccount ID
		{
			name:                "Success when subaccount ID is NOT provided",
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				return labelRepo
			},
		},
		{
			name:                "Error when determining label-able object type fail",
			formationAssignment: faWithInvalidTargetType,
			expectedErrMessage:  fmt.Sprintf("Couldn't determine the label-able object type from assignment type: %q", invalidTargetType),
		},
		// unit tests WITH provided subaccount ID and formation assignment target type is application
		{
			name:                     "Success when subaccount ID is provided and it is consumer",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				return labelRepo
			},
		},
		{
			name:                     "Success when subaccount ID is NOT consumer and the FA target type is app",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetApp,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(testApp, nil).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				labelRepo.On("ListForGlobalObject", emptyCtx, model.AppTemplateLabelableObject, appTemplateID).Return(subaccountnLbl, nil).Once()
				return labelRepo
			},
		},
		{
			name:                     "Success when subaccount ID is NOT consumer, the FA target type is app and global_subaccount_id is not the expected one, but the validation is skipped",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetApp,
			skipValidation:           true,
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is app and getting app fail",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetApp,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(nil, testErr).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				return labelRepo
			},
			expectedErrMessage: testErr.Error(),
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is app and app template ID is missing",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetApp,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(testAppWithoutTmplID, nil).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				return labelRepo
			},
			expectedErrMessage: fmt.Sprintf("The application template ID for application ID: %q should not be empty", appID),
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is app and listing labels fail",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetApp,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(testApp, nil).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				labelRepo.On("ListForGlobalObject", emptyCtx, model.AppTemplateLabelableObject, appTemplateID).Return(nil, testErr).Once()
				return labelRepo
			},
			expectedErrMessage: fmt.Sprintf("while getting labels for application template with ID: %q", appTemplateID),
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is app and global_subaccount_id label is missing",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetApp,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(testApp, nil).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				labelRepo.On("ListForGlobalObject", emptyCtx, model.AppTemplateLabelableObject, appTemplateID).Return(emptyLblMap, nil).Once()
				return labelRepo
			},
			expectedErrMessage: fmt.Sprintf("%q label should exist as part of the provider application template with ID: %q", destinationcreator.GlobalSubaccountLabelKey, appTemplateID),
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is app and global_subaccount_id label has invalid type",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetApp,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(testApp, nil).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				labelRepo.On("ListForGlobalObject", emptyCtx, model.AppTemplateLabelableObject, appTemplateID).Return(subaccountnLblWithInvalidType, nil).Once()
				return labelRepo
			},
			expectedErrMessage: fmt.Sprintf("unexpected type of %q label, expect: string, got: %T", destinationcreator.GlobalSubaccountLabelKey, invalidLblValue),
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is app and global_subaccount_id is not the expected one",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetApp,
			appRepoFn: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(testApp, nil).Once()
				return appRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				labelRepo.On("ListForGlobalObject", emptyCtx, model.AppTemplateLabelableObject, appTemplateID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				return labelRepo
			},
			expectedErrMessage: fmt.Sprintf("The provided destination subaccount is different from the owner subaccount of the application template with ID: %q", appTemplateID),
		},
		// unit tests WITH provided subaccount ID and formation assignment target type is runtime
		{
			name:                     "Success when subaccount ID is NOT consumer and the FA target type is runtime",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetRuntime,
			runtimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExists", emptyCtx, destinationInternalSubaccountID, testTargetID).Return(true, nil).Once()
				return runtimeRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is runtime and getting tenant fail",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetRuntime,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(nil, testErr).Once()
				return tenantRepo
			},
			expectedErrMessage: testErr.Error(),
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is runtime and tenant type is not valid",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetRuntime,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(accountTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("The provided destination external tenant ID: %q has invalid type, expected: %q, got: %q", destinationExternalSubaccountID, tenant.Subaccount, tenant.Account),
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is runtime and owner exists check fail",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetRuntime,
			runtimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExists", emptyCtx, destinationInternalSubaccountID, testTargetID).Return(false, testErr).Once()
				return runtimeRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: testErr.Error(),
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is runtime and runtime is not owner",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetRuntime,
			runtimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExists", emptyCtx, destinationInternalSubaccountID, testTargetID).Return(false, nil).Once()
				return runtimeRepo
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
			expectedErrMessage: fmt.Sprintf("The provided destination external subaccount: %q is not provider of the runtime with ID: %q", destinationExternalSubaccountID, testTargetID),
		},
		// unit tests WITH provided subaccount ID and formation assignment target type is runtime context
		{
			name:                     "Success when subaccount ID is NOT consumer and the FA target type is runtime context",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetRuntimeCtx,
			runtimeRepoFn: func() *automock.RuntimeRepository {
				runtimeRepo := &automock.RuntimeRepository{}
				runtimeRepo.On("OwnerExists", emptyCtx, destinationInternalSubaccountID, runtimeID).Return(true, nil).Once()
				return runtimeRepo
			},
			runtimeCtxRepoFn: func() *automock.RuntimeCtxRepository {
				runtimeCtxMock := &automock.RuntimeCtxRepository{}
				runtimeCtxMock.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(runtimeCtx, nil).Once()
				return runtimeCtxMock
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeContextLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				return labelRepo
			},
			tenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", emptyCtx, destinationExternalSubaccountID).Return(subaccTenant, nil).Once()
				return tenantRepo
			},
		},
		{
			name:                     "Error when subaccount ID is NOT consumer, the FA target type is runtime context and getting runtime ctx fail",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithSourceAppAndTargetRuntimeCtx,
			runtimeCtxRepoFn: func() *automock.RuntimeCtxRepository {
				runtimeCtxMock := &automock.RuntimeCtxRepository{}
				runtimeCtxMock.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(nil, testErr).Once()
				return runtimeCtxMock
			},
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.RuntimeContextLabelableObject, testTargetID).Return(subaccountnLblWithInvalidIDValue, nil).Once()
				return labelRepo
			},
			expectedErrMessage: testErr.Error(),
		},
		{
			name:                     "Error when formation assignment target type is invalid",
			externalDestSubaccountID: destinationExternalSubaccountID,
			formationAssignment:      faWithInvalidTargetType,
			expectedErrMessage:       fmt.Sprintf("Unknown formation assignment type: %q", invalidTargetType),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			appRepo := fixUnusedAppRepo()
			if testCase.appRepoFn != nil {
				appRepo = testCase.appRepoFn()
			}

			runtimeRepo := fixUnusedRuntimeRepo()
			if testCase.runtimeRepoFn != nil {
				runtimeRepo = testCase.runtimeRepoFn()
			}

			runtimeCtxRepo := fixUnusedRuntimeCtxRepo()
			if testCase.runtimeCtxRepoFn != nil {
				runtimeCtxRepo = testCase.runtimeCtxRepoFn()
			}

			labelRepo := fixUnusedLabelRepo()
			if testCase.labelRepoFn != nil {
				labelRepo = testCase.labelRepoFn()
			}

			tenantRepo := fixUnusedTenantRepo()
			if testCase.tenantRepoFn != nil {
				tenantRepo = testCase.tenantRepoFn()
			}

			defer mock.AssertExpectationsForObjects(t, appRepo, runtimeRepo, runtimeCtxRepo, labelRepo, tenantRepo)

			svc := destinationcreator.NewService(nil, nil, appRepo, runtimeRepo, runtimeCtxRepo, labelRepo, tenantRepo)

			result, err := svc.DetermineDestinationSubaccount(emptyCtx, testCase.externalDestSubaccountID, testCase.formationAssignment, testCase.skipValidation)
			if testCase.expectedErrMessage != "" {
				require.Empty(t, result)
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NotEmpty(t, result)
				require.NoError(t, err)
				require.Equal(t, destinationExternalSubaccountID, result)
			}
		})
	}
}

func Test_PrepareBasicRequestBody(t *testing.T) {
	basicCredsWithoutURL := fixBasicAuthCreds("", basicDestUser, basicDestPassword)

	basicDestDetailsWithoutURL := fixBasicDestinationDetails()
	basicDestDetailsWithoutURL.URL = ""

	basicDestDetailsWithInvalidAuth := fixBasicDestinationDetails()
	basicDestDetailsWithInvalidAuth.Authentication = invalidDestAuthType

	basicDestDetailsWithoutName := fixBasicDestinationDetails()
	basicDestDetailsWithoutName.Name = ""

	testCases := []struct {
		name                 string
		appRepo              func() *automock.ApplicationRepository
		destinationDetails   operators.Destination
		basicAuthCreds       operators.BasicAuthentication
		formationAssignment  *model.FormationAssignment
		expectedErrMessage   string
		expectedBasicReqBody *destinationcreator.BasicAuthDestinationRequestBody
	}{
		{
			name:                 "Success",
			formationAssignment:  faWithSourceAppAndTargetApp,
			destinationDetails:   basicDestDetails,
			basicAuthCreds:       basicAuthCreds,
			expectedBasicReqBody: fixBasicRequestBody(destinationURL),
		},
		{
			name:                 "Success when the url is missing in the destination details and it's in the basic creds",
			formationAssignment:  faWithSourceAppAndTargetApp,
			destinationDetails:   basicDestDetailsWithoutURL,
			basicAuthCreds:       basicAuthCreds,
			expectedBasicReqBody: fixBasicRequestBody(basicDestURL),
		},
		{
			name: "Success when the url is missing in both destination details and basic creds and it's retrieve from app base url",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(testApp, nil).Once()
				return appRepo
			},
			formationAssignment:  faWithSourceAppAndTargetApp,
			destinationDetails:   basicDestDetailsWithoutURL,
			basicAuthCreds:       basicCredsWithoutURL,
			expectedBasicReqBody: fixBasicRequestBody(appBaseURL),
		},
		{
			name: "Error when the url is missing in both destination details and basic creds and getting app fail",
			appRepo: func() *automock.ApplicationRepository {
				appRepo := &automock.ApplicationRepository{}
				appRepo.On("GetByID", emptyCtx, testTenantID, testTargetID).Return(nil, testErr).Once()
				return appRepo
			},
			formationAssignment: faWithSourceAppAndTargetApp,
			destinationDetails:  basicDestDetailsWithoutURL,
			basicAuthCreds:      basicCredsWithoutURL,
			expectedErrMessage:  testErr.Error(),
		},
		{
			name:                "Error when the the destination authentication type is invalid",
			formationAssignment: faWithSourceAppAndTargetApp,
			destinationDetails:  basicDestDetailsWithInvalidAuth,
			basicAuthCreds:      basicAuthCreds,
			expectedErrMessage:  fmt.Sprintf("The provided authentication type: %s in the destination details is invalid. It should be %s", invalidDestAuthType, destinationcreatorpkg.AuthTypeBasic),
		},
		{
			name:                "Error when the the destination request body is invalid",
			formationAssignment: faWithSourceAppAndTargetApp,
			destinationDetails:  basicDestDetailsWithoutName,
			basicAuthCreds:      basicAuthCreds,
			expectedErrMessage:  "while validating basic destination request body",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			appRepo := fixUnusedAppRepo()
			if testCase.appRepo != nil {
				appRepo = testCase.appRepo()
			}

			svc := destinationcreator.NewService(nil, destConfig, appRepo, nil, nil, nil, nil)

			basicReqBody, err := svc.PrepareBasicRequestBody(emptyCtx, testCase.destinationDetails, testCase.basicAuthCreds, testCase.formationAssignment, correlationIDs)
			if testCase.expectedErrMessage != "" {
				require.Empty(t, basicReqBody)
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, basicReqBody)
				require.Equal(t, testCase.expectedBasicReqBody, basicReqBody)
			}
		})
	}
}

func Test_GetConsumerTenant(t *testing.T) {
	testCases := []struct {
		name                string
		formationAssignment *model.FormationAssignment
		labelRepoFn         func() *automock.LabelRepository
		expectedErrMessage  string
	}{
		{
			name:                "Success",
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLbl, nil).Once()
				return labelRepo
			},
		},
		{
			name:                "Error when determining label-able object type fail",
			formationAssignment: faWithInvalidTargetType,
			expectedErrMessage:  fmt.Sprintf("Couldn't determine the label-able object type from assignment type: %q", invalidTargetType),
		},
		{
			name:                "Error when listing labels fail",
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(nil, testErr).Once()
				return labelRepo
			},
			expectedErrMessage: testErr.Error(),
		},
		{
			name:                "Error when global_subaccount_id label is missing",
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(emptyLblMap, nil).Once()
				return labelRepo
			},
			expectedErrMessage: fmt.Sprintf("%q label does not exists for: %q with ID: %q", destinationcreator.GlobalSubaccountLabelKey, model.FormationAssignmentTypeApplication, testTargetID),
		},
		{
			name:                "Error when global_subaccount_id label has invalid type",
			formationAssignment: faWithSourceAppAndTargetApp,
			labelRepoFn: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", emptyCtx, testTenantID, model.ApplicationLabelableObject, testTargetID).Return(subaccountnLblWithInvalidType, nil).Once()
				return labelRepo
			},
			expectedErrMessage: fmt.Sprintf("unexpected type of %q label, expect: string, got: %T", destinationcreator.GlobalSubaccountLabelKey, 0),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			labelRepo := fixUnusedLabelRepo()
			if testCase.labelRepoFn != nil {
				labelRepo = testCase.labelRepoFn()
			}
			defer mock.AssertExpectationsForObjects(t, labelRepo)

			svc := destinationcreator.NewService(nil, nil, nil, nil, nil, labelRepo, nil)

			result, err := svc.GetConsumerTenant(emptyCtx, testCase.formationAssignment)
			if testCase.expectedErrMessage != "" {
				require.Empty(t, result)
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMessage)
			} else {
				require.NotEmpty(t, result)
				require.NoError(t, err)
				require.Equal(t, destinationExternalSubaccountID, result)
			}
		})
	}
}
