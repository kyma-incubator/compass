package tenantmapping

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/hydrator/internal/director"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	"github.com/sirupsen/logrus"
)

// DirectorClient missing godoc
//
//go:generate mockery --name=DirectorClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type DirectorClient interface {
	GetTenantByExternalID(ctx context.Context, tenantID string) (*schema.Tenant, error)
	GetSystemAuthByID(ctx context.Context, authID string) (*model.SystemAuth, error)
	UpdateSystemAuth(ctx context.Context, sysAuth *model.SystemAuth) (director.UpdateAuthResult, error)
}

// ScopesGetter missing godoc
//
//go:generate mockery --name=ScopesGetter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ScopesGetter interface {
	GetRequiredScopes(scopesDefinition string) ([]string, error)
}

// ReqDataParser missing godoc
//
//go:generate mockery --name=ReqDataParser --output=automock --outpkg=automock --case=underscore --disable-version-string
type ReqDataParser interface {
	Parse(req *http.Request) (oathkeeper.ReqData, error)
}

// ObjectContextProvider missing godoc
//
//go:generate mockery --name=ObjectContextProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ObjectContextProvider interface {
	GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error)
	Match(ctx context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error)
}

// ClientInstrumenter collects metrics for different client and auth flows.
//
//go:generate mockery --name=ClientInstrumenter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ClientInstrumenter interface {
	InstrumentClient(clientID string, authFlow string, details string)
}

// Handler missing godoc
type Handler struct {
	reqDataParser              ReqDataParser
	objectContextProviders     map[string]ObjectContextProvider
	clientInstrumenter         ClientInstrumenter
	directorClient             DirectorClient
	tenantSubstitutionLabelKey string
}

// NewHandler missing godoc
func NewHandler(
	reqDataParser ReqDataParser,
	objectContextProviders map[string]ObjectContextProvider,
	clientInstrumenter ClientInstrumenter, directorClient DirectorClient, tenantSubstitutionLabelKey string) *Handler {
	return &Handler{
		reqDataParser:              reqDataParser,
		objectContextProviders:     objectContextProviders,
		clientInstrumenter:         clientInstrumenter,
		directorClient:             directorClient,
		tenantSubstitutionLabelKey: tenantSubstitutionLabelKey,
	}
}

// ServeHTTP missing godoc
func (h *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method != http.MethodPost {
		err := fmt.Sprintf("Bad request method. Got %s, expected POST", req.Method)
		log.C(ctx).Error(err)
		http.Error(writer, err, http.StatusBadRequest)
		return
	}

	reqData, err := h.reqDataParser.Parse(req)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while parsing request: %v", err)
		respond(ctx, writer, reqData.Body)
		return
	}

	body := h.processRequest(ctx, reqData)

	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumerId": body.Extra["consumers"],
	})

	newCtx := log.ContextWithLogger(ctx, logger)

	respond(newCtx, writer, body)
}

func (h Handler) processRequest(ctx context.Context, reqData oathkeeper.ReqData) oathkeeper.ReqBody {
	log.C(ctx).Debug("Getting object context")
	objCtxs, err := h.getObjectContexts(ctx, reqData)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while getting object context: %v", err)
		return reqData.Body
	}

	if len(objCtxs) == 0 {
		log.C(ctx).Error("An error occurred while determining the auth details for the request: no object contexts were found")
		return reqData.Body
	}

	objCtxNames := make([]string, 0)
	for i := range objCtxs {
		objCtxNames = append(objCtxNames, objCtxs[i].ContextProvider)
	}
	log.C(ctx).Infof("Matched object contexts: [%s]", strings.Join(objCtxNames, ","))

	tenants, err := h.calculateTenants(ctx, objCtxs)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while calculating tenants: %v", err)
		return reqData.Body
	}

	if err = h.addTenantsToExtra(tenants, reqData); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while adding tenants to extra: %v", err)
		return reqData.Body
	}

	addScopesToExtra(objCtxs, reqData)

	if err := addConsumersToExtra(objCtxs, reqData); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while adding consumers to extra: %v", err)
		reqData.Body.Extra = make(map[string]interface{}) // ensure that no tenant context is propagated from addTenantsToExtra
		return reqData.Body
	}

	return reqData.Body
}

