package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types/tenantmapping"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"net/http"
	"regexp"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
)

const (
	readyState         = "READY"
	createErrorState   = "CREATE_ERROR"
	deleteErrorState   = "DELETE_ERROR"
	configPendingState = "CONFIG_PENDING"

	assignOperation = "assign"

	inboundCommunicationPath  = "inboundCommunication"
	serviceInstancesKey       = "serviceInstances"
	serviceBindingKey         = "serviceBinding"
	serviceInstanceServiceKey = "service"
	serviceInstancePlanKey    = "plan"
	parametersKey             = "parameters"
	nameKey                   = "name"
	assignmentIDKey           = "assignment_id"
	currentWaveHashKey        = "current_wave_hash"

	locationHeader             = "Location"
	contentTypeHeaderKey       = "Content-Type"
	contentTypeApplicationJSON = "application/json;charset=UTF-8"
)

// Client is used to call SM
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	RetrieveResource(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceMatchParams resources.ResourceMatchParameters) (string, error)
	RetrieveResourceByID(ctx context.Context, region, subaccountID string, resource resources.Resource) (resources.Resource, error)
	RetrieveRawResourceByID(ctx context.Context, region, subaccountID string, resource resources.Resource) (map[string]interface{}, error)
	RetrieveMultipleResources(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceMatchParams resources.ResourceMatchParameters) ([]string, error)
	RetrieveMultipleResourcesIDsByLabels(ctx context.Context, region, subaccountID string, resources resources.Resources, labels map[string][]string) ([]string, error)
	CreateResource(ctx context.Context, region, subaccountID string, resourceReqBody resources.ResourceRequestBody, resource resources.Resource) (string, error)
	DeleteResource(ctx context.Context, region, subaccountID string, resource resources.Resource) error
	DeleteMultipleResources(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceMatchParams resources.ResourceMatchParameters) error
	DeleteMultipleResourcesByIDs(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceIDs []string) error
}

//go:generate mockery --exported --name=mtlsHTTPClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type mtlsHTTPClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// InstanceCreatorHandler processes received requests
type InstanceCreatorHandler struct {
	SMClient       Client
	mtlsHTTPClient mtlsHTTPClient
}

// NewHandler creates an InstanceCreatorHandler
func NewHandler(smClient Client, mtlsHTTPClient mtlsHTTPClient) *InstanceCreatorHandler {
	return &InstanceCreatorHandler{
		SMClient:       smClient,
		mtlsHTTPClient: mtlsHTTPClient,
	}
}

// HandlerFunc is the implementation of InstanceCreatorHandler
func (i InstanceCreatorHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.C(ctx).Info("Instance Creator Handler was hit...")

	uclStatusAPIUrl := r.Header.Get(locationHeader)

	// respond with 202 to the UCL call
	httputils.Respond(w, http.StatusAccepted)

	correlationID := correlation.CorrelationIDFromContext(ctx)

	log.C(ctx).Info("Instance Creator Handler handles instance creation...")
	go i.handleInstances(r, uclStatusAPIUrl, correlationID)
}

func (i *InstanceCreatorHandler) handleInstances(r *http.Request, statusAPIURL, correlationID string) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// Decoding the body
	log.C(ctx).Info("Decoding the request body...")
	var reqBody tenantmapping.Body
	if err := decodeJSONBody(r, &reqBody); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, "", err)
		} else {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, "", errors.Wrap(err, "while decoding json request body"))
		}
		return
	}

	// Validating the body
	log.C(ctx).Info("Validating tenant mapping request body...")
	if err := reqBody.Validate(); err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, "", errors.Wrapf(err, "while validating the request body"))
		return
	}

	if reqBody.Context.Operation == assignOperation {
		i.handleInstanceCreation(ctx, &reqBody, statusAPIURL, correlationID)
	} else {
		i.handleInstanceDeletion(ctx, &reqBody, statusAPIURL, correlationID)
	}
}

