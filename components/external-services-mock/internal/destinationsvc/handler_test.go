package destinationsvc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"

	"github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/destinationsvc"
	destcreatorpkg "github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"
	"github.com/stretchr/testify/require"
)

var (
	testSecretKey                                              = []byte("testSecretKey")
	noAuthDestIdentifierWithSubaccountID                       = fmt.Sprintf(destinationsvc.UniqueEntityNameIdentifier, noAuthDestName, testSubaccountID, "")
	samlAssertionDestIdentifierWithSubaccountIDAndInstanceID   = fmt.Sprintf(destinationsvc.UniqueEntityNameIdentifier, samlAssertionDestName, testSubaccountID, testServiceInstanceID)
	destinationCertIdentifierWithSubaccountID                  = fmt.Sprintf(destinationsvc.UniqueEntityNameIdentifier, testDestinationCertName, testSubaccountID, "")
	destinationCertIdentifierWithSubaccountIDAndInstanceID     = fmt.Sprintf(destinationsvc.UniqueEntityNameIdentifier, testDestinationCertName, testSubaccountID, testServiceInstanceID)
	samlDestinationCertIdentifierWithSubaccountIDAndInstanceID = fmt.Sprintf(destinationsvc.UniqueEntityNameIdentifier, testDestKeyStoreLocation, testSubaccountID, testServiceInstanceID)
)