func (h *Handler) getObjectContexts(ctx context.Context, reqData oathkeeper.ReqData) ([]ObjectContext, error) {
	log.C(ctx).Infof("Attempting to get object contexts")

	objectContexts := make([]ObjectContext, 0, len(h.objectContextProviders))
	authDetails := make([]*oathkeeper.AuthDetails, 0, len(h.objectContextProviders))
	for name, provider := range h.objectContextProviders {
		match, details, err := provider.Match(ctx, reqData)
		if err != nil {
			log.C(ctx).Warningf("Provider %s failed to match: %s", name, err.Error())
		}
		if match && err == nil {
			log.C(ctx).Infof("Provider %s attempting to get object context", name)
			authDetails = append(authDetails, details)

			objectContext, err := provider.GetObjectContext(ctx, reqData, *details)
			if err != nil {
				return nil, errors.Wrap(err, "while getting objectContexts: ")
			}

			objectContexts = append(objectContexts, objectContext)
			log.C(ctx).Infof("Provider %s successfully provided object context", name)
		}
	}

	h.instrumentClient(objectContexts, authDetails)

	return objectContexts, nil
}

func respond(ctx context.Context, writer http.ResponseWriter, body oathkeeper.ReqBody) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while encoding data: %v", err)
	}
}

func (h *Handler) instrumentClient(objectContexts []ObjectContext, authDetails []*oathkeeper.AuthDetails) {
	details := oathkeeper.AuthDetails{}

	if len(objectContexts) == 1 {
		details = *authDetails[0]
	} else {
		for _, d := range authDetails {
			if d.CertIssuer != "" {
				details = *d
				break
			}
		}
	}

	flowDetails := details.CertIssuer
	if details.Authenticator != nil {
		flowDetails = details.Authenticator.Name
	}

	h.clientInstrumenter.InstrumentClient(details.AuthID, string(details.AuthFlow), flowDetails)
}

func (h *Handler) calculateTenants(ctx context.Context, objectContexts []ObjectContext) (map[string]string, error) {
	var substituteTenantID string
	tenants := make(map[string]string)
	for _, objCtx := range objectContexts {
		tenants[objCtx.TenantKey] = objCtx.Tenant.InternalID
		tenants[objCtx.ExternalTenantKey] = objCtx.Tenant.ID

		val, ok := objCtx.Tenant.Labels[h.tenantSubstitutionLabelKey].(string)
		if ok {
			log.C(ctx).Infof("Found label %s with value %s of tenant with external ID %s", h.tenantSubstitutionLabelKey, val, objCtx.Tenant.ID)
			substituteTenantID = val
		}
	}

	_, consumerExists := tenants[tenantmapping.ConsumerTenantKey]
	_, externalExists := tenants[tenantmapping.ExternalTenantKey]
	if !consumerExists && !externalExists {
		tenants[tenantmapping.ConsumerTenantKey] = tenants[tenantmapping.ProviderTenantKey]
		tenants[tenantmapping.ExternalTenantKey] = tenants[tenantmapping.ProviderExternalTenantKey]
	}

	if substituteTenantID != "" {
		subtituteTenant, err := h.directorClient.GetTenantByExternalID(ctx, substituteTenantID)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("while fetching root tenant for tenant with external ID: %s", substituteTenantID)
			return nil, errors.Wrapf(err, "while fetching root tenant for tenant with external ID: %s", substituteTenantID)
		}

		log.C(ctx).Infof("The caller tenant has %s label with value %s, substituting the caller tenant %s with customer tenant Root parent with external ID %s and internal ID %s", h.tenantSubstitutionLabelKey, substituteTenantID, tenants[tenantmapping.ExternalTenantKey], subtituteTenant.ID, subtituteTenant.InternalID)
		tenants[tenantmapping.ProviderTenantKey] = tenants[tenantmapping.ConsumerTenantKey]
		tenants[tenantmapping.ProviderExternalTenantKey] = tenants[tenantmapping.ExternalTenantKey]

		tenants[tenantmapping.ConsumerTenantKey] = subtituteTenant.InternalID
		tenants[tenantmapping.ExternalTenantKey] = subtituteTenant.ID
	}
	return tenants, nil
}
func (h *Handler) addTenantsToExtra(tenants map[string]string, reqData oathkeeper.ReqData) error {
	tenantsJSON, err := json.Marshal(tenants)
	if err != nil {
		return errors.Wrap(err, "while marshaling tenants")
	}

	tenantsStr := string(tenantsJSON)

	escaped := strings.ReplaceAll(tenantsStr, `"`, `\"`)
	reqData.Body.Extra["tenant"] = escaped

	return nil
}

