package subscription

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
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
	staticMappingClaims map[string]oauth.ClaimsGetterFunc
}

func NewHandler(tenantConfig Config, providerConfig ProviderConfig, externalSvcMockURL, oauthTokenPath, clientID, clientSecret string, staticMappingClaims map[string]oauth.ClaimsGetterFunc) *handler {
	return &handler{
		tenantConfig:        tenantConfig,
		providerConfig:      providerConfig,
		externalSvcMockURL:  externalSvcMockURL,
		oauthTokenPath:      oauthTokenPath,
		clientID:            clientID,
		clientSecret:        clientSecret,
		staticMappingClaims: staticMappingClaims,
	}
}

func (h *handler) Subscription(writer http.ResponseWriter, r *http.Request) {
	if err, statusCode := h.executeSubscriptionRequest(r, http.MethodPut); err != nil {
		log.C(r.Context()).Errorf("while executing subscription request: %v", err)
		httphelpers.WriteError(writer, errors.Wrap(err, "while executing subscription request"), statusCode)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func (h *handler) Deprovisioning(writer http.ResponseWriter, r *http.Request) {
	if err, statusCode := h.executeSubscriptionRequest(r, http.MethodDelete); err != nil {
		log.C(r.Context()).WithError(err).Errorf("while executing unsubscription request: %v", err)
		httphelpers.WriteError(writer, errors.Wrap(err, "while executing unsubscription request"), statusCode)
		return
	}
	writer.WriteHeader(http.StatusOK)
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

	consumerSubaccountID := mux.Vars(r)["tenant_id"]
	if consumerSubaccountID == "" {
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

	log.C(ctx).Infof("Creating/Removing subscription for consumer with subaccount id: %s", consumerSubaccountID)
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

// TODO:: Consider deleting this
func (h *handler) GetToken(writer http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log.C(ctx).Info("Issuing token...")

	data := url.Values{}
	data.Add("grant_type", "client_credentials")
	data.Add("client_id", h.clientID)

	//req, err := http.NewRequest(http.MethodPost, h.externalSvcMockURL+h.oauthTokenPath, strings.NewReader(data.Encode()))
	req, err := http.NewRequest(http.MethodPost, h.externalSvcMockURL+h.oauthTokenPath, bytes.NewBuffer([]byte(data.Encode())))
	if err != nil {
		fmt.Println(err)
	}

	req.SetBasicAuth(h.clientID, h.clientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.C(ctx).Errorf("while executing request for token: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while executing request for token"), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while closing response body: %v", err)
		}
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while reading response body: %v", err)
		httphelpers.WriteError(writer, errors.Wrap(err, "An error has occurred while reading response body"), http.StatusInternalServerError)
		return
	}

	token := gjson.GetBytes(b, "access_token")
	if !token.Exists() {
		log.C(ctx).WithError(err).Errorf("An error has occurred while reading response body: %v", err)
		httphelpers.WriteError(writer, errors.Wrap(err, "An error has occurred while reading response body"), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write([]byte(token.String()))
	if err != nil {
		log.C(ctx).Errorf("while writing response: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while writing response"), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
	log.C(ctx).Info("Successfully issued token")
}

func (h *handler) OnSubscription(writer http.ResponseWriter, r *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(compassURL)); err != nil {
		log.C(r.Context()).WithError(err).Errorf("Failed to write response body: %v", err)
	}
}

func (h *handler) DependenciesConfigure(writer http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.C(r.Context()).Errorf("while reading request body: %s", err.Error())
		httphelpers.WriteError(writer, errors.Wrap(err, "while reading request body"), http.StatusInternalServerError)
		return
	}

	if string(body) == "" {
		log.C(r.Context()).Error("The request body is empty")
		httphelpers.WriteError(writer, errors.New("The request body is empty"), http.StatusInternalServerError)
		return
	}

	h.xsappnameClone = string(body)
	writer.Header().Set("Content-Type", "application/json")
	if _, err := writer.Write(body); err != nil {
		log.C(r.Context()).WithError(err).Errorf("Failed to write response body for dependencies configure request")
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func (h *handler) Dependencies(writer http.ResponseWriter, r *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	if _, err := writer.Write([]byte(h.xsappnameClone)); err != nil {
		log.C(r.Context()).WithError(err).Errorf("Failed to write response body for dependencies request")
		return
	}
	writer.WriteHeader(http.StatusOK)
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

// TODO:: Consider deleting this
func tokenFromExternalSvcMockWithClaims(externalSvcMockUrl, oauthTokenPath, clientID, clientSecret string, claims map[string]interface{}) (string, error) {
	data, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, externalSvcMockUrl+oauthTokenPath, bytes.NewBuffer(data))
	req.SetBasicAuth(clientID, clientSecret)
	if err != nil {
		return "", err
	}

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.C(req.Context()).WithError(err).Errorf("An error has occurred while closing response body: %v", err)
		}
	}()

	tkn := gjson.GetBytes(body, "access_token")
	if !tkn.Exists() {
		return "", errors.New("did not found any token in the response")
	}

	return tkn.String(), nil
}