// Core Instance Creation Logic
func (i *InstanceCreatorHandler) handleInstanceCreation(ctx context.Context, reqBody *tenantmapping.Body, statusAPIURL, correlationID string) {
	// If receiver tenant outboundCommunication is missing - create it. It's here because if this code fails it's better to be before the instances are created.
	if err := reqBody.AddReceiverTenantOutboundCommunicationIfMissing(); err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, err)
		return
	}

	assignedTenantInboundCommunication, err := reqBody.GetAssignedTenantInboundCommunication()
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while getting assigned tenant configuration inbound communication"))
		return
	}

	//region := reqBody.ReceiverTenant.Region
	//subaccount := reqBody.ReceiverTenant.SubaccountID

	serviceInstancesPath := tenantmapping.FindKeyPath(assignedTenantInboundCommunication, serviceInstancesKey)
	if serviceInstancesPath == "" {
		i.reportToUCLWithSuccess(ctx, statusAPIURL, correlationID, configPendingState, fmt.Sprintf("Service instances details are missing. Returning %q...", configPendingState), []byte(""))
		return
	}

	var assignedTenantConfiguration interface{}
	if err = json.Unmarshal(reqBody.AssignedTenant.Configuration, &assignedTenantConfiguration); err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrap(err, "while unmarshalling assigned tenant configuration to interface"))
		return
	}

	// Go through global service instances(if exist)
	if globalServiceInstances, globalServiceInstancesExist := assignedTenantInboundCommunication[serviceInstancesKey]; globalServiceInstancesExist {
		// Handle global service instances creation
		var globalServiceInstancesInterface []interface{}
		if err = json.Unmarshal(globalServiceInstances, &globalServiceInstancesInterface); err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrap(err, "while unmarshalling global instances in interface"))
			return
		}

		createdGlobalServiceInstances, err := i.createServiceInstances(ctx, reqBody, globalServiceInstancesInterface, assignedTenantConfiguration, inboundCommunicationPath)
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.New("while creating service instances"))
			return
		}

		createdGlobalServiceInstancesJSON, err := json.Marshal(createdGlobalServiceInstances)
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.New("while marshalling created global service instances"))
			return
		}

		// Save the service instances
		assignedTenantInboundCommunication[serviceInstancesKey] = createdGlobalServiceInstancesJSON
	}

	// Go through every type of inbound communication(auth)
	for auth, assignedTenantAuth := range assignedTenantInboundCommunication {
		// Skip if the auth don't have service instances
		if gjson.GetBytes(assignedTenantAuth, serviceInstancesKey).Exists() == false {
			continue
		}

		// Convert the assignedTenantAuth to map[string]interface{}
		var assignedTenantAuthMap map[string]interface{}
		if err := json.Unmarshal(assignedTenantAuth, &assignedTenantAuthMap); err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while unmarshalling assigned tenant %q auth in map of string to interface", auth))
			return
		}

		serviceInstancesArray, ok := assignedTenantAuthMap[serviceInstancesKey].([]interface{})
		if !ok {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.New("while casting service instances interface to array"))
			return
		}

		createdServiceInstances, err := i.createServiceInstances(ctx, reqBody, serviceInstancesArray, assignedTenantConfiguration, auth)
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.New("while creating service instances"))
			return
		}

		// Save the service instances
		assignedTenantAuthMap[serviceInstancesKey] = createdServiceInstances
		assignedTenantAuthMapMarshalled, err := json.Marshal(assignedTenantAuthMap)
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Errorf("while marshalling %q auth", auth))
			return
		}
		assignedTenantInboundCommunication[auth] = assignedTenantAuthMapMarshalled

		//// Go through all service instances(we know they exist because we checked with the if at the start of the for loop)
		//serviceInstancesArray, ok := assignedTenantAuthMap[serviceInstancesKey].([]interface{})
		//if !ok {
		//	i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.New("while casting service instances interface to array"))
		//	return
		//}
		//
		//for idx, serviceInstance := range serviceInstancesArray {
		//	serviceInstanceMap, ok := serviceInstance.(map[string]interface{})
		//	if !ok {
		//		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.New("while casting service instance interface to map"))
		//		return
		//	}
		//
		//	// Set the Service Instance name to service-plan-assignmentId
		//	//serviceInstanceName := strings.Join([]string{serviceInstanceMap[serviceInstanceServicePath].(string), serviceInstanceMap[serviceInstancePlanPath].(string), reqBody.AssignedTenant.AssignmentID}, "-")
		//	serviceInstanceName := getServiceInstanceName(serviceInstanceMap)
		//	serviceInstanceMap[nameKey] = serviceInstanceName
		//
		//	// Save the service instance binding in a var and delete it from the service instance map.
		//	// That way later we will be able to substitute only the service instance json paths
		//	serviceInstanceBinding := serviceInstanceMap[serviceBindingKey]
		//	delete(serviceInstanceMap, serviceBindingKey)
		//
		//	// Substitute the service instance json paths(without the service binding)
		//	substituteJSON(serviceInstanceMap, assignedTenantConfiguration)
		//
		//	// Extract Service Offering catalog name and Service Plan catalog name from the service instance
		//	serviceOfferingCatalogName, servicePlanCatalogName, err := extractServiceOfferingAndPlanCatalogNames(serviceInstance)
		//	if err != nil {
		//		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, err)
		//		return
		//	}
		//
		//	serviceInstanceParameters, err := extractParameters(serviceInstance)
		//	if err != nil {
		//		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while extracting the parameters of service instance with id: %d", idx))
		//		return
		//	}
		//
		//	// Get the Service Offering ID with catalog name(service from the contract)
		//	serviceOfferingID, err := i.SMClient.RetrieveResource(ctx, region, subaccount, &types.ServiceOfferings{}, &types.ServiceOfferingMatchParameters{CatalogName: serviceOfferingCatalogName})
		//	if err != nil {
		//		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Errorf("while retrieving service offering with catalog name %q", "catalogName"))
		//		return
		//	}
		//
		//	// Get the Service Plan ID with the Service Offering ID + Service Plan Catalog Name(plan from the contract)
		//	servicePlanID, err := i.SMClient.RetrieveResource(ctx, region, subaccount, &types.ServicePlans{}, &types.ServicePlanMatchParameters{PlanName: servicePlanCatalogName, OfferingID: serviceOfferingID})
		//	if err != nil {
		//		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Errorf("while retrieving service plan with catalog name %q and offering ID %q", "catalogName", serviceOfferingID))
		//		return
		//	}
		//
		//	// Create the service instance - params from the contract
		//	serviceInstanceID, err := i.SMClient.CreateResource(ctx, region, subaccount, &types.ServiceInstanceReqBody{Name: serviceInstanceName, ServicePlanID: servicePlanID, Parameters: serviceInstanceParameters, Labels: map[string][]string{assignmentIDKey: {reqBody.AssignedTenant.AssignmentID}}}, &types.ServiceInstance{})
		//	instanceAlreadyExists := strings.Contains(err.Error(), fmt.Sprintf("status: %d", http.StatusConflict))
		//	if err != nil && !instanceAlreadyExists {
		//		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Errorf("while creating service instance with name %q", serviceInstanceName))
		//		return
		//	}
		//
		//	// Recreating it if it already exists - resync case
		//	if instanceAlreadyExists {
		//		if err = i.recreateServiceInstance(ctx, region, subaccount, serviceInstanceName, servicePlanID, serviceInstanceParameters); err != nil {
		//			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Errorf("while recreating service instance with name %q", serviceInstanceName))
		//			return
		//		}
		//	}
		//
		//	// Retrieve the service instance by ID
		//	serviceInstanceRaw, err := i.SMClient.RetrieveRawResourceByID(ctx, region, subaccount, &types.ServiceInstance{ID: serviceInstanceID})
		//	if err != nil {
		//		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Errorf("while retrieving service instance with ID %q", serviceInstanceID))
		//		return
		//	}
		//
		//	// Save the service instance
		//	serviceInstancesArray[idx] = serviceInstanceRaw
		//
		//	// Substitute the service binding json paths
		//	substituteJSON(serviceInstanceBinding, assignedTenantConfiguration)
		//
		//	serviceBindingParameters, err := extractParameters(serviceInstanceBinding)
		//	if err != nil {
		//		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while extracting the parameters of service binding for a service instance with id: %d", idx))
		//		return
		//	}
		//
		//	// Create the service instance binding
		//	serviceBindingID, err := i.SMClient.CreateResource(ctx, region, subaccount, &types.ServiceKeyReqBody{Name: serviceInstanceName, ServiceKeyID: serviceInstanceID, Parameters: serviceBindingParameters}, &types.ServiceKey{})
		//	if err != nil {
		//		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Errorf("while creating service instance binding for service instance with ID %q", serviceInstanceID))
		//		return
		//	}
		//
		//	// Retrieve the service instance binding by ID
		//	serviceBindingRaw, err := i.SMClient.RetrieveRawResourceByID(ctx, region, subaccount, &types.ServiceInstance{ID: serviceBindingID})
		//	if err != nil {
		//		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Errorf("while retrieving service instance binding with ID %q", serviceBindingID))
		//		return
		//	}
		//
		//	// Save the service instance binding
		//	serviceInstanceRaw[serviceBindingKey] = serviceBindingRaw
		//	serviceInstance = serviceInstanceRaw

		// Substitute the top most json paths
		substituteJSON(assignedTenantAuthMap, assignedTenantConfiguration)

		// Remove the service instances from the json
		//delete(assignedTenantAuthMap, serviceInstancesPath)

		// Update the Receiver Tenant outbound communication with the Assigned Tenant inbound communication.
		// If the same auth exists in the receiver tenant, merge the two jsons.
		// If the auth doesn't exist - add it.
		receiverTenantOutboundCommunication, err := reqBody.GetReceiverTenantOutboundCommunication()
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while getting receiver tenant outbound communication"))
			return
		}

		mergedConfiguration := make(map[string]interface{}, len(receiverTenantOutboundCommunication[auth])+len(assignedTenantAuthMap))
		gjson.ParseBytes(receiverTenantOutboundCommunication[auth]).ForEach(func(key, value gjson.Result) bool {
			mergedConfiguration[key.Str] = value.Raw
			return true
		})
		for k, v := range assignedTenantAuthMap {
			mergedConfiguration[k] = v
		}

		if err := reqBody.SetReceiverTenantAuth(auth, mergedConfiguration); err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while updating receiver tenant %q auth", auth))
			return
		}
	}

	responseConfig, err := json.Marshal(reqBody.ReceiverTenant.Configuration)
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrap(err, "while marshalling receiver tenant configuration"))
		return
	}

	// Report to UCL with success
	i.reportToUCLWithSuccess(ctx, statusAPIURL, correlationID, readyState, "Successfully processed Service Instance creation.", responseConfig)
}

