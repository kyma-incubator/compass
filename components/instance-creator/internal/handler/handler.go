package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/PaesslerAG/jsonpath"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types/tenantmapping"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/persistence"
	"github.com/pkg/errors"
)

const (
	readyState         = "READY"
	createErrorState   = "CREATE_ERROR"
	deleteErrorState   = "DELETE_ERROR"
	configPendingState = "CONFIG_PENDING"

	assignOperation = "assign"

	inboundCommunicationKey       = "inboundCommunication"
	serviceInstancesKey           = "serviceInstances"
	serviceBindingKey             = "serviceBinding"
	serviceInstanceServiceBinding = "service"
	serviceInstancePlanKey        = "plan"
	configurationKey              = "configuration"
	nameKey                       = "name"
	assignmentIDKey               = "assignment_id"
	currentWaveHashKey            = "current_wave_hash"
	reverseKey                    = "reverse"

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
	ctx := log.ContextWithLogger(r.Context(), log.LoggerWithCorrelationID(r))

	log.C(ctx).Info("Instance Creator Handler was hit...")

	uclStatusAPIUrl := r.Header.Get(locationHeader)

	log.C(ctx).Info("Decoding the request body...")
	var reqBody tenantmapping.Body
	if err := decodeJSONBody(r, &reqBody); err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			respondWithError(ctx, w, http.StatusBadRequest, err)
		} else {
			respondWithError(ctx, w, http.StatusBadRequest, errors.Wrap(err, "while decoding json request body"))
		}
		return
	}

	log.C(ctx).Info("Validating tenant mapping request body...")
	if err := reqBody.Validate(); err != nil {
		respondWithError(ctx, w, http.StatusBadRequest, errors.Wrapf(err, "while validating the request body"))
		return
	}

	correlationID := correlation.CorrelationIDFromContext(ctx)

	// respond with 202 to the UCL call
	httputils.Respond(w, http.StatusAccepted)

	log.C(ctx).Info("Instance Creator Handler handles instance creation...")
	go i.handleInstances(correlationID, &reqBody, uclStatusAPIUrl)
}

func (i *InstanceCreatorHandler) handleInstances(correlationID string, reqBody *tenantmapping.Body, statusAPIURL string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	correlationIDKey := correlation.RequestIDHeaderKey
	ctx = correlation.SaveCorrelationIDHeaderToContext(ctx, &correlationIDKey, &correlationID)

	logger := log.C(ctx).WithField(correlationIDKey, correlationID)
	ctx = log.ContextWithLogger(ctx, logger)

	if reqBody.Context.Operation == assignOperation {
		i.handleAssign(ctx, reqBody, statusAPIURL)
	} else {
		i.handleUnassign(ctx, reqBody, statusAPIURL)
	}
}

func (i *InstanceCreatorHandler) handleAssign(ctx context.Context, reqBody *tenantmapping.Body, statusAPIURL string) {
	log.C(ctx).Debug("Getting a single DB connection for instance creation...")
	connection, err := i.connector.GetConnection(ctx)
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrap(err, "while trying to get database connection"))
		return
	}
	defer func() {
		log.C(ctx).Debug("Closing a DB connection for instance creation...")
		if err := connection.Close(); err != nil {
			log.C(ctx).WithError(err).Error("Error while closing the database connection")
		}
	}()

	assignmentID := reqBody.ReceiverTenant.AssignmentID

	advisoryLocker := connection.GetAdvisoryLocker()

	log.C(ctx).Debugf("Trying to get an advisory lock with (assignmentID, operation): (%q, %q)...", assignmentID, reqBody.Context.Operation)
	// This lock prevents multiple assign operations to execute simultaneously
	locked, err := advisoryLocker.TryLock(ctx, assignmentID+reqBody.Context.Operation)
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrap(err, "while trying to acquire postgres advisory lock in the beginning of instance creation"))
		return
	}
	if !locked {
		log.C(ctx).Debugf("Another instance creator is handling %q operation and assignment with ID %q", assignOperation, assignmentID)
		return
	}
	defer func() {
		log.C(ctx).Debugf("Unlocking an advisory lock with (assignmentID, operation): (%q, %q)...", assignmentID, reqBody.Context.Operation)
		if err := advisoryLocker.Unlock(ctx, assignmentID+reqBody.Context.Operation); err != nil {
			log.C(ctx).WithError(err).Error("Error while releasing a previously-acquired advisory lock")
		}
	}()

	log.C(ctx).Debugf("Locking an advisory lock with assignmentID %q...", assignmentID)
	// This lock prevents assign and unassign operations to execute simultaneously
	if err := advisoryLocker.Lock(ctx, assignmentID); err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrap(err, "while trying to acquire postgres advisory lock in the beginning of instance creation"))
		return
	}
	defer func() {
		log.C(ctx).Debugf("Unlocking an advisory lock with assignmentID %q...", assignmentID)
		if err := advisoryLocker.Unlock(ctx, assignmentID); err != nil {
			log.C(ctx).WithError(err).Error("Error while releasing a previously-acquired advisory lock")
		}
	}()

	log.C(ctx).Debug("Handling instance creation...")
	i.handleInstanceCreation(ctx, reqBody, statusAPIURL)
}

