package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types/tenantmapping"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/handler"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/handler/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	inboundCommunicationKey   = "inboundCommunication"
	outboundCommunicationKey  = "outboundCommunication"
	serviceInstancesKey       = "serviceInstances"
	serviceBindingKey         = "serviceBinding"
	serviceInstanceServiceKey = "service"
	serviceInstancePlanKey    = "plan"
	parametersKey             = "parameters"
	nameKey                   = "name"
	assignmentIDKey           = "assignment_id"
	currentWaveHashKey        = "current_wave_hash"
	reverseKey                = "reverse"
)

func Test_HandlerFunc(t *testing.T) {
	url := "https://target-url.com"
	apiPath := fmt.Sprintf("/")
	statusUrl := "localhost"

	testErr := errors.New("test error")

	formationID := "formation-id"
	assignmentID := "assignment-id"
	region := "region"
	subaccount := "subaccount"
	serviceInstancesIDs := []string{"instance-id-1", "instance-id-2"}
	serviceInstancesNames := []string{"instance-name-1", "instance-name-2"}
	serviceInstancesBindingsIDs := []string{"binding-id-1", "binding-id-2", "binding-id-3", "binding-id-4"}
	serviceInstanceBindingsNames := []string{"binding-name-1", "binding-name-2"}

	emptyJSON := `{}`

	reqBodyFormatter := `{
	 "context": %s,
	 "receiverTenant": %s,
	 "assignedTenant": %s
	}`

	reqBodyContextFormatter := `{"uclFormationId": %q, "operation": %q}`
	reqBodyContextWithAssign := fmt.Sprintf(reqBodyContextFormatter, formationID, "assign")
	reqBodyContextWithUnassign := fmt.Sprintf(reqBodyContextFormatter, formationID, "unassign")

	assignedTenantFormatter := `{
		"uclAssignmentId": %q,
		"configuration": %s
	}`

	serviceOfferingIDs := []string{"service-offering-id-1", "service-offering-id-2"}
	servicePlanIDs := []string{"service-plan-id-1", "service-plan-id-2"}

	assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths := `{
	      "credentials": {
            "inboundCommunication": {
			  "serviceInstances": [
                {
                  "service": "procurement-service",
                  "plan": "apiaccess",
                  "parameters": {},
                  "serviceBinding": {
                    "parameters": {}
                  }
                },
                {
                  "service": "identity",
                  "plan": "application",
                  "parameters": {
                    "consumed-services": [],
                    "xsuaa-cross-consumption": true
                  },
                  "serviceBinding": {
                    "parameters": {
                      "credential-type": "X509_PROVIDED",
                      "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
                    }
                  }
                }
              ],
			  "no-instances-auth-method": {
                "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  },
			  "refering-global-instances-auth-method": {
                "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  }
            }
          }
	    }`
	expectedResponseForGlobalInstances := `{"state":"READY","configuration":{"credentials":{"outboundCommunication":{"no-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"},"refering-global-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"}}}}}`

	assignedTenantConfigurationWithGlobalInstancesWithJsonpaths := `{
	     "credentials": {
	       "inboundCommunication": {
			  "serviceInstances": [
	           {
	             "service": "procurement-service",
	             "plan": "apiaccess",
	             "parameters": {},
	             "serviceBinding": {
	               "parameters": {
				      "service-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan"
			        }
	             }
	           },
	           {
	             "service": "identity",
	             "plan": "application",
	             "parameters": {
	               "consumed-services": [
	                 {
	                   "first-service-instance-service": "$.credentials.inboundCommunication.serviceInstances[0].service"
	                 }
	               ],
	               "xsuaa-cross-consumption": true
	             },
	             "serviceBinding": {
	               "parameters": {
	                 "credential-type": "X509_PROVIDED",
	                 "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
				      "first-service-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan",
				      "service-instance-plan": "$.credentials.inboundCommunication.serviceInstances[1].plan"
	               }
	             }
	           }
	         ],
			  "no-instances-auth-method": {
	           "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  },
			  "refering-global-instances-auth-method": {
				"certificate": "$.credentials.inboundCommunication.serviceInstances[1].serviceBinding.parameters.certificate"
			  }
	       }
	     }
	   }`
	substitutedAssignedTenantConfigurationWithGlobalInstancesWithJsonpaths := `{
	     "credentials": {
	       "inboundCommunication": {
			  "serviceInstances": [
	           {
	             "service": "procurement-service",
	             "plan": "apiaccess",
	             "parameters": {},
	             "serviceBinding": {
	               "parameters": {
				      "service-instance-plan": "apiaccess"
			        }
	             }
	           },
	           {
	             "service": "identity",
	             "plan": "application",
	             "parameters": {
	               "consumed-services": [
	                 {
	                   "first-service-instance-service": "procurement-service"
	                 }
	               ],
	               "xsuaa-cross-consumption": true
	             },
	             "serviceBinding": {
	               "parameters": {
	                 "credential-type": "X509_PROVIDED",
	                 "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----",
				      "first-service-instance-plan": "apiaccess",
				      "service-instance-plan": "application"
	               }
	             }
	           }
	         ],
			  "no-instances-auth-method": {
	           "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  },
			  "refering-global-instances-auth-method": {
				"certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  }
	       }
	     }
	   }`

	receiverTenantFormatter := `{
		"deploymentRegion": %q,
		"subaccountId": %q,
		"configuration": %s
	}`

	receiverTenantConfigurationWithServiceInstanceDetails := `{
	     "credentials": {
	       "inboundCommunication": {
			  "serviceInstances": [
	           {
	             "service": "procurement-service",
	             "plan": "apiaccess",
	             "parameters": {}
	           },
	           {
	             "service": "identity",
	             "plan": "application",
	             "xsuaa-cross-consumption": true
	           }
	         ],
			  "refering-global-instances-auth-method": {
			   "global-instance-plan": "$.credentials.inboundCommunication.serviceInstances[0].plan",
	           "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  }
	       }
	     }
	   }`

	receiverTenantConfigurationWithServiceInstanceDetailsAndMethodWithoutInstances := `{
	    "credentials": {
	      "inboundCommunication": {
			  "serviceInstances": [
	          {
	            "service": "procurement-service",
	            "plan": "apiaccess",
	            "parameters": {}
	          },
	          {
	            "service": "identity",
	            "plan": "application",
	            "xsuaa-cross-consumption": true
	          }
	        ],
			  "no-instances-details-method": {
	           "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  }
	      }
	    }
	  }`
	expectedResponseForGlobalInstancesWithInbound := `{"state":"READY","configuration":{"credentials":{"inboundCommunication":{"no-instances-details-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"}},"outboundCommunication":{"no-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"},"refering-global-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"}}}}}`

	receiverTenantConfigurationWithServiceInstanceDetailsAndMethodWithoutInstancesAndReversePaths := `{
	    "credentials": {
	      "inboundCommunication": {
			  "serviceInstances": [
	          {
	            "service": "procurement-service",
	            "plan": "apiaccess",
	            "parameters": {}
	          },
	          {
	            "service": "identity",
	            "plan": "application",
	            "xsuaa-cross-consumption": true
	          }
	        ],
			  "no-instances-details-method": {
	            "certificate": "-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"
			  },
 			  "reverse-paths-method": {
	            "reverse-second-instance-plan": "$.reverse.credentials.inboundCommunication.serviceInstances[1].plan"
			  }
	      }
	    }
	  }`
	expectedResponseForGlobalInstancesWithInboundAndReverse := `{"state":"READY","configuration":{"credentials":{"inboundCommunication":{"no-instances-details-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"},"reverse-paths-method":{"reverse-second-instance-plan":"application"}},"outboundCommunication":{"no-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"},"refering-global-instances-auth-method":{"certificate":"-----BEGIN CERTIFICATE----- cert -----END CERTIFICATE-----"}}}}}`

	testCases := []struct {
		name                 string
		smClientFn           func() *automock.Client
		mtlsClientFn         func() *automock.MtlsHTTPClient
		requestBody          string
		expectedResponseCode int
	}{
		{
			name:        "Wrong json - fails on decoding",
			requestBody: `wrong json`,
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Request body contains badly-formed JSON")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing config(empty json) - fails on validation",
			requestBody: emptyJSON,
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Formation ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing config(empty context, receiverTenant and assignedTenant) - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, emptyJSON, emptyJSON, emptyJSON),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("while validating the request body")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing formation ID in the context - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, `{"operation": "assign"}`, emptyJSON, emptyJSON),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Formation ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing operation in the context - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, `{"uclFormationId": "formation-id", "operation": ""}`, emptyJSON, emptyJSON),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Operation is invalid")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Wrong operation in the context - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, `{"uclFormationId": "formation-id", "operation": "wrong-operation"}`, emptyJSON, emptyJSON),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Operation is invalid")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Formation assignment is missing in the assignedTenant - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, emptyJSON, emptyJSON),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Assigned Tenant Assignment ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Region is missing in the receiverTenant - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, emptyJSON, fmt.Sprintf(assignedTenantFormatter, assignmentID, emptyJSON)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Receiver Tenant Region should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Subaccount ID is missing in the receiverTenant - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, `{"deploymentRegion": "region"}`, fmt.Sprintf(assignedTenantFormatter, assignmentID, emptyJSON)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Receiver Tenant Subaccount ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is assign and inboundCommunication is missing in the assignedTenant configuration - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, emptyJSON)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Assigned tenant inbound communication is missing in the configuration")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is assign and receiverTenant has outboundCommunication but not in the same path as assignedTenant inboundCommunication - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(`Receiver tenant outbound communication should be in the same place as the assigned tenant inbound communication`)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is unassign and fails while retrieving service instances by assignment ID",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, map[string][]string{assignmentIDKey: {assignmentID}}).Return(nil, testErr).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(`while retrieving service instances for assignmentID`)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is unassign and fails while retrieving service instances bindings by service instances IDs",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, map[string][]string{assignmentIDKey: {assignmentID}}).Return(serviceInstancesIDs, nil).Once()
				client.On("RetrieveMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs}).Return(nil, testErr).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(fmt.Sprintf("while retrieving service bindings for service instaces with IDs: %v", serviceInstancesIDs))).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is unassign and fails while deleting service instances bindings by service instances IDs",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, map[string][]string{assignmentIDKey: {assignmentID}}).Return(serviceInstancesIDs, nil).Once()
				client.On("RetrieveMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs}).Return(serviceInstancesBindingsIDs, nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesBindingsIDs).Return(testErr).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(fmt.Sprintf("while deleting service bindings with IDs: %v", serviceInstancesBindingsIDs))).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is unassign and fails while deleting service instances by service instances IDs",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, map[string][]string{assignmentIDKey: {assignmentID}}).Return(serviceInstancesIDs, nil).Once()
				client.On("RetrieveMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs}).Return(serviceInstancesBindingsIDs, nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesBindingsIDs).Return(nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesIDs).Return(testErr).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(fmt.Sprintf("while deleting service instances with IDs: %v", serviceInstancesIDs))).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is unassign and successfully deletes instances",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, map[string][]string{assignmentIDKey: {assignmentID}}).Return(serviceInstancesIDs, nil).Once()
				client.On("RetrieveMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs}).Return(serviceInstancesBindingsIDs, nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesBindingsIDs).Return(nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesIDs).Return(nil).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(`{"state":"READY","configuration":null}`)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and service instances are missing. Expecting CONFIG_PENDING.",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(`{"state":"CONFIG_PENDING","configuration":null}`)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances without jsonpaths.",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				globalServiceInstances := Configuration(assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths).GetGlobalServiceInstances(inboundCommunicationKey).ToArray()
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, smLabelsThatHaveAssignmentID(assignmentID)).Return(nil, nil).Once()

				firstGlobalServiceInstance := globalServiceInstances[0]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: firstGlobalServiceInstance.GetService()}).Return(serviceOfferingIDs[0], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: firstGlobalServiceInstance.GetPlan(), OfferingID: serviceOfferingIDs[0]}).Return(servicePlanIDs[0], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(firstGlobalServiceInstance.GetName(), servicePlanIDs[0], assignmentID, firstGlobalServiceInstance.GetParameters()), mock.Anything).Return(serviceInstancesIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[0]}).Return(firstGlobalServiceInstance.WithName(serviceInstancesNames[0]).ToJSONRawMessage(), nil).Once()

				firstGlobalServiceInstanceBinding := firstGlobalServiceInstance.GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(firstGlobalServiceInstanceBinding.GetName(), serviceInstancesIDs[0], firstGlobalServiceInstanceBinding.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[0]}).Return(firstGlobalServiceInstanceBinding.WithName(serviceInstanceBindingsNames[0]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstance := globalServiceInstances[1]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: secondGlobalServiceInstance.GetService()}).Return(serviceOfferingIDs[1], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: secondGlobalServiceInstance.GetPlan(), OfferingID: serviceOfferingIDs[1]}).Return(servicePlanIDs[1], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(secondGlobalServiceInstance.GetName(), servicePlanIDs[1], assignmentID, secondGlobalServiceInstance.GetParameters()), mock.Anything).Return(serviceInstancesIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[1]}).Return(secondGlobalServiceInstance.WithName(serviceInstancesNames[1]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstanceBinding := globalServiceInstances[1].GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(secondGlobalServiceInstanceBinding.GetName(), serviceInstancesIDs[1], secondGlobalServiceInstanceBinding.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[1]}).Return(secondGlobalServiceInstanceBinding.WithName(serviceInstanceBindingsNames[1]).ToJSONRawMessage(), nil).Once()

				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(expectedResponseForGlobalInstances)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances with jsonpaths.",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				substitutedGlobalServiceInstances := Configuration(substitutedAssignedTenantConfigurationWithGlobalInstancesWithJsonpaths).GetGlobalServiceInstances(inboundCommunicationKey).ToArray()
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, smLabelsThatHaveAssignmentID(assignmentID)).Return(nil, nil).Once()

				// First Instance
				firstGlobalServiceInstanceSubstituted := substitutedGlobalServiceInstances[0]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: firstGlobalServiceInstanceSubstituted.GetService()}).Return(serviceOfferingIDs[0], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: firstGlobalServiceInstanceSubstituted.GetPlan(), OfferingID: serviceOfferingIDs[0]}).Return(servicePlanIDs[0], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(firstGlobalServiceInstanceSubstituted.GetName(), servicePlanIDs[0], assignmentID, firstGlobalServiceInstanceSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[0]}).Return(firstGlobalServiceInstanceSubstituted.WithName(serviceInstancesNames[0]).ToJSONRawMessage(), nil).Once()

				firstGlobalServiceInstanceBindingSubstituted := firstGlobalServiceInstanceSubstituted.GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(firstGlobalServiceInstanceBindingSubstituted.GetName(), serviceInstancesIDs[0], firstGlobalServiceInstanceBindingSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[0]}).Return(firstGlobalServiceInstanceBindingSubstituted.WithName(serviceInstanceBindingsNames[0]).ToJSONRawMessage(), nil).Once()

				// Second Instance
				secondGlobalServiceInstanceSubstituted := substitutedGlobalServiceInstances[1]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: secondGlobalServiceInstanceSubstituted.GetService()}).Return(serviceOfferingIDs[1], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: secondGlobalServiceInstanceSubstituted.GetPlan(), OfferingID: serviceOfferingIDs[1]}).Return(servicePlanIDs[1], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(secondGlobalServiceInstanceSubstituted.GetName(), servicePlanIDs[1], assignmentID, secondGlobalServiceInstanceSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[1]}).Return(secondGlobalServiceInstanceSubstituted.WithName(serviceInstancesNames[1]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstanceBindingSubstituted := secondGlobalServiceInstanceSubstituted.GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(secondGlobalServiceInstanceBindingSubstituted.GetName(), serviceInstancesIDs[1], secondGlobalServiceInstanceBindingSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[1]}).Return(secondGlobalServiceInstanceBindingSubstituted.WithName(serviceInstanceBindingsNames[1]).ToJSONRawMessage(), nil).Once()

				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(expectedResponseForGlobalInstances)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances with jsonpaths which must be recreated.",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, emptyJSON), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				substitutedGlobalServiceInstances := Configuration(substitutedAssignedTenantConfigurationWithGlobalInstancesWithJsonpaths).GetGlobalServiceInstances(inboundCommunicationKey).ToArray()
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, smLabelsThatHaveAssignmentID(assignmentID)).Return(serviceInstancesIDs, nil).Once()
				// Delete All Instances and Bindings
				client.On("RetrieveMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs}).Return(serviceInstancesBindingsIDs, nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesBindingsIDs).Return(nil).Once()
				client.On("DeleteMultipleResourcesByIDs", mock.Anything, region, subaccount, mock.Anything, serviceInstancesIDs).Return(nil).Once()

				// First Instance
				firstGlobalServiceInstanceSubstituted := substitutedGlobalServiceInstances[0]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: firstGlobalServiceInstanceSubstituted.GetService()}).Return(serviceOfferingIDs[0], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: firstGlobalServiceInstanceSubstituted.GetPlan(), OfferingID: serviceOfferingIDs[0]}).Return(servicePlanIDs[0], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(firstGlobalServiceInstanceSubstituted.GetName(), servicePlanIDs[0], assignmentID, firstGlobalServiceInstanceSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[0]}).Return(firstGlobalServiceInstanceSubstituted.WithName(serviceInstancesNames[0]).ToJSONRawMessage(), nil).Once()

				firstGlobalServiceInstanceBindingSubstituted := firstGlobalServiceInstanceSubstituted.GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(firstGlobalServiceInstanceBindingSubstituted.GetName(), serviceInstancesIDs[0], firstGlobalServiceInstanceBindingSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[0]}).Return(firstGlobalServiceInstanceBindingSubstituted.WithName(serviceInstanceBindingsNames[0]).ToJSONRawMessage(), nil).Once()

				// Second Instance
				secondGlobalServiceInstanceSubstituted := substitutedGlobalServiceInstances[1]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: secondGlobalServiceInstanceSubstituted.GetService()}).Return(serviceOfferingIDs[1], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: secondGlobalServiceInstanceSubstituted.GetPlan(), OfferingID: serviceOfferingIDs[1]}).Return(servicePlanIDs[1], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(secondGlobalServiceInstanceSubstituted.GetName(), servicePlanIDs[1], assignmentID, secondGlobalServiceInstanceSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[1]}).Return(secondGlobalServiceInstanceSubstituted.WithName(serviceInstancesNames[1]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstanceBindingSubstituted := secondGlobalServiceInstanceSubstituted.GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(secondGlobalServiceInstanceBindingSubstituted.GetName(), serviceInstancesIDs[1], secondGlobalServiceInstanceBindingSubstituted.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[1]}).Return(secondGlobalServiceInstanceBindingSubstituted.WithName(serviceInstanceBindingsNames[1]).ToJSONRawMessage(), nil).Once()

				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(expectedResponseForGlobalInstances)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances and with service details in receiver tenant inbound communication - check that the inbound is deleted",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, receiverTenantConfigurationWithServiceInstanceDetails), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				globalServiceInstances := Configuration(assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths).GetGlobalServiceInstances(inboundCommunicationKey).ToArray()
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, smLabelsThatHaveAssignmentID(assignmentID)).Return(nil, nil).Once()

				firstGlobalServiceInstance := globalServiceInstances[0]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: firstGlobalServiceInstance.GetService()}).Return(serviceOfferingIDs[0], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: firstGlobalServiceInstance.GetPlan(), OfferingID: serviceOfferingIDs[0]}).Return(servicePlanIDs[0], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(firstGlobalServiceInstance.GetName(), servicePlanIDs[0], assignmentID, firstGlobalServiceInstance.GetParameters()), mock.Anything).Return(serviceInstancesIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[0]}).Return(firstGlobalServiceInstance.WithName(serviceInstancesNames[0]).ToJSONRawMessage(), nil).Once()

				firstGlobalServiceInstanceBinding := firstGlobalServiceInstance.GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(firstGlobalServiceInstanceBinding.GetName(), serviceInstancesIDs[0], firstGlobalServiceInstanceBinding.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[0]}).Return(firstGlobalServiceInstanceBinding.WithName(serviceInstanceBindingsNames[0]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstance := globalServiceInstances[1]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: secondGlobalServiceInstance.GetService()}).Return(serviceOfferingIDs[1], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: secondGlobalServiceInstance.GetPlan(), OfferingID: serviceOfferingIDs[1]}).Return(servicePlanIDs[1], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(secondGlobalServiceInstance.GetName(), servicePlanIDs[1], assignmentID, secondGlobalServiceInstance.GetParameters()), mock.Anything).Return(serviceInstancesIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[1]}).Return(secondGlobalServiceInstance.WithName(serviceInstancesNames[1]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstanceBinding := globalServiceInstances[1].GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(secondGlobalServiceInstanceBinding.GetName(), serviceInstancesIDs[1], secondGlobalServiceInstanceBinding.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[1]}).Return(secondGlobalServiceInstanceBinding.WithName(serviceInstanceBindingsNames[1]).ToJSONRawMessage(), nil).Once()

				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(expectedResponseForGlobalInstances)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances and with service details in receiver tenant inbound communication - check that the inbound is left only with auth methods without service instance details",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, receiverTenantConfigurationWithServiceInstanceDetailsAndMethodWithoutInstances), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				globalServiceInstances := Configuration(assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths).GetGlobalServiceInstances(inboundCommunicationKey).ToArray()
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, smLabelsThatHaveAssignmentID(assignmentID)).Return(nil, nil).Once()

				firstGlobalServiceInstance := globalServiceInstances[0]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: firstGlobalServiceInstance.GetService()}).Return(serviceOfferingIDs[0], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: firstGlobalServiceInstance.GetPlan(), OfferingID: serviceOfferingIDs[0]}).Return(servicePlanIDs[0], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(firstGlobalServiceInstance.GetName(), servicePlanIDs[0], assignmentID, firstGlobalServiceInstance.GetParameters()), mock.Anything).Return(serviceInstancesIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[0]}).Return(firstGlobalServiceInstance.WithName(serviceInstancesNames[0]).ToJSONRawMessage(), nil).Once()

				firstGlobalServiceInstanceBinding := firstGlobalServiceInstance.GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(firstGlobalServiceInstanceBinding.GetName(), serviceInstancesIDs[0], firstGlobalServiceInstanceBinding.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[0]}).Return(firstGlobalServiceInstanceBinding.WithName(serviceInstanceBindingsNames[0]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstance := globalServiceInstances[1]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: secondGlobalServiceInstance.GetService()}).Return(serviceOfferingIDs[1], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: secondGlobalServiceInstance.GetPlan(), OfferingID: serviceOfferingIDs[1]}).Return(servicePlanIDs[1], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(secondGlobalServiceInstance.GetName(), servicePlanIDs[1], assignmentID, secondGlobalServiceInstance.GetParameters()), mock.Anything).Return(serviceInstancesIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[1]}).Return(secondGlobalServiceInstance.WithName(serviceInstancesNames[1]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstanceBinding := globalServiceInstances[1].GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(secondGlobalServiceInstanceBinding.GetName(), serviceInstancesIDs[1], secondGlobalServiceInstanceBinding.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[1]}).Return(secondGlobalServiceInstanceBinding.WithName(serviceInstanceBindingsNames[1]).ToJSONRawMessage(), nil).Once()

				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(expectedResponseForGlobalInstancesWithInbound)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success - Operation is assign and there are only global service instances and with service details in receiver tenant inbound communication - check that the inbound is left only with auth methods without service instance details",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, receiverTenantConfigurationWithServiceInstanceDetailsAndMethodWithoutInstancesAndReversePaths), fmt.Sprintf(assignedTenantFormatter, assignmentID, assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				globalServiceInstances := Configuration(assignedTenantConfigurationWithGlobalInstancesWithoutJsonpaths).GetGlobalServiceInstances(inboundCommunicationKey).ToArray()
				client.On("RetrieveMultipleResourcesIDsByLabels", mock.Anything, region, subaccount, mock.Anything, smLabelsThatHaveAssignmentID(assignmentID)).Return(nil, nil).Once()

				firstGlobalServiceInstance := globalServiceInstances[0]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: firstGlobalServiceInstance.GetService()}).Return(serviceOfferingIDs[0], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: firstGlobalServiceInstance.GetPlan(), OfferingID: serviceOfferingIDs[0]}).Return(servicePlanIDs[0], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(firstGlobalServiceInstance.GetName(), servicePlanIDs[0], assignmentID, firstGlobalServiceInstance.GetParameters()), mock.Anything).Return(serviceInstancesIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[0]}).Return(firstGlobalServiceInstance.WithName(serviceInstancesNames[0]).ToJSONRawMessage(), nil).Once()

				firstGlobalServiceInstanceBinding := firstGlobalServiceInstance.GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(firstGlobalServiceInstanceBinding.GetName(), serviceInstancesIDs[0], firstGlobalServiceInstanceBinding.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[0], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[0]}).Return(firstGlobalServiceInstanceBinding.WithName(serviceInstanceBindingsNames[0]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstance := globalServiceInstances[1]
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServiceOfferingMatchParameters{CatalogName: secondGlobalServiceInstance.GetService()}).Return(serviceOfferingIDs[1], nil).Once()
				client.On("RetrieveResource", mock.Anything, region, subaccount, mock.Anything, &types.ServicePlanMatchParameters{PlanName: secondGlobalServiceInstance.GetPlan(), OfferingID: serviceOfferingIDs[1]}).Return(servicePlanIDs[1], nil).Once()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceInstanceReqBody(secondGlobalServiceInstance.GetName(), servicePlanIDs[1], assignmentID, secondGlobalServiceInstance.GetParameters()), mock.Anything).Return(serviceInstancesIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceInstance{ID: serviceInstancesIDs[1]}).Return(secondGlobalServiceInstance.WithName(serviceInstancesNames[1]).ToJSONRawMessage(), nil).Once()

				secondGlobalServiceInstanceBinding := globalServiceInstances[1].GetServiceBinding()
				client.On("CreateResource", mock.Anything, region, subaccount, serviceBindingReqBody(secondGlobalServiceInstanceBinding.GetName(), serviceInstancesIDs[1], secondGlobalServiceInstanceBinding.GetParameters()), mock.Anything).Return(serviceInstancesBindingsIDs[1], nil).Once()
				client.On("RetrieveRawResourceByID", mock.Anything, region, subaccount, &types.ServiceKey{ID: serviceInstancesBindingsIDs[1]}).Return(secondGlobalServiceInstanceBinding.WithName(serviceInstanceBindingsNames[1]).ToJSONRawMessage(), nil).Once()

				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(expectedResponseForGlobalInstancesWithInboundAndReverse)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//GIVEN
			smClient := &automock.Client{}
			if testCase.smClientFn != nil {
				smClient = testCase.smClientFn()
			}
			mtlsClient := &automock.MtlsHTTPClient{}
			if testCase.mtlsClientFn != nil {
				mtlsClient = testCase.mtlsClientFn()
			}
			defer mock.AssertExpectationsForObjects(t, smClient)

			req, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(testCase.requestBody)))
			require.NoError(t, err)
			req.Header.Set("Location", statusUrl)

			h := handler.NewHandler(smClient, mtlsClient)
			recorder := httptest.NewRecorder()

			//WHEN
			h.HandlerFunc(recorder, req)
			resp := recorder.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, testCase.expectedResponseCode, resp.StatusCode, string(body))
			require.Eventually(t, func() bool {
				return mtlsClient.AssertExpectations(t)
			}, time.Second*15, 50*time.Millisecond)
		})
	}
}

