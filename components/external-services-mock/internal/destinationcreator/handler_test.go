package destinationcreator_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/destinationcreator"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/stretchr/testify/require"
)

func TestHandler_CreateDestinations(t *testing.T) {
	destinationCreatorPath := fmt.Sprintf("/regions/%s/subaccounts/%s/destinations", testRegion, testSubaccountID)

	testCases := []struct {
		Name                                      string
		RequestBody                               string
		ExpectedResponseCode                      int
		ExpectedDestinationCreatorSvcDestinations map[string]json.RawMessage
		ExpectedDestinationSvcDestinations        map[string]json.RawMessage
		Region                                    string
		SubaccountID                              string
		ExistingDestination                       map[string]json.RawMessage
		MissingContentTypeHeader                  bool
		MissingClientUserHeader                   bool
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
			Name:                 "Success when creating no auth destinations",
			RequestBody:          destinationCreatorNoAuthDestReqBody,
			ExpectedResponseCode: http.StatusCreated,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
			ExpectedDestinationCreatorSvcDestinations: fixDestinationMappings(noAuthDestName, json.RawMessage(destinationCreatorNoAuthDestReqBody)),
			ExpectedDestinationSvcDestinations:        fixDestinationMappings(noAuthDestName, json.RawMessage(destinationServiceNoAuthDestReqBody)),
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
			ExistingDestination:  fixDestinationMappings(noAuthDestName, json.RawMessage(destinationCreatorNoAuthDestReqBody)),
		},
		// Basic Destinations unit tests
		{
			Name:                 "Success when creating basic destinations",
			RequestBody:          destinationCreatorBasicAuthDestReqBody,
			ExpectedResponseCode: http.StatusCreated,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
			ExpectedDestinationCreatorSvcDestinations: fixDestinationMappings(basicAuthDestName, json.RawMessage(destinationCreatorBasicAuthDestReqBody)),
			ExpectedDestinationSvcDestinations:        fixDestinationMappings(basicAuthDestName, json.RawMessage(destinationServiceBasicAuthReqBody)),
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
			ExistingDestination:  fixDestinationMappings(basicAuthDestName, json.RawMessage(destinationCreatorBasicAuthDestReqBody)),
		},
		// SAML Assertion Destinations unit tests
		{
			Name:                 "Success when creating SAML Assertion destinations",
			RequestBody:          destinationCreatorSAMLAssertionDestReqBody,
			ExpectedResponseCode: http.StatusCreated,
			Region:               testRegion,
			SubaccountID:         testSubaccountID,
			ExpectedDestinationCreatorSvcDestinations: fixDestinationMappings(samlAssertionDestName, json.RawMessage(destinationCreatorSAMLAssertionDestReqBody)),
			ExpectedDestinationSvcDestinations:        fixDestinationMappings(samlAssertionDestName, json.RawMessage(destinationServiceSAMLAssertionReqBody)),
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
			ExistingDestination:  fixDestinationMappings(samlAssertionDestName, json.RawMessage(destinationCreatorSAMLAssertionDestReqBody)),
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
				urlVars[regionParam] = testCase.Region
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.SubaccountID != "" {
				urlVars[subaccountIDParam] = testCase.SubaccountID
				req = mux.SetURLVars(req, urlVars)
			}

			config := &destinationcreator.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationcreator.DestinationAPIConfig{
					RegionParam:       regionParam,
					SubaccountIDParam: subaccountIDParam,
				},
			}

			h := destinationcreator.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingDestination != nil {
				h.DestinationCreatorSvcDestinations = testCase.ExistingDestination
			}

			// WHEN
			h.CreateDestinations(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedDestinationCreatorSvcDestinations != nil {
				require.Equal(t, testCase.ExpectedDestinationCreatorSvcDestinations, h.DestinationCreatorSvcDestinations)
			}

			if testCase.ExpectedDestinationSvcDestinations != nil {
				require.Equal(t, testCase.ExpectedDestinationSvcDestinations, h.DestinationSvcDestinations)
			}
		})
	}
}

