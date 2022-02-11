package subscription

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

const compassURL = "https://github.com/kyma-incubator/compass"

type handler struct {
	tenantConfig        Config
	providerConfig      ProviderConfig
	externalSvcMockURL  string
	oauthTokenPath      string
	clientID            string
	clientSecret        string
	xsappnameClone      string
	jobID               string
	staticMappingClaims map[string]oauth.ClaimsGetterFunc
}

type JobStatus struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

type Dependency struct {
	Xsappname string `json:"xsappname"`
}

func NewHandler(tenantConfig Config, providerConfig ProviderConfig, externalSvcMockURL, oauthTokenPath, clientID, clientSecret, jobID string, staticMappingClaims map[string]oauth.ClaimsGetterFunc) *handler {
	return &handler{
		tenantConfig:        tenantConfig,
		providerConfig:      providerConfig,
		externalSvcMockURL:  externalSvcMockURL,
		oauthTokenPath:      oauthTokenPath,
		clientID:            clientID,
		clientSecret:        clientSecret,
		jobID:               jobID,
		staticMappingClaims: staticMappingClaims,
	}
}

func (h *handler) Subscription(writer http.ResponseWriter, r *http.Request) {
	if err, statusCode := h.executeSubscriptionRequest(r, http.MethodPut); err != nil {
		log.C(r.Context()).Errorf("while executing subscription request: %v", err)
		httphelpers.WriteError(writer, errors.Wrap(err, "while executing subscription request"), statusCode)
		return
	}
	writer.Header().Set("Location", fmt.Sprintf("/api/v1/jobs/%s", h.jobID))
	writer.WriteHeader(http.StatusAccepted)
}

func (h *handler) Deprovisioning(writer http.ResponseWriter, r *http.Request) {
	if err, statusCode := h.executeSubscriptionRequest(r, http.MethodDelete); err != nil {
		log.C(r.Context()).Errorf("while executing unsubscription request: %v", err)
		httphelpers.WriteError(writer, errors.Wrap(err, "while executing unsubscription request"), statusCode)
		return
	}
	writer.Header().Set("Location", fmt.Sprintf("/api/v1/jobs/%s", h.jobID))
	writer.WriteHeader(http.StatusAccepted)
}

func (h *handler) JobStatus(writer http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	jobStatus := &JobStatus{
		ID:    h.jobID,
		State: "SUCCEEDED",
	}

	ctx := r.Context()
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
}

func (h *handler) OnSubscription(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("Handling on subscription request...")
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

func (h *handler) executeSubscriptionRequest(r *http.Request, httpMethod string) (error, int) {
	ctx := r.Context()
	authorization := r.Header.Get("Authorization")

	if len(authorization) == 0 {
		return errors.New("authorization header is required"), http.StatusBadRequest
	}

	token := strings.TrimPrefix(authorization, "Bearer ")

	if !strings.HasPrefix(authorization, "Bearer ") || len(token) == 0 {
		return errors.New("token value is required"), http.StatusBadRequest
	}

	consumerTenantID := mux.Vars(r)["tenant_id"]
	if consumerTenantID == "" {
		log.C(ctx).Error("parameter [tenant_id] not provided")
		return errors.New("parameter [tenant_id] not provided"), http.StatusBadRequest
	}

	// Build a request for consumer subscription/unsubscription
	BuildTenantFetcherRegionalURL(&h.tenantConfig)
	request, err := h.createTenantRequest(httpMethod, h.tenantConfig.TenantFetcherFullRegionalURL, token)
	if err != nil {
		log.C(ctx).Errorf("while creating subscription request: %s", err.Error())
		return errors.Wrap(err, "while creating subscription request"), http.StatusInternalServerError
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	log.C(ctx).Infof("Creating/Removing subscription for consumer with tenant id: %s and subaccount id: %s", consumerTenantID, h.tenantConfig.TestConsumerSubaccountID)
	resp, err := httpClient.Do(request)
	if err != nil {
		log.C(ctx).Errorf("while executing subscription request: %s", err.Error())
		return err, http.StatusInternalServerError
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.C(ctx).Errorf("wrong status code while executing subscription request, got [%d], expected [%d]", resp.StatusCode, http.StatusOK)
		return errors.Wrapf(err, "wrong status code while executing subscription request, got [%d], expected [%d]", resp.StatusCode, http.StatusOK), http.StatusInternalServerError
	}

	return nil, http.StatusOK
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
