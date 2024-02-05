package service_manager

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	url                  = "https://target-url.com"
	path                 = "/instance-creator/v1/tenantMappings/{tenant-id}"
	subaccountQueryParam = "subaccount_id"
	labelsQueryParam     = "labelQuery"

	subaccount = "subaccount"
	token      = "Bearer 123"
)

// Service Offerings Tests
func TestHandler_ServiceOfferingsList(t *testing.T) {
	serviceOfferingsPath := "/v1/service_offerings"

	testCases := []struct {
		Name                     string
		Subaccount               string
		AuthorizationToken       string
		ExpectedServiceOfferings string
		ExpectedErrorMessage     string
		ExpectedResponseCode     int
	}{
		{
			Name:                     "Success",
			Subaccount:               subaccount,
			AuthorizationToken:       token,
			ExpectedResponseCode:     http.StatusOK,
			ExpectedServiceOfferings: `{"num_items":2,"items":[{"id":"first-service-offering-id","catalog_name":"first-service-offering-test"},{"id":"second-service-offering-id","catalog_name":"second-service-offering-test"}]}`,
		},
		{
			Name:                 "Error when authorization value is empty",
			AuthorizationToken:   "",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "Missing authorization header",
		},
		{
			Name:                 "Error when authorization token is empty",
			AuthorizationToken:   "Bearer ",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "The token value cannot be empty",
		},
		{
			Name:                 "Error when subaccount is empty",
			Subaccount:           "",
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get subaccount from query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url+serviceOfferingsPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			values := req.URL.Query()
			values.Add(subaccountQueryParam, testCase.Subaccount)
			req.URL.RawQuery = values.Encode()

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, testCase.AuthorizationToken)
			}

			config := Config{
				Path:                 path,
				SubaccountQueryParam: subaccountQueryParam,
				LabelsQueryParam:     labelsQueryParam,
			}

			h := NewServiceManagerHandler(config)
			r := httptest.NewRecorder()

			// WHEN
			h.HandleServiceOfferingsList(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			if testCase.ExpectedErrorMessage != "" {
				require.Contains(t, string(body), testCase.ExpectedErrorMessage)
			} else {
				require.JSONEq(t, testCase.ExpectedServiceOfferings, string(body))
			}

		})
	}
}

// Service Plans Tests
func TestHandler_ServicePlansList(t *testing.T) {
	servicePlansPath := "/v1/service_plans"

	testCases := []struct {
		Name                 string
		Subaccount           string
		AuthorizationToken   string
		ExpectedServicePlans string
		ExpectedErrorMessage string
		ExpectedResponseCode int
	}{
		{
			Name:                 "Success",
			Subaccount:           subaccount,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusOK,
			ExpectedServicePlans: `{"num_items":2,"items":[{"id":"1","catalog_name":"first-catalog-name","service_offering_id":"first-service-offering-id"},{"id":"2","catalog_name":"second-catalog-name","service_offering_id":"second-service-offering-id"}]}`,
		},
		{
			Name:                 "Error when authorization value is empty",
			AuthorizationToken:   "",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "Missing authorization header",
		},
		{
			Name:                 "Error when authorization token is empty",
			AuthorizationToken:   "Bearer ",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "The token value cannot be empty",
		},
		{
			Name:                 "Error when subaccount is empty",
			Subaccount:           "",
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get subaccount from query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url+servicePlansPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			values := req.URL.Query()
			values.Add(subaccountQueryParam, testCase.Subaccount)
			req.URL.RawQuery = values.Encode()

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, testCase.AuthorizationToken)
			}

			config := Config{
				Path:                 path,
				SubaccountQueryParam: subaccountQueryParam,
				LabelsQueryParam:     labelsQueryParam,
			}

			h := NewServiceManagerHandler(config)
			r := httptest.NewRecorder()

			// WHEN
			h.HandleServicePlansList(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			if testCase.ExpectedErrorMessage != "" {
				require.Contains(t, string(body), testCase.ExpectedErrorMessage)
			} else {
				require.JSONEq(t, testCase.ExpectedServicePlans, string(body))
			}

		})
	}
}

