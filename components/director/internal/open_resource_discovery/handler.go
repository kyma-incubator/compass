package ord

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// AggregationResources holds ids of resources for ord data aggregation
type AggregationResources struct {
	ApplicationID         string `json:"applicationID"`
	ApplicationTemplateID string `json:"applicationTemplateID"`
}

type handler struct {
	opMgr           *operationsmanager.OperationsManager
	webhookSvc      WebhookService
	cfg             MetricsConfig
	transact        persistence.Transactioner
	onDemandChannel chan string
}

// NewORDAggregatorHTTPHandler returns a new HTTP handler, responsible for handling HTTP requests
func NewORDAggregatorHTTPHandler(opMgr *operationsmanager.OperationsManager, webhookSvc WebhookService, cfg MetricsConfig, transact persistence.Transactioner, onDemandChannel chan string) *handler {
	return &handler{
		opMgr:           opMgr,
		webhookSvc:      webhookSvc,
		cfg:             cfg,
		transact:        transact,
		onDemandChannel: onDemandChannel,
	}
}

// ScheduleAggregationForORDData validates the payload, checks if such an operation already exists.
// If it does, it reschedules the existing operation; otherwise, it creates a new operation with high priority.
func (h *handler) ScheduleAggregationForORDData(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	payload := AggregationResources{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to parse request body")
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	if payload.ApplicationID == "" && payload.ApplicationTemplateID == "" {
		log.C(ctx).Errorf("Invalid data provided for ORD aggregation")
		http.Error(writer, "Invalid payload, neither Application ID and Application Template ID are provided.", http.StatusBadRequest)
		return
	}

	var operationID string

	operation, err := h.opMgr.FindOperationByData(ctx, NewOrdOperationData(payload.ApplicationID, payload.ApplicationTemplateID))
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			log.C(ctx).WithError(err).Errorf("Loading Operation for ORD data aggregation failed")
			http.Error(writer, "Loading Operation for ORD data aggregation failed", http.StatusInternalServerError)
			return
		}

		// Check if the provided application has an ORD webhook
		if payload.ApplicationID != "" && payload.ApplicationTemplateID == "" {
			if _, err := h.getORDWebhookByIDGlobal(ctx, payload.ApplicationID, model.ApplicationWebhookReference); err != nil {
				if apperrors.IsNotFoundError(err) {
					log.C(ctx).WithError(err).Errorf("Application with id %q does not have ORD webhook", payload.ApplicationID)
					http.Error(writer, "The provided Application does not have ORD webhook", http.StatusBadRequest)
					return
				}
				log.C(ctx).WithError(err).Errorf("Loading ORD webhooks of application with id %q for ORD aggregation failed", payload.ApplicationID)
				http.Error(writer, "Loading ORD webhooks of Application for ORD data aggregation failed", http.StatusInternalServerError)
				return
			}
		}

		// Check if the provided application template has an ORD webhook
		if payload.ApplicationID == "" && payload.ApplicationTemplateID != "" {
			if _, err := h.getORDWebhookByIDGlobal(ctx, payload.ApplicationTemplateID, model.ApplicationTemplateWebhookReference); err != nil {
				if apperrors.IsNotFoundError(err) {
					log.C(ctx).WithError(err).Errorf("ApplicationTemplate with id %q does not have ORD webhook", payload.ApplicationTemplateID)
					http.Error(writer, "The provided ApplicationTemplate does not have ORD webhook", http.StatusBadRequest)
					return
				}
				log.C(ctx).WithError(err).Errorf("Loading ORD webhooks of application template with id %q for ORD aggregation failed", payload.ApplicationTemplateID)
				http.Error(writer, "Loading ORD webhooks of Application Template for ORD data aggregation failed", http.StatusInternalServerError)
				return
			}
		}

		// Check if the provided application has ORD webhook. If it does not - check if the application template has
		if payload.ApplicationID != "" && payload.ApplicationTemplateID != "" {
			if _, err := h.getORDWebhookByIDGlobal(ctx, payload.ApplicationID, model.ApplicationWebhookReference); err != nil {
				if !apperrors.IsNotFoundError(err) {
					log.C(ctx).WithError(err).Errorf("Loading ORD webhooks of application with id %q for ORD aggregation failed", payload.ApplicationID)
					http.Error(writer, "Loading ORD webhooks of Application for ORD data aggregation failed", http.StatusInternalServerError)
					return
				}

				if _, err := h.getORDWebhookByIDGlobal(ctx, payload.ApplicationTemplateID, model.ApplicationTemplateWebhookReference); err != nil {
					if apperrors.IsNotFoundError(err) {
						log.C(ctx).WithError(err).Errorf("ApplicationTemplate with id %q does not have ORD webhook", payload.ApplicationTemplateID)
						http.Error(writer, "The provided ApplicationTemplate does not have ORD webhook", http.StatusBadRequest)
						return
					}
					log.C(ctx).WithError(err).Errorf("Loading ORD webhooks of application template with id %q for ORD aggregation failed", payload.ApplicationTemplateID)
					http.Error(writer, "Loading ORD webhooks of Application Template for ORD data aggregation failed", http.StatusInternalServerError)
					return
				}
			}
		}

		now := time.Now()
		data := NewOrdOperationData(payload.ApplicationID, payload.ApplicationTemplateID)
		rawData, err := data.GetData()
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Preparing Operation for ORD data aggregation failed")
			http.Error(writer, "Preparing Operation for ORD data aggregation failed", http.StatusInternalServerError)
			return
		}

		newOperationInput := &model.OperationInput{
			OpType:    model.OperationTypeOrdAggregation,
			Status:    model.OperationStatusScheduled,
			Data:      json.RawMessage(rawData),
			Priority:  int(operationsmanager.HighOperationPriority),
			CreatedAt: &now,
		}

		operationID, err = h.opMgr.CreateOperation(ctx, newOperationInput)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Creating Operation for ORD data aggregation failed")
			http.Error(writer, "Creating Operation for ORD data aggregation failed", http.StatusInternalServerError)
			return
		}
	}

	if operation != nil {
		operationID = operation.ID
	}

	if len(operationID) != 0 {
		err = h.opMgr.RescheduleOperation(ctx, operationID)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Failed to reschedule operation with ID %s", operationID)
			http.Error(writer, "Scheduling Operation for ORD data aggregation failed", http.StatusInternalServerError)
			return
		}
		// Notify OperationProcessors for new operation
		h.onDemandChannel <- operationID
	}
	writer.WriteHeader(http.StatusOK)
}

func (h *handler) getORDWebhookByIDGlobal(ctx context.Context, objectID string, objectType model.WebhookReferenceObjectType) (*model.Webhook, error) {
	tx, err := h.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	ordWebhook, err := h.webhookSvc.GetByIDAndWebhookTypeGlobal(ctx, objectID, objectType, model.WebhookTypeOpenResourceDiscovery)
	if err != nil {
		return nil, err
	}

	return ordWebhook, tx.Commit()
}
