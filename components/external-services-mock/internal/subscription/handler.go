package subscription

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

const (
	tenantTokenClaimsKey = "tenant"
	// subscribedProviderIdValue, subscribedProviderAppNameValue and subscribedProviderSubaccountIdValue are used when CMP is an indirect dependency in a subscription flow
	subscribedProviderIdValue           = "subscribedProviderID"
	subscribedProviderAppNameValue      = "subscribedProviderAppName"
	subscribedProviderSubaccountIdValue = "subscribedProviderSubaccountID"
)

type handler struct {
	httpClient       *http.Client
	tenantConfig     Config
	providerConfig   ProviderConfig
	jobID            string
	tenantsHierarchy map[string]string // maps consumerSubaccount to consumerAccount
}

type JobStatus struct {
	Status string `json:"status"`
}

type ProviderSubscriptionInfo struct {
	ProviderSubaccountID     string
	ProviderSubscriptionsIds []string
}

var Subscriptions = make(map[string]*ProviderSubscriptionInfo)

// NewHandler returns new subscription handler responsible to subscribe and unsubscribe tenants
func NewHandler(httpClient *http.Client, tenantConfig Config, providerConfig ProviderConfig, jobID string) *handler {
	return &handler{
		httpClient:       httpClient,
		tenantConfig:     tenantConfig,
		providerConfig:   providerConfig,
		jobID:            jobID,
		tenantsHierarchy: map[string]string{tenantConfig.TestConsumerSubaccountIDTenantHierarchy: tenantConfig.TestConsumerAccountIDTenantHierarchy, tenantConfig.TestConsumerSubaccountID: tenantConfig.TestConsumerAccountID},
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
		Status: "COMPLETED",
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

	appName := mux.Vars(r)["app_name"]
	if appName == "" {
		log.C(ctx).Error("parameter [app_name] not provided")
		return http.StatusBadRequest, errors.New("parameter [app_name] not provided")
	}
	providerSubaccID := r.Header.Get(h.tenantConfig.PropagatedProviderSubaccountHeader)
	isIndirectDependency := r.Header.Get("isIndirectDependency")
	subscriptionID := getSubscriptionID(httpMethod, appName)
	if subscriptionID == "" {
		return http.StatusOK, nil
	}

	// Build a request for consumer subscribe/unsubscribe
	BuildTenantFetcherRegionalURL(&h.tenantConfig)
	request, err := h.createTenantRequest(httpMethod, h.tenantConfig.TenantFetcherFullRegionalURL, token, providerSubaccID, subscriptionID, isIndirectDependency)
	if err != nil {
		log.C(ctx).Errorf("while creating subscription request: %s", err.Error())
		return http.StatusInternalServerError, errors.Wrap(err, "while creating subscription request")
	}

	if httpMethod == http.MethodPut {
		log.C(ctx).Infof("Creating subscription for application with name %s", appName)
	} else {
		log.C(ctx).Infof("Removing subscription for application with name %s", appName)
	}
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.C(ctx).Errorf("while reading response body: %s", err.Error())
		return http.StatusInternalServerError, err
	}

	if resp.StatusCode != http.StatusOK {
		log.C(ctx).Errorf("wrong status code while executing subscription request, got [%d], expected [%d]", resp.StatusCode, http.StatusOK)
		return http.StatusInternalServerError, errors.New(fmt.Sprintf("wrong status code while executing subscription request, got [%d], expected [%d], reason: [%s]", resp.StatusCode, http.StatusOK, body))
	}
	if httpMethod == http.MethodPut {
		addSubscription(appName, providerSubaccID, subscriptionID)
	} else if httpMethod == http.MethodDelete {
		removeSubscription(appName)
	}

	return http.StatusOK, nil
}

