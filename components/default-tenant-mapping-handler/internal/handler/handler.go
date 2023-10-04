package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	readyState                 = "READY"
	locationHeader             = "Location"
	contentTypeHeaderKey       = "Content-Type"
	contentTypeApplicationJSON = "application/json;charset=UTF-8"
)

//go:generate mockery --exported --name=mtlsHTTPClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type mtlsHTTPClient interface {
	Do(request *http.Request) (*http.Response, error)
}

// DefaultTenantMappingHandler processes received requests
type DefaultTenantMappingHandler struct {
	mtlsHTTPClient mtlsHTTPClient
}

// NewHandler creates an DefaultTenantMappingHandler
func NewHandler(mtlsHTTPClient mtlsHTTPClient) *DefaultTenantMappingHandler {
	return &DefaultTenantMappingHandler{
		mtlsHTTPClient: mtlsHTTPClient,
	}
}

// HandlerFunc is the implementation of DefaultTenantMappingHandler
func (tmh *DefaultTenantMappingHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.C(ctx).Info("Default Tenant Mapping Handler was hit...")

	uclStatusAPIUrl := r.Header.Get(locationHeader)

	// respond with 202 to the UCL call
	httputils.Respond(w, http.StatusAccepted)

	correlationID := correlation.CorrelationIDFromContext(ctx)

	log.C(ctx).Info("Default Tenant Mapping Handler reports status to the UCL status API...")
	go tmh.callUCLStatusAPI(uclStatusAPIUrl, correlationID)
}

func (tmh *DefaultTenantMappingHandler) callUCLStatusAPI(statusAPIURL, correlationID string) {
	time.Sleep(5 * time.Second) // todo::: temporary workaround for validation purposes
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	correlationIDKey := correlation.RequestIDHeaderKey
	ctx = correlation.SaveCorrelationIDHeaderToContext(ctx, &correlationIDKey, &correlationID)

	logger := log.C(ctx).WithField(correlationIDKey, correlationID)
	ctx = log.ContextWithLogger(ctx, logger)

	reqBodyBytes, err := json.Marshal(SuccessResponse{State: readyState})
	if err != nil {
		log.C(ctx).WithError(err).Error("error while marshalling request body")
		return
	}

	if statusAPIURL == "" {
		log.C(ctx).WithError(err).Error("status API URL is empty...")
		return
	}

	req, err := http.NewRequest(http.MethodPatch, statusAPIURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		log.C(ctx).WithError(err).Error("error while building status API request")
		return
	}
	req.Header.Set(contentTypeHeaderKey, contentTypeApplicationJSON)
	req = req.WithContext(ctx)

	resp, err := tmh.mtlsHTTPClient.Do(req)
	if err != nil {
		log.C(ctx).WithError(err).Error("error while executing request to the status API")
		return
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		log.C(ctx).WithError(err).Errorf("status API returned unexpected non OK status code: %d", resp.StatusCode)
		return
	}
}