// Core Instance Deletion Logic
func (i *InstanceCreatorHandler) handleInstanceDeletion(ctx context.Context, reqBody *tenantmapping.Body, statusAPIURL, correlationID string) {
	assignmentID := reqBody.AssignedTenant.AssignmentID

	// Get the service instance with labels using the formation-id label key
	serviceInstanceIDs, err := i.SMClient.RetrieveMultipleResourcesIDsByLabels(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceInstances{}, map[string][]string{assignmentIDKey: {assignmentID}})
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, deleteErrorState, errors.Wrapf(err, "while retrieving service instances for assignmentID: %q", assignmentID))
		return
	}

	// Retrieve all service instances bindings
	serviceBindingsIDs, err := i.SMClient.RetrieveMultipleResources(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceKeys{}, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstanceIDs})
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, deleteErrorState, errors.Wrapf(err, "while retrieving service bindings for service instaces with IDs: %v", serviceInstanceIDs))
		return
	}

	// Delete all service instances bindings for the service instances
	if err = i.SMClient.DeleteMultipleResourcesByIDs(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceKeys{}, serviceBindingsIDs); err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, deleteErrorState, errors.Wrapf(err, "while deleting service bindings with IDs: %v", serviceBindingsIDs))
		return
	}

	// Delete all service instances with serviceInstanceIDs
	if err := i.SMClient.DeleteMultipleResourcesByIDs(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceInstances{}, serviceInstanceIDs); err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, deleteErrorState, errors.Wrapf(err, "while deleting service instances with IDs: %v", serviceInstanceIDs))
		return
	}

	// Report to UCL with success
	i.reportToUCLWithSuccess(ctx, statusAPIURL, correlationID, readyState, "Successfully processed Service Instance deletion.", nil)
}