func requestThatHasBody(expectedBody string) interface{} {
	return mock.MatchedBy(func(actualReq *http.Request) bool {
		bytes, err := io.ReadAll(actualReq.Body)
		if err != nil {
			return false
		}
		fmt.Printf("Expected Body %q\n", string(bytes))
		return strings.Contains(string(bytes), expectedBody)
	})
}

func fixHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func smLabelsThatHaveAssignmentID(expectedAssignmentID string) interface{} {
	return mock.MatchedBy(func(actualMap map[string][]string) bool {
		actualLabel, ok := actualMap[assignmentIDKey]
		return ok && len(actualLabel) == 1 && actualLabel[0] == expectedAssignmentID
	})
}

func serviceInstanceReqBody(name, planID, assignmentID, parameters string) interface{} {
	return mock.MatchedBy(func(actualReqBody *types.ServiceInstanceReqBody) bool {
		actualLabel, ok := actualReqBody.Labels[assignmentIDKey]
		return (name == "" || name == actualReqBody.Name) &&
			planID == actualReqBody.ServicePlanID &&
			parameters == string(actualReqBody.Parameters) &&
			ok && len(actualLabel) == 1 && assignmentID == actualLabel[0]
	})
}

func serviceBindingReqBody(name, serviceInstanceID, parameters string) interface{} {
	return mock.MatchedBy(func(actualReqBody *types.ServiceKeyReqBody) bool {
		return (name == "" || name == actualReqBody.Name) &&
			parameters == string(actualReqBody.Parameters) &&
			serviceInstanceID == actualReqBody.ServiceKeyID
	})
}

