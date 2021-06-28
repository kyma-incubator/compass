package healthz

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	UP   = "UP"
	DOWN = "DOWN"
)

type Health struct {
	ctx        context.Context
	indicators []Indicator
	config     map[string]IndicatorConfig
}

// New returns new health with the given context
func New(ctx context.Context, healthCfg Config) *Health {
	cfg := make(map[string]IndicatorConfig)
	for _, i := range healthCfg.Indicators {
		cfg[i.Name] = i
	}

	return &Health{
		ctx:        ctx,
		config:     cfg,
		indicators: make([]Indicator, 0),
	}
}

// RegisterIndicator registers indicator, if config is present for the provided one
// it is used, unless uses default config
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
func (h *Health) Start() (*Health, error) {
	for _, ind := range h.indicators {
		if err := ind.Run(h.ctx); err != nil {
			log.C(h.ctx).Errorf("Error when starting indicator %s: %s", ind.Name(), err.Error())
			return nil, err
		}
	}
	return h, nil
}

// ReportStatus reports the status of all indicators
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

// NewHealthHandler returns new health handler func
// with the provided health instance
func NewHealthHandler(h *Health) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		responseCode := http.StatusOK

		state := h.ReportStatus()
		if state == DOWN {
			responseCode = http.StatusInternalServerError
		}

		writer.WriteHeader(responseCode)
		_, err := writer.Write([]byte(state))
		if err != nil {
			log.C(request.Context()).WithError(err).Errorf("An error has occurred while writing to response body: %v", err)
		}
	}
}