func (i *InstanceCreatorHandler) callUCLStatusAPI(statusAPIURL, correlationID string, response interface{}) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	correlationIDKey := correlation.RequestIDHeaderKey
	ctx = correlation.SaveCorrelationIDHeaderToContext(ctx, &correlationIDKey, &correlationID)

	logger := log.C(ctx).WithField(correlationIDKey, correlationID)
	ctx = log.ContextWithLogger(ctx, logger)

	reqBodyBytes, err := json.Marshal(response)
	if err != nil {
		log.C(ctx).WithError(err).Error("error while marshalling request body")
		return
	}

	if statusAPIURL == "" {
		log.C(ctx).WithError(err).Error("status API URL is empty...")
		return
	}

	req, err := http.NewRequest(http.MethodPatch, statusAPIURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		log.C(ctx).WithError(err).Error("error while building status API request")
		return
	}
	req.Header.Set(contentTypeHeaderKey, contentTypeApplicationJSON)
	req = req.WithContext(ctx)

	resp, err := i.mtlsHTTPClient.Do(req)
	if err != nil {
		log.C(ctx).WithError(err).Error("error while executing request to the status API")
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		log.C(ctx).WithError(err).Errorf("status API returned unexpected non OK status code: %d", resp.StatusCode)
		return
	}
}

