package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/kyma-incubator/compass/components/default-tenant-mapping-handler/internal/types"
	"github.com/pkg/errors"

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

// NewHandler creates a DefaultTenantMappingHandler
func NewHandler(mtlsHTTPClient mtlsHTTPClient) *DefaultTenantMappingHandler {
	return &DefaultTenantMappingHandler{
		mtlsHTTPClient: mtlsHTTPClient,
	}
}

// HandlerFunc is the implementation of DefaultTenantMappingHandler
func (tmh *DefaultTenantMappingHandler) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	log.C(ctx).Info("Default Tenant Mapping Handler was hit...")
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).Errorf("Failed to read request body: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusBadRequest, errors.New("Failed to read request body"))
		return
	}

	var tm types.TenantMapping
	err = json.Unmarshal(reqBody, &tm)
	if err != nil {
		log.C(ctx).Errorf("Failed to unmarshal request body: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusBadRequest, errors.New("Invalid json"))
		return
	}
	log.C(ctx).Info(tm.String())

	uclStatusAPIUrl := r.Header.Get(locationHeader)

	// respond with 202 to the UCL call
	httputils.Respond(w, http.StatusAccepted)

	correlationID := correlation.CorrelationIDFromContext(ctx)
	traceID := correlation.TraceIDFromContext(ctx)
	spanID := correlation.SpanIDFromContext(ctx)
	parentSpanID := correlation.ParentSpanIDFromContext(ctx)

	go tmh.callUCLStatusAPI(uclStatusAPIUrl, correlationID, traceID, spanID, parentSpanID)
}

func (tmh *DefaultTenantMappingHandler) callUCLStatusAPI(statusAPIURL, correlationID, traceID, spanID, parentSpanID string) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	ctx = correlation.AddCorrelationIDsToContext(ctx, correlationID, traceID, spanID, parentSpanID)

	logger := log.AddCorrelationIDsToLogger(ctx, correlationID, traceID, spanID, parentSpanID)
	ctx = log.ContextWithLogger(ctx, logger)

	reqBodyBytes, err := json.Marshal(types.SuccessResponse{State: readyState})
	if err != nil {
		log.C(ctx).WithError(err).Error("error while marshalling request body")
		return
	}

	if statusAPIURL == "" {
		log.C(ctx).Error("status API URL cannot be empty")
		return
	}

	req, err := http.NewRequest(http.MethodPatch, statusAPIURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		log.C(ctx).WithError(err).Error("error while building status API request")
		return
	}
	req.Header.Set(contentTypeHeaderKey, contentTypeApplicationJSON)
	req = req.WithContext(ctx)

	log.C(ctx).Infof("Default Tenant Mapping Handler reports notification status response to the status API URL: %s", statusAPIURL)
	resp, err := tmh.mtlsHTTPClient.Do(req)
	if err != nil {
		log.C(ctx).WithError(err).Error("error while executing request to the status API")
		return
	}
	defer closeResponseBody(ctx, resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while reading status API response")
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.C(ctx).WithError(err).Errorf("An error occurred while calling UCL status API. Received status: %d and body: %s", resp.StatusCode, body)
		return
	}
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("An error has occurred while closing response body: %v", err)
	}
}
