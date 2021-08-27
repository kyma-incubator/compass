package identification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	doathkeeper "github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
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
	ctx := req.Context()

	if req.Method != http.MethodPost {
		err := fmt.Sprintf("Bad request method. Got %s, expected POST", req.Method)
		log.C(ctx).Errorf(err)
		http.Error(writer, err, http.StatusBadRequest)
		return
	}

	reqData, err := doathkeeper.NewReqDataParser().Parse(req)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while parsing request: %v", err)
		respond(ctx, writer, reqData.Body)
		return
	}

	clientID := GetFromRequest(req)
	log.C(ctx).Infof("Request coming from client with ID '%s'", clientID)
	h.metricsCollector.InstrumentClientIdentification(clientID)

	respond(ctx, writer, reqData.Body)
}

func respond(ctx context.Context, writer http.ResponseWriter, body doathkeeper.ReqBody) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while encoding data: %v", err)
	}
}