// Core Instance Creation Logic
func (i *InstanceCreatorHandler) handleInstanceCreation(ctx context.Context, reqBody *tenantmapping.Body, statusAPIURL string) {
	log.C(ctx).Debug("Adding receiver tenant outbound communication if missing...")
	err := reqBody.AddReceiverTenantOutboundCommunicationIfMissing()
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, err)
		return
	}

	assignedTenantInboundCommunication := reqBody.GetTenantCommunication(tenantmapping.AssignedTenantType, inboundCommunicationKey)
	assignedTenantConfiguration := reqBody.AssignedTenant.Configuration

	serviceInstancesPath := tenantmapping.FindKeyPath(assignedTenantInboundCommunication.Value(), serviceInstancesKey)
	if serviceInstancesPath == "" {
		i.reportToUCLWithSuccess(ctx, statusAPIURL, configPendingState, fmt.Sprintf("Service instances details are missing. Returning %q...", configPendingState), nil)
		return
	}

	globalServiceInstances := gjson.Get(assignedTenantInboundCommunication.Raw, serviceInstancesKey)

	if globalServiceInstances.Exists() && globalServiceInstances.IsArray() && len(globalServiceInstances.Array()) > 0 {
		log.C(ctx).Debug("Handle global service instances creation...")
		currentPath := fmt.Sprintf("%s.%s", tenantmapping.FindKeyPath(gjson.ParseBytes(assignedTenantConfiguration).Value(), inboundCommunicationKey), serviceInstancesKey)

		assignedTenantConfiguration, err = i.createServiceInstances(ctx, reqBody, globalServiceInstances.Raw, assignedTenantConfiguration, currentPath)
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrapf(err, "while creating service instances"))
			return
		}
	}

	log.C(ctx).Debug("Handle local service instances creation...")
	gjson.Parse(assignedTenantInboundCommunication.Raw).ForEach(func(auth, assignedTenantAuth gjson.Result) bool {
		currentPath := fmt.Sprintf("%s.%s", tenantmapping.FindKeyPath(gjson.ParseBytes(assignedTenantConfiguration).Value(), inboundCommunicationKey), auth)

		if !gjson.Get(assignedTenantAuth.Raw, serviceInstancesKey).Exists() {
			log.C(ctx).Debugf("Auth method %q doesn't have local service instances. Substituting its jsonpaths(if they exist) and proceeding with the next auth method...", auth)
			assignedTenantAuth, err = SubstituteGJSON(ctx, gjson.GetBytes(assignedTenantConfiguration, currentPath), gjson.ParseBytes(assignedTenantConfiguration).Value())
			if err != nil {
				return false
			}

			assignedTenantConfiguration, err = sjson.SetBytes(assignedTenantConfiguration, currentPath, assignedTenantAuth.Value())
			return err == nil
		}

		localServiceInstances := gjson.Get(assignedTenantAuth.Raw, serviceInstancesKey)

		log.C(ctx).Debugf("Handle local service instances creation for auth method %q...", auth)
		assignedTenantConfiguration, err = i.createServiceInstances(ctx, reqBody, localServiceInstances.Raw, assignedTenantConfiguration, fmt.Sprintf("%s.%s", currentPath, serviceInstancesKey))
		if err != nil {
			return false
		}

		log.C(ctx).Debugf("Substitute the topmost jsonpaths for auth method %q...", auth)
		assignedTenantAuth, err = SubstituteGJSON(ctx, gjson.GetBytes(assignedTenantConfiguration, currentPath), gjson.ParseBytes(assignedTenantConfiguration).Value())
		if err != nil {
			return false
		}

		assignedTenantConfiguration, err = sjson.SetBytes(assignedTenantConfiguration, currentPath, assignedTenantAuth.Value())
		return err == nil
	})
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrapf(err, "while creating service instances for auth methods"))
		return
	}

	receiverTenantConfiguration := reqBody.ReceiverTenant.Configuration
	receiverTenantInboundCommunication := reqBody.GetTenantCommunication(tenantmapping.ReceiverTenantType, inboundCommunicationKey)
	if receiverTenantInboundCommunication.Exists() {
		log.C(ctx).Debugf("Removing service instance details(if they exist) from receiver tenant inbound communication...")
		inboundCommunicationPath := tenantmapping.FindKeyPath(gjson.ParseBytes(receiverTenantConfiguration).Value(), inboundCommunicationKey)

		log.C(ctx).Debugf("Removing global service instance details(if they exist) from receiver tenant inbound communication...")
		receiverTenantConfiguration, err = sjson.DeleteBytes(receiverTenantConfiguration, fmt.Sprintf("%s.%s", inboundCommunicationPath, serviceInstancesKey))
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrapf(err, "while removing global service instances from receiver tenant inboundCommunication"))
			return
		}

		log.C(ctx).Debugf("Removing auth methods with service instance details(if they exist) from receiver tenant inbound communication...")
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
			i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrapf(err, "while deleting auth methods with local instances or refering global instances"))
			return
		}

		log.C(ctx).Debugf("Creating temporary config with reverse field which will be used to populate the reverse paths...")
		receiverTenantConfigurationWithReverse, err := sjson.SetBytes(receiverTenantConfiguration, reverseKey, gjson.ParseBytes(assignedTenantConfiguration).Value())
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrapf(err, "while setting reverse object in receiver tenant configuration"))
			return
		}

		log.C(ctx).Debugf("Substituting the reverse jsonpaths...")
		receiverTenantConfigurationGJSONResult, err := SubstituteGJSON(ctx, gjson.ParseBytes(receiverTenantConfiguration), gjson.ParseBytes(receiverTenantConfigurationWithReverse).Value())
		if err != nil {
			i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrapf(err, "while converting receiver tenant configuration to gjson.Result"))
			return
		}

		receiverTenantConfiguration = []byte(receiverTenantConfigurationGJSONResult.Raw)

		log.C(ctx).Debugf("Removing receiver tenant inbound communication if it is left empty...")
		receiverTenantInboundCommunication = gjson.GetBytes(receiverTenantConfiguration, inboundCommunicationPath)
		if (receiverTenantInboundCommunication.IsObject() && len(receiverTenantInboundCommunication.Map()) == 0) || (receiverTenantInboundCommunication.IsArray() && len(receiverTenantInboundCommunication.Array()) == 0) {
			receiverTenantConfiguration, err = sjson.DeleteBytes(receiverTenantConfiguration, inboundCommunicationPath)
			if err != nil {
				i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrapf(err, "while removing the whole receiver tenant inbound communication"))
				return
			}
		}
	}

	log.C(ctx).Debugf("Removing service instance details from assigned tenant inbound communication before populating the receiver tenant outbound communication...")
	assignedTenantInboundCommunicationPath := tenantmapping.FindKeyPath(gjson.ParseBytes(assignedTenantConfiguration).Value(), inboundCommunicationKey) // Receiver outbound Path == Assigned inbound Path

	log.C(ctx).Debugf("Removing assigned tenant global service instances from inbound communication...")
	assignedTenantConfiguration, err = sjson.DeleteBytes(assignedTenantConfiguration, fmt.Sprintf("%s.%s", assignedTenantInboundCommunicationPath, serviceInstancesKey))
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrapf(err, "while removing global service instances from assigned tenant inbound communication"))
		return
	}

	log.C(ctx).Debugf("Removing assigned tenant local service instances from inbound communication...")
	assignedTenantInboundCommunication = gjson.GetBytes(assignedTenantConfiguration, assignedTenantInboundCommunicationPath)
	gjson.Parse(assignedTenantInboundCommunication.Raw).ForEach(func(auth, assignedTenantAuth gjson.Result) bool {
		assignedTenantConfiguration, err = sjson.DeleteBytes(assignedTenantConfiguration, fmt.Sprintf("%s.%s.%s", assignedTenantInboundCommunicationPath, auth.Str, serviceInstancesKey))
		return err == nil
	})
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrapf(err, "while deleting service instances for auth methods"))
		return
	}

	receiverTenantOutboundCommunicationPath := tenantmapping.FindKeyPath(gjson.ParseBytes(receiverTenantConfiguration).Value(), "outboundCommunication") // Receiver outbound Path == Assigned inbound Path

	receiverTenantOutboundCommunication := gjson.GetBytes(receiverTenantConfiguration, receiverTenantOutboundCommunicationPath)
	assignedTenantInboundCommunication = gjson.GetBytes(assignedTenantConfiguration, assignedTenantInboundCommunicationPath)

	log.C(ctx).Debugf("Merging assigned tenant inbound communication with receiver tenant outbound communication...")
	mergedReceiverTenantOutboundCommunication := DeepMergeJSON(assignedTenantInboundCommunication, receiverTenantOutboundCommunication)

	responseConfig, err := sjson.SetBytes(receiverTenantConfiguration, receiverTenantOutboundCommunicationPath, mergedReceiverTenantOutboundCommunication.Value())
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrapf(err, "while setting merged receiver tenant outboundCommunication with assigned tenant inboundCommunication in receiver tenant"))
		return
	}

	// Report to UCL with success
	i.reportToUCLWithSuccess(ctx, statusAPIURL, readyState, "Successfully processed Service Instance creation.", responseConfig)
}