func TestHandler_DeleteDestinations(t *testing.T) {
	destinationCreatorDeletionPath := fmt.Sprintf("/regions/%s/subaccounts/%s/destinations/%s", testRegion, testSubaccountID, testDestName)

	testCases := []struct {
		Name                                      string
		ExpectedResponseCode                      int
		ExpectedDestinationCreatorSvcDestinations map[string]json.RawMessage
		ExpectedDestinationSvcDestinations        map[string]json.RawMessage
		Region                                    string
		SubaccountID                              string
		DestName                                  string
		ExistingDestination                       map[string]json.RawMessage
		MissingContentTypeHeader                  bool
		MissingClientUserHeader                   bool
	}{
		{
			Name:                 "Success when deleting destinations",
			ExpectedResponseCode: http.StatusNoContent,
			ExpectedDestinationCreatorSvcDestinations: make(map[string]json.RawMessage),
			ExpectedDestinationSvcDestinations:        make(map[string]json.RawMessage),
			Region:                                    testRegion,
			SubaccountID:                              testSubaccountID,
			DestName:                                  testDestName,
			ExistingDestination:                       fixDestinationMappings(testDestName, json.RawMessage(destinationCreatorNoAuthDestReqBody)),
		},
		{
			Name:                 "Success when there are no destinations to be deleted",
			ExpectedResponseCode: http.StatusNoContent,
			ExpectedDestinationCreatorSvcDestinations: make(map[string]json.RawMessage),
			ExpectedDestinationSvcDestinations:        make(map[string]json.RawMessage),
			Region:                                    testRegion,
			SubaccountID:                              testSubaccountID,
			DestName:                                  testDestName,
			ExistingDestination:                       make(map[string]json.RawMessage),
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
			if testCase.Region != "" {
				urlVars[regionParam] = testCase.Region
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.SubaccountID != "" {
				urlVars[subaccountIDParam] = testCase.SubaccountID
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.DestName != "" {
				urlVars[destNameParam] = testCase.DestName
				req = mux.SetURLVars(req, urlVars)
			}

			config := &destinationcreator.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationcreator.DestinationAPIConfig{
					RegionParam:          regionParam,
					SubaccountIDParam:    subaccountIDParam,
					DestinationNameParam: destNameParam,
				},
			}

			h := destinationcreator.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingDestination != nil {
				h.DestinationCreatorSvcDestinations = testCase.ExistingDestination
				h.DestinationSvcDestinations = testCase.ExistingDestination
			}

			// WHEN
			h.DeleteDestinations(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedDestinationCreatorSvcDestinations != nil {
				require.Equal(t, testCase.ExpectedDestinationCreatorSvcDestinations, h.DestinationCreatorSvcDestinations)
			}

			if testCase.ExpectedDestinationSvcDestinations != nil {
				require.Equal(t, testCase.ExpectedDestinationSvcDestinations, h.DestinationSvcDestinations)
			}
		})
	}
}

func TestHandler_CreateCertificate(t *testing.T) {
	destinationCreatorCertificatePath := fmt.Sprintf("/regions/%s/subaccounts/%s/certificates", testRegion, testSubaccountID)

	testCases := []struct {
		Name                                      string
		RequestBody                               string
		ExpectedResponseCode                      int
		ExpectedDestinationCreatorSvcCertificates map[string]json.RawMessage
		ExpectedDestinationSvcCertificates        map[string]json.RawMessage
		Region                                    string
		SubaccountID                              string
		ExistingCertificate                       map[string]json.RawMessage
		MissingContentTypeHeader                  bool
		MissingClientUserHeader                   bool
	}{
		{
			Name:                 "Success when creating certificate",
			RequestBody:          destinationCreatorCertReqBody,
			ExpectedResponseCode: http.StatusCreated,
			ExpectedDestinationCreatorSvcCertificates: fixCertMappings(testCertName, json.RawMessage(destinationCreatorCertResponseBody)),
			ExpectedDestinationSvcCertificates:        fixCertMappings(testCertFileName, json.RawMessage(destinationServiceCertResponseBody)),
			Region:                                    testRegion,
			SubaccountID:                              testSubaccountID,
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
			ExistingCertificate:  fixCertMappings(testCertName, json.RawMessage(destinationCreatorCertResponseBody)),
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
				urlVars[regionParam] = testCase.Region
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.SubaccountID != "" {
				urlVars[subaccountIDParam] = testCase.SubaccountID
				req = mux.SetURLVars(req, urlVars)
			}

			config := &destinationcreator.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationcreator.DestinationAPIConfig{
					RegionParam:       regionParam,
					SubaccountIDParam: subaccountIDParam,
				},
			}

			h := destinationcreator.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingCertificate != nil {
				h.DestinationCreatorSvcCertificates = testCase.ExistingCertificate
			}

			// WHEN
			h.CreateCertificate(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedDestinationCreatorSvcCertificates != nil {
				require.Equal(t, testCase.ExpectedDestinationCreatorSvcCertificates, h.DestinationCreatorSvcCertificates)
			}

			if testCase.ExpectedDestinationSvcCertificates != nil {
				require.Equal(t, testCase.ExpectedDestinationSvcCertificates, h.DestinationSvcCertificates)
			}
		})
	}
}