// Service Bindings Tests
func TestHandler_ServiceBindingsList(t *testing.T) {
	serviceBindingsPath := "/v1/service_bindings"

	testCases := []struct {
		Name                       string
		Subaccount                 string
		AuthorizationToken         string
		ExistingServiceBindingsMap map[string]ServiceBindingsMock
		ExpectedServiceBindings    string
		ExpectedErrorMessage       string
		ExpectedResponseCode       int
	}{
		{
			Name:               "Success - with existing service bindings",
			Subaccount:         subaccount,
			AuthorizationToken: token,
			ExistingServiceBindingsMap: map[string]ServiceBindingsMock{
				subaccount: {
					NumItems: 2,
					Items: []*ServiceBindingMock{
						{
							ID:                "id-1",
							Name:              "name-1",
							ServiceInstanceID: "instance-id-1",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
						{
							ID:                "id-2",
							Name:              "name-2",
							ServiceInstanceID: "instance-id-2",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
					},
				},
			},
			ExpectedResponseCode:    http.StatusOK,
			ExpectedServiceBindings: `{"num_items":2,"items":[{"id":"id-1","name":"name-1","service_instance_id":"instance-id-1","credentials":"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"},{"id":"id-2","name":"name-2","service_instance_id":"instance-id-2","credentials":"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"}]}`,
		},
		{
			Name:                    "Success - without service bindings",
			Subaccount:              subaccount,
			AuthorizationToken:      token,
			ExpectedResponseCode:    http.StatusOK,
			ExpectedServiceBindings: `{"num_items":0,"items":null}`,
		},
		{
			Name:                 "Error when authorization value is empty",
			AuthorizationToken:   "",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "Missing authorization header",
		},
		{
			Name:                 "Error when authorization token is empty",
			AuthorizationToken:   "Bearer ",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "The token value cannot be empty",
		},
		{
			Name:                 "Error when subaccount is empty",
			Subaccount:           "",
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get subaccount from query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url+serviceBindingsPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			values := req.URL.Query()
			values.Add(subaccountQueryParam, testCase.Subaccount)
			req.URL.RawQuery = values.Encode()

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, testCase.AuthorizationToken)
			}

			config := Config{
				Path:                 path,
				SubaccountQueryParam: subaccountQueryParam,
				LabelsQueryParam:     labelsQueryParam,
			}

			h := NewServiceManagerHandler(config)
			if len(testCase.ExistingServiceBindingsMap) != 0 {
				h.ServiceBindingsMap = testCase.ExistingServiceBindingsMap
			}
			r := httptest.NewRecorder()

			// WHEN
			h.HandleServiceBindingsList(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			if testCase.ExpectedErrorMessage != "" {
				require.Contains(t, string(body), testCase.ExpectedErrorMessage)
			} else {
				require.JSONEq(t, testCase.ExpectedServiceBindings, string(body))
			}

		})
	}
}

