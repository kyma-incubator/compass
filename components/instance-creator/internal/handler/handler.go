package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types/tenantmapping"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/persistence"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/PaesslerAG/jsonpath"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
)

const (
	readyState         = "READY"
	createErrorState   = "CREATE_ERROR"
	deleteErrorState   = "DELETE_ERROR"
	configPendingState = "CONFIG_PENDING"

	assignOperation = "assign"

	inboundCommunicationKey   = "inboundCommunication"
	serviceInstancesKey       = "serviceInstances"
	serviceBindingKey         = "serviceBinding"
	serviceInstanceServiceKey = "service"
	serviceInstancePlanKey    = "plan"
	configurationKey          = "configuration"
	nameKey                   = "name"
	assignmentIDKey           = "assignment_id"
	currentWaveHashKey        = "current_wave_hash"
	reverseKey                = "reverse"

	locationHeader             = "Location"
	contentTypeHeaderKey       = "Content-Type"
	contentTypeApplicationJSON = "application/json;charset=UTF-8"

	subaccountIsMissingFormatter = "Subaccount with id %s is not known to CIS"
)

// Client is used to call SM
//
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	RetrieveResource(ctx context.Context, region, subaccountID string, resources resources.Resources, resourceMatchParams resources.ResourceMatchParameters) (string, error)
	RetrieveResourceByID(ctx context.Context, region, subaccountID string, resource resources.Resource) (resources.Resource, error)
	RetrieveRawResourceByID(ctx context.Context, region, subaccountID string, resource resources.Resource) (json.RawMessage, error)
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
	connector      persistence.DatabaseConnector
}

// NewHandler creates an InstanceCreatorHandler
func NewHandler(smClient Client, mtlsHTTPClient mtlsHTTPClient, connector persistence.DatabaseConnector) *InstanceCreatorHandler {
	return &InstanceCreatorHandler{
		SMClient:       smClient,
		mtlsHTTPClient: mtlsHTTPClient,
		connector:      connector,
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
		i.handleAssign(ctx, &reqBody, statusAPIURL, correlationID)
	} else {
		i.handleUnassign(ctx, &reqBody, statusAPIURL, correlationID)
	}
}

// Core Instance Creation Logic
func (i *InstanceCreatorHandler) handleAssign(ctx context.Context, reqBody *tenantmapping.Body, statusAPIURL, correlationID string) {
	// Get a single DB session
	connection, err := i.connector.GetConnection(ctx)
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrap(err, "while trying to get database connection"))
		return
	}
	defer func() {
		if err := connection.Close(); err != nil {
			log.C(ctx).WithError(err).Error("Error while closing the database connection")
		}
	}()

	assignmentID := reqBody.AssignedTenant.AssignmentID

	advisoryLocker := connection.GetAdvisoryLocker()

	// This lock prevents multiple assign operations to execute simultaneously
	locked, err := advisoryLocker.TryLock(ctx, assignmentID+reqBody.Context.Operation)
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrap(err, "while trying to acquire postgres advisory lock in the beginning of instance creation"))
		return
	}
	if !locked {
		log.C(ctx).Debugf("Another instance creator is handling %q operation and assignment with ID %q", assignOperation, assignmentID)
		return
	}
	defer func() {
		if err := advisoryLocker.Unlock(ctx, assignmentID+reqBody.Context.Operation); err != nil {
			log.C(ctx).WithError(err).Error("Error while releasing a previously-acquired advisory lock")
		}
	}()

	// This lock prevents assign and unassign operations to execute simultaneously
	if err := advisoryLocker.Lock(ctx, assignmentID); err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrap(err, "while trying to acquire postgres advisory lock in the beginning of instance creation"))
		return
	}
	defer func() {
		if err := advisoryLocker.Unlock(ctx, assignmentID); err != nil {
			log.C(ctx).WithError(err).Error("Error while releasing a previously-acquired advisory lock")
		}
	}()

	i.handleInstanceCreation(ctx, reqBody, statusAPIURL, correlationID)
}