func addScopesToExtra(objectContexts []ObjectContext, reqData oathkeeper.ReqData) {
	objScopes := make([]string, 0)
	intersectableScopes := make([][]string, 0)

	for _, objCtx := range objectContexts {
		currentScopes := strings.Split(objCtx.Scopes, " ")
		switch objCtx.ScopesMergeStrategy {
		case overrideAllScopes:
			reqData.Body.Extra["scope"] = strings.Join(currentScopes, " ")
			return
		case mergeWithOtherScopes:
			objScopes = append(objScopes, currentScopes...)
		default: // Intersect
			intersectableScopes = append(intersectableScopes, currentScopes)
		}
	}

	objScopes = removeDuplicateValues(objScopes)

	for _, currentScopes := range intersectableScopes {
		if len(objScopes) > 0 {
			objScopes = intersect(objScopes, currentScopes)
		} else {
			objScopes = currentScopes
		}
	}

	joined := strings.Join(objScopes, " ")
	reqData.Body.Extra["scope"] = joined
}

func removeDuplicateValues(scopes []string) []string {
	keys := make(map[string]struct{})
	result := make([]string, 0, len(scopes))

	for _, entry := range scopes {
		if _, exists := keys[entry]; !exists {
			keys[entry] = struct{}{}
			result = append(result, entry)
		}
	}
	return result
}

func addConsumersToExtra(objectContexts []ObjectContext, reqData oathkeeper.ReqData) error {
	region := deriveRegionFromObjectContexts(objectContexts)

	c := consumer.Consumer{}
	if len(objectContexts) == 1 {
		c.ConsumerID = objectContexts[0].ConsumerID
		c.Type = objectContexts[0].ConsumerType
		c.Flow = objectContexts[0].AuthFlow
	} else {
		c = getCertServiceObjectContextProviderConsumer(objectContexts)
		c.OnBehalfOf = getOnBehalfConsumer(objectContexts)

		if c.OnBehalfOf != "" { // i.e. make sure that regions match only during consumer-provider flow
			for _, objCtx := range objectContexts {
				if objCtx.Tenant.InternalID != "" && objCtx.Region != region {
					return errors.Errorf("mismatched region for consumer ID REDACTED_%x: actual %s, expected: %s)", sha256.Sum256([]byte(objCtx.ConsumerID)), objCtx.Region, region)
				}
			}
		}
	}

	reqData.Body.Extra["consumerID"] = c.ConsumerID
	reqData.Body.Extra["consumerType"] = c.Type
	reqData.Body.Extra["flow"] = c.Flow
	reqData.Body.Extra["onBehalfOf"] = c.OnBehalfOf
	reqData.Body.Extra["region"] = region
	reqData.Body.Extra["tokenClientID"] = getClientIDFromConsumerToken(objectContexts)

	return nil
}

// deriveRegionFromObjectContexts makes sure to find the region from an existing tenant which is matched by some previous obj ctx.
// This is ensured by checking the objCtx.TenantID field, because it will be populated by the corresponding obj ctx provider once it is matched.
// This is necessary due to the fact that some obj ctx providers might result with a non-existing tenant for which TenantID will be empty (conversely the region for them would also be empty).
func deriveRegionFromObjectContexts(objectContext []ObjectContext) string {
	for _, objCtx := range objectContext {
		if objCtx.Region != "" {
			return objCtx.Region
		}
	}
	return ""
}

func getCertServiceObjectContextProviderConsumer(objectContexts []ObjectContext) consumer.Consumer {
	c := consumer.Consumer{}
	for _, objCtx := range objectContexts {
		if objCtx.ContextProvider == tenantmapping.CertServiceObjectContextProvider {
			c.ConsumerID = objCtx.ConsumerID
			c.Type = objCtx.ConsumerType
			c.Flow = objCtx.AuthFlow
		}
	}
	return c
}

func getOnBehalfConsumer(objectContexts []ObjectContext) string {
	for _, objCtx := range objectContexts {
		if objCtx.ContextProvider != tenantmapping.CertServiceObjectContextProvider {
			return objCtx.ConsumerID
		}
	}
	return ""
}

func getClientIDFromConsumerToken(objectContexts []ObjectContext) string {
	for _, objCtx := range objectContexts {
		if objCtx.ContextProvider == tenantmapping.AuthenticatorObjectContextProvider || objCtx.ContextProvider == tenantmapping.ConsumerProviderObjectContextProvider {
			return objCtx.OauthClientID
		}
	}
	return ""
}

func intersect(s1 []string, s2 []string) []string {
	h := make(map[string]bool, len(s1))
	for _, s := range s1 {
		h[s] = true
	}

	var intersection []string
	for _, s := range s2 {
		if h[s] {
			intersection = append(intersection, s)
		}
	}
	return intersection
}