func TestHandler_CreateDestinations(t *testing.T) {
	destinationCreatorPath := fmt.Sprintf("/regions/%s/subaccounts/%s/destinations", testRegion, testSubaccountID)
	basicDestIdentifierWithSubaccountID := fmt.Sprintf(destinationsvc.UniqueEntityNameIdentifier, basicAuthDestName, testSubaccountID, "")
	samlAssertionDestIdentifierWithSubaccountID := fmt.Sprintf(destinationsvc.UniqueEntityNameIdentifier, samlAssertionDestName, testSubaccountID, "")

	testCases := []struct {
		Name                               string
		RequestBody                        string
		ExpectedResponseCode               int
		ExpectedDestinationSvcDestinations map[string]destcreatorpkg.Destination
		Region                             string
		SubaccountID                       string
		ExistingDestination                map[string]destcreatorpkg.Destination
		MissingContentTypeHeader           bool
		MissingClientUserHeader            bool
	}{
		// Common unit tests
		{
			Name:                 "Error when authentication type is unknown",
			RequestBody:          `{"authenticationType":"unknown"}`,
			ExpectedResponseCode: http.StatusInternalServerError,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
		},
		{
			Name:                     "Error when request content type header is invalid",
			RequestBody:              destinationServiceBasicAuthReqBody,
			ExpectedResponseCode:     http.StatusUnsupportedMediaType,
			Region:                   testRegion,
			SubaccountID:             testSubaccountID,
			MissingContentTypeHeader: true,
		},
		{
			Name:                    "Error when client_user header is missing",
			RequestBody:             destinationServiceBasicAuthReqBody,
			ExpectedResponseCode:    http.StatusBadRequest,
			Region:                  testRegion,
			SubaccountID:            testSubaccountID,
			MissingClientUserHeader: true,
		},
		{
			Name:                 "Error when request path params are missing",
			RequestBody:          destinationServiceBasicAuthReqBody,
			ExpectedResponseCode: http.StatusBadRequest,
		},
		{
			Name:                 "Error when authenticationType is missing from request body",
			RequestBody:          destinationCreatorReqBodyWithoutAuthType,
			ExpectedResponseCode: http.StatusBadRequest,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
		},
		// No Authentication Destinations unit tests
		{
			Name:                               "Success when creating no auth destinations",
			RequestBody:                        destinationCreatorNoAuthDestReqBody,
			ExpectedResponseCode:               http.StatusCreated,
			Region:                             testRegion,
			SubaccountID:                       testSubaccountID,
			ExpectedDestinationSvcDestinations: fixDestinationMappings(noAuthDestIdentifierWithSubaccountID, fixNoAuthDestination(noAuthDestName)),
		},
		{
			Name:                 "Error when creating no auth destinations and the unmarshalling of the req body fails",
			RequestBody:          `{"authenticationType":"NoAuthentication", "invalid": }`,
			ExpectedResponseCode: http.StatusInternalServerError,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
		},
		{
			Name:                 "Error when creating no auth destinations and the validation of the req body fails",
			RequestBody:          `{"authenticationType":"NoAuthentication"}`,
			ExpectedResponseCode: http.StatusBadRequest,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
		},
		{
			Name:                 "Error when creating no auth destinations and there is already such existing destination",
			RequestBody:          destinationCreatorNoAuthDestReqBody,
			ExpectedResponseCode: http.StatusConflict,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
			ExistingDestination:  fixDestinationMappings(noAuthDestIdentifierWithSubaccountID, fixNoAuthDestination(noAuthDestName)),
		},
		//Basic Destinations unit tests
		{
			Name:                               "Success when creating basic destinations",
			RequestBody:                        destinationCreatorBasicAuthDestReqBody,
			ExpectedResponseCode:               http.StatusCreated,
			Region:                             testRegion,
			SubaccountID:                       testSubaccountID,
			ExpectedDestinationSvcDestinations: fixDestinationMappings(basicDestIdentifierWithSubaccountID, fixBasicDestination(basicAuthDestName)),
		},
		{
			Name:                 "Error when creating basic auth destinations and the unmarshalling of the req body fails",
			RequestBody:          `{"authenticationType":"BasicAuthentication", "invalid": }`,
			ExpectedResponseCode: http.StatusInternalServerError,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
		},
		{
			Name:                 "Error when creating basic auth destinations and the validation of the req body fails",
			RequestBody:          `{"authenticationType":"BasicAuthentication"}`,
			ExpectedResponseCode: http.StatusBadRequest,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
		},
		{
			Name:                 "Error when creating basic auth destinations and there is already such existing destination",
			RequestBody:          destinationCreatorBasicAuthDestReqBody,
			ExpectedResponseCode: http.StatusConflict,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
			ExistingDestination:  fixDestinationMappings(basicDestIdentifierWithSubaccountID, fixBasicDestination(basicAuthDestName)),
		},
		// SAML Assertion Destinations unit tests
		{
			Name:                               "Success when creating SAML Assertion destinations",
			RequestBody:                        destinationCreatorSAMLAssertionDestReqBody,
			ExpectedResponseCode:               http.StatusCreated,
			Region:                             testRegion,
			SubaccountID:                       testSubaccountID,
			ExpectedDestinationSvcDestinations: fixDestinationMappings(samlAssertionDestIdentifierWithSubaccountID, fixSAMLAssertionDestination(samlAssertionDestName)),
		},
		{
			Name:                 "Error when creating SAML Assertion destinations and the unmarshalling of the req body fails",
			RequestBody:          `{"authenticationType":"SAMLAssertion", "invalid": }`,
			ExpectedResponseCode: http.StatusInternalServerError,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
		},
		{
			Name:                 "Error when creating SAML Assertion destinations and the validation of the req body fails",
			RequestBody:          `{"authenticationType":"SAMLAssertion"}`,
			ExpectedResponseCode: http.StatusBadRequest,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
		},
		{
			Name:                 "Error when creating SAML Assertion destinations and there is already such existing destination",
			RequestBody:          destinationCreatorSAMLAssertionDestReqBody,
			ExpectedResponseCode: http.StatusConflict,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
			ExistingDestination:  fixDestinationMappings(samlAssertionDestIdentifierWithSubaccountID, fixSAMLAssertionDestination(samlAssertionDestName)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodPost, url+destinationCreatorPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)

			if !testCase.MissingContentTypeHeader {
				req.Header.Add("Content-Type", "application/json;charset=UTF-8")
			}

			if !testCase.MissingClientUserHeader {
				req.Header.Add("CLIENT_USER", "test")
			}

			urlVars := make(map[string]string)
			if testCase.Region != "" {
				urlVars[regionParamValue] = testCase.Region
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.SubaccountID != "" {
				urlVars[subaccountIDParamValue] = testCase.SubaccountID
				req = mux.SetURLVars(req, urlVars)
			}

			config := &destinationsvc.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationsvc.DestinationAPIConfig{
					RegionParam:       regionParamValue,
					SubaccountIDParam: subaccountIDParamValue,
				},
			}

			h := destinationsvc.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingDestination != nil {
				h.DestinationSvcDestinations = testCase.ExistingDestination
			}

			// WHEN
			h.CreateDestinations(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedDestinationSvcDestinations != nil {
				require.Equal(t, testCase.ExpectedDestinationSvcDestinations, h.DestinationSvcDestinations)
			}
		})
	}
}

