package ord

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"
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

//go:generate mockery --name=ORDService --output=automock --outpkg=automock --case=underscore --disable-version-string
// ORDService missing godoc
type ORDService interface {
	ProcessApplications(ctx context.Context, cfg MetricsConfig, appIDs []string) error
	ProcessApplicationTemplates(ctx context.Context, cfg MetricsConfig, appTemplateIDs []string) error
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

	if err := h.ordSvc.ProcessApplications(ctx, h.cfg, resources.ApplicationIDs); err != nil {
		log.C(ctx).WithError(err).Errorf("ORD data aggregation failed for one or more applications")
		http.Error(writer, "ORD data aggregation failed for one or more applications", http.StatusInternalServerError)
		return
	}

	if err := h.ordSvc.ProcessApplicationTemplates(ctx, h.cfg, resources.ApplicationTemplateIDs); err != nil {
		log.C(ctx).WithError(err).Errorf("ORD data aggregation failed for one or more application templates")
		http.Error(writer, "ORD data aggregation failed for one or more application templates", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
