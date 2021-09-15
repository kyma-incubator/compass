package tenantfetchersvc

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	labelutils "github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	compassURL          = "https://github.com/kyma-incubator/compass"
	InternalServerError = "Internal Server Error"
)

// TenantProvisioner is used to create all related to the incoming request tenants, and build their hierarchy;
//go:generate mockery --name=TenantProvisioner --output=automock --outpkg=automock --case=underscore
type TenantProvisioner interface {
	ProvisionTenants(context.Context, TenantSubscriptionRequest) error
	ProvisionRegionalTenants(context.Context, TenantSubscriptionRequest) error
}

// RuntimeService is used to interact with runtimes
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore
type RuntimeService interface {
	SetLabel(context.Context, *model.LabelInput) error
	GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error)
	ListByFiltersGlobal(context.Context, []*labelfilter.LabelFilter) ([]*model.Runtime, error)
}

// HandlerConfig is the configuration required by the tenant handler.
// It includes configurable parameters for incoming requests, including different tenant IDs json properties, and path parameters.
type HandlerConfig struct {
	HandlerEndpoint               string `envconfig:"APP_HANDLER_ENDPOINT,default=/v1/callback/{tenantId}"`
	RegionalHandlerEndpoint       string `envconfig:"APP_REGIONAL_HANDLER_ENDPOINT,default=/v1/regional/{region}/callback/{tenantId}"`
	DependenciesEndpoint          string `envconfig:"APP_DEPENDENCIES_ENDPOINT,default=/v1/dependencies"`
	TenantPathParam               string `envconfig:"APP_TENANT_PATH_PARAM,default=tenantId"`
	RegionPathParam               string `envconfig:"APP_REGION_PATH_PARAM,default=region"`
	RegionLabelKey                string `envconfig:"APP_REGION_LABEL_KEY,default=region_key"`
	SubscriptionConsumerLabelKey  string `envconfig:"APP_SUBSCRIPTION_CONSUMER_LABEL_KEY,default=subscription_consumer_id"`
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
	SubscriptionConsumerIDProperty string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_CONSUMER_ID_PROPERTY,default=subscriptionConsumerId"`
}

type handler struct {
	provisioner    TenantProvisioner
	runtimeService RuntimeService
	transact       persistence.Transactioner
	config         HandlerConfig
}

// NewTenantsHTTPHandler returns a new HTTP handler, responsible for creation and deletion of regional and non-regional tenants.
func NewTenantsHTTPHandler(provisioner TenantProvisioner, runtimeService RuntimeService, transact persistence.Transactioner, config HandlerConfig) *handler {
	return &handler{
		provisioner:    provisioner,
		runtimeService: runtimeService,
		transact:       transact,
		config:         config,
	}
}

// Create handles creation of non-regional tenants.
func (h *handler) Create(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to read tenant information from request body")
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}

	subscriptionRequest, err := h.getSubscriptionRequest(body, "")
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to extract tenant information from request body")
		http.Error(writer, fmt.Sprintf("Failed to extract tenant information from request body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	mainTenantID := subscriptionRequest.MainTenantID()

	tx, err := h.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while opening db transaction")
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err = h.provisionTenants(ctx, subscriptionRequest, ""); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to provision tenant with ID %s", mainTenantID)
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while committing db transaction")
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}

	respondSuccess(ctx, writer, mainTenantID)
}

// SubscribeTenant handles subscription for tenant. If tenant does not exist, will create it first.
func (h *handler) SubscribeTenant(writer http.ResponseWriter, request *http.Request) {
	h.applySubscriptionChange(writer, request, true)
}

