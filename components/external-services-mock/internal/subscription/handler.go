package subscription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

const compassURL = "https://github.com/kyma-incubator/compass"

type handler struct {
	httpClient     *http.Client
	tenantConfig   Config
	providerConfig ProviderConfig
	xsappnameClone string
	jobID          string
}

type JobStatus struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

type Dependency struct {
	Xsappname string `json:"xsappname"`
}

func NewHandler(httpClient *http.Client, tenantConfig Config, providerConfig ProviderConfig, jobID string) *handler {
	return &handler{
		httpClient:     httpClient,
		tenantConfig:   tenantConfig,
		providerConfig: providerConfig,
		jobID:          jobID,
	}
}

func (h *handler) Subscribe(writer http.ResponseWriter, r *http.Request) {
	if statusCode, err := h.executeSubscriptionRequest(r, http.MethodPut); err != nil {
		log.C(r.Context()).Errorf("while executing subscribe request: %v", err)
		httphelpers.WriteError(writer, errors.Wrap(err, "while executing subscribe request"), statusCode)
		return
	}
	writer.Header().Set("Location", fmt.Sprintf("/api/v1/jobs/%s", h.jobID))
	writer.WriteHeader(http.StatusAccepted)
}

func (h *handler) Unsubscribe(writer http.ResponseWriter, r *http.Request) {
	if statusCode, err := h.executeSubscriptionRequest(r, http.MethodDelete); err != nil {
		log.C(r.Context()).Errorf("while executing unsubscribe request: %v", err)
		httphelpers.WriteError(writer, errors.Wrap(err, "while executing unsubscribe request"), statusCode)
		return
	}
	writer.Header().Set("Location", fmt.Sprintf("/api/v1/jobs/%s", h.jobID))
	writer.WriteHeader(http.StatusAccepted)
}

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

func (h *handler) OnSubscription(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("Handling on subscription request...")

	if r.Method != http.MethodPut && r.Method != http.MethodDelete {
		log.C(ctx).Errorf("expected %s or %s method but got: %s", http.MethodPut, http.MethodDelete, r.Method)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(compassURL)); err != nil {
		log.C(ctx).Errorf("while writing response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
	log.C(ctx).Info("Successfully handled on subscription request")
}

func (h *handler) DependenciesConfigure(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("Configuring subscription dependency...")
	if r.Method != http.MethodPost {
		log.C(ctx).Errorf("expected %s method but got: %s", http.MethodPost, r.Method)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(ctx).Errorf("while reading request body: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while reading request body"), http.StatusInternalServerError)
		return
	}

	if string(body) == "" {
		log.C(ctx).Error("The request body is empty")
		httphelpers.WriteError(writer, errors.New("The request body is empty"), http.StatusInternalServerError)
		return
	}

	h.xsappnameClone = string(body)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(body); err != nil {
		log.C(ctx).Errorf("while writing response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
	log.C(ctx).Infof("Successfully configured subscription dependency: %s", h.xsappnameClone)
}

func (h *handler) Dependencies(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("Handling dependency request...")

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

	deps := []*Dependency{{Xsappname: h.xsappnameClone}}
	depsMarshalled, err := json.Marshal(deps)
	if err != nil {
		log.C(ctx).Errorf("while marshalling subscription dependencies: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while marshalling subscription dependencies"), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(depsMarshalled); err != nil {
		log.C(ctx).Errorf("while writing response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
	log.C(ctx).Info("Successfully handled dependency request")
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

	// Build a request for consumer subscribe/unsubscribe
	BuildTenantFetcherRegionalURL(&h.tenantConfig)
	request, err := h.createTenantRequest(httpMethod, h.tenantConfig.TenantFetcherFullRegionalURL, token)
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

func (h *handler) createTenantRequest(httpMethod, tenantFetcherUrl, token string) (*http.Request, error) {
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

	request, err := http.NewRequest(httpMethod, tenantFetcherUrl, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	return request, nil
}