func TestHandler_ServiceBindingsGet(t *testing.T) {
	serviceBindingsPath := fmt.Sprintf("/v1/service_bindings")

	testServiceBindingID := "id-1"

	testCases := []struct {
		Name                       string
		Subaccount                 string
		ServiceBindingID           string
		AuthorizationToken         string
		ExistingServiceBindingsMap map[string]ServiceBindingsMock
		ExpectedServiceBinding     string
		ExpectedErrorMessage       string
		ExpectedResponseCode       int
	}{
		{
			Name:               "Success",
			Subaccount:         subaccount,
			ServiceBindingID:   testServiceBindingID,
			AuthorizationToken: token,
			ExistingServiceBindingsMap: map[string]ServiceBindingsMock{
				subaccount: {
					NumItems: 2,
					Items: []*ServiceBindingMock{
						{
							ID:                testServiceBindingID,
							Name:              "name-1",
							ServiceInstanceID: "instance-id-1",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
						{
							ID:                "id-2",
							Name:              "name-2",
							ServiceInstanceID: "instance-id-2",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
					},
				},
			},
			ExpectedResponseCode:   http.StatusOK,
			ExpectedServiceBinding: `{"id":"id-1","name":"name-1","service_instance_id":"instance-id-1","credentials":"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"}`,
		},
		{
			Name:                 "Error - service binding not found",
			Subaccount:           subaccount,
			ServiceBindingID:     testServiceBindingID,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusNotFound,
			ExpectedErrorMessage: "Service binding not found",
		},
		{
			Name:                 "Error when authorization value is empty",
			AuthorizationToken:   "",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "Missing authorization header",
		},
		{
			Name:                 "Error when authorization token is empty",
			AuthorizationToken:   "Bearer ",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "The token value cannot be empty",
		},
		{
			Name:                 "Error - service binding id not found in url vars",
			Subaccount:           subaccount,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get service binding id from url",
		},
		{
			Name:                 "Error when subaccount is empty",
			Subaccount:           "",
			ServiceBindingID:     testServiceBindingID,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get subaccount from query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url+serviceBindingsPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			values := req.URL.Query()
			values.Add(subaccountQueryParam, testCase.Subaccount)
			req.URL.RawQuery = values.Encode()

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, testCase.AuthorizationToken)
			}

			if testCase.ServiceBindingID != "" {
				req = mux.SetURLVars(req, map[string]string{ServiceBindingIDPath: testCase.ServiceBindingID})
			}

			config := Config{
				Path:                 path,
				SubaccountQueryParam: subaccountQueryParam,
				LabelsQueryParam:     labelsQueryParam,
			}

			h := NewServiceManagerHandler(config)
			if len(testCase.ExistingServiceBindingsMap) != 0 {
				h.ServiceBindingsMap = testCase.ExistingServiceBindingsMap
			}
			r := httptest.NewRecorder()

			// WHEN
			h.HandleServiceBindingGet(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			if testCase.ExpectedErrorMessage != "" {
				require.Contains(t, string(body), testCase.ExpectedErrorMessage)
			} else {
				require.JSONEq(t, testCase.ExpectedServiceBinding, string(body))
			}
		})
	}
}

