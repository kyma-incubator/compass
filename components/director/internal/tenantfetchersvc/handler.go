package tenantfetchersvc

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

const (
	compassURL                  = "https://github.com/kyma-incubator/compass"
	tenantCreationFailureMsgFmt = "Failed to create tenant with ID %s"

	autogeneratedTenantProvider = "autogenerated"
)

//go:generate mockery --name=TenantProvisioner --output=automock --outpkg=automock --case=underscore
type TenantProvisioner interface {
	ProvisionTenant(ctx context.Context, tenant model.BusinessTenantMappingInput) error
	ProvisionRegionalTenant(ctx context.Context, tenant model.BusinessTenantMappingInput, region string) error
}

type HandlerConfig struct {
	HandlerEndpoint         string `envconfig:"APP_HANDLER_ENDPOINT,default=/v1/callback/{tenantId}"`
	RegionalHandlerEndpoint string `envconfig:"APP_REGIONAL_HANDLER_ENDPOINT,default=/v1/regional/{region}/callback/{tenantId}"`
	TenantPathParam         string `envconfig:"APP_TENANT_PATH_PARAM,default=tenantId"`
	RegionPathParam         string `envconfig:"APP_REGION_PATH_PARAM,default=region"`

	TenantProviderConfig

	JWKSSyncPeriod            time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone       bool          `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=true"`
	JwksEndpoint              string        `envconfig:"APP_JWKS_ENDPOINT"`
	SubscriptionCallbackScope string        `envconfig:"APP_SUBSCRIPTION_CALLBACK_SCOPE"`
}

type TenantProviderConfig struct {
	TenantIdProperty           string `envconfig:"APP_TENANT_PROVIDER_TENANT_ID_PROPERTY"`
	SubaccountTenantIdProperty string `envconfig:"APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY"`
	CustomerIdProperty         string `envconfig:"APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY"`
	SubdomainProperty          string `envconfig:"APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY"`
	TenantProvider             string `envconfig:"APP_TENANT_PROVIDER"`
}

type handler struct {
	provisioner TenantProvisioner
	transact    persistence.Transactioner
	config      HandlerConfig
}

func NewTenantsHTTPHandler(provisioner TenantProvisioner, transact persistence.Transactioner, config HandlerConfig) *handler {
	return &handler{
		provisioner: provisioner,
		transact:    transact,
		config:      config,
	}
}

func (h *handler) Create(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	body, err := readBody(request)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to read tenant information from request body: %v", err)
		http.Error(writer, "Failed to read tenant information from request body", http.StatusInternalServerError)
		return
	}
	accountTenant, err := h.tenantFromBody(body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to extract tenant information from request body: %v", err)
		http.Error(writer, fmt.Sprintf("Failed to extract tenant information from request body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if err := h.provisionTenant(ctx, *accountTenant, ""); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to provision tenant with ID %s: %v", accountTenant.ExternalTenant, err)
		http.Error(writer, fmt.Sprintf(tenantCreationFailureMsgFmt, accountTenant.ExternalTenant), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(compassURL)); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write response body for tenant request creation for tenant %s: %v", accountTenant.ExternalTenant, err)
	}
}

func (h *handler) DeleteByExternalID(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	body, err := readBody(req)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to read tenant information from delete request body: %v", err)
		writer.WriteHeader(http.StatusOK)
		return
	}

	if tenantID := gjson.GetBytes(body, h.config.TenantIdProperty).String(); len(tenantID) > 0 {
		log.C(ctx).Infof("Received delete request for tenant with external tenant ID %s, returning 200 OK", tenantID)
	} else {
		log.C(ctx).Infof("External tenant ID property %q is missing from delete request body", h.config.TenantIdProperty)
	}

	writer.WriteHeader(http.StatusOK)
}

func (h *handler) CreateRegional(writer http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	vars := mux.Vars(request)
	region, ok := vars[h.config.RegionPathParam]
	if !ok {
		log.C(ctx).Errorf("Region path parameter is missing from request")
		http.Error(writer, "Region path parameter is missing from request", http.StatusBadRequest)
		return
	}

	body, err := readBody(request)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to read tenant information from request body: %v", err)
		http.Error(writer, "Failed to read tenant information from request body", http.StatusInternalServerError)
		return
	}
	regionalTenant, err := h.tenantFromBody(body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to extract tenant information from request body: %v", err)
		http.Error(writer, fmt.Sprintf("Failed to extract tenant information from request body: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if err := h.provisionTenant(ctx, *regionalTenant, region); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to provision tenant with ID %s: %v", regionalTenant.ExternalTenant, err)
		http.Error(writer, fmt.Sprintf(tenantCreationFailureMsgFmt, regionalTenant.ExternalTenant), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write([]byte(compassURL)); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to write response body for tenant request creation for tenant %s: %v", regionalTenant.ExternalTenant, err)
	}
}

func (h *handler) tenantFromBody(body []byte) (*model.BusinessTenantMappingInput, error) {
	properties, err := getProperties(body, map[string]bool{
		h.config.TenantIdProperty:           true,
		h.config.SubdomainProperty:          true,
		h.config.SubaccountTenantIdProperty: false,
		h.config.CustomerIdProperty:         false,
	})
	if err != nil {
		return nil, err
	}

	if prop, ok := properties[h.config.SubaccountTenantIdProperty]; ok && len(prop) > 0 {
		return &model.BusinessTenantMappingInput{
			Name:           properties[h.config.SubaccountTenantIdProperty],
			ExternalTenant: properties[h.config.SubaccountTenantIdProperty],
			Parent:         properties[h.config.TenantIdProperty],
			Subdomain:      properties[h.config.SubdomainProperty],
			Type:           tenantEntity.TypeToStr(tenantEntity.Subaccount),
			Provider:       h.config.TenantProvider,
		}, nil
	}

	return &model.BusinessTenantMappingInput{
		Name:           properties[h.config.TenantIdProperty],
		ExternalTenant: properties[h.config.TenantIdProperty],
		Parent:         properties[h.config.CustomerIdProperty],
		Subdomain:      properties[h.config.SubdomainProperty],
		Type:           tenantEntity.TypeToStr(tenantEntity.Account),
		Provider:       h.config.TenantProvider,
	}, nil
}

func (h *handler) provisionTenant(ctx context.Context, tenant model.BusinessTenantMappingInput, region string) error {
	externalTenantID := tenant.ExternalTenant
	tx, err := h.transact.Begin()
	if err != nil {
		return errors.Wrapf(err, "while starting DB transaction")
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if len(region) > 0 {
		err = h.provisioner.ProvisionRegionalTenant(ctx, tenant, region)
	} else {
		err = h.provisioner.ProvisionTenant(ctx, tenant)
	}
	if err != nil && !apperrors.IsNotUniqueError(err) {
		return errors.Wrapf(err, "while provisioning tenant with external ID %s", externalTenantID)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrapf(err, "failed to commit transaction while storing tenant with external ID %s", externalTenantID)
	}

	return nil
}

func getProperties(body []byte, props map[string]bool) (map[string]string, error) {
	resultProps := map[string]string{}
	for propName, mandatory := range props {
		result := gjson.GetBytes(body, propName).String()
		if mandatory && len(result) <= 0 {
			return nil, fmt.Errorf("mandatory property %q is missing from request body", propName)
		}
		resultProps[propName] = result
	}

	return resultProps, nil
}

func readBody(r *http.Request) ([]byte, error) {
	ctx := r.Context()

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.C(ctx).WithError(err).Errorf("Unable to close request body: %v", err)
		}
	}()

	return buf, nil
}
