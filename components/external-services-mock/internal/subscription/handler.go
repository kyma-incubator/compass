package subscription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

type handler struct {
	httpClient     *http.Client
	tenantConfig   Config
	providerConfig ProviderConfig
	jobID          string
}

type JobStatus struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

// NewHandler returns new subscription handler responsible to subscribe and unsubscribe tenants
func NewHandler(httpClient *http.Client, tenantConfig Config, providerConfig ProviderConfig, jobID string) *handler {
	return &handler{
		httpClient:     httpClient,
		tenantConfig:   tenantConfig,
		providerConfig: providerConfig,
		jobID:          jobID,
	}
}

// Subscribe build and execute subscribe request to tenant fetcher. This method is invoked on local setup,
// on real environment an external service with the same path but different host is called and then the request is propagated to tenant fetcher component as callbacks
func (h *handler) Subscribe(writer http.ResponseWriter, r *http.Request) {
	if statusCode, err := h.executeSubscriptionRequest(r, http.MethodPut); err != nil {
		log.C(r.Context()).Errorf("while executing subscribe request: %v", err)
		httphelpers.WriteError(writer, errors.Wrap(err, "while executing subscribe request"), statusCode)
		return
	}
	writer.Header().Set("Location", fmt.Sprintf("/api/v1/jobs/%s", h.jobID))
	writer.WriteHeader(http.StatusAccepted)
}

// Unsubscribe build and execute unsubscribe request to tenant fetcher. This method is invoked on local setup,
// on real environment an external service with the same path but different host is called and then the request is propagated to tenant fetcher component as callbacks
func (h *handler) Unsubscribe(writer http.ResponseWriter, r *http.Request) {
	if statusCode, err := h.executeSubscriptionRequest(r, http.MethodDelete); err != nil {
		log.C(r.Context()).Errorf("while executing unsubscribe request: %v", err)
		httphelpers.WriteError(writer, errors.Wrap(err, "while executing unsubscribe request"), statusCode)
		return
	}
	writer.Header().Set("Location", fmt.Sprintf("/api/v1/jobs/%s", h.jobID))
	writer.WriteHeader(http.StatusAccepted)
}

// JobStatus returns mock status of the asynchronous subscription job for testing purposes
func (h *handler) JobStatus(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("Handling subscription job status request...")

	authorization := r.Header.Get("Authorization")
	if len(authorization) == 0 {
		log.C(ctx).Error("authorization header is required")
		httphelpers.WriteError(writer, errors.New("authorization header is required"), http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authorization, "Bearer ")

	if !strings.HasPrefix(authorization, "Bearer ") || len(token) == 0 {
		log.C(ctx).Error("token value is required")
		httphelpers.WriteError(writer, errors.New("token value is required"), http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	jobStatus := &JobStatus{
		ID:    h.jobID,
		State: "SUCCEEDED",
	}

	payload, err := json.Marshal(jobStatus)
	if err != nil {
		log.C(ctx).Errorf("while marshalling response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling response"), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err = writer.Write(payload); err != nil {
		log.C(ctx).Errorf("while writing response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
	log.C(ctx).Info("Successfully handled subscription job status request")
}

func (h *handler) executeSubscriptionRequest(r *http.Request, httpMethod string) (int, error) {
	ctx := r.Context()
	authorization := r.Header.Get("Authorization")

	if len(authorization) == 0 {
		return http.StatusUnauthorized, errors.New("authorization header is required")
	}

	token := strings.TrimPrefix(authorization, "Bearer ")

	if !strings.HasPrefix(authorization, "Bearer ") || len(token) == 0 {
		return http.StatusUnauthorized, errors.New("token value is required")
	}

	consumerTenantID := mux.Vars(r)["tenant_id"]
	if consumerTenantID == "" {
		log.C(ctx).Error("parameter [tenant_id] not provided")
		return http.StatusBadRequest, errors.New("parameter [tenant_id] not provided")
	}
	providerSubaccID := r.Header.Get(h.tenantConfig.PropagatedProviderSubaccountHeader)

	// Build a request for consumer subscribe/unsubscribe
	BuildTenantFetcherRegionalURL(&h.tenantConfig)
	request, err := h.createTenantRequest(httpMethod, h.tenantConfig.TenantFetcherFullRegionalURL, token, providerSubaccID)
	if err != nil {
		log.C(ctx).Errorf("while creating subscription request: %s", err.Error())
		return http.StatusInternalServerError, errors.Wrap(err, "while creating subscription request")
	}

	log.C(ctx).Infof("Creating/Removing subscription for consumer with tenant id: %s and subaccount id: %s", consumerTenantID, h.tenantConfig.TestConsumerSubaccountID)
	resp, err := h.httpClient.Do(request)
	if err != nil {
		log.C(ctx).Errorf("while executing subscription request: %s", err.Error())
		return http.StatusInternalServerError, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.C(ctx).Errorf("wrong status code while executing subscription request, got [%d], expected [%d]", resp.StatusCode, http.StatusOK)
		return http.StatusInternalServerError, errors.New(fmt.Sprintf("wrong status code while executing subscription request, got [%d], expected [%d]", resp.StatusCode, http.StatusOK))
	}

	return http.StatusOK, nil
}

func (h *handler) createTenantRequest(httpMethod, tenantFetcherUrl, token, providerSubaccID string) (*http.Request, error) {
	var (
		body = "{}"
		err  error
	)

	if len(h.tenantConfig.TestConsumerAccountID) > 0 {
		body, err = sjson.Set(body, h.providerConfig.TenantIDProperty, h.tenantConfig.TestConsumerAccountID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
		}
	}
	if len(h.tenantConfig.TestConsumerSubaccountID) > 0 {
		body, err = sjson.Set(body, h.providerConfig.SubaccountTenantIDProperty, h.tenantConfig.TestConsumerSubaccountID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
		}
	}

	body, err = sjson.Set(body, h.providerConfig.SubdomainProperty, DefaultSubdomain)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
	}

	if len(h.tenantConfig.SubscriptionProviderID) > 0 {
		body, err = sjson.Set(body, h.providerConfig.SubscriptionProviderIDProperty, h.tenantConfig.SubscriptionProviderID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
		}
	}

	if len(providerSubaccID) > 0 { //Not present in unsubscription requests
		body, err = sjson.Set(body, h.providerConfig.ProviderSubaccountIDProperty, providerSubaccID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
		}
	}

	request, err := http.NewRequest(httpMethod, tenantFetcherUrl, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	return request, nil
}
