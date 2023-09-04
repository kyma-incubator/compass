package ord

import (
	"encoding/json"
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
	opMgr      *operationsmanager.OperationsManager
	webhookSvc WebhookService
	cfg        MetricsConfig
}

// NewORDAggregatorHTTPHandler returns a new HTTP handler, responsible for handling HTTP requests
func NewORDAggregatorHTTPHandler(opMgr *operationsmanager.OperationsManager, webhookSvc WebhookService, cfg MetricsConfig) *handler {
	return &handler{
		opMgr:      opMgr,
		webhookSvc: webhookSvc,
		cfg:        cfg,
	}
}

// TODO - to process this asynchronously and to not return error
func (h *handler) ScheduleAggregationForORDData(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	payload := AggregationResources{}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to parse request body")
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	if payload.ApplicationID == "" && payload.ApplicationTemplateID == "" {
		// Invalid data as input - return error
		log.C(ctx).Errorf("Invalid data provided for ORD aggregation")
		http.Error(writer, "Invalid payload, neither Application ID and Application Template ID are provided.", http.StatusBadRequest)
		return
	}

	var operationID string

	operation, err := h.opMgr.FindOperationByData(ctx, NewOrdOperationData(payload.ApplicationID, payload.ApplicationTemplateID))
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			if payload.ApplicationID != "" && payload.ApplicationTemplateID == "" {
				ordWebhook, err := h.webhookSvc.GetByIDAndWebhookTypeGlobal(ctx, payload.ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeOpenResourceDiscovery)
				if err != nil {
					log.C(ctx).WithError(err).Errorf("Loading ORD webhooks of applicaiton with id %q for ORD aggregation failed", payload.ApplicationID)
					http.Error(writer, "Loading ORD webhooks of Application for ORD data aggregation failed", http.StatusInternalServerError)
					return
				}
				if ordWebhook == nil {
					// Exit with OK as this applicaiotn does not seem to have ORD webhooks
					writer.WriteHeader(http.StatusOK)
				}
				// Proceed to process application webhook
			}

			// If there are AppID and AppTemplateID defined in the operation data - process application template static ord and process the app in te context of appTmpl
			if payload.ApplicationID != "" && payload.ApplicationTemplateID != "" {
				// Check if app has ORD webhook and if not check if app-template has
				appOrdWebhook, err := h.webhookSvc.GetByIDAndWebhookTypeGlobal(ctx, payload.ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeOpenResourceDiscovery)
				if err != nil {
					log.C(ctx).WithError(err).Errorf("Loading ORD webhooks of applicaiton with id %q for ORD aggregation failed", payload.ApplicationID)
					http.Error(writer, "Loading ORD webhooks of Application for ORD data aggregation failed", http.StatusInternalServerError)
					return
				}
				if appOrdWebhook == nil {
					appTemplateOrdWebhook, err := h.webhookSvc.GetByIDAndWebhookTypeGlobal(ctx, payload.ApplicationTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscovery)
					if err != nil {
						log.C(ctx).WithError(err).Errorf("Loading ORD webhooks of applicaiton template with id %q for ORD aggregation failed", payload.ApplicationTemplateID)
						http.Error(writer, "Loading ORD webhooks of Application Template for ORD data aggregation failed", http.StatusInternalServerError)
						return
					}
					if appTemplateOrdWebhook == nil {
						// Exit with OK as this application and application template does not seem to have ORD webhooks
						writer.WriteHeader(http.StatusOK)
					}
					// Proceed to process application template webhook
				}
				// Proceed to process application webhook
			}

			// Aggregate only static ord
			if payload.ApplicationID == "" && payload.ApplicationTemplateID != "" {
				ordWebhook, err := h.webhookSvc.GetByIDAndWebhookTypeGlobal(ctx, payload.ApplicationTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscovery)
				if err != nil {
					log.C(ctx).WithError(err).Errorf("Loading ORD webhooks of applicaiton template with id %q for ORD aggregation failed", payload.ApplicationTemplateID)
					http.Error(writer, "Loading ORD webhooks of Application Template for ORD data aggregation failed", http.StatusInternalServerError)
					return
				}
				if ordWebhook == nil {
					// Exit with OK as this applicaiotn does not seem to have ORD webhooks
					writer.WriteHeader(http.StatusOK)
				}
				// Proceed to process application template webhook
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
		} else {
			log.C(ctx).WithError(err).Errorf("Loading Operation for ORD data aggregation failed")
			http.Error(writer, "Loading Operation for ORD data aggregation failed", http.StatusInternalServerError)
			return
		}
	} else {
		operationID = operation.ID
	}

	if len(operationID) != 0 {
		err = h.opMgr.RescheduleOperation(ctx, operationID)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Failed to reschedule operation with ID %s", operationID)
			http.Error(writer, "Sheduling Operation for ORD data aggregation failed", http.StatusInternalServerError)
			return
		}
	}

	writer.WriteHeader(http.StatusOK)
	// TODO notify OperationProcessors for new operation
}