func TestHandler_ServiceBindingsDelete(t *testing.T) {
	serviceBindingsPath := fmt.Sprintf("/v1/service_bindings")

	testServiceBindingID := "id-1"

	testCases := []struct {
		Name                        string
		Subaccount                  string
		ServiceBindingID            string
		AuthorizationToken          string
		ExistingServiceBindingsMap  map[string]ServiceBindingsMock
		ExpectedServiceBindingsLeft map[string]ServiceBindingsMock
		ExpectedErrorMessage        string
		ExpectedResponseCode        int
	}{
		{
			Name:               "Success",
			Subaccount:         subaccount,
			ServiceBindingID:   testServiceBindingID,
			AuthorizationToken: token,
			ExistingServiceBindingsMap: map[string]ServiceBindingsMock{
				subaccount: {
					NumItems: 2,
					Items: []*ServiceBindingMock{
						{
							ID:                testServiceBindingID,
							Name:              "name-1",
							ServiceInstanceID: "instance-id-1",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
						{
							ID:                "id-2",
							Name:              "name-2",
							ServiceInstanceID: "instance-id-2",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
					},
				},
			},
			ExpectedResponseCode: http.StatusOK,
			ExpectedServiceBindingsLeft: map[string]ServiceBindingsMock{
				subaccount: {
					NumItems: 1,
					Items: []*ServiceBindingMock{
						{
							ID:                "id-2",
							Name:              "name-2",
							ServiceInstanceID: "instance-id-2",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
					},
				},
			},
		},
		{
			Name:                 "Error - service binding not found",
			Subaccount:           subaccount,
			ServiceBindingID:     testServiceBindingID,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusNotFound,
			ExpectedErrorMessage: "Service binding not found",
		},
		{
			Name:                 "Error when authorization value is empty",
			AuthorizationToken:   "",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "Missing authorization header",
		},
		{
			Name:                 "Error when authorization token is empty",
			AuthorizationToken:   "Bearer ",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "The token value cannot be empty",
		},
		{
			Name:                 "Error - service binding id not found in url vars",
			Subaccount:           subaccount,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get service binding id from url",
		},
		{
			Name:                 "Error when subaccount is empty",
			Subaccount:           "",
			ServiceBindingID:     testServiceBindingID,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get subaccount from query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodPost, url+serviceBindingsPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			values := req.URL.Query()
			values.Add(subaccountQueryParam, testCase.Subaccount)
			req.URL.RawQuery = values.Encode()

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, testCase.AuthorizationToken)
			}

			if testCase.ServiceBindingID != "" {
				req = mux.SetURLVars(req, map[string]string{ServiceBindingIDPath: testCase.ServiceBindingID})
			}

			config := Config{
				Path:                 path,
				SubaccountQueryParam: subaccountQueryParam,
				LabelsQueryParam:     labelsQueryParam,
			}

			h := NewServiceManagerHandler(config)
			if len(testCase.ExistingServiceBindingsMap) != 0 {
				h.ServiceBindingsMap = testCase.ExistingServiceBindingsMap
			}
			r := httptest.NewRecorder()

			// WHEN
			h.HandleServiceBindingDelete(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			if testCase.ExpectedErrorMessage != "" {
				require.Contains(t, string(body), testCase.ExpectedErrorMessage)
			} else {
				require.Equal(t, testCase.ExpectedServiceBindingsLeft, h.ServiceBindingsMap)
			}
		})
	}
}

func TestHandler_ServiceBindingsCreate(t *testing.T) {
	serviceBindingsPath := fmt.Sprintf("/v1/service_bindings")

	testServiceBindingID := "id-1"

	testCases := []struct {
		Name                       string
		Subaccount                 string
		RequestBody                string
		AuthorizationToken         string
		ExistingServiceBindingsMap map[string]ServiceBindingsMock
		ExpectedNewServiceBinding  ServiceBindingMock
		ExpectedErrorMessage       string
		ExpectedResponseCode       int
	}{
		{
			Name:               "Success",
			Subaccount:         subaccount,
			RequestBody:        `{"name": "name-2", "service_instance_id": "instance-id-2", "credentials": "-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"}`,
			AuthorizationToken: token,
			ExistingServiceBindingsMap: map[string]ServiceBindingsMock{
				subaccount: {
					NumItems: 1,
					Items: []*ServiceBindingMock{
						{
							ID:                testServiceBindingID,
							Name:              "name-1",
							ServiceInstanceID: "instance-id-1",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
					},
				},
			},
			ExpectedResponseCode: http.StatusOK,
			ExpectedNewServiceBinding: ServiceBindingMock{
				ID:                "id-2",
				Name:              "name-2",
				ServiceInstanceID: "instance-id-2",
				Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
			},
		},
		{
			Name:                 "Error when authorization value is empty",
			AuthorizationToken:   "",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "Missing authorization header",
		},
		{
			Name:                 "Error when authorization token is empty",
			AuthorizationToken:   "Bearer ",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "The token value cannot be empty",
		},
		{
			Name:                 "Error when subaccount is empty",
			Subaccount:           "",
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get subaccount from query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodPost, url+serviceBindingsPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)

			values := req.URL.Query()
			values.Add(subaccountQueryParam, testCase.Subaccount)
			req.URL.RawQuery = values.Encode()

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, testCase.AuthorizationToken)
			}

			config := Config{
				Path:                 path,
				SubaccountQueryParam: subaccountQueryParam,
				LabelsQueryParam:     labelsQueryParam,
			}

			h := NewServiceManagerHandler(config)
			if len(testCase.ExistingServiceBindingsMap) != 0 {
				h.ServiceBindingsMap = testCase.ExistingServiceBindingsMap
			}
			r := httptest.NewRecorder()

			// WHEN
			h.HandleServiceBindingCreate(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			if testCase.ExpectedErrorMessage != "" {
				require.Contains(t, string(body), testCase.ExpectedErrorMessage)
			} else {
				found := false
				for _, binding := range h.ServiceBindingsMap[subaccount].Items {
					if binding.ServiceInstanceID == testCase.ExpectedNewServiceBinding.ServiceInstanceID {
						require.Equal(t, testCase.ExpectedNewServiceBinding.ServiceInstanceID, binding.ServiceInstanceID)
						require.Equal(t, testCase.ExpectedNewServiceBinding.Name, binding.Name)
						require.Equal(t, testCase.ExpectedNewServiceBinding.Credentials, binding.Credentials)
						found = true
					}
				}
				require.True(t, found)
			}
		})
	}
}

