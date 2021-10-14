package tenantmapping

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/sirupsen/logrus"
)

const (
	// UserObjectContextProvider missing godoc
	UserObjectContextProvider = "UserObjectContextProvider"
	// SystemAuthObjectContextProvider missing godoc
	SystemAuthObjectContextProvider = "SystemAuthObjectContextProvider"
	// AuthenticatorObjectContextProvider missing godoc
	AuthenticatorObjectContextProvider = "AuthenticatorObjectContextProvider"
	// CertServiceObjectContextProvider missing godoc
	CertServiceObjectContextProvider = "CertServiceObjectContextProvider"
	// ConsumerTenantKey key for consumer tenant id in Claims.Tenant
	ConsumerTenantKey = "consumerTenant"
	// ExternalTenantKey key for external tenant id in Claims.Tenant
	ExternalTenantKey = "externalTenant"
	// ProviderTenantKey key for provider tenant id in Claims.Tenant
	ProviderTenantKey = "providerTenant"
	// ProviderExternalTenantKey key for external provider tenant id in Claims.Tenant
	ProviderExternalTenantKey = "providerExternalTenant"
)

// ScopesGetter missing godoc
//go:generate mockery --name=ScopesGetter --output=automock --outpkg=automock --case=underscore
type ScopesGetter interface {
	GetRequiredScopes(scopesDefinition string) ([]string, error)
}

// ReqDataParser missing godoc
//go:generate mockery --name=ReqDataParser --output=automock --outpkg=automock --case=underscore
type ReqDataParser interface {
	Parse(req *http.Request) (oathkeeper.ReqData, error)
}

// ObjectContextProvider missing godoc
//go:generate mockery --name=ObjectContextProvider --output=automock --outpkg=automock --case=underscore
type ObjectContextProvider interface {
	GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error)
	Match(ctx context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error)
}

// TenantRepository missing godoc
//go:generate mockery --name=TenantRepository --output=automock --outpkg=automock --case=underscore
type TenantRepository interface {
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
}

// ClientInstrumenter collects metrics for different client and auth flows.
//go:generate mockery --name=ClientInstrumenter --output=automock --outpkg=automock --case=underscore
type ClientInstrumenter interface {
	InstrumentClient(clientID string, authFlow string, details string)
}

// Handler missing godoc
type Handler struct {
	reqDataParser          ReqDataParser
	transact               persistence.Transactioner
	objectContextProviders map[string]ObjectContextProvider
	clientInstrumenter     ClientInstrumenter
}

// NewHandler missing godoc
func NewHandler(
	reqDataParser ReqDataParser,
	transact persistence.Transactioner,
	objectContextProviders map[string]ObjectContextProvider,
	clientInstrumenter ClientInstrumenter) *Handler {
	return &Handler{
		reqDataParser:          reqDataParser,
		transact:               transact,
		objectContextProviders: objectContextProviders,
		clientInstrumenter:     clientInstrumenter,
	}
}

// ServeHTTP missing godoc
func (h *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method != http.MethodPost {
		err := fmt.Sprintf("Bad request method. Got %s, expected POST", req.Method)
		log.C(ctx).Errorf(err)
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
	tx, err := h.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening db transaction: %v", err)
		return reqData.Body
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	newCtx := persistence.SaveToContext(ctx, tx)

	log.C(ctx).Debug("Getting object context")
	objCtxs, err := h.getObjectContexts(newCtx, reqData)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while getting object context: %v", err)
		return reqData.Body
	}

	if len(objCtxs) == 0 {
		log.C(ctx).Error("An error occurred while determining the auth details for the request: no object contexts were found")
		return reqData.Body
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while committing transaction: %v", err)
		return reqData.Body
	}

	if err := addTenantsToExtra(objCtxs, reqData); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while adding tenants to extra: %v", err)
		return reqData.Body
	}

	addScopesToExtra(objCtxs, reqData)

	addConsumersToExtra(objCtxs, reqData)

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

func (h *Handler) instrumentClient(objectContexts []ObjectContext, authDetails []*oathkeeper.AuthDetails) {
	var flowDetails string
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

	flowDetails = details.CertIssuer
	if details.Authenticator != nil {
		flowDetails = details.Authenticator.Name
	}

	h.clientInstrumenter.InstrumentClient(details.AuthID, string(details.AuthFlow), flowDetails)
}

func respond(ctx context.Context, writer http.ResponseWriter, body oathkeeper.ReqBody) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while encoding data: %v", err)
	}
}

func addTenantsToExtra(objectContexts []ObjectContext, reqData oathkeeper.ReqData) error {
	tenants := make(map[string]string)
	for _, objCtx := range objectContexts {
		tenants[objCtx.TenantKey] = objCtx.TenantID
		tenants[objCtx.ExternalTenantKey] = objCtx.ExternalTenantID
	}

	_, consumerExists := tenants[ConsumerTenantKey]
	_, externalExists := tenants[ExternalTenantKey]
	if !consumerExists && !externalExists {
		tenants[ConsumerTenantKey] = tenants[ProviderTenantKey]
		tenants[ExternalTenantKey] = tenants[ProviderExternalTenantKey]
	}

	tenantsJSON, err := json.Marshal(tenants)
	if err != nil {
		return errors.Wrap(err, "While marshaling tenants")
	}

	tenantsStr := string(tenantsJSON)

	escaped := strings.ReplaceAll(tenantsStr, `"`, `\"`)
	reqData.Body.Extra["tenant"] = escaped

	return nil
}

func addScopesToExtra(objectContexts []ObjectContext, reqData oathkeeper.ReqData) {
	objScopes := make([][]string, 0, len(objectContexts))
	for _, objCtx := range objectContexts {
		objScopes = append(objScopes, strings.Split(objCtx.Scopes, " "))
	}

	intersection := objScopes[0]
	for _, s := range objScopes[1:] {
		intersection = intersect(intersection, s)
	}
	joined := strings.Join(intersection, " ")
	reqData.Body.Extra["scope"] = joined
}

func addConsumersToExtra(objectContexts []ObjectContext, reqData oathkeeper.ReqData) {
	consumer := consumer.Consumer{}
	if len(objectContexts) == 1 {
		consumer.ConsumerID = objectContexts[0].ConsumerID
		consumer.ConsumerType = objectContexts[0].ConsumerType
		consumer.Flow = objectContexts[0].AuthFlow
	} else {
		consumer = getCertServiceObjectContextProviderConsumer(objectContexts)
		consumer.OnBehalfOf = getOnBehalfConsumer(objectContexts)
	}

	reqData.Body.Extra["consumerID"] = consumer.ConsumerID
	reqData.Body.Extra["consumerType"] = consumer.ConsumerType
	reqData.Body.Extra["flow"] = consumer.Flow
	reqData.Body.Extra["onBehalfOf"] = consumer.OnBehalfOf
}

func getCertServiceObjectContextProviderConsumer(objectContexts []ObjectContext) consumer.Consumer {
	consumer := consumer.Consumer{}
	for _, objCtx := range objectContexts {
		if objCtx.ContextProvider == CertServiceObjectContextProvider {
			consumer.ConsumerID = objCtx.ConsumerID
			consumer.ConsumerType = objCtx.ConsumerType
			consumer.Flow = objCtx.AuthFlow
		}
	}
	return consumer
}

func getOnBehalfConsumer(objectContexts []ObjectContext) string {
	for _, objCtx := range objectContexts {
		if objCtx.ContextProvider != CertServiceObjectContextProvider {
			return objCtx.ConsumerID
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
