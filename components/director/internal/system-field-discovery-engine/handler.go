package systemfielddiscoveryengine

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/data"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// OperationsManager provides methods for operations management
//
//go:generate mockery --name=OperationsManager --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationsManager interface {
	FindOperationByData(ctx context.Context, data interface{}) (*model.Operation, error)
	CreateOperation(ctx context.Context, in *model.OperationInput) (string, error)
	RescheduleOperation(ctx context.Context, operationID string) error
}

// SystemFieldDiscoveryResources holds application id and registry for system field discovery
type SystemFieldDiscoveryResources struct {
	ApplicationID string `json:"applicationID"`
	TenantID      string `json:"tenantID"`
}

type handler struct {
	opMgr           OperationsManager
	onDemandChannel chan string
}

// NewSystemFieldDiscoveryHTTPHandler returns a new HTTP handler, responsible for handling HTTP requests
func NewSystemFieldDiscoveryHTTPHandler(opMgr OperationsManager, onDemandChannel chan string) *handler {
	return &handler{
		opMgr:           opMgr,
		onDemandChannel: onDemandChannel,
	}
}

// ScheduleSaaSRegistryDiscoveryForSystemFieldDiscoveryData validates the payload, checks if such an operation already exists.
// If it does, it reschedules the existing operation; otherwise, it creates a new operation with high priority.
func (h *handler) ScheduleSaaSRegistryDiscoveryForSystemFieldDiscoveryData(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	payload := SystemFieldDiscoveryResources{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to parse request body")
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	if payload.ApplicationID == "" || payload.TenantID == "" {
		log.C(ctx).Errorf("Invalid data provided for system field discovery aggregation")
		http.Error(writer, "Invalid payload, Application ID or Tenant ID are not provided.", http.StatusBadRequest)
		return
	}

	log.C(ctx).Infof("Rescheduling system field discovery data aggregation for application with id %q  and tenant with id %q for saas registry", payload.ApplicationID, payload.TenantID)
	operation, err := h.opMgr.FindOperationByData(ctx, data.NewSystemFieldDiscoveryOperationData(payload.ApplicationID, payload.TenantID))
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			log.C(ctx).WithError(err).Errorf("Loading Operation for system field discovery data aggregation failed")
			http.Error(writer, "Loading Operation for system field discovery data aggregation failed", http.StatusInternalServerError)
			return
		}

		log.C(ctx).Infof("Operation with ApplicationID %q and TenantID %q does not exist. Trying to create...", payload.ApplicationID, payload.TenantID)
		now := time.Now()
		data := data.NewSystemFieldDiscoveryOperationData(payload.ApplicationID, payload.TenantID)
		rawData, err := data.GetData()
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Preparing Operation for system field discovery data aggregation failed")
			http.Error(writer, "Preparing Operation for system field discovery data aggregation failed", http.StatusInternalServerError)
			return
		}

		newOperationInput := &model.OperationInput{
			OpType:        model.OperationTypeSaasRegistryDiscovery,
			Status:        model.OperationStatusScheduled,
			Data:          json.RawMessage(rawData),
			ErrorSeverity: model.OperationErrorSeverityNone,
			Priority:      int(operationsmanager.HighOperationPriority),
			CreatedAt:     &now,
		}

		opID, err := h.opMgr.CreateOperation(ctx, newOperationInput)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Creating Operation for system field discovery data aggregation failed")
			http.Error(writer, "Creating Operation for system field discovery data aggregation failed", http.StatusInternalServerError)
			return
		}
		log.C(ctx).Infof("Successfully created operation with ApplicationID %q and TenantID %q", payload.ApplicationID, payload.TenantID)

		// Notify OperationProcessors for new operation
		h.onDemandChannel <- opID

		writer.WriteHeader(http.StatusOK)
		return
	}

	if err = h.opMgr.RescheduleOperation(ctx, operation.ID); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to reschedule operation with ID %s", operation.ID)
		http.Error(writer, "Scheduling Operation for system field discovery data aggregation failed", http.StatusInternalServerError)
		return
	}
	// Notify OperationProcessors for new operation
	h.onDemandChannel <- operation.ID

	writer.WriteHeader(http.StatusOK)
}