type Configuration string
type ServiceInstances string
type ServiceInstance string
type ServiceBinding string

func (c Configuration) GetCommunication(typeCommunication string) string {
	return gjson.Get(string(c), tenantmapping.FindKeyPath(gjson.Parse(string(c)).Value(), typeCommunication)).String()
}

func (c Configuration) GetGlobalServiceInstances(typeCommunication string) ServiceInstances {
	return ServiceInstances(gjson.Get(string(c), fmt.Sprintf("%s.%s", tenantmapping.FindKeyPath(gjson.Parse(string(c)).Value(), typeCommunication), serviceInstancesKey)).String())
}

func (c Configuration) GetLocalServiceInstances(typeCommunication, authMethod string) ServiceInstances {
	return ServiceInstances(gjson.Get(string(c), fmt.Sprintf("%s.%s.%s", tenantmapping.FindKeyPath(gjson.Parse(string(c)).Value(), typeCommunication), authMethod, serviceInstancesKey)).String())
}

func (sis ServiceInstances) ToArray() []ServiceInstance {
	arr := gjson.Parse(string(sis)).Array()
	result := make([]ServiceInstance, 0, len(arr))
	for _, el := range arr {
		result = append(result, ServiceInstance(el.String()))
	}
	return result
}

func (sis ServiceInstances) ToString() string {
	return string(sis)
}

