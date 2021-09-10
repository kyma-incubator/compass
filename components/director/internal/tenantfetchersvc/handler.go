package tenantfetchersvc

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

const (
	compassURL                  = "https://github.com/kyma-incubator/compass"
	tenantCreationFailureMsgFmt = "Failed to create tenant with ID %s"
)

// TenantProvisioner is used to create all related to the incoming request tenants, and build their hierarchy;
//go:generate mockery --name=TenantProvisioner --output=automock --outpkg=automock --case=underscore
type TenantProvisioner interface {
	ProvisionTenants(context.Context, TenantProvisioningRequest) error
	ProvisionRegionalTenants(context.Context, TenantProvisioningRequest) error
}

// HandlerConfig is the configuration required by the tenant handler.
// It includes configurable parameters for incoming requests, including different tenant IDs json properties, and path parameters.
type HandlerConfig struct {
	HandlerEndpoint         string `envconfig:"APP_HANDLER_ENDPOINT,default=/v1/callback/{tenantId}"`
	RegionalHandlerEndpoint string `envconfig:"APP_REGIONAL_HANDLER_ENDPOINT,default=/v1/regional/{region}/callback/{tenantId}"`
	DependenciesEndpoint    string `envconfig:"APP_DEPENDENCIES_ENDPOINT,default=/v1/dependencies"`
	TenantPathParam         string `envconfig:"APP_TENANT_PATH_PARAM,default=tenantId"`
	RegionPathParam         string `envconfig:"APP_REGION_PATH_PARAM,default=region"`
	TenantProviderConfig
}

// TenantProviderConfig includes the configuration for tenant providers - the tenant ID json property names, the subdomain property name, and the tenant provider name.
type TenantProviderConfig struct {
	TenantIDProperty           string `envconfig:"APP_TENANT_PROVIDER_TENANT_ID_PROPERTY,default=tenantId"`
	SubaccountTenantIDProperty string `envconfig:"APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY,default=subaccountTenantId"`
	CustomerIDProperty         string `envconfig:"APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY,default=customerId"`
	SubdomainProperty          string `envconfig:"APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY,default=subdomain"`
	TenantProvider             string `envconfig:"APP_TENANT_PROVIDER,default=external-provider"`
}

type handler struct {
	provisioner TenantProvisioner
	transact    persistence.Transactioner
	config      HandlerConfig
}

// NewTenantsHTTPHandler returns a new HTTP handler, responsible for creation and deletion of regional and non-regional tenants.
func NewTenantsHTTPHandler(provisioner TenantProvisioner, transact persistence.Transactioner, config HandlerConfig) *handler {
	return &handler{
		provisioner: provisioner,
		transact:    transact,
		config:      config,
	}
}

// Create handles creation of non-regional tenants.
func (h *handler) Create(writer http.ResponseWriter, request *http.Request) {
	h.handleTenantCreationRequest(writer, request, "")
}

// CreateRegional handles creation of regional tenants.
func (h *handler) CreateRegional(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	vars := mux.Vars(request)
	region, ok := vars[h.config.RegionPathParam]
	if !ok {
		log.C(ctx).Errorf("Region path parameter is missing from request")
		http.Error(writer, "Region path parameter is missing from request", http.StatusBadRequest)
		return
	}

	h.handleTenantCreationRequest(writer, request, region)
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

func (h *handler) handleTenantCreationRequest(writer http.ResponseWriter, request *http.Request, region string) {
	ctx := request.Context()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to read tenant information from request body: %v", err)
		http.Error(writer, "Failed to read tenant information from request body", http.StatusInternalServerError)
		return
	}
	provisioningReq, err := h.getProvisioningRequest(body, region)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to extract tenant information from request body: %v", err)
		http.Error(writer, fmt.Sprintf("Failed to extract tenant information from request body: %s", err.Error()), http.StatusBadRequest)
		return
	}
	mainTenantID := provisioningReq.MainTenantID()
	if err := h.provisionTenants(ctx, provisioningReq, region); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to provision tenant with ID %s: %v", mainTenantID, err)
		http.Error(writer, fmt.Sprintf(tenantCreationFailureMsgFmt, mainTenantID), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(compassURL)); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write response body for tenant request creation for tenant %s: %v", mainTenantID, err)
	}
}

func (h *handler) getProvisioningRequest(body []byte, region string) (*TenantProvisioningRequest, error) {
	properties, err := getProperties(body, map[string]bool{
		h.config.TenantIDProperty:           true,
		h.config.SubaccountTenantIDProperty: false,
		h.config.SubdomainProperty:          true,
		h.config.CustomerIDProperty:         false,
	})
	if err != nil {
		return nil, err
	}

	return &TenantProvisioningRequest{
		AccountTenantID:    properties[h.config.TenantIDProperty],
		SubaccountTenantID: properties[h.config.SubaccountTenantIDProperty],
		CustomerTenantID:   properties[h.config.CustomerIDProperty],
		Subdomain:          properties[h.config.SubdomainProperty],
		Region:             region,
	}, nil
}

func (h *handler) provisionTenants(ctx context.Context, request *TenantProvisioningRequest, region string) error {
	tx, err := h.transact.Begin()
	if err != nil {
		return errors.Wrapf(err, "while starting DB transaction")
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if len(region) > 0 {
		err = h.provisioner.ProvisionRegionalTenants(ctx, *request)
	} else {
		err = h.provisioner.ProvisionTenants(ctx, *request)
	}
	if err != nil && !apperrors.IsNotUniqueError(err) {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "failed to commit transaction while storing tenant")
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