func (i *InstanceCreatorHandler) handleInstanceCreation(ctx context.Context, reqBody *tenantmapping.Body, statusAPIURL, correlationID string) {
	// If receiver tenant outboundCommunication is missing - create it. It's here because if this code fails it's better to be before the instances are created.
	err := reqBody.AddReceiverTenantOutboundCommunicationIfMissing()
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, err)
		return
	}

	assignedTenantInboundCommunication := reqBody.GetAssignedTenantInboundCommunication()
	assignedTenantConfiguration := reqBody.AssignedTenant.Configuration

	serviceInstancesPath := tenantmapping.FindKeyPath(assignedTenantInboundCommunication.Value(), serviceInstancesKey)
	if serviceInstancesPath == "" {
		i.reportToUCLWithSuccess(ctx, statusAPIURL, correlationID, configPendingState, fmt.Sprintf("Service instances details are missing. Returning %q...", configPendingState), nil)
		return
	}

	globalServiceInstances := gjson.Get(assignedTenantInboundCommunication.Raw, serviceInstancesKey)

	// Go through global service instances(if exist)
	if globalServiceInstances.Exists() && globalServiceInstances.IsArray() && len(globalServiceInstances.Array()) > 0 {
		// Handle global service instances creation
		currentPath := fmt.Sprintf("%s.%s", tenantmapping.FindKeyPath(gjson.ParseBytes(assignedTenantConfiguration).Value(), inboundCommunicationKey), serviceInstancesKey)

		assignedTenantConfiguration, err = i.createServiceInstances(ctx, reqBody, globalServiceInstances.Raw, assignedTenantConfiguration, currentPath)
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while creating service instances"))
			return
		}
	}

	// Go through every auth methods and create their instances
	gjson.Parse(assignedTenantInboundCommunication.Raw).ForEach(func(auth, assignedTenantAuth gjson.Result) bool {
		currentPath := fmt.Sprintf("%s.%s", tenantmapping.FindKeyPath(gjson.ParseBytes(assignedTenantConfiguration).Value(), inboundCommunicationKey), auth)

		if gjson.Get(assignedTenantAuth.Raw, serviceInstancesKey).Exists() == false {
			// Substitute auth methods referring global instances
			assignedTenantAuth, err = SubstituteGJSON(gjson.GetBytes(assignedTenantConfiguration, currentPath), gjson.ParseBytes(assignedTenantConfiguration).Value())
			if err != nil {
				return false
			}

			assignedTenantConfiguration, err = sjson.SetBytes(assignedTenantConfiguration, currentPath, assignedTenantAuth.Value())
			if err != nil {
				return false
			}

			return true
		}

		localServiceInstances := gjson.Get(assignedTenantAuth.Raw, serviceInstancesKey)

		assignedTenantConfiguration, err = i.createServiceInstances(ctx, reqBody, localServiceInstances.Raw, assignedTenantConfiguration, fmt.Sprintf("%s.%s", currentPath, serviceInstancesKey))
		if err != nil {
			return false
		}

		// Substitute the top most json paths
		assignedTenantAuth, err = SubstituteGJSON(gjson.GetBytes(assignedTenantConfiguration, currentPath), gjson.ParseBytes(assignedTenantConfiguration).Value())
		if err != nil {
			return false
		}

		assignedTenantConfiguration, err = sjson.SetBytes(assignedTenantConfiguration, currentPath, assignedTenantAuth.Value())
		if err != nil {
			return false
		}

		return true
	})
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while creating service instances for auth methods"))
		return
	}

	// Remove Receiver Tenant inboundCommunication with service instance details and substitute the reverse json paths
	receiverTenantConfiguration := reqBody.ReceiverTenant.Configuration
	receiverTenantInboundCommunication := reqBody.GetReceiverTenantInboundCommunication()
	if receiverTenantInboundCommunication.Exists() {
		inboundCommunicationPath := tenantmapping.FindKeyPath(gjson.ParseBytes(receiverTenantConfiguration).Value(), inboundCommunicationKey)

		// Remove global instances
		receiverTenantConfiguration, err = sjson.DeleteBytes(receiverTenantConfiguration, fmt.Sprintf("%s.%s", inboundCommunicationPath, serviceInstancesKey))
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while removing global service instances from receiver tenant inboundCommunication"))
			return
		}

		// Remove auth methods with local instances or reffering global instances
		receiverTenantInboundCommunication = gjson.GetBytes(receiverTenantConfiguration, inboundCommunicationPath)
		gjson.Parse(receiverTenantInboundCommunication.Raw).ForEach(func(auth, assignedTenantAuth gjson.Result) bool {
			if gjson.Get(assignedTenantAuth.Raw, serviceInstancesKey).Exists() || strings.Contains(assignedTenantAuth.Raw, fmt.Sprintf("$.%s.%s", inboundCommunicationPath, serviceInstancesKey)) {
				receiverTenantConfiguration, err = sjson.DeleteBytes(receiverTenantConfiguration, fmt.Sprintf("%s.%s", inboundCommunicationPath, auth))
				if err != nil {
					return false
				}
			}

			return true
		})
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while deleting auth methods with local instances or refering global instances"))
			return
		}

		// Create temporary config with reverse field which is needed to populate the reverse paths below
		receiverTenantConfigurationWithReverse, err := sjson.SetBytes(receiverTenantConfiguration, reverseKey, gjson.ParseBytes(assignedTenantConfiguration).Value())
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while setting reverse object in receiver tenant configuration"))
			return
		}

		// Substitute reverse json paths
		receiverTenantConfigurationGJSONResult, err := SubstituteGJSON(gjson.ParseBytes(receiverTenantConfiguration), gjson.ParseBytes(receiverTenantConfigurationWithReverse).Value())
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while converting receiver tenant configuration to gjson.Result"))
			return
		}

		receiverTenantConfiguration = []byte(receiverTenantConfigurationGJSONResult.Raw)

		// Delete receiver tenant inbound communication if it is empty
		receiverTenantInboundCommunication = gjson.GetBytes(receiverTenantConfiguration, inboundCommunicationPath)
		if (receiverTenantInboundCommunication.IsObject() && len(receiverTenantInboundCommunication.Map()) == 0) || (receiverTenantInboundCommunication.IsArray() && len(receiverTenantInboundCommunication.Array()) == 0) {
			receiverTenantConfiguration, err = sjson.DeleteBytes(receiverTenantConfiguration, inboundCommunicationPath)
			if err != nil {
				i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while deleting the whole receiver tenant inbound communication"))
				return
			}
		}
	}

	// Populate Receiver tenant outbound communication
	// Remove service instance details from inbound communication before populating the receiver tenant outbound communication
	assignedTenantInboundCommunicationPath := tenantmapping.FindKeyPath(gjson.ParseBytes(assignedTenantConfiguration).Value(), inboundCommunicationKey) // Receiver outbound Path == Assigned inbound Path

	assignedTenantConfiguration, err = sjson.DeleteBytes(assignedTenantConfiguration, fmt.Sprintf("%s.%s", assignedTenantInboundCommunicationPath, serviceInstancesKey))
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while deleting global service instances from assigned tenant inboundCommunication"))
		return
	}

	assignedTenantInboundCommunication = gjson.GetBytes(assignedTenantConfiguration, assignedTenantInboundCommunicationPath)
	gjson.Parse(assignedTenantInboundCommunication.Raw).ForEach(func(auth, assignedTenantAuth gjson.Result) bool {
		assignedTenantConfiguration, err = sjson.DeleteBytes(assignedTenantConfiguration, fmt.Sprintf("%s.%s.%s", assignedTenantInboundCommunicationPath, auth.Str, serviceInstancesKey))
		if err != nil {
			return false
		}

		return true
	})
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while deleting service instances for auth methods"))
		return
	}

	receiverTenantOutboundCommunicationPath := tenantmapping.FindKeyPath(gjson.ParseBytes(receiverTenantConfiguration).Value(), "outboundCommunication") // Receiver outbound Path == Assigned inbound Path

	receiverTenantOutboundCommunication := gjson.GetBytes(receiverTenantConfiguration, receiverTenantOutboundCommunicationPath)
	assignedTenantInboundCommunication = gjson.GetBytes(assignedTenantConfiguration, assignedTenantInboundCommunicationPath)

	mergedReceiverTenantOutboundCommunication := DeepMergeJSON(assignedTenantInboundCommunication, receiverTenantOutboundCommunication)

	responseConfig, err := sjson.SetBytes(receiverTenantConfiguration, receiverTenantOutboundCommunicationPath, mergedReceiverTenantOutboundCommunication.Value())
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrapf(err, "while setting merged receiver tenant outboundCommunication with assigned tenant inboundCommunication in receiver tenant"))
		return
	}

	// Report to UCL with success
	i.reportToUCLWithSuccess(ctx, statusAPIURL, correlationID, readyState, "Successfully processed Service Instance creation.", responseConfig)
}