// Service Instances Tests
func TestHandler_ServiceInstancesList(t *testing.T) {
	serviceInstancesPath := "/v1/service_instances"

	testCases := []struct {
		Name                        string
		Subaccount                  string
		AuthorizationToken          string
		ExistingServiceInstancesMap map[string]ServiceInstancesMock
		ExpectedServiceInstances    string
		ExpectedErrorMessage        string
		ExpectedResponseCode        int
	}{
		{
			Name:               "Success - with existing instances",
			Subaccount:         subaccount,
			AuthorizationToken: token,
			ExistingServiceInstancesMap: map[string]ServiceInstancesMock{
				subaccount: {
					NumItems: 2,
					Items: []*ServiceInstanceMock{
						{
							ID:            "123",
							Name:          "name-1",
							ServicePlanID: "plan-1",
							PlatformID:    "platform-1",
							Labels:        map[string][]string{"label-key-1": {"label-val-1", "label-val-2"}},
						},
						{
							ID:            "456",
							Name:          "name-2",
							ServicePlanID: "plan-2",
							PlatformID:    "platform-2",
							Labels:        map[string][]string{"label-key-1": {"label-val-1", "label-val-2"}},
						},
					},
				},
			},
			ExpectedResponseCode:     http.StatusOK,
			ExpectedServiceInstances: `{"num_items":2,"items":[{"id":"123","name":"name-1","service_plan_id":"plan-1","platform_id":"platform-1","labels":{"label-key-1":["label-val-1","label-val-2"]}},{"id":"456","name":"name-2","service_plan_id":"plan-2","platform_id":"platform-2","labels":{"label-key-1":["label-val-1","label-val-2"]}}]}`,
		},
		{
			Name:                     "Success - without instances",
			Subaccount:               subaccount,
			AuthorizationToken:       token,
			ExpectedResponseCode:     http.StatusOK,
			ExpectedServiceInstances: `{"num_items":0,"items":null}`,
		},
		{
			Name:                 "Error when authorization value is empty",
			AuthorizationToken:   "",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "Missing authorization header",
		},
		{
			Name:                 "Error when authorization token is empty",
			AuthorizationToken:   "Bearer ",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "The token value cannot be empty",
		},
		{
			Name:                 "Error when subaccount is empty",
			Subaccount:           "",
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get subaccount from query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url+serviceInstancesPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			values := req.URL.Query()
			values.Add(subaccountQueryParam, testCase.Subaccount)
			req.URL.RawQuery = values.Encode()

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, testCase.AuthorizationToken)
			}

			config := Config{
				Path:                 path,
				SubaccountQueryParam: subaccountQueryParam,
				LabelsQueryParam:     labelsQueryParam,
			}

			h := NewServiceManagerHandler(config)
			if len(testCase.ExistingServiceInstancesMap) != 0 {
				h.ServiceInstancesMap = testCase.ExistingServiceInstancesMap
			}
			r := httptest.NewRecorder()

			// WHEN
			h.HandleServiceInstancesList(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			if testCase.ExpectedErrorMessage != "" {
				require.Contains(t, string(body), testCase.ExpectedErrorMessage)
			} else {
				require.JSONEq(t, testCase.ExpectedServiceInstances, string(body))
			}

		})
	}
}