func (i *InstanceCreatorHandler) createServiceInstances(ctx context.Context, reqBody *tenantmapping.Body, serviceInstancesArray []interface{}, assignedTenantConfiguration interface{}, parentScope string) ([]interface{}, error) {
	region := reqBody.ReceiverTenant.Region
	subaccount := reqBody.ReceiverTenant.SubaccountID

	for idx, serviceInstance := range serviceInstancesArray {
		serviceInstanceMap, ok := serviceInstance.(map[string]interface{})
		if !ok {
			return nil, errors.New("while casting service instance interface to map")
		}

		// Get current wave hash
		currentWaveHash, err := getCurrentWaveHash(parentScope, serviceInstancesArray)
		if err != nil {
			return nil, errors.Wrap(err, "while getting current wave hash")
		}

		// Get the service instances from this wave, if they exist - it is resync case, and we need to delete them, so we can create them from scratch
		smLabels := map[string][]string{
			assignmentIDKey:    {reqBody.AssignedTenant.AssignmentID},
			currentWaveHashKey: {currentWaveHash},
		}
		existentServiceInstancesIDs, err := i.SMClient.RetrieveMultipleResourcesIDsByLabels(ctx, region, subaccount, &types.ServiceInstances{}, smLabels)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting service instances for the current wave with hash: %q", currentWaveHash)
		}

		if len(existentServiceInstancesIDs) > 0 {
			// Retrieve all service instances bindings
			serviceBindingsIDs, err := i.SMClient.RetrieveMultipleResources(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceKeys{}, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: existentServiceInstancesIDs})
			if err != nil {
				return nil, errors.Wrapf(err, "while retrieving service bindings for service instaces with IDs: %v", existentServiceInstancesIDs)
			}

			// Delete all service instances bindings for the service instances
			if err = i.SMClient.DeleteMultipleResourcesByIDs(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceKeys{}, serviceBindingsIDs); err != nil {
				return nil, errors.Wrapf(err, "while deleting service bindings with IDs: %v", serviceBindingsIDs)
			}

			// Delete all service instances with serviceInstanceIDs
			if err := i.SMClient.DeleteMultipleResourcesByIDs(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceInstances{}, existentServiceInstancesIDs); err != nil {
				return nil, errors.Wrapf(err, "while deleting service instances with IDs: %v", existentServiceInstancesIDs)
			}
		}

		// Set the Service Instance name to service-plan-assignmentId
		serviceInstanceName := getServiceInstanceName(serviceInstanceMap)
		serviceInstanceMap[nameKey] = serviceInstanceName

		// Save the service instance binding in a var and delete it from the service instance map.
		// That way later we will be able to substitute only the service instance json paths
		serviceInstanceBinding := serviceInstanceMap[serviceBindingKey]
		delete(serviceInstanceMap, serviceBindingKey)

		// Substitute the service instance json paths(without the service binding)
		substituteJSON(serviceInstanceMap, assignedTenantConfiguration)

		// Extract Service Offering catalog name and Service Plan catalog name from the service instance
		serviceOfferingCatalogName, servicePlanCatalogName, err := extractServiceOfferingAndPlanCatalogNames(serviceInstance)
		if err != nil {
			return nil, err
		}

		serviceInstanceParameters, err := extractParameters(serviceInstance)
		if err != nil {
			return nil, errors.Wrapf(err, "while extracting the parameters of service instance with id: %d", idx)
		}

		// Get the Service Offering ID with catalog name(service from the contract)
		serviceOfferingID, err := i.SMClient.RetrieveResource(ctx, region, subaccount, &types.ServiceOfferings{}, &types.ServiceOfferingMatchParameters{CatalogName: serviceOfferingCatalogName})
		if err != nil {
			return nil, errors.Errorf("while retrieving service offering with catalog name %q", "catalogName")
		}

		// Get the Service Plan ID with the Service Offering ID + Service Plan Catalog Name(plan from the contract)
		servicePlanID, err := i.SMClient.RetrieveResource(ctx, region, subaccount, &types.ServicePlans{}, &types.ServicePlanMatchParameters{PlanName: servicePlanCatalogName, OfferingID: serviceOfferingID})
		if err != nil {
			return nil, errors.Errorf("while retrieving service plan with catalog name %q and offering ID %q", "catalogName", serviceOfferingID)
		}

		// Create the service instance - params from the contract
		serviceInstanceID, err := i.SMClient.CreateResource(ctx, region, subaccount, &types.ServiceInstanceReqBody{Name: serviceInstanceName, ServicePlanID: servicePlanID, Parameters: serviceInstanceParameters, Labels: smLabels}, &types.ServiceInstance{})
		if err != nil {
			return nil, errors.Errorf("while creating service instance with name %q", serviceInstanceName)
		}

		// Retrieve the service instance by ID
		serviceInstanceRaw, err := i.SMClient.RetrieveRawResourceByID(ctx, region, subaccount, &types.ServiceInstance{ID: serviceInstanceID})
		if err != nil {
			return nil, errors.Errorf("while retrieving service instance with ID %q", serviceInstanceID)
		}

		// Save the service instance
		serviceInstancesArray[idx] = serviceInstanceRaw

		// Substitute the service binding json paths
		// TODO:: THIS CONFIG WILL NOT HAVE THE NEW SERVICE INSTANCE INFO IN IT
		substituteJSON(serviceInstanceBinding, assignedTenantConfiguration)

		serviceBindingParameters, err := extractParameters(serviceInstanceBinding)
		if err != nil {
			return nil, errors.Wrapf(err, "while extracting the parameters of service binding for a service instance with id: %d", idx)
		}

		// Create the service instance binding
		serviceBindingID, err := i.SMClient.CreateResource(ctx, region, subaccount, &types.ServiceKeyReqBody{Name: serviceInstanceName, ServiceKeyID: serviceInstanceID, Parameters: serviceBindingParameters}, &types.ServiceKey{})
		if err != nil {
			return nil, errors.Errorf("while creating service instance binding for service instance with ID %q", serviceInstanceID)
		}

		// Retrieve the service instance binding by ID
		serviceBindingRaw, err := i.SMClient.RetrieveRawResourceByID(ctx, region, subaccount, &types.ServiceInstance{ID: serviceBindingID})
		if err != nil {
			return nil, errors.Errorf("while retrieving service instance binding with ID %q", serviceBindingID)
		}

		// Save the service instance binding
		serviceInstanceRaw[serviceBindingKey] = serviceBindingRaw
		serviceInstance = serviceInstanceRaw
	}

	return serviceInstancesArray, nil
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("an error has occurred while closing response body: %v", err)
	}
}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

