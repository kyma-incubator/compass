package systemfetcher

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

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

// AggregationResource holds id of tenant for systems fetching
type AggregationResource struct {
	TenantIDs      []string `json:"tenantIDs"`
	SkipReschedule bool     `json:"skipReschedule"`
}

type handler struct {
	opMgr                    OperationsManager
	businessTenantMappingSvc BusinessTenantMappingService
	transact                 persistence.Transactioner
	onDemandChannel          chan string
	workerPool               chan struct{}
}

// NewSystemFetcherAggregatorHTTPHandler returns a new HTTP handler, responsible for handling HTTP requests
func NewSystemFetcherAggregatorHTTPHandler(opMgr OperationsManager, businessTenantMappingSvc BusinessTenantMappingService, transact persistence.Transactioner, onDemandChannel chan string, workers chan struct{}) *handler {
	return &handler{
		opMgr:                    opMgr,
		businessTenantMappingSvc: businessTenantMappingSvc,
		transact:                 transact,
		onDemandChannel:          onDemandChannel,
		workerPool:               workers,
	}
}

// ScheduleAggregationForSystemFetcherData validates the payload, checks if such an operation already exists.
// If it does, it reschedules the existing operation; otherwise, it creates a new operation with high priority.
func (h *handler) ScheduleAggregationForSystemFetcherData(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	payload := AggregationResource{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		log.C(ctx).WithError(err).Error("Failed to parse request body")
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(payload.TenantIDs) == 0 {
		log.C(ctx).Error("Invalid data provided for System Fetcher aggregation")
		http.Error(writer, "Invalid payload, TenantIDs is not provided or it is empty.", http.StatusBadRequest)
		return
	}

	log.C(ctx).Infof("Rescheduling system fetcher data aggregation for tenants %v", payload.TenantIDs)
	writer.WriteHeader(http.StatusAccepted)

	h.workerPool <- struct{}{}

	entry := log.LoggerFromContext(ctx)
	opCtx := log.ContextWithLogger(context.Background(), entry)

	go func(ctx context.Context) {
		defer func() {
			<-h.workerPool
		}()

		h.scheduleOperations(opCtx, payload)
	}(opCtx)
}

func (h *handler) scheduleOperations(ctx context.Context, payload AggregationResource) {
	for _, tenantID := range payload.TenantIDs {
		h.scheduleOperation(ctx, tenantID, payload.SkipReschedule)
	}
}

func (h *handler) scheduleOperation(ctx context.Context, tenantID string, skipReschedule bool) {
	operation, err := h.opMgr.FindOperationByData(ctx, NewSystemFetcherOperationData(tenantID))
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			log.C(ctx).WithError(err).Error("Loading Operation for System Fetcher data aggregation failed")
			return
		}

		log.C(ctx).Infof("Operation with TenantID %q does not exist. Trying to create...", tenantID)

		businessTenantMapping, err := h.getBusinessTenantMappingByID(ctx, tenantID)
		if err != nil {
			if !apperrors.IsNotFoundError(err) {
				log.C(ctx).WithError(err).Errorf("Getting tenant by internal id %q failed", tenantID)
				return
			}

			businessTenantMapping, err = h.getBusinessTenantMappingByExternalID(ctx, tenantID)
			if err != nil {
				if apperrors.IsNotFoundError(err) {
					log.C(ctx).WithError(err).Errorf("External tenant with id %q not found", tenantID)
					return
				}
				log.C(ctx).WithError(err).Errorf("Getting external tenant with id %q failed", tenantID)
				return
			}
		}

		if businessTenantMapping == nil {
			log.C(ctx).Error("Loading Business Tenant Mapping for System Fetcher data aggregation failed")
			return
		}

		if businessTenantMapping.Type != tenant.Account && businessTenantMapping.Type != tenant.Customer {
			log.C(ctx).Infof("Tenant with ID %q is of type %q - operations are created only for tenants of type Account and Customer.", businessTenantMapping.ID, businessTenantMapping.Type)
			return
		}

		opID, err := h.createSystemFetchingOperation(ctx, businessTenantMapping.ID)
		if err != nil {
			log.C(ctx).WithError(err).Error("Creating Operation for System Fetcher data aggregation failed")
			return
		}

		log.C(ctx).Infof("Successfully created operation with id %q for TenantID %q", opID, businessTenantMapping.ID)

		// Notify OperationProcessors for new operation
		h.onDemandChannel <- opID
		return
	}

	if skipReschedule {
		log.C(ctx).Debugf("SkipReschedule is true. Skipping reschedule for tenant with ID %q.", tenantID)
		return
	}

	if err = h.opMgr.RescheduleOperation(ctx, operation.ID); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to reschedule operation with ID %s", operation.ID)
		return
	}

	// Notify OperationProcessors for new operation
	h.onDemandChannel <- operation.ID
}

func (h *handler) createSystemFetchingOperation(ctx context.Context, tenantID string) (string, error) {
	now := time.Now()
	data := NewSystemFetcherOperationData(tenantID)
	rawData, err := data.GetData()
	if err != nil {
		return "", errors.Wrap(err, "while preparing system fetcher operation data")
	}

	newOperationInput := &model.OperationInput{
		OpType:    model.OperationTypeSystemFetching,
		Status:    model.OperationStatusScheduled,
		Data:      json.RawMessage(rawData),
		Priority:  int(operationsmanager.HighOperationPriority),
		CreatedAt: &now,
	}

	opID, err := h.opMgr.CreateOperation(ctx, newOperationInput)
	if err != nil {
		return "", errors.Wrap(err, "while creating system fetcher operation")
	}

	return opID, nil
}

func (h *handler) getBusinessTenantMappingByID(ctx context.Context, tenantID string) (*model.BusinessTenantMapping, error) {
	tx, err := h.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	businessTenantMapping, err := h.businessTenantMappingSvc.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return businessTenantMapping, tx.Commit()
}

func (h *handler) getBusinessTenantMappingByExternalID(ctx context.Context, externalID string) (*model.BusinessTenantMapping, error) {
	tx, err := h.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	businessTenantMapping, err := h.businessTenantMappingSvc.GetTenantByExternalID(ctx, externalID)
	if err != nil {
		return nil, err
	}
	return businessTenantMapping, tx.Commit()
}