func TestHandler_DeleteDestinations(t *testing.T) {
	destinationCreatorDeletionPath := fmt.Sprintf("/regions/%s/subaccounts/%s/destinations/%s", testRegion, testSubaccountID, noAuthDestName)

	testCases := []struct {
		Name                               string
		ExpectedResponseCode               int
		ExpectedDestinationSvcDestinations map[string]destcreatorpkg.Destination
		RegionParam                        string
		SubaccountIDParam                  string
		DestNameParam                      string
		ExistingDestination                map[string]destcreatorpkg.Destination
		MissingContentTypeHeader           bool
		MissingClientUserHeader            bool
	}{
		{
			Name:                               "Success when deleting destinations",
			ExpectedResponseCode:               http.StatusNoContent,
			ExpectedDestinationSvcDestinations: make(map[string]destcreatorpkg.Destination),
			RegionParam:                        testRegion,
			SubaccountIDParam:                  testSubaccountID,
			DestNameParam:                      noAuthDestName,
			ExistingDestination:                fixDestinationMappings(noAuthDestIdentifierWithSubaccountID, fixNoAuthDestination(noAuthDestName)),
		},
		{
			Name:                               "Success when there are no destinations to be deleted",
			ExpectedResponseCode:               http.StatusNoContent,
			ExpectedDestinationSvcDestinations: make(map[string]destcreatorpkg.Destination),
			RegionParam:                        testRegion,
			SubaccountIDParam:                  testSubaccountID,
			DestNameParam:                      noAuthDestName,
			ExistingDestination:                make(map[string]destcreatorpkg.Destination),
		},
		{
			Name:                     "Error when content type header is invalid",
			ExpectedResponseCode:     http.StatusUnsupportedMediaType,
			MissingContentTypeHeader: true,
		},
		{
			Name:                    "Error when client_user header is missing",
			ExpectedResponseCode:    http.StatusBadRequest,
			MissingClientUserHeader: true,
		},
		{
			Name:                 "Error when path params are missing",
			ExpectedResponseCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodDelete, url+destinationCreatorDeletionPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			if !testCase.MissingContentTypeHeader {
				req.Header.Add("Content-Type", "application/json;charset=UTF-8")
			}

			if !testCase.MissingClientUserHeader {
				req.Header.Add("CLIENT_USER", "test")
			}

			urlVars := make(map[string]string)
			if testCase.RegionParam != "" {
				urlVars[regionParamValue] = testCase.RegionParam
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.SubaccountIDParam != "" {
				urlVars[subaccountIDParamValue] = testCase.SubaccountIDParam
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.DestNameParam != "" {
				urlVars[destNameParamKey] = testCase.DestNameParam
				req = mux.SetURLVars(req, urlVars)
			}

			config := &destinationsvc.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationsvc.DestinationAPIConfig{
					RegionParam:          regionParamValue,
					SubaccountIDParam:    subaccountIDParamValue,
					DestinationNameParam: destNameParamKey,
				},
			}

			h := destinationsvc.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingDestination != nil {
				h.DestinationSvcDestinations = testCase.ExistingDestination
			}

			// WHEN
			h.DeleteDestinations(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedDestinationSvcDestinations != nil {
				require.Equal(t, testCase.ExpectedDestinationSvcDestinations, h.DestinationSvcDestinations)
			}
		})
	}
}

func TestHandler_CreateCertificate(t *testing.T) {
	destinationCreatorCertificatePath := fmt.Sprintf("/regions/%s/subaccounts/%s/certificates", testRegion, testSubaccountID)

	testCases := []struct {
		Name                               string
		RequestBody                        string
		ExpectedResponseCode               int
		ExpectedDestinationSvcCertificates map[string]json.RawMessage
		Region                             string
		SubaccountID                       string
		ExistingCertificate                map[string]json.RawMessage
		MissingContentTypeHeader           bool
		MissingClientUserHeader            bool
	}{
		{
			Name:                               "Success when creating certificate",
			RequestBody:                        destinationCreatorCertReqBody,
			ExpectedResponseCode:               http.StatusCreated,
			ExpectedDestinationSvcCertificates: fixCertMappings(destinationCertIdentifierWithSubaccountID, json.RawMessage(destinationServiceCertResponseBody)),
			Region:                             testRegion,
			SubaccountID:                       testSubaccountID,
		},
		{
			Name:                     "Error when content type header is invalid",
			ExpectedResponseCode:     http.StatusUnsupportedMediaType,
			MissingContentTypeHeader: true,
		},
		{
			Name:                    "Error when client_user header is missing",
			ExpectedResponseCode:    http.StatusBadRequest,
			MissingClientUserHeader: true,
		},
		{
			Name:                 "Error when path params are missing",
			ExpectedResponseCode: http.StatusBadRequest,
		},
		{
			Name:                 "Error when unmarshalling request body",
			RequestBody:          "invalid",
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
			ExpectedResponseCode: http.StatusInternalServerError,
		},
		{
			Name:                 "Error when validating request body",
			RequestBody:          `{"invalidKey": "value"}`,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
			ExpectedResponseCode: http.StatusBadRequest,
		},
		{
			Name:                 "Return conflict where there is already existing certificate",
			RequestBody:          destinationCreatorCertReqBody,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
			ExistingCertificate:  fixCertMappings(destinationCertIdentifierWithSubaccountID, json.RawMessage(destinationServiceCertResponseBody)),
			ExpectedResponseCode: http.StatusConflict,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodPost, url+destinationCreatorCertificatePath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)

			if !testCase.MissingContentTypeHeader {
				req.Header.Add("Content-Type", "application/json;charset=UTF-8")
			}

			if !testCase.MissingClientUserHeader {
				req.Header.Add("CLIENT_USER", "test")
			}

			urlVars := make(map[string]string)
			if testCase.Region != "" {
				urlVars[regionParamValue] = testCase.Region
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.SubaccountID != "" {
				urlVars[subaccountIDParamValue] = testCase.SubaccountID
				req = mux.SetURLVars(req, urlVars)
			}

			config := &destinationsvc.Config{
				CorrelationIDsKey: correlationIDsKey,
				CertificateAPIConfig: &destinationsvc.CertificateAPIConfig{
					RegionParam:       regionParamValue,
					SubaccountIDParam: subaccountIDParamValue,
				},
			}

			h := destinationsvc.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingCertificate != nil {
				h.DestinationSvcCertificates = testCase.ExistingCertificate
				testCase.ExpectedDestinationSvcCertificates = testCase.ExistingCertificate
			}

			// WHEN
			h.CreateCertificate(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedDestinationSvcCertificates != nil {
				require.Equal(t, testCase.ExpectedDestinationSvcCertificates, h.DestinationSvcCertificates)
			} else {
				require.Equal(t, make(map[string]json.RawMessage), h.DestinationSvcCertificates)
			}
		})
	}
}

func TestHandler_DeleteCertificate(t *testing.T) {
	destinationCreatorCertificateDeletionPath := fmt.Sprintf("/regions/%s/subaccounts/%s/certificates/%s", testRegion, testSubaccountID, testDestinationCertName)

	testCases := []struct {
		Name                               string
		ExpectedResponseCode               int
		ExpectedDestinationSvcCertificates map[string]json.RawMessage
		RegionParam                        string
		SubaccountIDParam                  string
		CertNameParam                      string
		ExistingCertificateDestinationSvc  map[string]json.RawMessage
		MissingContentTypeHeader           bool
		MissingClientUserHeader            bool
	}{
		{
			Name:                               "Success when deleting certificates",
			ExpectedResponseCode:               http.StatusNoContent,
			ExpectedDestinationSvcCertificates: make(map[string]json.RawMessage),
			RegionParam:                        testRegion,
			SubaccountIDParam:                  testSubaccountID,
			CertNameParam:                      testDestinationCertName,
			ExistingCertificateDestinationSvc:  fixCertMappings(destinationCertIdentifierWithSubaccountID, json.RawMessage(destinationServiceCertResponseBody)),
		},
		{
			Name:                               "Success when there are no certificates to be deleted",
			ExpectedResponseCode:               http.StatusNoContent,
			ExpectedDestinationSvcCertificates: make(map[string]json.RawMessage),
			RegionParam:                        testRegion,
			SubaccountIDParam:                  testSubaccountID,
			CertNameParam:                      testDestinationCertName,
			ExistingCertificateDestinationSvc:  make(map[string]json.RawMessage),
		},
		{
			Name:                     "Error when content type header is invalid",
			ExpectedResponseCode:     http.StatusUnsupportedMediaType,
			MissingContentTypeHeader: true,
		},
		{
			Name:                    "Error when client_user header is missing",
			ExpectedResponseCode:    http.StatusBadRequest,
			MissingClientUserHeader: true,
		},
		{
			Name:                 "Error when path params are missing",
			ExpectedResponseCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodDelete, url+destinationCreatorCertificateDeletionPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			if !testCase.MissingContentTypeHeader {
				req.Header.Add("Content-Type", "application/json;charset=UTF-8")
			}

			if !testCase.MissingClientUserHeader {
				req.Header.Add("CLIENT_USER", "test")
			}

			urlVars := make(map[string]string)
			if testCase.RegionParam != "" {
				urlVars[regionParamValue] = testCase.RegionParam
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.SubaccountIDParam != "" {
				urlVars[subaccountIDParamValue] = testCase.SubaccountIDParam
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.CertNameParam != "" {
				urlVars[certNameParamKey] = testCase.CertNameParam
				req = mux.SetURLVars(req, urlVars)
			}

			config := &destinationsvc.Config{
				CorrelationIDsKey: correlationIDsKey,
				CertificateAPIConfig: &destinationsvc.CertificateAPIConfig{
					RegionParam:          regionParamValue,
					SubaccountIDParam:    subaccountIDParamValue,
					CertificateNameParam: certNameParamKey,
				},
			}

			h := destinationsvc.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingCertificateDestinationSvc != nil {
				h.DestinationSvcCertificates = testCase.ExistingCertificateDestinationSvc
			}

			// WHEN
			h.DeleteCertificate(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedDestinationSvcCertificates != nil {
				require.Equal(t, testCase.ExpectedDestinationSvcCertificates, h.DestinationSvcCertificates)
			}
		})
	}
}

func TestHandler_FindDestinationByNameFromDestinationSvc(t *testing.T) {
	destinationSvcPath := fmt.Sprintf("/destination-configuration/v1/destinations/%s", noAuthDestName)
	tokenWithSubaccountIDAndInstanceID := generateJWT(t, testSubaccountID, testServiceInstanceID)
	tokenOnlyWithServiceInstanceID := generateJWT(t, "", testServiceInstanceID)

	testCases := []struct {
		Name                      string
		AuthorizationToken        string
		ExpectedResponseCode      int
		DestNameParam             string
		ExistingDestination       map[string]destcreatorpkg.Destination
		ExistingCertificate       map[string]json.RawMessage
		ExpectedResponse          json.RawMessage
		MissingAuthorizationToken bool
		MissingUserToken          bool
	}{
		{
			Name:                 "Success when getting destination by name from Destination Service",
			AuthorizationToken:   tokenWithSubaccountIDAndInstanceID,
			ExpectedResponseCode: http.StatusOK,
			DestNameParam:        samlAssertionDestName,
			ExistingDestination:  fixDestinationMappings(samlAssertionDestIdentifierWithSubaccountIDAndInstanceID, fixSAMLAssertionDestination(samlAssertionDestName)),
			ExistingCertificate:  fixCertMappings(samlDestinationCertIdentifierWithSubaccountIDAndInstanceID, json.RawMessage(destinationServiceSAMLDestCertResponseBody)),
			ExpectedResponse:     json.RawMessage(destinationServiceFindAPIResponseBodyForSAMLAssertionDest),
		},
		{
			Name:                 "Error when missing authorization token",
			ExpectedResponseCode: http.StatusUnauthorized,
		},
		{
			Name:                      "Error when authorization token value is empty",
			AuthorizationToken:        "",
			ExpectedResponseCode:      http.StatusUnauthorized,
			MissingAuthorizationToken: true,
		},
		{
			Name:                 "Error when path param is missing",
			AuthorizationToken:   tokenWithSubaccountIDAndInstanceID,
			ExpectedResponseCode: http.StatusBadRequest,
		},
		{
			Name:                 "Not Found when destination doesn't exist",
			AuthorizationToken:   tokenWithSubaccountIDAndInstanceID,
			ExpectedResponseCode: http.StatusNotFound,
			DestNameParam:        "invalid",
		},
		{
			Name:                 "Error when subaccount ID is missing from the authorization token",
			AuthorizationToken:   tokenOnlyWithServiceInstanceID,
			ExpectedResponseCode: http.StatusInternalServerError,
			DestNameParam:        samlAssertionDestName,
		},
		{
			Name:                 "Error when user token header is missing",
			AuthorizationToken:   tokenWithSubaccountIDAndInstanceID,
			ExpectedResponseCode: http.StatusInternalServerError,
			DestNameParam:        samlAssertionDestName,
			ExistingDestination:  fixDestinationMappings(samlAssertionDestIdentifierWithSubaccountIDAndInstanceID, fixSAMLAssertionDestination(samlAssertionDestName)),
			ExistingCertificate:  fixCertMappings(samlDestinationCertIdentifierWithSubaccountIDAndInstanceID, json.RawMessage(destinationServiceSAMLDestCertResponseBody)),
			MissingUserToken:     true,
		},
		{
			Name:                 "Error when certificate is missing",
			AuthorizationToken:   tokenWithSubaccountIDAndInstanceID,
			ExpectedResponseCode: http.StatusInternalServerError,
			DestNameParam:        samlAssertionDestName,
			ExistingDestination:  fixDestinationMappings(samlAssertionDestIdentifierWithSubaccountIDAndInstanceID, fixSAMLAssertionDestination(samlAssertionDestName)),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url+destinationSvcPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", testCase.AuthorizationToken))
			}

			if testCase.MissingAuthorizationToken {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, "Bearer ")
			}

			if !testCase.MissingUserToken {
				req.Header.Add(httphelpers.UserTokenHeaderKey, "test")
			}

			urlVars := make(map[string]string)
			if testCase.DestNameParam != "" {
				urlVars[nameParamKey] = testCase.DestNameParam
				req = mux.SetURLVars(req, urlVars)
			}

			h := destinationsvc.NewHandler(&destinationsvc.Config{})
			r := httptest.NewRecorder()

			if testCase.ExistingDestination != nil {
				h.DestinationSvcDestinations = testCase.ExistingDestination
			}

			if testCase.ExistingCertificate != nil {
				h.DestinationSvcCertificates = testCase.ExistingCertificate
			}

			// WHEN
			h.FindDestinationByNameFromDestinationSvc(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedResponse != nil {
				require.Equal(t, testCase.ExpectedResponse, json.RawMessage(body))
			}
		})
	}
}

func TestHandler_GetDestinationCertificateByNameFromDestinationSvc(t *testing.T) {
	destinationSvcCertificatePath := fmt.Sprintf("/destination-configuration/v1/subaccountCertificates%s", testDestinationCertName)
	tokenWithSubaccountIDAndInstanceID := generateJWT(t, testSubaccountID, testServiceInstanceID)
	tokenOnlyWithServiceInstanceID := generateJWT(t, "", testServiceInstanceID)

	testCases := []struct {
		Name                      string
		AuthorizationToken        string
		ExpectedResponseCode      int
		CertNameParam             string
		ExistingCertificate       map[string]json.RawMessage
		ExpectedCertificate       json.RawMessage
		MissingAuthorizationToken bool
	}{
		{
			Name:                 "Success when getting certificate by name from Destination Service",
			AuthorizationToken:   tokenWithSubaccountIDAndInstanceID,
			ExpectedResponseCode: http.StatusOK,
			CertNameParam:        testDestinationCertName,
			ExistingCertificate:  fixCertMappings(destinationCertIdentifierWithSubaccountIDAndInstanceID, json.RawMessage(destinationServiceCertResponseBody)),
			ExpectedCertificate:  json.RawMessage(destinationServiceCertResponseBody),
		},
		{
			Name:                 "Error when missing authorization token",
			ExpectedResponseCode: http.StatusUnauthorized,
		},
		{
			Name:                      "Error when authorization token value is empty",
			AuthorizationToken:        "",
			ExpectedResponseCode:      http.StatusUnauthorized,
			MissingAuthorizationToken: true,
		},
		{
			Name:                 "Error when path param is missing",
			AuthorizationToken:   tokenWithSubaccountIDAndInstanceID,
			ExpectedResponseCode: http.StatusBadRequest,
		},
		{
			Name:                 "Not Found when certificate doesn't exist",
			AuthorizationToken:   tokenWithSubaccountIDAndInstanceID,
			ExpectedResponseCode: http.StatusNotFound,
			CertNameParam:        testDestinationCertName,
		},
		{
			Name:                 "Error when subaccount ID is missing from the authorization token",
			AuthorizationToken:   tokenOnlyWithServiceInstanceID,
			ExpectedResponseCode: http.StatusInternalServerError,
			CertNameParam:        testDestinationCertName,
		},
		{
			Name:                 "Error when marshalling",
			AuthorizationToken:   tokenWithSubaccountIDAndInstanceID,
			ExpectedResponseCode: http.StatusInternalServerError,
			CertNameParam:        testDestinationCertName,
			ExistingCertificate:  map[string]json.RawMessage{destinationCertIdentifierWithSubaccountIDAndInstanceID: json.RawMessage("invalid-json")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url+destinationSvcCertificatePath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, fmt.Sprintf("Bearer %s", testCase.AuthorizationToken))
			}

			if testCase.MissingAuthorizationToken {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, "Bearer ")
			}

			urlVars := make(map[string]string)
			if testCase.CertNameParam != "" {
				urlVars[nameParamKey] = testCase.CertNameParam
				req = mux.SetURLVars(req, urlVars)
			}

			h := destinationsvc.NewHandler(&destinationsvc.Config{})
			r := httptest.NewRecorder()

			if testCase.ExistingCertificate != nil {
				h.DestinationSvcCertificates = testCase.ExistingCertificate
			}

			// WHEN
			h.GetDestinationCertificateByNameFromDestinationSvc(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedCertificate != nil {
				require.Equal(t, testCase.ExpectedCertificate, json.RawMessage(body))
			}
		})
	}
}