func TestHandler_DeleteCertificate(t *testing.T) {
	destinationCreatorCertificateDeletionPath := fmt.Sprintf("/regions/%s/subaccounts/%s/certificates/%s", testRegion, testSubaccountID, testCertName)

	testCases := []struct {
		Name                                      string
		ExpectedResponseCode                      int
		ExpectedDestinationCreatorSvcCertificates map[string]json.RawMessage
		ExpectedDestinationSvcCertificates        map[string]json.RawMessage
		Region                                    string
		SubaccountID                              string
		CertName                                  string
		ExistingCertificateDestinationCreator     map[string]json.RawMessage
		ExistingCertificateDestinationSvc         map[string]json.RawMessage
		MissingContentTypeHeader                  bool
		MissingClientUserHeader                   bool
	}{
		{
			Name:                 "Success when deleting certificates",
			ExpectedResponseCode: http.StatusNoContent,
			ExpectedDestinationCreatorSvcCertificates: make(map[string]json.RawMessage),
			ExpectedDestinationSvcCertificates:        make(map[string]json.RawMessage),
			Region:                                    testRegion,
			SubaccountID:                              testSubaccountID,
			CertName:                                  testCertName,
			ExistingCertificateDestinationCreator:     fixCertMappings(testCertName, json.RawMessage(destinationCreatorCertResponseBody)),
			ExistingCertificateDestinationSvc:         fixCertMappings(testCertFileName, json.RawMessage(destinationServiceCertResponseBody)),
		},
		{
			Name:                 "Success when there are no certificates to be deleted",
			ExpectedResponseCode: http.StatusNoContent,
			ExpectedDestinationCreatorSvcCertificates: make(map[string]json.RawMessage),
			ExpectedDestinationSvcCertificates:        make(map[string]json.RawMessage),
			Region:                                    testRegion,
			SubaccountID:                              testSubaccountID,
			CertName:                                  testCertName,
			ExistingCertificateDestinationCreator:     make(map[string]json.RawMessage),
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
			if testCase.Region != "" {
				urlVars[regionParam] = testCase.Region
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.SubaccountID != "" {
				urlVars[subaccountIDParam] = testCase.SubaccountID
				req = mux.SetURLVars(req, urlVars)
			}

			if testCase.CertName != "" {
				urlVars[certNameParam] = testCase.CertName
				req = mux.SetURLVars(req, urlVars)
			}

			config := &destinationcreator.Config{
				CorrelationIDsKey: correlationIDsKey,
				DestinationAPIConfig: &destinationcreator.DestinationAPIConfig{
					RegionParam:          regionParam,
					SubaccountIDParam:    subaccountIDParam,
					DestinationNameParam: destNameParam,
				},
				CertificateAPIConfig: &destinationcreator.CertificateAPIConfig{
					CertificateNameParam: certNameParam,
				},
			}

			h := destinationcreator.NewHandler(config)
			r := httptest.NewRecorder()

			if testCase.ExistingCertificateDestinationCreator != nil {
				h.DestinationCreatorSvcCertificates = testCase.ExistingCertificateDestinationCreator
			}

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

			if testCase.ExpectedDestinationCreatorSvcCertificates != nil {
				require.Equal(t, testCase.ExpectedDestinationCreatorSvcCertificates, h.DestinationCreatorSvcCertificates)
			}

			if testCase.ExpectedDestinationSvcCertificates != nil {
				require.Equal(t, testCase.ExpectedDestinationSvcCertificates, h.DestinationSvcCertificates)
			}
		})
	}
}

