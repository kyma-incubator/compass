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
	opMgr *operationsmanager.OperationsManager
	cfg   MetricsConfig
}

// NewORDAggregatorHTTPHandler returns a new HTTP handler, responsible for handling HTTP requests
func NewORDAggregatorHTTPHandler(opMgr *operationsmanager.OperationsManager, cfg MetricsConfig) *handler {
	return &handler{
		opMgr: opMgr,
		cfg:   cfg,
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

	var operationID string

	operation, err := h.opMgr.FindOperationByData(ctx, NewOrdOperationData(payload.ApplicationID, payload.ApplicationTemplateID))
	if err != nil {
		if apperrors.IsNotFoundError(err) {
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

	err = h.opMgr.RescheduleOperation(ctx, operationID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to reschedule operation with ID %s", operationID)
		http.Error(writer, "Sheduling Operation for ORD data aggregation failed", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	// TODO notify OperationProcessors for new operation
}