func (i *InstanceCreatorHandler) handleUnassign(ctx context.Context, reqBody *tenantmapping.Body, statusAPIURL string) {
	log.C(ctx).Debug("Getting a single DB connection for instance deletion...")
	connection, err := i.connector.GetConnection(ctx)
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrap(err, "while trying to get database connection"))
		return
	}
	defer func() {
		log.C(ctx).Debug("Closing a DB connection for instance creation...")
		if err := connection.Close(); err != nil {
			log.C(ctx).WithError(err).Error("Error while closing the database connection")
		}
	}()

	advisoryLocker := connection.GetAdvisoryLocker()

	assignmentID := reqBody.ReceiverTenant.AssignmentID

	log.C(ctx).Debugf("Trying to get an advisory lock with (assignmentID, operation): (%q, %q)...", assignmentID, reqBody.Context.Operation)
	// This lock prevents multiple unassign operations to execute simultaneously
	locked, err := advisoryLocker.TryLock(ctx, assignmentID+reqBody.Context.Operation)
	if err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrap(err, "while trying to acquire postgres advisory lock in the beginning of instance creation"))
		return
	}
	if !locked {
		log.C(ctx).Debugf("Another instance creator is handling %q operation and assignment with ID %q", "unassign", assignmentID)
		return
	}
	defer func() {
		log.C(ctx).Debugf("Unlocking an advisory lock with (assignmentID, operation): (%q, %q)...", assignmentID, reqBody.Context.Operation)
		if err := advisoryLocker.Unlock(ctx, assignmentID+reqBody.Context.Operation); err != nil {
			log.C(ctx).WithError(err).Error("Error while releasing a previously-acquired advisory lock")
		}
	}()

	log.C(ctx).Debugf("Locking an advisory lock with assignmentID %q...", assignmentID)
	// This lock prevents unassign and assign operations to execute simultaneously
	if err := advisoryLocker.Lock(ctx, assignmentID); err != nil {
		i.reportToUCLWithError(ctx, statusAPIURL, createErrorState, errors.Wrap(err, "while trying to acquire postgres advisory lock in the beginning of instance creation"))
		return
	}
	defer func() {
		log.C(ctx).Debugf("Unlocking an advisory lock with assignmentID %q...", assignmentID)
		if err := advisoryLocker.Unlock(ctx, assignmentID); err != nil {
			log.C(ctx).WithError(err).Error("Error while releasing a previously-acquired advisory lock")
		}
	}()

	log.C(ctx).Debug("Handling instance deletion...")
	i.handleInstanceDeletion(ctx, reqBody, statusAPIURL)
}

