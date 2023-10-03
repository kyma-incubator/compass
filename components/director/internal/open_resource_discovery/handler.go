package ord

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processors"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// OperationsManager missing godoc
//
//go:generate mockery --name=OperationsManager --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationsManager interface {
	FindOperationByData(ctx context.Context, data interface{}) (*model.Operation, error)
	CreateOperation(ctx context.Context, in *model.OperationInput) (string, error)
	RescheduleOperation(ctx context.Context, operationID string) error
}

// AggregationResources holds ids of resources for ord data aggregation
type AggregationResources struct {
	ApplicationID         string `json:"applicationID"`
	ApplicationTemplateID string `json:"applicationTemplateID"`
}

type handler struct {
	opMgr           OperationsManager
	appSvc          ApplicationService
	webhookSvc      processors.WebhookService
	transact        persistence.Transactioner
	onDemandChannel chan string
}

// NewORDAggregatorHTTPHandler returns a new HTTP handler, responsible for handling HTTP requests
func NewORDAggregatorHTTPHandler(opMgr OperationsManager, appSvc ApplicationService, webhookSvc processors.WebhookService, transact persistence.Transactioner, onDemandChannel chan string) *handler {
	return &handler{
		opMgr:           opMgr,
		appSvc:          appSvc,
		webhookSvc:      webhookSvc,
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

	log.C(ctx).Infof("Rescheduling ord data aggregation for application with id %q and application template with id %q", payload.ApplicationID, payload.ApplicationTemplateID)
	operation, err := h.opMgr.FindOperationByData(ctx, NewOrdOperationData(payload.ApplicationID, payload.ApplicationTemplateID))
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			log.C(ctx).WithError(err).Errorf("Loading Operation for ORD data aggregation failed")
			http.Error(writer, "Loading Operation for ORD data aggregation failed", http.StatusInternalServerError)
			return
		}

		log.C(ctx).Infof("Operation with ApplicationID %q and ApplicationTemplateID %q does not exist. Trying to create...", payload.ApplicationID, payload.ApplicationTemplateID)
		// Check if the provided application template has static Open Resource Discovery webhook
		if payload.ApplicationID == "" && payload.ApplicationTemplateID != "" {
			staticORDWebhook, err := h.getWebhookByObjectIDAndType(ctx, payload.ApplicationTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscoveryStatic)
			if err != nil {
				if apperrors.IsNotFoundError(err) {
					log.C(ctx).WithError(err).Errorf("ApplicationTemplate with id %q does not have static ORD webhook", payload.ApplicationTemplateID)
					http.Error(writer, "The provided ApplicationTemplate does not have static ORD webhook", http.StatusBadRequest)
					return
				}
				log.C(ctx).WithError(err).Errorf("Getting webhook of type %q for application template with id %q failed", model.WebhookTypeOpenResourceDiscoveryStatic, payload.ApplicationTemplateID)
				http.Error(writer, "Getting static Open Resource Discovery webhook for application template failed", http.StatusInternalServerError)
				return
			}
			log.C(ctx).Debugf("Static Open Resource Discovery webhook with id %q was found for Application Template with id %q", staticORDWebhook.ID, payload.ApplicationTemplateID)
		}

		// Check if the provided application has an ORD webhook
		if payload.ApplicationID != "" && payload.ApplicationTemplateID == "" {
			ordWebhook, err := h.getWebhookByObjectIDAndType(ctx, payload.ApplicationID, model.ApplicationWebhookReference, model.WebhookTypeOpenResourceDiscovery)
			if err != nil {
				if apperrors.IsNotFoundError(err) {
					log.C(ctx).WithError(err).Errorf("Application with id %q does not have ORD webhook", payload.ApplicationID)
					http.Error(writer, "The provided Application does not have ORD webhook", http.StatusBadRequest)
					return
				}
				log.C(ctx).WithError(err).Errorf("Getting webhook of type %q for application with id %q failed", model.WebhookTypeOpenResourceDiscovery, payload.ApplicationID)
				http.Error(writer, "Getting Open Resource Discovery webhook for application failed", http.StatusInternalServerError)
				return
			}
			log.C(ctx).Debugf("Open Resource Discovery webhook with id %q was found", ordWebhook.ID)
		}

		// Check if the provided application template has ORD webhook. If so, check if the application is created from the provided app template
		if payload.ApplicationID != "" && payload.ApplicationTemplateID != "" {
			ordWebhook, err := h.getWebhookByObjectIDAndType(ctx, payload.ApplicationTemplateID, model.ApplicationTemplateWebhookReference, model.WebhookTypeOpenResourceDiscovery)
			if err != nil {
				if apperrors.IsNotFoundError(err) {
					log.C(ctx).WithError(err).Errorf("ApplicationTemplate with id %q does not have ORD webhook", payload.ApplicationTemplateID)
					http.Error(writer, "The provided ApplicationTemplate does not have ORD webhook", http.StatusBadRequest)
					return
				}
				log.C(ctx).WithError(err).Errorf("Getting webhook of type %q for application template with id %q failed", model.WebhookTypeOpenResourceDiscovery, payload.ApplicationTemplateID)
				http.Error(writer, "Getting Open Resource Discovery webhook for application template failed", http.StatusInternalServerError)
				return
			}
			log.C(ctx).Debugf("Open Resource Discovery webhook with id %q was found for Application Template with id %q", ordWebhook.ID, payload.ApplicationTemplateID)

			app, err := h.getAppByID(ctx, payload.ApplicationID)
			if err != nil {
				if apperrors.IsNotFoundError(err) {
					log.C(ctx).WithError(err).Errorf("Application with id %q does not exist", payload.ApplicationID)
					http.Error(writer, "The provided Application does not exist", http.StatusBadRequest)
					return
				}
				log.C(ctx).WithError(err).Errorf("Getting application with id %q failed", payload.ApplicationTemplateID)
				http.Error(writer, "Getting Application failed", http.StatusInternalServerError)
				return
			}

			if *app.ApplicationTemplateID != payload.ApplicationTemplateID {
				log.C(ctx).WithError(err).Errorf("The provided Application with id %q is not created from the provided Application Template with id %q", payload.ApplicationID, payload.ApplicationTemplateID)
				http.Error(writer, "The provided Application is not created from the provided Application Template", http.StatusBadRequest)
				return
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

		opID, err := h.opMgr.CreateOperation(ctx, newOperationInput)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Creating Operation for ORD data aggregation failed")
			http.Error(writer, "Creating Operation for ORD data aggregation failed", http.StatusInternalServerError)
			return
		}
		log.C(ctx).Infof("Successfully created operation with ApplicationID %q and ApplicationTemplateID %q", payload.ApplicationID, payload.ApplicationTemplateID)

		// Notify OperationProcessors for new operation
		h.onDemandChannel <- opID

		writer.WriteHeader(http.StatusOK)
		return
	}

	if err = h.opMgr.RescheduleOperation(ctx, operation.ID); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to reschedule operation with ID %s", operation.ID)
		http.Error(writer, "Scheduling Operation for ORD data aggregation failed", http.StatusInternalServerError)
		return
	}
	// Notify OperationProcessors for new operation
	h.onDemandChannel <- operation.ID

	writer.WriteHeader(http.StatusOK)
}

func (h *handler) getWebhookByObjectIDAndType(ctx context.Context, objectID string, objectType model.WebhookReferenceObjectType, webhookType model.WebhookType) (*model.Webhook, error) {
	tx, err := h.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	ordWebhook, err := h.webhookSvc.GetByIDAndWebhookTypeGlobal(ctx, objectID, objectType, webhookType)
	if err != nil {
		return nil, err
	}

	return ordWebhook, tx.Commit()
}

func (h *handler) getAppByID(ctx context.Context, appID string) (*model.Application, error) {
	tx, err := h.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	app, err := h.appSvc.GetGlobalByID(ctx, appID)
	if err != nil {
		return nil, err
	}

	return app, tx.Commit()
}
