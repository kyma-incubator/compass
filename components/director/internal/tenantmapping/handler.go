package tenantmapping

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

const (
	UserObjectContextProvider          = "UserObjectContextProvider"
	SystemAuthObjectContextProvider    = "SystemAuthObjectContextProvider"
	AuthenticatorObjectContextProvider = "AuthenticatorObjectContextProvider"
)

//go:generate mockery -name=ScopesGetter -output=automock -outpkg=automock -case=underscore
type ScopesGetter interface {
	GetRequiredScopes(scopesDefinition string) ([]string, error)
}

//go:generate mockery -name=ReqDataParser -output=automock -outpkg=automock -case=underscore
type ReqDataParser interface {
	Parse(req *http.Request) (oathkeeper.ReqData, error)
}

//go:generate mockery -name=ObjectContextProvider -output=automock -outpkg=automock -case=underscore
type ObjectContextProvider interface {
	GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error)
}

//go:generate mockery -name=TenantRepository -output=automock -outpkg=automock -case=underscore
type TenantRepository interface {
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
}

type Handler struct {
	authenticators         []authenticator.Config
	reqDataParser          ReqDataParser
	transact               persistence.Transactioner
	objectContextProviders map[string]ObjectContextProvider
}

func NewHandler(
	authenticators []authenticator.Config,
	reqDataParser ReqDataParser,
	transact persistence.Transactioner,
	objectContextProviders map[string]ObjectContextProvider) *Handler {
	return &Handler{
		authenticators:         authenticators,
		reqDataParser:          reqDataParser,
		transact:               transact,
		objectContextProviders: objectContextProviders,
	}
}

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
		log.C(ctx).WithError(err).Errorf("An error occurred while parsing request.")
		respond(ctx, writer, reqData.Body)
		return
	}

	authDetails, err := reqData.GetAuthIDWithAuthenticators(ctx, h.authenticators)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while determining the auth details for the request.")
		respond(ctx, writer, reqData.Body)
		return
	}

	logger := log.C(ctx).WithFields(logrus.Fields{
		"authID":        authDetails.AuthID,
		"authFlow":      authDetails.AuthFlow,
		"authenticator": authDetails.Authenticator,
	})

	newCtx := log.ContextWithLogger(ctx, logger)

	body := h.processRequest(newCtx, reqData, *authDetails)
	respond(newCtx, writer, body)
}

func (h Handler) processRequest(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) oathkeeper.ReqBody {
	tx, err := h.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while opening db transaction.")
		return reqData.Body
	}
	defer h.transact.RollbackUnlessCommitted(ctx, tx)

	newCtx := persistence.SaveToContext(ctx, tx)

	log.C(ctx).Debug("Getting object context")
	objCtx, err := h.getObjectContext(newCtx, reqData, authDetails)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while getting object context.")
		return reqData.Body
	}

	if err := tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while committing transaction.")
		return reqData.Body
	}

	reqData.Body.Extra["tenant"] = objCtx.TenantID
	reqData.Body.Extra["externalTenant"] = objCtx.ExternalTenantID
	reqData.Body.Extra["scope"] = objCtx.Scopes
	reqData.Body.Extra["consumerID"] = objCtx.ConsumerID
	reqData.Body.Extra["consumerType"] = objCtx.ConsumerType

	return reqData.Body
}

func (h *Handler) getObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	log.C(ctx).Infof("Attempting to get object context for %s flow and authID=%s", authDetails.AuthFlow, authDetails.AuthID)

	var provider ObjectContextProvider
	switch authDetails.AuthFlow {
	case oathkeeper.JWTAuthFlow:
		if authDetails.Authenticator != nil {
			provider = h.objectContextProviders[AuthenticatorObjectContextProvider]
		} else {
			provider = h.objectContextProviders[UserObjectContextProvider]
		}
	case oathkeeper.OAuth2Flow, oathkeeper.CertificateFlow, oathkeeper.OneTimeTokenFlow:
		provider = h.objectContextProviders[SystemAuthObjectContextProvider]
	default:
		return ObjectContext{}, fmt.Errorf("unknown authentication flow (%s)", authDetails.AuthFlow)
	}

	return provider.GetObjectContext(ctx, reqData, authDetails)
}

func respond(ctx context.Context, writer http.ResponseWriter, body oathkeeper.ReqBody) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while encoding data.")
	}
}