func TestHandler_ServiceInstancesGet(t *testing.T) {
	serviceInstancesPath := fmt.Sprintf("/v1/service_instances")

	testServiceInstanceID := "id-1"

	testCases := []struct {
		Name                        string
		Subaccount                  string
		ServiceInstanceID           string
		LabelQuery                  string
		AuthorizationToken          string
		ExistingServiceInstancesMap map[string]ServiceInstancesMock
		ExpectedServiceInstance     string
		ExpectedErrorMessage        string
		ExpectedResponseCode        int
	}{
		{
			Name:               "Success",
			Subaccount:         subaccount,
			ServiceInstanceID:  testServiceInstanceID,
			AuthorizationToken: token,
			ExistingServiceInstancesMap: map[string]ServiceInstancesMock{
				subaccount: {
					NumItems: 2,
					Items: []*ServiceInstanceMock{
						{
							ID:            testServiceInstanceID,
							Name:          "name-1",
							ServicePlanID: "plan-1",
							PlatformID:    "platform-1",
						},
						{
							ID:            "id-2",
							Name:          "name-2",
							ServicePlanID: "plan-2",
							PlatformID:    "platform-2",
						},
					},
				},
			},
			ExpectedResponseCode:    http.StatusOK,
			ExpectedServiceInstance: `{"id":"id-1","name":"name-1","service_plan_id":"plan-1","platform_id":"platform-1"}`,
		},
		{
			Name:               "Success - with label filtering",
			Subaccount:         subaccount,
			ServiceInstanceID:  testServiceInstanceID,
			AuthorizationToken: token,
			ExistingServiceInstancesMap: map[string]ServiceInstancesMock{
				subaccount: {
					NumItems: 2,
					Items: []*ServiceInstanceMock{
						{
							ID:            testServiceInstanceID,
							Name:          "name-1",
							ServicePlanID: "plan-1",
							PlatformID:    "platform-1",
							Labels:        map[string][]string{},
						},
						{
							ID:            "id-2",
							Name:          "name-2",
							ServicePlanID: "plan-2",
							PlatformID:    "platform-2",
						},
					},
				},
			},
			ExpectedResponseCode:    http.StatusOK,
			ExpectedServiceInstance: `{"id":"id-1","name":"name-1","service_plan_id":"plan-1","platform_id":"platform-1"}`,
		},
		{
			Name:                 "Error - service instance not found",
			Subaccount:           subaccount,
			ServiceInstanceID:    testServiceInstanceID,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusNotFound,
			ExpectedErrorMessage: "Service instance not found",
		},
		{
			Name:                 "Error when authorization value is empty",
			AuthorizationToken:   "",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "Missing authorization header",
		},
		{
			Name:                 "Error when authorization token is empty",
			AuthorizationToken:   "Bearer ",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "The token value cannot be empty",
		},
		{
			Name:                 "Error - service instance id not found in url vars",
			Subaccount:           subaccount,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get service instance id from url",
		},
		{
			Name:                 "Error when subaccount is empty",
			Subaccount:           "",
			ServiceInstanceID:    testServiceInstanceID,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get subaccount from query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodGet, url+serviceInstancesPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			values := req.URL.Query()
			values.Add(subaccountQueryParam, testCase.Subaccount)
			if testCase.LabelQuery != "" {
				values.Add(labelsQueryParam, testCase.LabelQuery)
			}
			req.URL.RawQuery = values.Encode()

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, testCase.AuthorizationToken)
			}

			if testCase.ServiceInstanceID != "" {
				req = mux.SetURLVars(req, map[string]string{ServiceInstanceIDPath: testCase.ServiceInstanceID})
			}

			config := Config{
				Path:                 path,
				SubaccountQueryParam: subaccountQueryParam,
				LabelsQueryParam:     labelsQueryParam,
			}

			h := NewServiceManagerHandler(config)
			if len(testCase.ExistingServiceInstancesMap) != 0 {
				h.ServiceInstancesMap = testCase.ExistingServiceInstancesMap
			}
			r := httptest.NewRecorder()

			// WHEN
			h.HandleServiceInstanceGet(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			if testCase.ExpectedErrorMessage != "" {
				require.Contains(t, string(body), testCase.ExpectedErrorMessage)
			} else {
				require.JSONEq(t, testCase.ExpectedServiceInstance, string(body))
			}
		})
	}
}