// SubscribeTenant handles subscription for tenant. If tenant does not exist, will create it first.
func (h *handler) UnSubscribeTenant(writer http.ResponseWriter, request *http.Request) {
	h.applySubscriptionChange(writer, request, false)
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

func (h *handler) applySubscriptionChange(writer http.ResponseWriter, request *http.Request, isSubscriptionFlow bool) {
	ctx := request.Context()

	vars := mux.Vars(request)
	region, ok := vars[h.config.RegionPathParam]
	if !ok {
		log.C(ctx).Error("Region path parameter is missing from request")
		http.Error(writer, "Region path parameter is missing from request", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to read tenant information from request body")
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}

	subscriptionRequest, err := h.getSubscriptionRequest(body, region)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to extract tenant information from request body")
		http.Error(writer, fmt.Sprintf("Failed to extract tenant information from request body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	mainTenantID := subscriptionRequest.MainTenantID()

	tx, err := h.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while opening db transaction")
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if isSubscriptionFlow {
		if err = h.provisionTenants(ctx, subscriptionRequest, region); err != nil {
			log.C(ctx).WithError(err).Errorf("Failed to provision tenant with ID %s", mainTenantID)
			http.Error(writer, InternalServerError, http.StatusInternalServerError)
			return
		}
	}

	if err = h.applyRuntimesSubscriptionChange(ctx, subscriptionRequest.SubscriptionConsumerID, subscriptionRequest.SubaccountTenantID, region, isSubscriptionFlow); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while subscribing tenant with id %q for runtimes with labels: %q and %q", mainTenantID, subscriptionRequest.SubscriptionConsumerID, region)
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while committing db transaction")
		http.Error(writer, InternalServerError, http.StatusInternalServerError)
		return
	}

	respondSuccess(ctx, writer, mainTenantID)
}

func (h *handler) applyRuntimesSubscriptionChange(ctx context.Context, subscriptionConsumerID, subaccountTenantID, region string, isSubscriptionFlow bool) error {
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(h.config.SubscriptionConsumerLabelKey, fmt.Sprintf("\"%s\"", subscriptionConsumerID)),
		labelfilter.NewForKeyWithQuery(h.config.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}

	runtimes, err := h.runtimeService.ListByFiltersGlobal(ctx, filters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil
		}

		return errors.Wrap(err, fmt.Sprintf("Failed to get runtimes for labels %s: %s and %s: %s", h.config.RegionLabelKey, region, h.config.SubscriptionConsumerLabelKey, subscriptionConsumerID))
	}

	for _, runtime := range runtimes {
		ctx = tenant.SaveToContext(ctx, runtime.Tenant, "")

		labelOldValue := make([]string, 0)
		label, err := h.runtimeService.GetLabel(ctx, runtime.ID, h.config.ConsumerSubaccountIDsLabelKey)
		if err != nil {
			if !apperrors.IsNotFoundError(err) {
				return errors.Wrap(err, fmt.Sprintf("Failed to get label for runtime with id: %s and key: %s", runtime.ID, h.config.ConsumerSubaccountIDsLabelKey))
			}
			// if the error is not found, do nothing and continue
		} else {
			if labelOldValue, err = labelutils.ValueToStringsSlice(label.Value); err != nil {
				return errors.Wrap(err, fmt.Sprintf("Failed to parse label values for label with id: %s", label.ID))
			}
		}

		labelNewValue := make([]string, 0)
		if isSubscriptionFlow {
			labelNewValue = append(labelOldValue, subaccountTenantID)
		} else {
			labelNewValue = removeElement(labelOldValue, subaccountTenantID)
		}

		if err := h.runtimeService.SetLabel(ctx, &model.LabelInput{
			Key:        h.config.ConsumerSubaccountIDsLabelKey,
			Value:      labelNewValue,
			ObjectType: model.RuntimeLabelableObject,
			ObjectID:   runtime.ID,
		}); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to set label for runtime with id: %s", runtime.ID))
		}
	}

	return nil
}

func (h *handler) getSubscriptionRequest(body []byte, region string) (*TenantSubscriptionRequest, error) {
	properties, err := getProperties(body, map[string]bool{
		h.config.TenantIDProperty:               true,
		h.config.SubaccountTenantIDProperty:     false,
		h.config.SubdomainProperty:              true,
		h.config.CustomerIDProperty:             false,
		h.config.SubscriptionConsumerIDProperty: true,
	})
	if err != nil {
		return nil, err
	}

	return &TenantSubscriptionRequest{
		AccountTenantID:        properties[h.config.TenantIDProperty],
		SubaccountTenantID:     properties[h.config.SubaccountTenantIDProperty],
		CustomerTenantID:       properties[h.config.CustomerIDProperty],
		Subdomain:              properties[h.config.SubdomainProperty],
		SubscriptionConsumerID: properties[h.config.SubscriptionConsumerIDProperty],
		Region:                 region,
	}, nil
}

func (h *handler) provisionTenants(ctx context.Context, request *TenantSubscriptionRequest, region string) error {
	var err error

	if len(region) > 0 {
		err = h.provisioner.ProvisionRegionalTenants(ctx, *request)
	} else {
		err = h.provisioner.ProvisionTenants(ctx, *request)
	}
	if err != nil && !apperrors.IsNotUniqueError(err) {
		return err
	}

	return nil
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

func removeElement(slice []string, elem string) []string {
	result := make([]string, 0)
	for _, e := range slice {
		if e != elem {
			result = append(result, e)
		}
	}
	return result
}
