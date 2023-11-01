package destinationcreator_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/destinationcreator"
	destcreatorpkg "github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"
	"github.com/stretchr/testify/require"
)

var (
	testSecretKey                                              = []byte("testSecretKey")
	noAuthDestIdentifierWithSubaccountID                       = fmt.Sprintf(destinationcreator.UniqueEntityNameIdentifier, noAuthDestName, testSubaccountID, "")
	samlAssertionDestIdentifierWithSubaccountIDAndInstanceID   = fmt.Sprintf(destinationcreator.UniqueEntityNameIdentifier, samlAssertionDestName, testSubaccountID, testServiceInstanceID)
	destinationCertIdentifierWithSubaccountID                  = fmt.Sprintf(destinationcreator.UniqueEntityNameIdentifier, testDestinationCertWithExtension, testSubaccountID, "")
	destinationCertIdentifierWithSubaccountIDAndInstanceID     = fmt.Sprintf(destinationcreator.UniqueEntityNameIdentifier, testDestinationCertWithExtension, testSubaccountID, testServiceInstanceID)
	samlDestinationCertIdentifierWithSubaccountIDAndInstanceID = fmt.Sprintf(destinationcreator.UniqueEntityNameIdentifier, testDestKeyStoreLocation, testSubaccountID, testServiceInstanceID)
)

func TestHandler_CreateDestinations(t *testing.T) {
	destinationCreatorPath := fmt.Sprintf("/regions/%s/subaccounts/%s/destinations", testRegion, testSubaccountID)
	basicDestIdentifierWithSubaccountID := fmt.Sprintf(destinationcreator.UniqueEntityNameIdentifier, basicAuthDestName, testSubaccountID, "")
	samlAssertionDestIdentifierWithSubaccountID := fmt.Sprintf(destinationcreator.UniqueEntityNameIdentifier, samlAssertionDestName, testSubaccountID, "")

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

			config := &destinationcreator.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationcreator.DestinationAPIConfig{
					RegionParam:       regionParamValue,
					SubaccountIDParam: subaccountIDParamValue,
				},
			}

			h := destinationcreator.NewHandler(config)
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

			config := &destinationcreator.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationcreator.DestinationAPIConfig{
					RegionParam:          regionParamValue,
					SubaccountIDParam:    subaccountIDParamValue,
					DestinationNameParam: destNameParamKey,
				},
			}

			h := destinationcreator.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingDestination != nil {
				h.DestinationSvcDestinations = testCase.ExistingDestination
			}

			// WHEN
			h.DeleteDestinations(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
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

			config := &destinationcreator.Config{
				CorrelationIDsKey: correlationIDsKey,
				CertificateAPIConfig: &destinationcreator.CertificateAPIConfig{
					RegionParam:       regionParamValue,
					SubaccountIDParam: subaccountIDParamValue,
				},
			}

			h := destinationcreator.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingCertificate != nil {
				h.DestinationSvcCertificates = testCase.ExistingCertificate
				testCase.ExpectedDestinationSvcCertificates = testCase.ExistingCertificate
			}

			// WHEN
			h.CreateCertificate(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
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

			config := &destinationcreator.Config{
				CorrelationIDsKey: correlationIDsKey,
				CertificateAPIConfig: &destinationcreator.CertificateAPIConfig{
					RegionParam:          regionParamValue,
					SubaccountIDParam:    subaccountIDParamValue,
					CertificateNameParam: certNameParamKey,
				},
			}

			h := destinationcreator.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingCertificateDestinationSvc != nil {
				h.DestinationSvcCertificates = testCase.ExistingCertificateDestinationSvc
			}

			// WHEN
			h.DeleteCertificate(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
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

			h := destinationcreator.NewHandler(&destinationcreator.Config{})
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
			CertNameParam:        testDestinationCertWithExtension,
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
			CertNameParam:        testDestinationCertWithExtension,
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

			h := destinationcreator.NewHandler(&destinationcreator.Config{})
			r := httptest.NewRecorder()

			if testCase.ExistingCertificate != nil {
				h.DestinationSvcCertificates = testCase.ExistingCertificate
			}

			// WHEN
			h.GetDestinationCertificateByNameFromDestinationSvc(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedCertificate != nil {
				require.Equal(t, testCase.ExpectedCertificate, json.RawMessage(body))
			}
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