func TestHandler_ServiceInstancesDelete(t *testing.T) {
	serviceBindingsPath := fmt.Sprintf("/v1/service_instances")

	testServiceBindingID := "id-1"

	testCases := []struct {
		Name                        string
		Subaccount                  string
		ServiceBindingID            string
		AuthorizationToken          string
		ExistingServiceBindingsMap  map[string]ServiceBindingsMock
		ExpectedServiceBindingsLeft map[string]ServiceBindingsMock
		ExpectedErrorMessage        string
		ExpectedResponseCode        int
	}{
		{
			Name:               "Success",
			Subaccount:         subaccount,
			ServiceBindingID:   testServiceBindingID,
			AuthorizationToken: token,
			ExistingServiceBindingsMap: map[string]ServiceBindingsMock{
				subaccount: {
					NumItems: 2,
					Items: []*ServiceBindingMock{
						{
							ID:                testServiceBindingID,
							Name:              "name-1",
							ServiceInstanceID: "instance-id-1",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
						{
							ID:                "id-2",
							Name:              "name-2",
							ServiceInstanceID: "instance-id-2",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
					},
				},
			},
			ExpectedResponseCode: http.StatusOK,
			ExpectedServiceBindingsLeft: map[string]ServiceBindingsMock{
				subaccount: {
					NumItems: 1,
					Items: []*ServiceBindingMock{
						{
							ID:                "id-2",
							Name:              "name-2",
							ServiceInstanceID: "instance-id-2",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
					},
				},
			},
		},
		{
			Name:                 "Error - service binding not found",
			Subaccount:           subaccount,
			ServiceBindingID:     testServiceBindingID,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusNotFound,
			ExpectedErrorMessage: "Service binding not found",
		},
		{
			Name:                 "Error when authorization value is empty",
			AuthorizationToken:   "",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "Missing authorization header",
		},
		{
			Name:                 "Error when authorization token is empty",
			AuthorizationToken:   "Bearer ",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "The token value cannot be empty",
		},
		{
			Name:                 "Error - service binding id not found in url vars",
			Subaccount:           subaccount,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get service binding id from url",
		},
		{
			Name:                 "Error when subaccount is empty",
			Subaccount:           "",
			ServiceBindingID:     testServiceBindingID,
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get subaccount from query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodPost, url+serviceBindingsPath, bytes.NewBuffer([]byte{}))
			require.NoError(t, err)

			values := req.URL.Query()
			values.Add(subaccountQueryParam, testCase.Subaccount)
			req.URL.RawQuery = values.Encode()

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, testCase.AuthorizationToken)
			}

			if testCase.ServiceBindingID != "" {
				req = mux.SetURLVars(req, map[string]string{ServiceBindingIDPath: testCase.ServiceBindingID})
			}

			config := Config{
				Path:                 path,
				SubaccountQueryParam: subaccountQueryParam,
				LabelsQueryParam:     labelsQueryParam,
			}

			h := NewServiceManagerHandler(config)
			if len(testCase.ExistingServiceBindingsMap) != 0 {
				h.ServiceBindingsMap = testCase.ExistingServiceBindingsMap
			}
			r := httptest.NewRecorder()

			// WHEN
			h.HandleServiceBindingDelete(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			if testCase.ExpectedErrorMessage != "" {
				require.Contains(t, string(body), testCase.ExpectedErrorMessage)
			} else {
				require.Equal(t, testCase.ExpectedServiceBindingsLeft, h.ServiceBindingsMap)
			}
		})
	}
}

