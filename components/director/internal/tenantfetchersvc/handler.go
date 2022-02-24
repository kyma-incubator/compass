package tenantfetchersvc

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/tidwall/gjson"
)

const (
	// InternalServerError message
	InternalServerError = "Internal Server Error"
	compassURL          = "https://github.com/kyma-incubator/compass"
)

// TenantSubscriber is used to apply subscription changes for tenants;
//go:generate mockery --name=TenantSubscriber --output=automock --outpkg=automock --case=underscore
type TenantSubscriber interface {
	Subscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest, region string) error
	Unsubscribe(ctx context.Context, tenantSubscriptionRequest *TenantSubscriptionRequest, region string) error
}

// HandlerConfig is the configuration required by the tenant handler.
// It includes configurable parameters for incoming requests, including different tenant IDs json properties, and path parameters.
type HandlerConfig struct {
	HandlerEndpoint               string `envconfig:"APP_HANDLER_ENDPOINT,default=/v1/callback/{tenantId}"`
	RegionalHandlerEndpoint       string `envconfig:"APP_REGIONAL_HANDLER_ENDPOINT,default=/v1/regional/{region}/callback/{tenantId}"`
	DependenciesEndpoint          string `envconfig:"APP_DEPENDENCIES_ENDPOINT,default=/v1/dependencies"`
	TenantPathParam               string `envconfig:"APP_TENANT_PATH_PARAM,default=tenantId"`
	RegionPathParam               string `envconfig:"APP_REGION_PATH_PARAM,default=region"`
	SubscriptionProviderLabelKey  string `envconfig:"APP_SUBSCRIPTION_PROVIDER_LABEL_KEY,default=subscriptionProviderId"`
	ConsumerSubaccountIDsLabelKey string `envconfig:"APP_CONSUMER_SUBACCOUNT_IDS_LABEL_KEY,default=consumer_subaccount_ids"`
	TenantProviderConfig
	features.Config
}

// TenantProviderConfig includes the configuration for tenant providers - the tenant ID json property names, the subdomain property name, and the tenant provider name.
type TenantProviderConfig struct {
	TenantIDProperty               string `envconfig:"APP_TENANT_PROVIDER_TENANT_ID_PROPERTY,default=tenantId"`
	SubaccountTenantIDProperty     string `envconfig:"APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY,default=subaccountTenantId"`
	CustomerIDProperty             string `envconfig:"APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY,default=customerId"`
	SubdomainProperty              string `envconfig:"APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY,default=subdomain"`
	TenantProvider                 string `envconfig:"APP_TENANT_PROVIDER,default=external-provider"`
	SubscriptionProviderIDProperty string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_ID_PROPERTY,default=subscriptionProviderId"`
}

type handler struct {
	subscriber TenantSubscriber
	transact   persistence.Transactioner
	config     HandlerConfig
}

// NewTenantsHTTPHandler returns a new HTTP handler, responsible for creation and deletion of regional and non-regional tenants.
func NewTenantsHTTPHandler(subscriber TenantSubscriber, transact persistence.Transactioner, config HandlerConfig) *handler {
	return &handler{
		subscriber: subscriber,
		transact:   transact,
		config:     config,
	}
}

// Create handles creation of non-regional tenants.
func (h *handler) Create(writer http.ResponseWriter, request *http.Request) {
	h.applySubscriptionChange(writer, request, h.subscriber.Subscribe, false)
}

// SubscribeTenant handles subscription for tenant. If tenant does not exist, will create it first.
func (h *handler) SubscribeTenant(writer http.ResponseWriter, request *http.Request) {
	h.applySubscriptionChange(writer, request, h.subscriber.Subscribe, true)
}

// UnSubscribeTenant handles unsubscription for tenant which will remove the tenant id label from the runtime
func (h *handler) UnSubscribeTenant(writer http.ResponseWriter, request *http.Request) {
	h.applySubscriptionChange(writer, request, h.subscriber.Unsubscribe, true)
}

// DeleteByExternalID handles both regional and non-regional tenant deletion requests.
func (h *handler) DeleteByExternalID(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to read tenant information from delete request body: %v", err)
		writer.WriteHeader(http.StatusOK)
		return
	}

	if tenantID := gjson.GetBytes(body, h.config.TenantIDProperty).String(); len(tenantID) > 0 {
		log.C(ctx).Infof("Received delete request for tenant with external tenant ID %s, returning 200 OK", tenantID)
	} else {
		log.C(ctx).Infof("External tenant ID property %q is missing from delete request body", h.config.TenantIDProperty)
	}

	writer.WriteHeader(http.StatusOK)
}

// Dependencies handler returns all external services where once created in Compass, the tenant should be created as well.
func (h *handler) Dependencies(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	if _, err := writer.Write([]byte("{}")); err != nil {
		log.C(request.Context()).WithError(err).Errorf("Failed to write response body for dependencies request")
		return
	}
}

func (h *handler) applySubscriptionChange(writer http.ResponseWriter, request *http.Request, subscriptionFunc subscriptionFunc, shouldExtractRegion bool) {
	var (
		ok     bool
		region string
		ctx    = request.Context()
	)

	if shouldExtractRegion {
		vars := mux.Vars(request)
		region, ok = vars[h.config.RegionPathParam]
		if !ok {
			log.C(ctx).Error("Region path parameter is missing from request")
			http.Error(writer, "Region path parameter is missing from request", http.StatusBadRequest)
			return
		}
	}

	body, err := ioutil.ReadAll(request.Body)
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

	tx, err := h.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening db transaction: %v", err)
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	mainTenantID := subscriptionRequest.MainTenantID()
	if err := subscriptionFunc(ctx, subscriptionRequest, region); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to apply subscription change for tenant %s: %v", mainTenantID, err)
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while committing db transaction: %v", err)
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}

	respondSuccess(ctx, writer, mainTenantID)
}

func (h *handler) getSubscriptionRequest(body []byte, region string) (*TenantSubscriptionRequest, error) {
	properties, err := getProperties(body, map[string]bool{
		h.config.TenantIDProperty:               true,
		h.config.SubaccountTenantIDProperty:     false,
		h.config.SubdomainProperty:              true,
		h.config.CustomerIDProperty:             false,
		h.config.SubscriptionProviderIDProperty: true,
	})
	if err != nil {
		return nil, err
	}

	req := &TenantSubscriptionRequest{
		AccountTenantID:        properties[h.config.TenantIDProperty],
		SubaccountTenantID:     properties[h.config.SubaccountTenantIDProperty],
		CustomerTenantID:       properties[h.config.CustomerIDProperty],
		Subdomain:              properties[h.config.SubdomainProperty],
		SubscriptionProviderID: properties[h.config.SubscriptionProviderIDProperty],
		Region:                 region,
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