func TestHandler_GetSensitiveData(t *testing.T) {
	testCases := []struct {
		Name                         string
		ExpectedResponseCode         int
		DestNameParam                string
		DestName                     string
		ExistingDestinationSensitive map[string][]byte
		ExpectedDestinationSensitive []byte
	}{
		{
			Name:                         "Success when getting destination",
			ExpectedResponseCode:         http.StatusOK,
			DestNameParam:                nameParamKey,
			DestName:                     testDestinationName,
			ExistingDestinationSensitive: fixSensitiveData(destinationsvc.GetDestinationPrefixNameIdentifier(testDestinationName), []byte(destinationServiceFindAPIResponseBodyForSAMLAssertionDest)),
			ExpectedDestinationSensitive: []byte(destinationServiceFindAPIResponseBodyForSAMLAssertionDest),
		},
		{
			Name:                         "Error when name param is missing",
			ExpectedResponseCode:         http.StatusBadRequest,
			ExpectedDestinationSensitive: []byte("Missing name parameter\n"),
		},
		{
			Name:                         "Error when a destination with the given name does not exist",
			ExpectedResponseCode:         http.StatusNotFound,
			DestNameParam:                nameParamKey,
			DestName:                     testDestinationName,
			ExpectedDestinationSensitive: []byte("Destination not found\n"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			urlVars := make(map[string]string)
			if testCase.DestNameParam != "" {
				urlVars[nameParamKey] = testCase.DestName
				req = mux.SetURLVars(req, urlVars)
			}

			config := &destinationsvc.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationsvc.DestinationAPIConfig{
					RegionParam:          regionParamValue,
					SubaccountIDParam:    subaccountIDParamValue,
					DestinationNameParam: destNameParamKey,
				},
			}

			h := destinationsvc.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingDestinationSensitive != nil {
				h.DestinationsSensitive = testCase.ExistingDestinationSensitive
			}

			// WHEN
			h.GetSensitiveData(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			require.Equal(t, testCase.ExpectedDestinationSensitive, body)
		})
	}
}