// reportToUCLWithError reports status to the UCL Status API with the JSON error wrapped in an ErrorResponse struct
func (i *InstanceCreatorHandler) reportToUCLWithError(ctx context.Context, statusAPIURL, correlationID string, state string, err error) {
	log.C(ctx).Error(err.Error())
	errorResponse := ErrorResponse{State: state, Message: err.Error()}
	i.callUCLStatusAPI(statusAPIURL, correlationID, errorResponse)
}

// reportToUCLWithSuccess reports status to the UCL Status API with the JSON success wrapped in an SuccessResponse struct
func (i *InstanceCreatorHandler) reportToUCLWithSuccess(ctx context.Context, statusAPIURL, correlationID, state, msg string, configuration json.RawMessage) {
	log.C(ctx).Info(msg)
	successResponse := SuccessResponse{State: state, Configuration: configuration}
	i.callUCLStatusAPI(statusAPIURL, correlationID, successResponse)
}

func substituteJSON(json interface{}, rootMap interface{}) {
	switch json.(type) {
	case map[string]interface{}:
		parseMap(json.(map[string]interface{}), rootMap)
	case []interface{}:
		parseArray(json.([]interface{}), rootMap)
	default:
		fmt.Println("Invalid JSON")
	}
}

func parseMap(aMap map[string]interface{}, rootMap interface{}) {
	for key, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			parseMap(val.(map[string]interface{}), rootMap)
		case []interface{}:
			parseArray(val.([]interface{}), rootMap)
		case string:
			if strings.Contains(concreteVal, "$.") {
				fmt.Printf("Before substituting: %q\n", concreteVal)
				regex := regexp.MustCompile(`\{(\$.*?)\}`)
				matches := regex.FindAllStringSubmatch(concreteVal, -1)
				if len(matches) == 0 { // the value is something like "$.<>"
					substitution, err := jsonpath.Get(concreteVal, rootMap)
					if err != nil {
						fmt.Println(err)
						return
					}
					aMap[key] = substitution
					fmt.Printf("After substituting: %q\n", substitution)
				} else { // the value contains jsonPaths with concatenated static strings. for example, "{$.<>}/string..."
					for _, match := range matches {
						substitution, err := jsonpath.Get(match[1], rootMap)
						if err != nil {
							fmt.Println(err)
							return
						}
						concreteVal = strings.ReplaceAll(concreteVal, match[0], substitution.(string))
					}

					aMap[key] = concreteVal
					fmt.Printf("After substituting: %q\n", concreteVal)
				}
			}
		}
	}
}