func (h *handler) createTenantRequest(httpMethod, tenantFetcherUrl, token, providerSubaccID, subscriptionID, isIndirectDependency string) (*http.Request, error) {
	var (
		body = "{}"
		err  error
	)

	consumerSubaccountID, err := extractValueFromTokenClaims(token, tenantTokenClaimsKey)
	if err != nil {
		return nil, errors.New("error occurred when extracting consumer subaccount from token claims")
	}

	consumerAccountID := h.tenantsHierarchy[consumerSubaccountID]

	if len(consumerAccountID) > 0 {
		body, err = sjson.Set(body, h.providerConfig.TenantIDProperty, consumerAccountID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
		}
	}
	if len(consumerSubaccountID) > 0 {
		body, err = sjson.Set(body, h.providerConfig.SubaccountTenantIDProperty, consumerSubaccountID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
		}
	}

	body, err = sjson.Set(body, h.providerConfig.SubdomainProperty, DefaultSubdomain)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
	}

	body, err = sjson.Set(body, h.providerConfig.DependentServiceInstancesInfoProperty, []map[string]string{{h.providerConfig.DependentServiceInstancesInfoAppIDProperty: h.tenantConfig.SubscriptionProviderID, h.providerConfig.DependentServiceInstancesInfoAppNameProperty: h.tenantConfig.SubscriptionProviderAppNameValue, h.providerConfig.DependentServiceInstancesInfoProviderSubaccountIDProperty: providerSubaccID}})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
	}

	body, err = sjson.Set(body, h.providerConfig.SubscriptionIDProperty, subscriptionID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
	}
	fmt.Printf("subscriptions: %v", Subscriptions)

	body, err = sjson.Set(body, h.providerConfig.LicenseTypeProperty, DefaultLicenseType)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
	}

	if len(h.tenantConfig.TestConsumerTenantID) > 0 {
		body, err = sjson.Set(body, h.providerConfig.ConsumerTenantIDProperty, h.tenantConfig.TestConsumerTenantID)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
		}
	}

	if isIndirectDependency != "" { // is indirect dependency
		body, err = h.setProviderValues(body, subscribedProviderIdValue, subscribedProviderAppNameValue, subscribedProviderSubaccountIdValue)
		if err != nil {
			return nil, err
		}
	} else { // is direct dependency
		body, err = h.setProviderValues(body, h.tenantConfig.SubscriptionProviderID, h.tenantConfig.SubscriptionProviderAppNameValue, providerSubaccID)
		if err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequest(httpMethod, tenantFetcherUrl, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	return request, nil
}

func extractValueFromTokenClaims(consumerToken, claimsKey string) (string, error) {
	// JWT format: <header>.<payload>.<signature>
	tokenParts := strings.Split(consumerToken, ".")
	if len(tokenParts) != 3 {
		return "", errors.New("invalid token format")
	}
	payload := tokenParts[1]

	consumerTokenPayload, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", err
	}

	var jsonMap map[string]interface{}
	err = json.Unmarshal(consumerTokenPayload, &jsonMap)
	if err != nil {
		return "", err
	}

	return jsonMap[claimsKey].(string), nil
}

func getSubscriptionID(httpMethod, appName string) string {
	if httpMethod == http.MethodPut {
		return uuid.New().String()
	}
	if httpMethod == http.MethodDelete {
		if provider, exists := Subscriptions[appName]; exists {
			return provider.ProviderSubscriptionsIds[0]
		}
	}
	return ""
}

func addSubscription(appName, providerSubaccountID, subscriptionID string) {
	if provider, exists := Subscriptions[appName]; !exists {
		Subscriptions[appName] = &ProviderSubscriptionInfo{
			ProviderSubaccountID:     providerSubaccountID,
			ProviderSubscriptionsIds: []string{subscriptionID},
		}
	} else {
		provider.ProviderSubscriptionsIds = append(provider.ProviderSubscriptionsIds, subscriptionID)
	}
}

func removeSubscription(appName string) {
	provider := Subscriptions[appName]
	if len(provider.ProviderSubscriptionsIds) == 1 {
		delete(Subscriptions, appName)
	} else {
		provider.ProviderSubscriptionsIds = provider.ProviderSubscriptionsIds[1:] //remove first element
	}
}

func (h *handler) setProviderValues(body, providerID, appName, subaccountID string) (string, error) {
	var err error
	if len(h.tenantConfig.SubscriptionProviderID) > 0 {
		body, err = sjson.Set(body, h.providerConfig.SubscriptionProviderIDProperty, providerID)
		if err != nil {
			return "", errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
		}
	}

	if len(h.tenantConfig.SubscriptionProviderAppNameValue) > 0 {
		body, err = sjson.Set(body, h.providerConfig.SubscriptionProviderAppNameProperty, appName)
		if err != nil {
			return "", errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
		}
	}

	body, err = sjson.Set(body, h.providerConfig.ProviderSubaccountIDProperty, subaccountID)
	if err != nil {
		return "", errors.New(fmt.Sprintf("An error occured when setting json value: %v", err))
	}

	return body, nil
}