func (si ServiceInstance) GetService() string {
	return gjson.Get(string(si), serviceInstanceServiceKey).String()
}

func (si ServiceInstance) GetPlan() string {
	return gjson.Get(string(si), serviceInstancePlanKey).String()
}

func (si ServiceInstance) GetName() string {
	return gjson.Get(string(si), nameKey).String()
}

func (si ServiceInstance) GetParameters() string {
	return gjson.Get(string(si), parametersKey).String()
}

func (si ServiceInstance) GetServiceBinding() ServiceBinding {
	return ServiceBinding(gjson.Get(string(si), serviceBindingKey).String())
}

func (si ServiceInstance) WithName(name string) ServiceInstance {
	instanceWithName, _ := sjson.Set(string(si), nameKey, name)
	return ServiceInstance(instanceWithName)
}

func (si ServiceInstance) ToJSONRawMessage() json.RawMessage {
	return []byte(si)
}

func (sb ServiceBinding) GetParameters() string {
	return gjson.Get(string(sb), parametersKey).String()
}

func (sb ServiceBinding) GetName() string {
	return gjson.Get(string(sb), nameKey).String()
}

func (sb ServiceBinding) WithName(name string) ServiceBinding {
	bindingWithName, _ := sjson.Set(string(sb), nameKey, name)
	return ServiceBinding(bindingWithName)
}

func (sb ServiceBinding) ToJSONRawMessage() json.RawMessage {
	return []byte(sb)
}