func (i *InstanceCreatorHandler) handleUnassign(ctx context.Context, reqBody *tenantmapping.Body, statusAPIURL, correlationID string) {
	// Get a single DB session
	connection, err := i.connector.GetConnection(ctx)
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrap(err, "while trying to get database connection"))
		return
	}
	defer func() {
		if err := connection.Close(); err != nil {
			log.C(ctx).WithError(err).Error("Error while closing the database connection")
		}
	}()

	advisoryLocker := connection.GetAdvisoryLocker()

	assignmentID := reqBody.AssignedTenant.AssignmentID

	// This lock prevents multiple unassign operations to execute simultaneously
	locked, err := advisoryLocker.TryLock(ctx, assignmentID+reqBody.Context.Operation)
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrap(err, "while trying to acquire postgres advisory lock in the beginning of instance creation"))
		return
	}
	if !locked {
		log.C(ctx).Debugf("Another instance creator is handling %q operation and assignment with ID %q", "unassign", assignmentID)
		return
	}
	defer func() {
		if err := advisoryLocker.Unlock(ctx, assignmentID+reqBody.Context.Operation); err != nil {
			log.C(ctx).WithError(err).Error("Error while releasing a previously-acquired advisory lock")
		}
	}()

	// This lock prevents unassign and assign operations to execute simultaneously
	if err := advisoryLocker.Lock(ctx, assignmentID); err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, createErrorState, errors.Wrap(err, "while trying to acquire postgres advisory lock in the beginning of instance creation"))
		return
	}
	defer func() {
		if err := advisoryLocker.Unlock(ctx, assignmentID); err != nil {
			log.C(ctx).WithError(err).Error("Error while releasing a previously-acquired advisory lock")
		}
	}()

	i.handleInstanceDeletion(ctx, reqBody, statusAPIURL, correlationID)
}