func TestHandler_CleanupDestinations(t *testing.T) {
	t.Run("Successfully delete destinations data", func(t *testing.T) {
		// GIVEN
		req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer([]byte{}))
		require.NoError(t, err)

		config := &destinationsvc.Config{
			CorrelationIDsKey: correlationIDsKey,
			DestinationAPIConfig: &destinationsvc.DestinationAPIConfig{
				RegionParam:          regionParamValue,
				SubaccountIDParam:    subaccountIDParamValue,
				DestinationNameParam: destNameParamKey,
			},
		}

		h := destinationsvc.NewHandler(config)
		r := httptest.NewRecorder()

		// WHEN
		h.CleanupDestinations(r, req)
		resp := r.Result()

		// THEN
		require.Equal(t, resp.StatusCode, http.StatusOK)
		require.Equal(t, h.DestinationSvcDestinations, make(map[string]destcreatorpkg.Destination))
		require.Equal(t, h.DestinationsSensitive, make(map[string][]byte))
	})
}

func TestHandler_CleanupDestinationCertificates(t *testing.T) {
	t.Run("Successfully delete destinations data", func(t *testing.T) {
		// GIVEN
		req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer([]byte{}))
		require.NoError(t, err)

		config := &destinationsvc.Config{
			CorrelationIDsKey: correlationIDsKey,
			DestinationAPIConfig: &destinationsvc.DestinationAPIConfig{
				RegionParam:          regionParamValue,
				SubaccountIDParam:    subaccountIDParamValue,
				DestinationNameParam: destNameParamKey,
			},
		}

		h := destinationsvc.NewHandler(config)
		r := httptest.NewRecorder()

		// WHEN
		h.CleanupDestinationCertificates(r, req)
		resp := r.Result()

		// THEN
		require.Equal(t, resp.StatusCode, http.StatusOK)
		require.Equal(t, h.DestinationSvcCertificates, make(map[string]json.RawMessage))
	})
}