func parseArray(anArray []interface{}, rootMap interface{}) {
	for i, val := range anArray {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			parseMap(val.(map[string]interface{}), rootMap)
		case []interface{}:
			parseArray(val.([]interface{}), rootMap)
		case string:
			if strings.Contains(concreteVal, "$.") {
				fmt.Printf("Before substituting: %q\n", concreteVal)
				regex := regexp.MustCompile(`\{(\$.*?)\}`)
				matches := regex.FindAllStringSubmatch(concreteVal, -1)
				if len(matches) == 0 { // the value is something like "$.<>"
					substitution, err := jsonpath.Get(concreteVal, rootMap)
					if err != nil {
						fmt.Println(err)
						return
					}
					anArray[i] = substitution
					fmt.Printf("After substituting: %q\n", substitution)
				} else { // the value contains jsonPaths with concatenated static strings. for example, "{$.<>}/string..."
					for _, match := range matches {
						substitution, err := jsonpath.Get(match[1], rootMap)
						if err != nil {
							fmt.Println(err)
							return
						}
						concreteVal = strings.ReplaceAll(concreteVal, match[0], substitution.(string))
					}

					anArray[i] = concreteVal
					fmt.Printf("After substituting: %q\n", concreteVal)
				}
			}
		}
	}
}

func extractServiceOfferingAndPlanCatalogNames(serviceInstance interface{}) (string, string, error) {
	serviceInstanceJSON, err := json.Marshal(serviceInstance)
	if err != nil {
		return "", "", errors.Wrapf(err, "while marshalling service instance")
	}

	serviceOfferingCatalogName := gjson.GetBytes(serviceInstanceJSON, serviceInstanceServiceKey)
	if !serviceOfferingCatalogName.Exists() {
		return "", "", errors.Errorf("%q is missing in service instance", serviceInstanceServiceKey)
	}
	servicePlanCatalogName := gjson.GetBytes(serviceInstanceJSON, serviceInstancePlanKey)
	if !servicePlanCatalogName.Exists() {
		return "", "", errors.Errorf("%q is missing in service instance", serviceInstancePlanKey)
	}

	return serviceOfferingCatalogName.String(), servicePlanCatalogName.String(), nil
}

func extractParameters(object interface{}) (json.RawMessage, error) {
	objectJSON, err := json.Marshal(object)
	if err != nil {
		return []byte(""), errors.Wrapf(err, "while marshalling object")
	}

	objectParameters := gjson.GetBytes(objectJSON, parametersKey)
	if !objectParameters.Exists() || objectParameters.Raw == "{}" {
		return []byte(""), nil
	}

	objectParametersJSON, err := json.Marshal(objectParameters)
	if err != nil {
		return []byte(""), errors.Wrapf(err, "while marshalling object parameters")
	}

	return objectParametersJSON, nil
}

func getServiceInstanceName(serviceInstanceMap map[string]interface{}) string {
	serviceInstanceName, ok := serviceInstanceMap[nameKey]
	if !ok {
		serviceInstanceName = uuid.New().String()
	}
	return serviceInstanceName.(string)
}

func getCurrentWaveHash(parentScope string, serviceInstances interface{}) (string, error) {
	currentWave := map[string]interface{}{parentScope: serviceInstances}

	// Marshal always orders keys which gives us the same hash
	marshalledInstances, err := json.Marshal(currentWave)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(marshalledInstances)
	res := hex.EncodeToString(sum[0:])
	return res, nil
}