func TestHandler_ServiceInstancesCreate(t *testing.T) {
	serviceBindingsPath := fmt.Sprintf("/v1/service_instances")

	testServiceBindingID := "id-1"

	testCases := []struct {
		Name                       string
		Subaccount                 string
		RequestBody                string
		AuthorizationToken         string
		ExistingServiceBindingsMap map[string]ServiceBindingsMock
		ExpectedNewServiceBinding  ServiceBindingMock
		ExpectedErrorMessage       string
		ExpectedResponseCode       int
	}{
		{
			Name:               "Success",
			Subaccount:         subaccount,
			RequestBody:        `{"name": "name-2", "service_instance_id": "instance-id-2", "credentials": "-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"}`,
			AuthorizationToken: token,
			ExistingServiceBindingsMap: map[string]ServiceBindingsMock{
				subaccount: {
					NumItems: 1,
					Items: []*ServiceBindingMock{
						{
							ID:                testServiceBindingID,
							Name:              "name-1",
							ServiceInstanceID: "instance-id-1",
							Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
						},
					},
				},
			},
			ExpectedResponseCode: http.StatusOK,
			ExpectedNewServiceBinding: ServiceBindingMock{
				ID:                "id-2",
				Name:              "name-2",
				ServiceInstanceID: "instance-id-2",
				Credentials:       []byte(`"-----BEGIN CERTIFICATE-----\n cert \n-----END CERTIFICATE-----\n"`),
			},
		},
		{
			Name:                 "Error when authorization value is empty",
			AuthorizationToken:   "",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "Missing authorization header",
		},
		{
			Name:                 "Error when authorization token is empty",
			AuthorizationToken:   "Bearer ",
			ExpectedResponseCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "The token value cannot be empty",
		},
		{
			Name:                 "Error when subaccount is empty",
			Subaccount:           "",
			AuthorizationToken:   token,
			ExpectedResponseCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "Failed to get subaccount from query",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			req, err := http.NewRequest(http.MethodPost, url+serviceBindingsPath, bytes.NewBuffer([]byte(testCase.RequestBody)))
			require.NoError(t, err)

			values := req.URL.Query()
			values.Add(subaccountQueryParam, testCase.Subaccount)
			req.URL.RawQuery = values.Encode()

			if testCase.AuthorizationToken != "" {
				req.Header.Add(httphelpers.AuthorizationHeaderKey, testCase.AuthorizationToken)
			}

			config := Config{
				Path:                 path,
				SubaccountQueryParam: subaccountQueryParam,
				LabelsQueryParam:     labelsQueryParam,
			}

			h := NewServiceManagerHandler(config)
			if len(testCase.ExistingServiceBindingsMap) != 0 {
				h.ServiceBindingsMap = testCase.ExistingServiceBindingsMap
			}
			r := httptest.NewRecorder()

			// WHEN
			h.HandleServiceBindingCreate(r, req)
			resp := r.Result()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// THEN
			require.Equal(t, testCase.ExpectedResponseCode, resp.StatusCode, string(body))
			if testCase.ExpectedErrorMessage != "" {
				require.Contains(t, string(body), testCase.ExpectedErrorMessage)
			} else {
				found := false
				for _, binding := range h.ServiceBindingsMap[subaccount].Items {
					if binding.ServiceInstanceID == testCase.ExpectedNewServiceBinding.ServiceInstanceID {
						require.Equal(t, testCase.ExpectedNewServiceBinding.ServiceInstanceID, binding.ServiceInstanceID)
						require.Equal(t, testCase.ExpectedNewServiceBinding.Name, binding.Name)
						require.Equal(t, testCase.ExpectedNewServiceBinding.Credentials, binding.Credentials)
						found = true
					}
				}
				require.True(t, found)
			}
		})
	}
}
