package systemfetcher

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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
	TenantID       string `json:"tenantID"`
	SkipReschedule bool   `json:"skipReschedule"`
}

type handler struct {
	opMgr                    OperationsManager
	businessTenantMappingSvc BusinessTenantMappingService
	transact                 persistence.Transactioner
	onDemandChannel          chan string
}

// NewSystemFetcherAggregatorHTTPHandler returns a new HTTP handler, responsible for handling HTTP requests
func NewSystemFetcherAggregatorHTTPHandler(opMgr OperationsManager, businessTenantMappingSvc BusinessTenantMappingService, transact persistence.Transactioner, onDemandChannel chan string) *handler {
	return &handler{
		opMgr:                    opMgr,
		businessTenantMappingSvc: businessTenantMappingSvc,
		transact:                 transact,
		onDemandChannel:          onDemandChannel,
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

	if payload.TenantID == "" {
		log.C(ctx).Error("Invalid data provided for System Fetcher aggregation")
		http.Error(writer, "Invalid payload, Tenant ID is not provided.", http.StatusBadRequest)
		return
	}

	log.C(ctx).Infof("Rescheduling system fetcher data aggregation for tenant with id %q", payload.TenantID)
	operation, err := h.opMgr.FindOperationByData(ctx, NewSystemFetcherOperationData(payload.TenantID))
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			log.C(ctx).WithError(err).Error("Loading Operation for System Fetcher data aggregation failed")
			http.Error(writer, "Loading Operation for System Fetcher data aggregation failed", http.StatusInternalServerError)
			return
		}

		log.C(ctx).Infof("Operation with TenantID %q does not exist. Trying to create...", payload.TenantID)

		// Check if the provided tenant exists.
		businessTenantMappingID := payload.TenantID
		err := h.existsBusinessTenantMappingByID(ctx, payload.TenantID)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				err = h.existsBusinessTenantMappingByExternalTenant(ctx, payload.TenantID)
				if err != nil {
					if apperrors.IsNotFoundError(err) {
						log.C(ctx).WithError(err).Errorf("External tenant with id %q not found", payload.TenantID)
						http.Error(writer, "External Tenant not found", http.StatusNotFound)
						return
					} else {
						log.C(ctx).WithError(err).Errorf("Check for external tenant with id %q failed", payload.TenantID)
						http.Error(writer, "Check for External Tenant failed", http.StatusInternalServerError)
						return
					}
				}
				businessTenantMappingID, err = h.getBusinessTenantMappingByExternalTenant(ctx, payload.TenantID)
				if err != nil {
					log.C(ctx).WithError(err).Errorf("Getting external tenant with id %q failed", payload.TenantID)
					http.Error(writer, "Getting External Tenant failed", http.StatusInternalServerError)
					return
				}
			} else {
				log.C(ctx).WithError(err).Errorf("Getting tenant with id %q failed", payload.TenantID)
				http.Error(writer, "Getting Tenant failed", http.StatusInternalServerError)
				return
			}
		}

		businessTenantMapping, err := h.getBusinessTenantMappingByID(ctx, businessTenantMappingID)
		if err != nil || businessTenantMapping == nil {
			if err != nil {
				log.C(ctx).WithError(err).Error("Loading Business Tenant Mapping for System Fetcher data aggregation failed")
			} else {
				log.C(ctx).Error("Loading Business Tenant Mapping for System Fetcher data aggregation failed")
			}
			http.Error(writer, "Loading Business Tenant Mapping for System Fetcher data aggregation failed", http.StatusInternalServerError)
			return
		}

		if businessTenantMapping.Type != tenant.Account && businessTenantMapping.Type != tenant.Customer {
			log.C(ctx).Infof("Tenant with ID %q is of type %q - operations are created only for tenants of type Account and Customer.", businessTenantMapping.ID, businessTenantMapping.Type)
			writer.WriteHeader(http.StatusOK)
			return
		}

		now := time.Now()
		data := NewSystemFetcherOperationData(businessTenantMappingID)
		rawData, err := data.GetData()
		if err != nil {
			log.C(ctx).WithError(err).Error("Preparing Operation for System Fetcher data aggregation failed")
			http.Error(writer, "Preparing Operation for System Fetcher data aggregation failed", http.StatusInternalServerError)
			return
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
			log.C(ctx).WithError(err).Error("Creating Operation for System Fetcher data aggregation failed")
			http.Error(writer, "Creating Operation for System Fetcher data aggregation failed", http.StatusInternalServerError)
			return
		}
		log.C(ctx).Infof("Successfully created operation with TenantID %q", businessTenantMappingID)

		// Notify OperationProcessors for new operation
		h.onDemandChannel <- opID

		writer.WriteHeader(http.StatusOK)
		return
	}

	if payload.SkipReschedule {
		log.C(ctx).Debugf("Skipping reschedule for tenant with ID %q.", payload.TenantID)
		writer.WriteHeader(http.StatusOK)
		return
	}

	if err = h.opMgr.RescheduleOperation(ctx, operation.ID); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to reschedule operation with ID %s", operation.ID)
		http.Error(writer, "Scheduling Operation for System Fetcher data aggregation failed", http.StatusInternalServerError)
		return
	}
	// Notify OperationProcessors for new operation
	h.onDemandChannel <- operation.ID

	writer.WriteHeader(http.StatusOK)
}

func (h *handler) existsBusinessTenantMappingByID(ctx context.Context, businessTenantMappingID string) error {
	tx, err := h.transact.Begin()
	if err != nil {
		return err
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	err = h.businessTenantMappingSvc.Exists(ctx, businessTenantMappingID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (h *handler) existsBusinessTenantMappingByExternalTenant(ctx context.Context, externalTenant string) error {
	tx, err := h.transact.Begin()
	if err != nil {
		return err
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	err = h.businessTenantMappingSvc.ExistsByExternalTenant(ctx, externalTenant)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (h *handler) getBusinessTenantMappingByExternalTenant(ctx context.Context, externalTenant string) (string, error) {
	tx, err := h.transact.Begin()
	if err != nil {
		return "", err
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	businessTenantMappingID, err := h.businessTenantMappingSvc.GetInternalTenant(ctx, externalTenant)
	if err != nil {
		return "", err
	}
	return businessTenantMappingID, tx.Commit()
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
