package ord

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// AggregationResources holds ids of resources for ord data aggregation
type AggregationResources struct {
	ApplicationIDs         []string `json:"applicationIDs"`
	ApplicationTemplateIDs []string `json:"applicationTemplateIDs"`
}

type handler struct {
	ordSvc ORDService
	cfg    MetricsConfig
}

// ORDService missing godoc
//
//go:generate mockery --name=ORDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ORDService interface {
	ProcessApplication(ctx context.Context, appID string) error
	ProcessApplicationTemplate(ctx context.Context, appTemplateID string) error
}

// NewORDAggregatorHTTPHandler returns a new HTTP handler, responsible for handling HTTP requests
func NewORDAggregatorHTTPHandler(svc ORDService, cfg MetricsConfig) *handler {
	return &handler{
		ordSvc: svc,
		cfg:    cfg,
	}
}

func (h *handler) AggregateORDData(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	resources := AggregationResources{}
	if err := json.NewDecoder(request.Body).Decode(&resources); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to parse request body")
		http.Error(writer, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, appID := range resources.ApplicationIDs {
		if err := h.ordSvc.ProcessApplication(ctx, appID); err != nil {
			log.C(ctx).WithError(err).Errorf("ORD data aggregation failed for one or more applications")
			http.Error(writer, "ORD data aggregation failed for one or more applications", http.StatusInternalServerError)
			return
		}
	}

	for _, appTemplateID := range resources.ApplicationTemplateIDs {
		if err := h.ordSvc.ProcessApplicationTemplate(ctx, appTemplateID); err != nil {
			log.C(ctx).WithError(err).Errorf("ORD data aggregation failed for one or more application templates")
			http.Error(writer, "ORD data aggregation failed for one or more application templates", http.StatusInternalServerError)
			return
		}
	}

	writer.WriteHeader(http.StatusOK)
}
