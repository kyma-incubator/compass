package ord

import (
	"fmt"
	"net/http"
)

type handler struct {
	ordSvc *Service
	cfg    MetricsConfig
}

// NewORDAggregatorHTTPHandler returns a new HTTP handler, responsible for handling HTTP requests
func NewORDAggregatorHTTPHandler(svc *Service, cfg MetricsConfig) *handler {
	return &handler{
		ordSvc: svc,
		cfg:    cfg,
	}
}

func (h *handler) ScheduleORDAggregation(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	//
	// tenantID, err := tenant.LoadFromContext(request.Context())
	// if err != nil {
	//	log.C(ctx).WithError(err).Error("Failed to fetch sensitive data for destinations")
	//	http.Error(writer, err.Error(), http.StatusBadRequest)
	//	return
	// }

	appID := request.URL.Query().Get("appID")
	appTemplateID := request.URL.Query().Get("appTemplateID")

	if appID == "" && appTemplateID == "" {
		err := fmt.Errorf("missing query parameter '%s' or '%s'", "appID", "appTemplateID")
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if appID != "" && appTemplateID != "" {
		err := fmt.Errorf("pass only one parameter -  '%s' or '%s'", "appID", "appTemplateID")
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if appID != "" {
		err := h.ordSvc.ProcessApp(ctx, h.cfg, appID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if appTemplateID != "" {
		err := h.ordSvc.ProcessAppTemplate(ctx, h.cfg, appTemplateID)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
	}

	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte("HTTP status code returned!")); err != nil {
		return
	}
}
