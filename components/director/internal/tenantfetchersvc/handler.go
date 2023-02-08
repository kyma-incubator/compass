package tenantfetchersvc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/tidwall/gjson"
)

const (
	// InternalServerError message
	InternalServerError = "Internal Server Error"
	compassURL          = "https://github.com/kyma-incubator/compass"
)

// TenantFetcher is used to fectch tenants for creation;
//
//go:generate mockery --name=TenantFetcher --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantFetcher interface {
	SynchronizeTenant(ctx context.Context, parentTenantID, tenantID string) error
}

// TenantSubscriber is used to apply subscription changes for tenants;
//
//go:generate mockery --name=TenantSubscriber --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantSubscriber interface {
	Subscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest) error
	Unsubscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest) error
}

// HandlerConfig is the configuration required by the tenant handler.
// It includes configurable parameters for incoming requests, including different tenant IDs json properties, and path parameters.
type HandlerConfig struct {
	TenantOnDemandHandlerEndpoint      string `envconfig:"APP_TENANT_ON_DEMAND_HANDLER_ENDPOINT,default=/v1/fetch/{parentTenantId}/{tenantId}"`
	RegionalHandlerEndpoint            string `envconfig:"APP_REGIONAL_HANDLER_ENDPOINT,default=/v1/regional/{region}/callback/{tenantId}"`
	DependenciesEndpoint               string `envconfig:"APP_REGIONAL_DEPENDENCIES_ENDPOINT,default=/v1/regional/{region}/dependencies"`
	TenantPathParam                    string `envconfig:"APP_TENANT_PATH_PARAM,default=tenantId"`
	ParentTenantPathParam              string `envconfig:"APP_PARENT_TENANT_PATH_PARAM,default=parentTenantId"`
	RegionPathParam                    string `envconfig:"APP_REGION_PATH_PARAM,default=region"`
	XsAppNamePathParam                 string `envconfig:"APP_TENANT_FETCHER_XSAPPNAME_PATH,default=xsappname"`
	OmitDependenciesCallbackParam      string `envconfig:"APP_TENANT_FETCHER_OMIT_PARAM_NAME"`
	OmitDependenciesCallbackParamValue string `envconfig:"APP_TENANT_FETCHER_OMIT_PARAM_VALUE"`

	Database persistence.DatabaseConfig

	DirectorGraphQLEndpoint     string        `envconfig:"APP_DIRECTOR_GRAPHQL_ENDPOINT"`
	ClientTimeout               time.Duration `envconfig:"default=60s"`
	HTTPClientSkipSslValidation bool          `envconfig:"APP_HTTP_CLIENT_SKIP_SSL_VALIDATION,default=false"`

	TenantProviderConfig

	MetricsPushEndpoint          string                  `envconfig:"optional,APP_METRICS_PUSH_ENDPOINT"`
	TenantDependenciesConfigPath string                  `envconfig:"APP_TENANT_REGION_DEPENDENCIES_CONFIG_PATH"`
	RegionToDependenciesConfig   map[string][]Dependency `envconfig:"-"`
}

// TenantProviderConfig includes the configuration for tenant providers - the tenant ID json property names, the subdomain property name, and the tenant provider name.
type TenantProviderConfig struct {
	TenantIDProperty                    string `envconfig:"APP_TENANT_PROVIDER_TENANT_ID_PROPERTY,default=tenantId"`
	SubaccountTenantIDProperty          string `envconfig:"APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY,default=subaccountTenantId"`
	CustomerIDProperty                  string `envconfig:"APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY,default=customerId"`
	SubdomainProperty                   string `envconfig:"APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY,default=subdomain"`
	TenantProvider                      string `envconfig:"APP_TENANT_PROVIDER,default=external-provider"`
	SubscriptionProviderIDProperty      string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_ID_PROPERTY,default=subscriptionProviderIdProperty"`
	ProviderSubaccountIDProperty        string `envconfig:"APP_TENANT_PROVIDER_PROVIDER_SUBACCOUNT_ID_PROPERTY,default=providerSubaccountIdProperty"`
	ConsumerTenantIDProperty            string `envconfig:"APP_TENANT_PROVIDER_CONSUMER_TENANT_ID_PROPERTY,default=consumerTenantIdProperty"`
	SubscriptionProviderAppNameProperty string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_APP_NAME_PROPERTY,default=subscriptionProviderAppNameProperty"`
}

