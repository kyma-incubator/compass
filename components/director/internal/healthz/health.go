package healthz

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	// UP missing godoc
	UP = "UP"
	// DOWN missing godoc
	DOWN = "DOWN"
)

// Health missing godoc
type Health struct {
	ctx        context.Context
	indicators []Indicator
	config     map[string]IndicatorConfig
}

// New returns new health with the given context
func New(ctx context.Context, healthCfg Config) (*Health, error) {
	if err := healthCfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "An error has occurred while validating indicator config")
	}

	cfg := make(map[string]IndicatorConfig)
	for _, i := range healthCfg.Indicators {
		cfg[i.Name] = i
	}

	return &Health{
		ctx:        ctx,
		config:     cfg,
		indicators: make([]Indicator, 0),
	}, nil
}

// RegisterIndicator registers the provided indicator - if configuration for indicator with the same name is present,
// it is used, otherwise a default indicator configuration is used
func (h *Health) RegisterIndicator(ind Indicator) *Health {
	cfg, ok := h.config[ind.Name()]
	if !ok {
		cfg = NewDefaultConfig()
	}
	ind.Configure(cfg)

	h.indicators = append(h.indicators, ind)
	return h
}

// Start will start all of the defined indicators.
// Each of the indicators run in their own goroutines
func (h *Health) Start() *Health {
	for _, ind := range h.indicators {
		ind.Run(h.ctx)
	}
	return h
}

// ReportStatus reports the status of all the indicators
func (h *Health) ReportStatus() string {
	state := UP
	for _, ind := range h.indicators {
		status := ind.Status()
		if status.Error() != nil {
			state = DOWN
			log.C(h.ctx).Errorf("Reporting status DOWN for indicator: %s, error: %s, details: %s", ind.Name(), status.Error(), status.Details())
		}
	}

	return state
}

// NewHealthHandler returns new health handler func with the provided health instance
func NewHealthHandler(h *Health) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		responseCode := http.StatusOK

		state := h.ReportStatus()
		if state == DOWN {
			responseCode = http.StatusInternalServerError
		}

		writer.WriteHeader(responseCode)
		if _, err := writer.Write([]byte(state)); err != nil {
			log.C(request.Context()).WithError(err).Errorf("An error has occurred while writing to response body: %v", err)
		}
	}
}