// Core Instance Deletion Logic
func (i *InstanceCreatorHandler) handleInstanceDeletion(ctx context.Context, reqBody *tenantmapping.Body, statusAPIURL, correlationID string) {
	assignmentID := reqBody.AssignedTenant.AssignmentID

	// Get the service instance with labels using the formation-id label key
	serviceInstancesIDs, err := i.SMClient.RetrieveMultipleResourcesIDsByLabels(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceInstances{}, map[string][]string{assignmentIDKey: {assignmentID}})
	if err != nil {
		// That's the case where the subaccount is deleted, and we are trying to delete its instances.
		// We should return READY and not fail.
		if strings.Contains(err.Error(), fmt.Sprintf(subaccountIsMissingFormatter, reqBody.ReceiverTenant.SubaccountID)) {
			i.reportToUCLWithSuccess(ctx, statusAPIURL, correlationID, readyState, "Successfully processed Service Instance deletion.", nil)
			return
		}

		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, deleteErrorState, errors.Wrapf(err, "while retrieving service instances for assignmentID: %q", assignmentID))
		return
	}

	// Retrieve all service instances bindings
	serviceBindingsIDs, err := i.SMClient.RetrieveMultipleResources(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceKeys{}, &types.ServiceKeyMatchParameters{ServiceInstancesIDs: serviceInstancesIDs})
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(subaccountIsMissingFormatter, reqBody.ReceiverTenant.SubaccountID)) {
			i.reportToUCLWithSuccess(ctx, statusAPIURL, correlationID, readyState, "Successfully processed Service Instance deletion.", nil)
			return
		}

		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, deleteErrorState, errors.Wrapf(err, "while retrieving service bindings for service instaces with IDs: %v", serviceInstancesIDs))
		return
	}

	// Delete all service instances bindings for the service instances
	if err = i.SMClient.DeleteMultipleResourcesByIDs(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceKeys{}, serviceBindingsIDs); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(subaccountIsMissingFormatter, reqBody.ReceiverTenant.SubaccountID)) {
			i.reportToUCLWithSuccess(ctx, statusAPIURL, correlationID, readyState, "Successfully processed Service Instance deletion.", nil)
			return
		}

		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, deleteErrorState, errors.Wrapf(err, "while deleting service bindings with IDs: %v", serviceBindingsIDs))
		return
	}

	// Delete all service instances with serviceInstanceIDs
	if err := i.SMClient.DeleteMultipleResourcesByIDs(ctx, reqBody.ReceiverTenant.Region, reqBody.ReceiverTenant.SubaccountID, &types.ServiceInstances{}, serviceInstancesIDs); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(subaccountIsMissingFormatter, reqBody.ReceiverTenant.SubaccountID)) {
			i.reportToUCLWithSuccess(ctx, statusAPIURL, correlationID, readyState, "Successfully processed Service Instance deletion.", nil)
			return
		}

		i.reportToUCLWithError(ctx, statusAPIURL, correlationID, deleteErrorState, errors.Wrapf(err, "while deleting service instances with IDs: %v", serviceInstancesIDs))
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
		log.C(ctx).Error("status API URL is empty...")
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