func TestHandler_PostDestination(t *testing.T) {
	headerWithSubaccountIDAndInstanceID := http.Header{}
	headerWithSubaccountIDAndInstanceID.Add("subaccount_id", testSubaccountID)
	headerWithSubaccountIDAndInstanceID.Add("instance_id", testServiceInstanceID)

	testCases := []struct {
		Name                         string
		ExpectedResponseCode         int
		ExpectedDestinationSensitive map[string][]byte
		ExistingDestinations         map[string]destcreatorpkg.Destination
		BodyData                     []byte
		ExpectedResponseBodyData     string
		Headers                      http.Header
		StatusCode                   int
	}{
		{
			Name:                 "Success when getting destination",
			ExpectedResponseCode: http.StatusCreated,
			ExpectedDestinationSensitive: map[string][]byte{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): []byte(destinationServiceFindAPIResponseBodyForBasicAssertionDest),
			},
			BodyData:                 []byte(gjson.Get(destinationServiceFindAPIResponseBodyForBasicAssertionDest, "destinationConfiguration").String()),
			Headers:                  headerWithSubaccountIDAndInstanceID,
			ExpectedResponseBodyData: "",
		},
		{
			Name:                         "Error when authentication type property is missing in the body",
			ExpectedResponseCode:         http.StatusBadRequest,
			ExpectedDestinationSensitive: map[string][]byte{},
			ExpectedResponseBodyData:     "{\"error\":\"The authenticationType field in the request body is required and it should not be empty. X-Request-Id: \"}\n",
			BodyData:                     []byte{},
			Headers:                      headerWithSubaccountIDAndInstanceID,
		},
		{
			Name:                         "Error when there are no headers",
			ExpectedResponseCode:         http.StatusBadRequest,
			ExpectedDestinationSensitive: map[string][]byte{},
			ExpectedResponseBodyData:     "{\"error\":\"missing subaccount_id and instance_id headers. X-Request-Id: \"}\n",
			Headers:                      http.Header{},
			BodyData:                     []byte(gjson.Get(destinationServiceFindAPIResponseBodyForBasicAssertionDest, "destinationConfiguration").String()),
		},
		{
			Name:                         "Error when destination type is invalid",
			ExpectedResponseCode:         http.StatusInternalServerError,
			ExpectedDestinationSensitive: map[string][]byte{},
			ExpectedResponseBodyData:     "{\"error\":\"The provided destination authentication type: invalid is invalid. X-Request-Id: \"}\n",
			BodyData:                     []byte(gjson.Get(invalidDestination, "destinationConfiguration").String()),
			Headers:                      headerWithSubaccountIDAndInstanceID,
		},
		{
			Name:                 "Error with conflict",
			ExpectedResponseCode: http.StatusOK,
			ExpectedDestinationSensitive: map[string][]byte{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): []byte(destinationServiceFindAPIResponseBodyForBasicAssertionDest),
			},
			ExistingDestinations: map[string]destcreatorpkg.Destination{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): nil,
			},
			BodyData:                 []byte(gjson.Get(destinationServiceFindAPIResponseBodyForBasicAssertionDest, "destinationConfiguration").String()),
			Headers:                  headerWithSubaccountIDAndInstanceID,
			ExpectedResponseBodyData: "[{\"name\":\"test-basic-dest\",\"status\":409,\"cause\":\"Destination name already taken\"},{\"name\":\"test-basic-dest\",\"status\":201}]",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			ctx := context.TODO()

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(testCase.BodyData))
			require.NoError(t, err)

			if testCase.Headers != nil {
				req.Header = testCase.Headers
			}
			req = req.WithContext(context.WithValue(ctx, correlation.RequestIDHeaderKey, "corr-id"))

			config := &destinationsvc.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationsvc.DestinationAPIConfig{
					RegionParam:          regionParamValue,
					SubaccountIDParam:    subaccountIDParamValue,
					DestinationNameParam: destNameParamKey,
				},
			}

			h := destinationsvc.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingDestinations != nil {
				h.DestinationSvcDestinations = testCase.ExistingDestinations
			}

			// WHEN
			h.PostDestination(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseBodyData, string(body))
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedDestinationSensitive, h.DestinationsSensitive)
		})
	}
}