// Dependency contains the xsappname to be used in the dependencies callback
type Dependency struct {
	Xsappname string `json:"xsappname"`
}

type handler struct {
	fetcher    TenantFetcher
	subscriber TenantSubscriber
	config     HandlerConfig
}

// NewTenantsHTTPHandler returns a new HTTP handler, responsible for creation and deletion of regional and non-regional tenants.
func NewTenantsHTTPHandler(subscriber TenantSubscriber, config HandlerConfig) *handler {
	return &handler{
		subscriber: subscriber,
		config:     config,
	}
}

// NewTenantFetcherHTTPHandler returns a new HTTP handler, responsible for creation of on-demand tenants.
func NewTenantFetcherHTTPHandler(fetcher TenantFetcher, config HandlerConfig) *handler {
	return &handler{
		fetcher: fetcher,
		config:  config,
	}
}

// FetchTenantOnDemand fetches External tenants registry events for a provided subaccount and creates a subaccount tenant
func (h *handler) FetchTenantOnDemand(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	vars := mux.Vars(request)
	tenantID, ok := vars[h.config.TenantPathParam]
	if !ok || len(tenantID) == 0 {
		log.C(ctx).Error("Tenant path parameter is missing from request")
		http.Error(writer, "Tenant path parameter is missing from request", http.StatusBadRequest)
		return
	}

	parentTenantID, ok := vars[h.config.ParentTenantPathParam]
	if !ok || len(parentTenantID) == 0 {
		log.C(ctx).Error("Parent tenant path parameter is missing from request")
		http.Error(writer, "Parent tenant ID path parameter is missing from request", http.StatusBadRequest)
		return
	}

	log.C(ctx).Infof("Fetching create event for tenant with ID %s", tenantID)

	if err := h.fetcher.SynchronizeTenant(ctx, parentTenantID, tenantID); err != nil {
		log.C(ctx).WithError(err).Errorf("Error while processing request for creation of tenant %s: %v", tenantID, err)
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}
	writeCreatedResponse(writer, ctx, tenantID)
}

func writeCreatedResponse(writer http.ResponseWriter, ctx context.Context, tenantID string) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(compassURL)); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write response body for request for creation of tenant %s: %v", tenantID, err)
	}
}

// SubscribeTenant handles subscription for tenant. If tenant does not exist, will create it first.
func (h *handler) SubscribeTenant(writer http.ResponseWriter, request *http.Request) {
	h.applySubscriptionChange(writer, request, h.subscriber.Subscribe)
}

// UnSubscribeTenant handles unsubscription for tenant which will remove the tenant id label from the runtime
func (h *handler) UnSubscribeTenant(writer http.ResponseWriter, request *http.Request) {
	h.applySubscriptionChange(writer, request, h.subscriber.Unsubscribe)
}

// Dependencies handler returns all external services where once created in Compass, the tenant should be created as well.
func (h *handler) Dependencies(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	vars := mux.Vars(request)
	region, ok := vars[h.config.RegionPathParam]
	if !ok {
		log.C(ctx).Error("Region path parameter is missing from request")
		http.Error(writer, "Region path parameter is missing from request", http.StatusBadRequest)
		return
	}

	var bytes []byte
	var err error

	if len(h.config.OmitDependenciesCallbackParam) > 0 && len(h.config.OmitDependenciesCallbackParamValue) > 0 {
		queryType, ok := request.URL.Query()[h.config.OmitDependenciesCallbackParam]
		if ok && queryType[0] == h.config.OmitDependenciesCallbackParamValue {
			bytes = []byte("[]")
		}
	}
	if bytes == nil {
		dependencies, ok := h.config.RegionToDependenciesConfig[region]
		if !ok {
			log.C(ctx).Errorf("Invalid region provided: %s", region)
			http.Error(writer, fmt.Sprintf("Invalid region provided: %s", region), http.StatusBadRequest)
			return
		}

		bytes, err = json.Marshal(dependencies)
		if err != nil {
			log.C(ctx).WithError(err).Error("Failed to marshal response body for dependencies request")
			http.Error(writer, InternalServerError, http.StatusInternalServerError)
			return
		}
	}

	writer.Header().Set("Content-Type", "application/json")
	if _, err = writer.Write(bytes); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write response body for dependencies request")
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}
}

