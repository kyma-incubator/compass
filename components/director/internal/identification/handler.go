package identification

import (
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

type Handler struct {
	metricsCollector *metrics.Collector
}

func NewHandler(metricsCollector *metrics.Collector) *Handler {
	return &Handler{
		metricsCollector: metricsCollector,
	}
}

func (h *Handler) ServerHTTP(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		err := fmt.Sprintf("Bad request method. Got %s, expected POST", req.Method)
		log.C(req.Context()).Errorf(err)
		http.Error(writer, err, http.StatusBadRequest)
		return
	}

	clientID := GetFromRequest(req)
	log.C(req.Context()).Infof("Request coming from client with ID '%s'", clientID)
	h.metricsCollector.InstrumentClientIdentification(clientID)

	writer.WriteHeader(http.StatusOK)
}