func TestHandler_DeleteDestination(t *testing.T) {
	headerWithSubaccountIDAndInstanceID := http.Header{}
	headerWithSubaccountIDAndInstanceID.Add("subaccount_id", testSubaccountID)
	headerWithSubaccountIDAndInstanceID.Add("instance_id", testServiceInstanceID)

	emptyHeaders := http.Header{}

	testCases := []struct {
		Name                          string
		ExpectedResponseCode          int
		DestNameParam                 string
		DestName                      string
		ExistingDestinations          map[string]destcreatorpkg.Destination
		ExistingDestinationsSensitive map[string][]byte
		ExpectedDestinations          map[string]destcreatorpkg.Destination
		ExpectedDestinationsSensitive map[string][]byte
		Headers                       http.Header
		StatusCode                    int
	}{
		{
			Name:                 "Success when getting destination",
			ExpectedResponseCode: http.StatusOK,
			ExistingDestinationsSensitive: map[string][]byte{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): []byte(destinationServiceFindAPIResponseBodyForBasicAssertionDest),
			},
			ExistingDestinations: map[string]destcreatorpkg.Destination{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): nil,
			},
			Headers:                       headerWithSubaccountIDAndInstanceID,
			ExpectedDestinationsSensitive: map[string][]byte{},
			ExpectedDestinations:          map[string]destcreatorpkg.Destination{},
			DestNameParam:                 nameParamKey,
			DestName:                      basicAuthDestName,
		},
		{
			Name:                 "Error when destination name is missing",
			ExpectedResponseCode: http.StatusBadRequest,
			ExistingDestinationsSensitive: map[string][]byte{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): []byte(destinationServiceFindAPIResponseBodyForBasicAssertionDest),
			},
			ExistingDestinations: map[string]destcreatorpkg.Destination{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): nil,
			},
			Headers: headerWithSubaccountIDAndInstanceID,
			ExpectedDestinationsSensitive: map[string][]byte{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): []byte(destinationServiceFindAPIResponseBodyForBasicAssertionDest),
			},
			ExpectedDestinations: map[string]destcreatorpkg.Destination{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): nil,
			},
			DestNameParam: nameParamKey,
			DestName:      "",
		},
		{
			Name:                 "Error when headers are missing",
			ExpectedResponseCode: http.StatusBadRequest,
			ExistingDestinationsSensitive: map[string][]byte{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): []byte(destinationServiceFindAPIResponseBodyForBasicAssertionDest),
			},
			ExistingDestinations: map[string]destcreatorpkg.Destination{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): nil,
			},
			Headers: emptyHeaders,
			ExpectedDestinationsSensitive: map[string][]byte{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): []byte(destinationServiceFindAPIResponseBodyForBasicAssertionDest),
			},
			ExpectedDestinations: map[string]destcreatorpkg.Destination{
				fmt.Sprintf("name_%s_subacc_%s_instance_%s", basicAuthDestName, testSubaccountID, testServiceInstanceID): nil,
			},
			DestNameParam: nameParamKey,
			DestName:      basicAuthDestName,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			if testCase.Headers != nil {
				req.Header = testCase.Headers
			}

			config := &destinationsvc.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationsvc.DestinationAPIConfig{
					RegionParam:          regionParamValue,
					SubaccountIDParam:    subaccountIDParamValue,
					DestinationNameParam: destNameParamKey,
				},
			}

			urlVars := make(map[string]string)
			if testCase.DestNameParam != "" {
				urlVars[nameParamKey] = testCase.DestName
				req = mux.SetURLVars(req, urlVars)
			}

			h := destinationsvc.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingDestinations != nil {
				h.DestinationSvcDestinations = testCase.ExistingDestinations
			}
			if testCase.ExistingDestinationsSensitive != nil {
				h.DestinationsSensitive = testCase.ExistingDestinationsSensitive
			}

			// WHEN
			h.DeleteDestination(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			require.Equal(t, testCase.ExpectedDestinationsSensitive, h.DestinationsSensitive)
			require.Equal(t, testCase.ExpectedDestinations, h.DestinationSvcDestinations)
		})
	}
}

func generateJWT(t *testing.T, subaccountID, serviceInstanceID string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["ext_attr"] = json.RawMessage(fmt.Sprintf(`{"subaccountid":"%s", "serviceinstanceid":"%s"}`, subaccountID, serviceInstanceID))
	claims["exp"] = time.Now().Add(time.Minute * 1).Unix()

	tokenString, err := token.SignedString(testSecretKey)
	require.NoError(t, err)

	return tokenString
}