func (h *handler) applySubscriptionChange(writer http.ResponseWriter, request *http.Request, subscriptionFunc subscriptionFunc) {
	ctx := request.Context()

	vars := mux.Vars(request)
	region, ok := vars[h.config.RegionPathParam]
	if !ok {
		log.C(ctx).Error("Region path parameter is missing from request")
		http.Error(writer, "Region path parameter is missing from request", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to read tenant information from request body: %v", err)
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}

	subscriptionRequest, err := h.getSubscriptionRequest(body, region)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to extract tenant information from request body: %v", err)
		http.Error(writer, fmt.Sprintf("Failed to extract tenant information from request body: %v", err), http.StatusBadRequest)
		return
	}

	mainTenantID := subscriptionRequest.MainTenantID()
	if err := subscriptionFunc(ctx, subscriptionRequest); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to apply subscription change for tenant %s: %v", mainTenantID, err)
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}

	respondSuccess(ctx, writer, mainTenantID)
}

func (h *handler) getSubscriptionRequest(body []byte, region string) (*TenantSubscriptionRequest, error) {
	properties, err := getProperties(body, map[string]bool{
		h.config.TenantIDProperty:                    true,
		h.config.SubaccountTenantIDProperty:          false,
		h.config.SubdomainProperty:                   true,
		h.config.CustomerIDProperty:                  false,
		h.config.SubscriptionProviderIDProperty:      true,
		h.config.ProviderSubaccountIDProperty:        true,
		h.config.ConsumerTenantIDProperty:            true,
		h.config.SubscriptionProviderAppNameProperty: true,
	})
	if err != nil {
		return nil, err
	}

	req := &TenantSubscriptionRequest{
		AccountTenantID:             properties[h.config.TenantIDProperty],
		SubaccountTenantID:          properties[h.config.SubaccountTenantIDProperty],
		CustomerTenantID:            properties[h.config.CustomerIDProperty],
		Subdomain:                   properties[h.config.SubdomainProperty],
		SubscriptionProviderID:      properties[h.config.SubscriptionProviderIDProperty],
		ProviderSubaccountID:        properties[h.config.ProviderSubaccountIDProperty],
		ConsumerTenantID:            properties[h.config.ConsumerTenantIDProperty],
		SubscriptionProviderAppName: properties[h.config.SubscriptionProviderAppNameProperty],
		Region:                      region,
		SubscriptionPayload:         string(body),
	}

	if req.AccountTenantID == req.SubaccountTenantID {
		req.SubaccountTenantID = ""
	}

	if req.AccountTenantID == req.CustomerTenantID {
		req.CustomerTenantID = ""
	}

	return req, nil
}

func getProperties(body []byte, props map[string]bool) (map[string]string, error) {
	resultProps := map[string]string{}
	for propName, mandatory := range props {
		result := gjson.GetBytes(body, propName).String()
		if mandatory && len(result) == 0 {
			return nil, fmt.Errorf("mandatory property %q is missing from request body", propName)
		}
		resultProps[propName] = result
	}

	return resultProps, nil
}

func respondSuccess(ctx context.Context, writer http.ResponseWriter, mainTenantID string) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(compassURL)); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write response body for tenant request creation for tenant %s: %v", mainTenantID, err)
	}
}