// Core Instance Deletion Logic
func (i *InstanceCreatorHandler) handleInstanceDeletion(ctx context.Context, reqBody *tenantmapping.Body, statusAPIURL string) {
	assignmentID := reqBody.ReceiverTenant.AssignmentID
	region := reqBody.ReceiverTenant.Region
	subaccount := reqBody.ReceiverTenant.SubaccountID
	labels := map[string][]string{assignmentIDKey: {assignmentID}}

	log.C(ctx).Debugf("Listing service instances with labels %v for region %q, subaccount %q...", labels, region, subaccount)
	serviceInstancesIDs, err := i.SMClient.RetrieveMultipleResourcesIDsByLabels(ctx, region, subaccount, &types.ServiceInstances{}, labels)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(subaccountIsMissingFormatter, subaccount)) {
			log.C(ctx).Debugf("Subaccount %q was deleted while we are trying to delete its instances. Returning...", subaccount)
			i.reportToUCLWithSuccess(ctx, statusAPIURL, readyState, "Successfully processed Service Instance deletion.", nil)
			return
		}

		i.reportToUCLWithError(ctx, statusAPIURL, deleteErrorState, errors.Wrapf(err, "while retrieving service instances for assignmentID: %q", assignmentID))
		return
	}

	log.C(ctx).Debugf("Listing service instances bindings for service instances with IDs %v, for region %q, subaccount %q ..", serviceInstancesIDs, region, subaccount)
	serviceBindingsIDs, err := i.SMClient.RetrieveMultipleResources(ctx, region, subaccount, &types.ServiceBindings{}, &types.ServiceBindingMatchParameters{ServiceInstancesIDs: serviceInstancesIDs})
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(subaccountIsMissingFormatter, subaccount)) {
			log.C(ctx).Debugf("Subaccount %q was deleted while we are trying to delete its instances. Returning...", subaccount)
			i.reportToUCLWithSuccess(ctx, statusAPIURL, readyState, "Successfully processed Service Instance deletion.", nil)
			return
		}

		i.reportToUCLWithError(ctx, statusAPIURL, deleteErrorState, errors.Wrapf(err, "while retrieving service bindings for service instaces with IDs: %v", serviceInstancesIDs))
		return
	}

	log.C(ctx).Debugf("Deleting service instances bindings with IDs %v, for region %q, subaccount %q ..", serviceBindingsIDs, region, subaccount)
	if err = i.SMClient.DeleteMultipleResourcesByIDs(ctx, region, subaccount, &types.ServiceBindings{}, serviceBindingsIDs); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(subaccountIsMissingFormatter, subaccount)) {
			log.C(ctx).Debugf("Subaccount %q was deleted while we are trying to delete its instances. Returning...", subaccount)
			i.reportToUCLWithSuccess(ctx, statusAPIURL, readyState, "Successfully processed Service Instance deletion.", nil)
			return
		}

		i.reportToUCLWithError(ctx, statusAPIURL, deleteErrorState, errors.Wrapf(err, "while deleting service bindings with IDs: %v", serviceBindingsIDs))
		return
	}

	log.C(ctx).Debugf("Deleting service instances with IDs %v, for region %q, subaccount %q ..", serviceInstancesIDs, region, subaccount)
	if err := i.SMClient.DeleteMultipleResourcesByIDs(ctx, region, subaccount, &types.ServiceInstances{}, serviceInstancesIDs); err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(subaccountIsMissingFormatter, subaccount)) {
			log.C(ctx).Debugf("Subaccount %q was deleted while we are trying to delete its instances. Returning...", subaccount)
			i.reportToUCLWithSuccess(ctx, statusAPIURL, readyState, "Successfully processed Service Instance deletion.", nil)
			return
		}

		i.reportToUCLWithError(ctx, statusAPIURL, deleteErrorState, errors.Wrapf(err, "while deleting service instances with IDs: %v", serviceInstancesIDs))
		return
	}

	// Report to UCL with success
	i.reportToUCLWithSuccess(ctx, statusAPIURL, readyState, "Successfully processed Service Instance deletion.", nil)
}