func (i *InstanceCreatorHandler) createServiceInstances(ctx context.Context, reqBody *tenantmapping.Body, serviceInstancesRaw string, assignedTenantConfiguration json.RawMessage, pathToServiceInstances string) (json.RawMessage, error) {
	region := reqBody.ReceiverTenant.Region
	subaccount := reqBody.ReceiverTenant.SubaccountID

	serviceInstancesArray := gjson.Parse(serviceInstancesRaw).Array()

	// Get current wave hash
	currentWaveHash, err := getCurrentWaveHash(pathToServiceInstances, gjson.Parse(serviceInstancesRaw).Value())
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

	for idx, serviceInstance := range serviceInstancesArray {
		currentPath := fmt.Sprintf("%s.%d", pathToServiceInstances, idx)
		serviceInstanceName := getResourceName(serviceInstance)

		assignedTenantConfiguration, err = sjson.SetBytes(assignedTenantConfiguration, fmt.Sprintf("%s.%s", currentPath, nameKey), serviceInstanceName)
		if err != nil {
			return nil, err
		}

		serviceInstanceBinding := gjson.Get(serviceInstance.Raw, serviceBindingKey)
		serviceBindingName := getResourceName(serviceInstanceBinding)

		assignedTenantConfiguration, err = sjson.SetBytes(assignedTenantConfiguration, fmt.Sprintf("%s.%s.%s", currentPath, serviceBindingKey, nameKey), serviceBindingName)
		if err != nil {
			return nil, err
		}

		serviceInstanceWithoutBinding, err := sjson.Delete(serviceInstance.Raw, serviceBindingKey)
		if err != nil {
			return nil, err
		}

		// Substitute the service instance json paths(without the service binding)
		if serviceInstance, err = SubstituteGJSON(gjson.Parse(serviceInstanceWithoutBinding), gjson.ParseBytes(assignedTenantConfiguration).Value()); err != nil {
			return nil, err
		}

		serviceOfferingCatalogName := gjson.Get(serviceInstance.Raw, serviceInstanceServiceKey).String()
		servicePlanCatalogName := gjson.Get(serviceInstance.Raw, serviceInstancePlanKey).String()
		serviceInstanceParameters := []byte(gjson.Get(serviceInstance.Raw, configurationKey).String())

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
		assignedTenantConfiguration, err = sjson.SetBytes(assignedTenantConfiguration, currentPath, gjson.ParseBytes(serviceInstanceRaw).Value())
		if err != nil {
			return nil, err
		}

		// Substitute the service binding json paths
		if serviceInstanceBinding, err = SubstituteGJSON(serviceInstanceBinding, gjson.ParseBytes(assignedTenantConfiguration).Value()); err != nil {
			return nil, err
		}

		serviceBindingParameters := []byte(gjson.Get(serviceInstanceBinding.Raw, configurationKey).String())
		if err != nil {
			return nil, errors.Wrapf(err, "while extracting the parameters of service binding for a service instance with id: %d", idx)
		}

		// Create the service instance binding
		serviceBindingID, err := i.SMClient.CreateResource(ctx, region, subaccount, &types.ServiceKeyReqBody{Name: serviceBindingName, ServiceKeyID: serviceInstanceID, Parameters: serviceBindingParameters}, &types.ServiceKey{})
		if err != nil {
			return nil, errors.Errorf("while creating service instance binding for service instance with ID %q", serviceInstanceID)
		}

		// Retrieve the service instance binding by ID
		serviceBindingRaw, err := i.SMClient.RetrieveRawResourceByID(ctx, region, subaccount, &types.ServiceKey{ID: serviceBindingID})
		if err != nil {
			return nil, errors.Errorf("while retrieving service instance binding with ID %q", serviceBindingID)
		}

		// Save the service instance binding
		assignedTenantConfiguration, err = sjson.SetBytes(assignedTenantConfiguration, fmt.Sprintf("%s.%s", currentPath, serviceBindingKey), gjson.ParseBytes(serviceBindingRaw).Value())
		if err != nil {
			return []byte(""), err
		}
	}

	return assignedTenantConfiguration, nil
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

func SubstituteGJSON(json gjson.Result, rootMap interface{}) (gjson.Result, error) {
	substitutedJson := json

	var err error
	var iterateAndSubstitute func(key string, value gjson.Result, path string)
	iterateAndSubstitute = func(key string, value gjson.Result, path string) {
		if value.IsObject() {
			for k, v := range value.Map() {
				iterateAndSubstitute(k, v, tenantmapping.NewCurrentPath(path, k))
			}
		} else if value.IsArray() {
			for i, el := range value.Array() {
				strI := fmt.Sprint(i)
				iterateAndSubstitute(strI, el, tenantmapping.NewCurrentPath(path, strI))
			}
		} else {
			concreteVal := value.String()
			if strings.Contains(concreteVal, "$.") {
				regex := regexp.MustCompile(`\{(\$.*?)\}`)
				matches := regex.FindAllStringSubmatch(concreteVal, -1)
				if len(matches) == 0 { // the value is something like "$.<>"
					var substitution interface{}
					substitution, err = jsonpath.Get(concreteVal, rootMap)
					if err != nil {
						log.D().Debugf("Error while substituting jsonpaths for %q with rootmap %v", concreteVal, rootMap)
						return
					}
					substitutedJsonStr, err := sjson.Set(substitutedJson.String(), path, substitution)
					if err != nil {
						log.D().Debugf("Error while setting %q key with value %q for json %q", path, value, substitutedJson)
						return
					}

					substitutedJson = gjson.Parse(substitutedJsonStr)
				} else { // the value contains jsonPaths with concatenated static strings. for example, "{$.<>}/string..."
					substitution := concreteVal
					for _, match := range matches {
						var currentSubstitution interface{}
						currentSubstitution, err = jsonpath.Get(match[1], rootMap)
						if err != nil {
							log.D().Debugf("Error while substituting jsonpaths for %q with rootmap %v", concreteVal, rootMap)
							return
						}
						substitution = strings.ReplaceAll(substitution, match[0], currentSubstitution.(string))
					}
					var substitutedJsonStr string
					substitutedJsonStr, err = sjson.Set(substitutedJson.String(), path, substitution)
					if err != nil {
						log.D().Debugf("Error while setting %q key with value %q for json %q", path, value, substitutedJson)
						return
					}

					substitutedJson = gjson.Parse(substitutedJsonStr)
				}

			}
		}
	}
	iterateAndSubstitute("", json, "")

	return substitutedJson, err
}

func getResourceName(resource gjson.Result) string {
	name := gjson.Get(resource.Raw, nameKey)
	if !name.Exists() {
		return uuid.New().String()
	}
	return name.String()
}

func getCurrentWaveHash(pathToServiceInstances string, serviceInstances interface{}) (string, error) {
	currentWave, err := sjson.SetBytes([]byte(`{}`), pathToServiceInstances, serviceInstances)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(currentWave)
	res := hex.EncodeToString(sum[0:])
	return res, nil
}

func DeepMergeJSON(src, dest gjson.Result) gjson.Result {
	res := dest

	var merge func(src, dest gjson.Result, path string)
	merge = func(src, dest gjson.Result, path string) {
		src.ForEach(func(srcKey, srcValue gjson.Result) bool {
			currentPath := path + "." + srcKey.String()
			if path == "" {
				currentPath = srcKey.String()
			}

			destValue := dest.Get(currentPath)
			if !destValue.Exists() || destValue.Type == gjson.Null {
				newJson, err := sjson.Set(res.Raw, currentPath, srcValue.Value())
				if err != nil {
					return false
				}

				res = gjson.Parse(newJson)
				return true
			}

			// If both are objects, merge recursively
			if srcValue.IsObject() && destValue.IsObject() {
				merge(srcValue, res, currentPath)
			} else if srcValue.IsArray() && destValue.IsArray() {
				merged := make([]interface{}, len(destValue.Array()))
				for i, destItem := range destValue.Array() {
					merged[i] = destItem.Value()
				}

				for _, srcItem := range srcValue.Array() {
					exists := false
					for _, destItem := range destValue.Array() {
						if srcItem.Raw == destItem.Raw {
							exists = true
							break
						}
					}
					if !exists {
						merged = append(merged, srcItem.Value())
					}
				}
				newJson, err := sjson.Set(res.Raw, currentPath, merged)
				if err != nil {
					return false
				}
				res = gjson.Parse(newJson)
			} else {
				newJson, err := sjson.Set(res.Raw, currentPath, srcValue.Value())
				if err != nil {
					return false
				}
				res = gjson.Parse(newJson)
			}
			return true
		})
	}

	merge(src, dest, "")

	return res
}
