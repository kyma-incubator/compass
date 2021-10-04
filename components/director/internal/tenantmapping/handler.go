package tenantmapping

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/pkg/errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
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
	// KeysHeader key for the Header that contains tenant and internalTenant keys that should be used in the idToken
	KeysHeader = "Extra-Keys"
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
	GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails, extraTenantKeys KeysExtra) (ObjectContext, error)
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
	authenticators         []authenticator.Config
	reqDataParser          ReqDataParser
	transact               persistence.Transactioner
	objectContextProviders map[string]ObjectContextProvider
	clientInstrumenter     ClientInstrumenter
}

// NewHandler missing godoc
func NewHandler(
	authenticators []authenticator.Config,
	reqDataParser ReqDataParser,
	transact persistence.Transactioner,
	objectContextProviders map[string]ObjectContextProvider,
	clientInstrumenter ClientInstrumenter) *Handler {
	return &Handler{
		authenticators:         authenticators,
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
		"consumers": body.Extra["consumers"],
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
		log.C(ctx).WithError(err).Errorf("An error occurred while determining the auth details for the request: %v", err)
		return reqData.Body
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while committing transaction: %v", err)
		return reqData.Body
	}

	addTenantsToExtra(objCtxs, reqData)

	addScopesToExtra(objCtxs, reqData)

	addConsumersToExtra(objCtxs, reqData)

	return reqData.Body
}

func (h *Handler) getObjectContexts(ctx context.Context, reqData oathkeeper.ReqData) ([]ObjectContext, error) {
	log.C(ctx).Infof("Attempting to get object contexts")

	var objectContexts []ObjectContext
	var authDetails []*oathkeeper.AuthDetails
	for name, provider := range h.objectContextProviders {
		match, details, err := provider.Match(ctx, reqData)
		authDetails = append(authDetails, details)
		if match && err == nil {
			keys, err := extractKeys(reqData, name)
			if err != nil {
				return nil, errors.Wrap(err, "while extracting keys: ")
			}

			objectContext, err := provider.GetObjectContext(ctx, reqData, *details, keys)
			if err != nil {
				return nil, errors.Wrap(err, "while getting objectContexts: ")
			}

			objectContexts = append(objectContexts, objectContext)
		}
	}

	h.instrumentClient(objectContexts, authDetails)

	return objectContexts, nil
}

func (h *Handler) instrumentClient(objectContexts []ObjectContext, authDetails []*oathkeeper.AuthDetails) {
	var flowDetails string
	var details *oathkeeper.AuthDetails

	if len(objectContexts) == 1 {
		details = authDetails[0]
	} else {
		for _, d := range authDetails {
			if d.CertIssuer != "" {
				details = d
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

func extractKeys(reqData oathkeeper.ReqData, objectContextProviderName string) (KeysExtra, error) {
	keysStringg := reqData.Body.Header[KeysHeader]
	if len(keysStringg) < 1 {
		return KeysExtra{}, errors.New(`missing "Extra-Keys" header`)
	}
	keysString := keysStringg[0]
	keysJSON, err := strconv.Unquote(keysString)
	if err != nil {
		return KeysExtra{}, err
	}

	var keys map[string]KeysExtra

	err = json.Unmarshal([]byte(keysJSON), &keys)
	if err != nil {
		return KeysExtra{}, err
	}

	return keys[objectContextProviderName], nil
}

func respond(ctx context.Context, writer http.ResponseWriter, body oathkeeper.ReqBody) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while encoding data: %v", err)
	}
}

func addTenantsToExtra(objectContexts []ObjectContext, reqData oathkeeper.ReqData) {
	tenants := make(map[string]string)
	for _, objCtx := range objectContexts {
		tenants[objCtx.TenantKey] = objCtx.TenantID
		tenants[objCtx.ExternalTenantKey] = objCtx.ExternalTenantID
	}

	if _, ok := tenants["consumerTenant"]; !ok {
		tenants["consumerTenant"] = tenants["providerTenant"]
	}

	tenantsJSON, err := json.Marshal(tenants)
	if err != nil {
	}

	tenantsStr := string(tenantsJSON)

	escaped := strings.ReplaceAll(tenantsStr, `"`, `\"`)
	reqData.Body.Extra["tenant"] = escaped
}

func addScopesToExtra(objectContexts []ObjectContext, reqData oathkeeper.ReqData) {
	var objScopes [][]string
	fmt.Println("CONTEXTS:           ", objectContexts)
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
	var consumers []consumer.Consumer
	for _, objCtx := range objectContexts {
		consumer := consumer.Consumer{
			ConsumerID:   objCtx.ConsumerID,
			ConsumerType: objCtx.ConsumerType,
			Flow:         objCtx.AuthFlow,
		}
		consumers = append(consumers, consumer)
	}

	sort.Slice(consumers, func(i, j int) bool {
		return consumers[i].ConsumerType < consumers[j].ConsumerType
	})

	consumersJSON, err := json.Marshal(consumers)
	if err != nil {
	}

	consumersStr := string(consumersJSON)
	escaped := strings.ReplaceAll(consumersStr, `"`, `\"`)

	reqData.Body.Extra["consumers"] = escaped
}

func intersect(s1 []string, s2 []string) []string {
	var intersection []string

	for _, s := range s1 {
		if contains(s2, s) {
			intersection = append(intersection, s)
		}
	}

	return intersection
}

func contains(slice []string, s string) bool {
	for _, ss := range slice {
		if s == ss {
			return true
		}
	}

	return false
}