func TestHandler_GetDestinationByNameFromDestinationSvc(t *testing.T) {
	destinationSvcPath := fmt.Sprintf("/destination-configuration/v1/subaccountDestinations/%s", testDestName)

	testCases := []struct {
		Name                       string
		ExpectedResponseCode       int
		DestName                   string
		ExistingDestination        map[string]json.RawMessage
		ExpectedDestination        json.RawMessage
		MissingAuthorizationHeader bool
		MissingAuthorizationToken  bool
	}{
		{
			Name:                 "Success when getting destination by name from Destination Service",
			ExpectedResponseCode: http.StatusOK,
			DestName:             noAuthDestName,
			ExistingDestination:  fixDestinationMappings(noAuthDestName, json.RawMessage(destinationCreatorNoAuthDestReqBody)),
			ExpectedDestination:  json.RawMessage(destinationCreatorNoAuthDestReqBody),
		},
		{
			Name:                       "Error when missing authorization token",
			ExpectedResponseCode:       http.StatusBadRequest,
			MissingAuthorizationHeader: true,
		},
		{
			Name:                       "Error when authorization token value is empty",
			ExpectedResponseCode:       http.StatusBadRequest,
			MissingAuthorizationHeader: true,
			MissingAuthorizationToken:  true,
		},
		{
			Name:                 "Error when path param is missing",
			ExpectedResponseCode: http.StatusBadRequest,
		},
		{
			Name:                 "Not Found when destination doesn't exist",
			ExpectedResponseCode: http.StatusNotFound,
			DestName:             noAuthDestName,
		},
		{
			Name:                 "Error when marshalling",
			ExpectedResponseCode: http.StatusBadRequest,
			DestName:             noAuthDestName,
			ExistingDestination:  map[string]json.RawMessage{noAuthDestName: json.RawMessage("invalid-json")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url+destinationSvcPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			if !testCase.MissingAuthorizationHeader {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, "Bearer token")
			}

			if testCase.MissingAuthorizationToken {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, "Bearer ")
			}

			urlVars := make(map[string]string)
			if testCase.DestName != "" {
				urlVars[nameParam] = testCase.DestName
				req = mux.SetURLVars(req, urlVars)
			}

			h := destinationcreator.NewHandler(&destinationcreator.Config{})
			r := httptest.NewRecorder()

			if testCase.ExistingDestination != nil {
				h.DestinationSvcDestinations = testCase.ExistingDestination
			}

			// WHEN
			h.GetDestinationByNameFromDestinationSvc(r, req)
			resp := r.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))

			if testCase.ExpectedDestination != nil {
				require.Equal(t, testCase.ExpectedDestination, json.RawMessage(body))
			}
		})
	}
}

func TestHandler_GetDestinationCertificateByNameFromDestinationSvc(t *testing.T) {
	destinationSvcCertificatePath := fmt.Sprintf("/destination-configuration/v1/subaccountCertificates%s", testCertName)

	testCases := []struct {
		Name                       string
		ExpectedResponseCode       int
		CertName                   string
		ExistingCertificate        map[string]json.RawMessage
		ExpectedCertificate        json.RawMessage
		MissingAuthorizationHeader bool
		MissingAuthorizationToken  bool
	}{
		{
			Name:                 "Success when getting certificate by name from Destination Service",
			ExpectedResponseCode: http.StatusOK,
			CertName:             testCertName,
			ExistingCertificate:  fixCertMappings(testCertName, json.RawMessage(destinationCreatorCertResponseBody)),
			ExpectedCertificate:  json.RawMessage(destinationCreatorCertResponseBody),
		},
		{
			Name:                       "Error when missing authorization token",
			ExpectedResponseCode:       http.StatusBadRequest,
			MissingAuthorizationHeader: true,
		},
		{
			Name:                       "Error authorization token value is empty",
			ExpectedResponseCode:       http.StatusBadRequest,
			MissingAuthorizationHeader: true,
			MissingAuthorizationToken:  true,
		},
		{
			Name:                 "Error when path param is missing",
			ExpectedResponseCode: http.StatusBadRequest,
		},
		{
			Name:                 "Not Found when certificate doesn't exist",
			ExpectedResponseCode: http.StatusNotFound,
			CertName:             testCertName,
		},
		{
			Name:                 "Error when marshalling",
			ExpectedResponseCode: http.StatusBadRequest,
			CertName:             testCertName,
			ExistingCertificate:  map[string]json.RawMessage{testCertName: json.RawMessage("invalid-json")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url+destinationSvcCertificatePath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			if !testCase.MissingAuthorizationHeader {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, "Bearer token")
			}

			if testCase.MissingAuthorizationToken {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, "Bearer ")
			}

			urlVars := make(map[string]string)
			if testCase.CertName != "" {
				urlVars[nameParam] = testCase.CertName
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