func (i *InstanceCreatorHandler) createServiceInstances(ctx context.Context, reqBody *tenantmapping.Body, serviceInstancesRaw string, assignedTenantConfiguration json.RawMessage, pathToServiceInstances string) (json.RawMessage, error) {
	region := reqBody.ReceiverTenant.Region
	subaccount := reqBody.ReceiverTenant.SubaccountID

	serviceInstancesArray := gjson.Parse(serviceInstancesRaw).Array()

	log.C(ctx).Debug("Getting current wave hash...")
	currentWaveHash, err := getCurrentWaveHash(pathToServiceInstances, gjson.Parse(serviceInstancesRaw).Value())
	if err != nil {
		return nil, errors.Wrap(err, "while getting current wave hash")
	}

	smLabels := map[string][]string{
		assignmentIDKey:    {reqBody.ReceiverTenant.AssignmentID},
		currentWaveHashKey: {currentWaveHash},
	}
	log.C(ctx).Debugf("Listing service instances with labels %v for region %q and subaccount %q...", smLabels, region, subaccount)
	existentServiceInstancesIDs, err := i.SMClient.RetrieveMultipleResourcesIDsByLabels(ctx, region, subaccount, &types.ServiceInstances{}, smLabels)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting service instances for the current wave with hash: %q", currentWaveHash)
	}

	if len(existentServiceInstancesIDs) > 0 {
		log.C(ctx).Debug("Service instances for this wave exist - it's the resync case. Recreating the service instances...")

		log.C(ctx).Debugf("Listing service instances bindings for service instances with IDs %v, for region %q, subaccount %q ..", existentServiceInstancesIDs, region, subaccount)
		serviceBindingsIDs, err := i.SMClient.RetrieveMultipleResources(ctx, region, subaccount, &types.ServiceBindings{}, &types.ServiceBindingMatchParameters{ServiceInstancesIDs: existentServiceInstancesIDs})
		if err != nil {
			return nil, errors.Wrapf(err, "while retrieving service bindings for service instaces with IDs: %v", existentServiceInstancesIDs)
		}

		log.C(ctx).Debugf("Deleting service instances bindings with IDs %v, for region %q, subaccount %q ..", serviceBindingsIDs, region, subaccount)
		if err = i.SMClient.DeleteMultipleResourcesByIDs(ctx, region, subaccount, &types.ServiceBindings{}, serviceBindingsIDs); err != nil {
			return nil, errors.Wrapf(err, "while deleting service bindings with IDs: %v", serviceBindingsIDs)
		}

		log.C(ctx).Debugf("Deleting service instances with IDs %v, for region %q, subaccount %q ..", existentServiceInstancesIDs, region, subaccount)
		if err := i.SMClient.DeleteMultipleResourcesByIDs(ctx, region, subaccount, &types.ServiceInstances{}, existentServiceInstancesIDs); err != nil {
			return nil, errors.Wrapf(err, "while deleting service instances with IDs: %v", existentServiceInstancesIDs)
		}
	}

	log.C(ctx).Debug("Iterating through service instances array...")
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

		log.C(ctx).Debugf("Substituting the service instance(with index %d and name %q) jsonpaths(withouth the jsonpaths of its binding)...", idx, serviceInstanceName)
		if serviceInstance, err = SubstituteGJSON(ctx, gjson.Parse(serviceInstanceWithoutBinding), gjson.ParseBytes(assignedTenantConfiguration).Value()); err != nil {
			return nil, err
		}

		serviceOfferingCatalogName := gjson.Get(serviceInstance.Raw, serviceInstanceServiceBinding).String()
		servicePlanCatalogName := gjson.Get(serviceInstance.Raw, serviceInstancePlanKey).String()
		serviceInstanceParameters := []byte(gjson.Get(serviceInstance.Raw, configurationKey).String())

		log.C(ctx).Debugf("Getting the service offering ID with catalog name('service' field from the contract) %q for region %q and subaccount %q...", serviceOfferingCatalogName, region, subaccount)
		serviceOfferingID, err := i.SMClient.RetrieveResource(ctx, region, subaccount, &types.ServiceOfferings{}, &types.ServiceOfferingMatchParameters{CatalogName: serviceOfferingCatalogName})
		if err != nil {
			return nil, errors.Errorf("while retrieving service offering with catalog name %q", serviceOfferingCatalogName)
		}

		log.C(ctx).Debugf("Getting the service plan ID with service offering ID %q and catalog name %q for region %q and subaccount %q...", serviceOfferingID, servicePlanCatalogName, region, subaccount)
		servicePlanID, err := i.SMClient.RetrieveResource(ctx, region, subaccount, &types.ServicePlans{}, &types.ServicePlanMatchParameters{PlanName: servicePlanCatalogName, OfferingID: serviceOfferingID})
		if err != nil {
			return nil, errors.Errorf("while retrieving service plan with catalog name %q and offering ID %q", serviceOfferingCatalogName, serviceOfferingID)
		}

		log.C(ctx).Debugf("Creating service instance with name %q, plan id %q, parameters %q and labels %v for subaccount %q and region %q...", serviceInstanceName, servicePlanID, serviceInstanceParameters, smLabels, region, subaccount)
		serviceInstanceID, err := i.SMClient.CreateResource(ctx, region, subaccount, &types.ServiceInstanceReqBody{Name: serviceInstanceName, ServicePlanID: servicePlanID, Parameters: serviceInstanceParameters, Labels: smLabels}, &types.ServiceInstance{})
		if err != nil {
			return nil, errors.Errorf("while creating service instance with name %q", serviceInstanceName)
		}

		log.C(ctx).Debugf("Getting raw service instance with id %q for subaccount %q and region %q...", serviceInstanceID, region, subaccount)
		serviceInstanceRaw, err := i.SMClient.RetrieveRawResourceByID(ctx, region, subaccount, &types.ServiceInstance{ID: serviceInstanceID})
		if err != nil {
			return nil, errors.Errorf("while retrieving service instance with ID %q", serviceInstanceID)
		}

		log.C(ctx).Debug("Saving the raw service instance in the assigned tenant configuration...")
		assignedTenantConfiguration, err = sjson.SetBytes(assignedTenantConfiguration, currentPath, gjson.ParseBytes(serviceInstanceRaw).Value())
		if err != nil {
			return nil, err
		}

		log.C(ctx).Debug("Substituting service binding jsonpaths...")
		if serviceInstanceBinding, err = SubstituteGJSON(ctx, serviceInstanceBinding, gjson.ParseBytes(assignedTenantConfiguration).Value()); err != nil {
			return nil, err
		}

		serviceBindingParameters := []byte(gjson.Get(serviceInstanceBinding.Raw, configurationKey).String())
		if err != nil {
			return nil, errors.Wrapf(err, "while extracting the parameters of service binding for a service instance with id: %d", idx)
		}

		log.C(ctx).Debugf("Creating service binding with name %q, service instance id %q and parameters %q for subaccount %q and region %q...", serviceBindingName, serviceInstanceID, serviceBindingParameters, region, subaccount)
		serviceBindingID, err := i.SMClient.CreateResource(ctx, region, subaccount, &types.ServiceBindingReqBody{Name: serviceBindingName, ServiceBindingID: serviceInstanceID, Parameters: serviceBindingParameters}, &types.ServiceBinding{})
		if err != nil {
			return nil, errors.Errorf("while creating service instance binding for service instance with ID %q", serviceInstanceID)
		}

		log.C(ctx).Debugf("Getting raw service binding with id %q for subaccount %q and region %q...", serviceBindingID, region, subaccount)
		serviceBindingRaw, err := i.SMClient.RetrieveRawResourceByID(ctx, region, subaccount, &types.ServiceBinding{ID: serviceBindingID})
		if err != nil {
			return nil, errors.Errorf("while retrieving service instance binding with ID %q", serviceBindingID)
		}

		log.C(ctx).Debug("Saving the raw service binding in the assigned tenant configuration...")
		assignedTenantConfiguration, err = sjson.SetBytes(assignedTenantConfiguration, fmt.Sprintf("%s.%s", currentPath, serviceBindingKey), gjson.ParseBytes(serviceBindingRaw).Value())
		if err != nil {
			return []byte(""), err
		}
	}

	return assignedTenantConfiguration, nil
}

func (i *InstanceCreatorHandler) callUCLStatusAPI(ctx context.Context, statusAPIURL string, response interface{}) {
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
func (i *InstanceCreatorHandler) reportToUCLWithError(ctx context.Context, statusAPIURL, state string, err error) {
	log.C(ctx).Error(err.Error())
	errorResponse := ErrorResponse{State: state, Message: err.Error()}
	i.callUCLStatusAPI(ctx, statusAPIURL, errorResponse)
}

// reportToUCLWithSuccess reports status to the UCL Status API with the JSON success wrapped in an SuccessResponse struct
func (i *InstanceCreatorHandler) reportToUCLWithSuccess(ctx context.Context, statusAPIURL, state, msg string, configuration json.RawMessage) {
	log.C(ctx).Info(msg)
	successResponse := SuccessResponse{State: state, Configuration: configuration}
	i.callUCLStatusAPI(ctx, statusAPIURL, successResponse)
}

// SubstituteGJSON substitutes the jsonpaths in a given json
func SubstituteGJSON(ctx context.Context, json gjson.Result, rootMap interface{}) (gjson.Result, error) {
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
						log.C(ctx).Debugf("Error while substituting jsonpaths for %q with rootmap %v", concreteVal, rootMap)
						return
					}
					substitutedJsonStr, err := sjson.Set(substitutedJson.String(), path, substitution)
					if err != nil {
						log.C(ctx).Debugf("Error while setting %q key with value %q for json %q", path, value, substitutedJson)
						return
					}

					substitutedJson = gjson.Parse(substitutedJsonStr)
				} else { // the value contains multiple jsonPaths with concatenated static strings. for example, "{$.<>}/string..."
					substitution := concreteVal
					for _, match := range matches {
						var currentSubstitution interface{}
						currentSubstitution, err = jsonpath.Get(match[1], rootMap)
						if err != nil {
							log.C(ctx).Debugf("Error while substituting jsonpaths for %q with rootmap %v", concreteVal, rootMap)
							return
						}
						substitution = strings.ReplaceAll(substitution, match[0], currentSubstitution.(string))
					}
					var substitutedJsonStr string
					substitutedJsonStr, err = sjson.Set(substitutedJson.String(), path, substitution)
					if err != nil {
						log.C(ctx).Debugf("Error while setting %q key with value %q for json %q", path, value, substitutedJson)
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

// respondWithError writes a http response using with the JSON encoded error wrapped in an Error struct
func respondWithError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	log.C(ctx).WithError(err).Errorf("Responding with error: %v", err)
	w.Header().Add(contentTypeHeaderKey, contentTypeApplicationJSON)
	w.WriteHeader(status)
	errorResponse := ErrorResponse{Message: err.Error()}
	encodingErr := json.NewEncoder(w).Encode(errorResponse)
	if encodingErr != nil {
		log.C(ctx).WithError(err).Errorf("Failed to encode error response: %v", err)
	}
}
