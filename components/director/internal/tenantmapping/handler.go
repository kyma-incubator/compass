package tenantmapping

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ScopesGetter -output=automock -outpkg=automock -case=underscore
type ScopesGetter interface {
	GetRequiredScopes(scopesDefinition string) ([]string, error)
}

//go:generate mockery -name=ReqDataParser -output=automock -outpkg=automock -case=underscore
type ReqDataParser interface {
	Parse(req *http.Request) (ReqData, error)
}

//go:generate mockery -name=ObjectContextForUserProvider -output=automock -outpkg=automock -case=underscore
type ObjectContextForUserProvider interface {
	GetObjectContext(ctx context.Context, reqData ReqData, authID string) (ObjectContext, error)
}

//go:generate mockery -name=ObjectContextForSystemAuthProvider -output=automock -outpkg=automock -case=underscore
type ObjectContextForSystemAuthProvider interface {
	GetObjectContext(ctx context.Context, reqData ReqData, authID string, authFlow AuthFlow) (ObjectContext, error)
}

//go:generate mockery -name=Logger -output=automock -outpkg=automock -case=underscore
type Logger interface {
	Error(args ...interface{})
}

//go:generate mockery -name=TenantRepository -output=automock -outpkg=automock -case=underscore
type TenantRepository interface {
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
}

type Handler struct {
	reqDataParser       ReqDataParser
	transact            persistence.Transactioner
	mapperForUser       ObjectContextForUserProvider
	mapperForSystemAuth ObjectContextForSystemAuthProvider
	logger              Logger
}

func NewHandler(
	reqDataParser ReqDataParser,
	transact persistence.Transactioner,
	mapperForUser ObjectContextForUserProvider,
	mapperForSystemAuth ObjectContextForSystemAuthProvider) *Handler {
	return &Handler{
		reqDataParser:       reqDataParser,
		transact:            transact,
		mapperForUser:       mapperForUser,
		mapperForSystemAuth: mapperForSystemAuth,
		logger:              logrus.WithField("component", "tenant-mapping-handler"),
	}
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(writer, fmt.Sprintf("Bad request method. Got %s, expected POST", req.Method), http.StatusBadRequest)
		return
	}

	reqData, err := h.reqDataParser.Parse(req)
	if err != nil {
		h.logError(err, "while parsing the request")
		h.respond(writer, reqData.Body)
		return
	}

	tx, err := h.transact.Begin()
	if err != nil {
		h.logError(err, "while opening the db transaction")
		h.respond(writer, reqData.Body)
		return
	}
	defer h.transact.RollbackUnlessCommited(tx)

	ctx := persistence.SaveToContext(req.Context(), tx)

	objCtx, err := h.getObjectContext(ctx, reqData)
	if err != nil {
		h.logError(err, "while getting object context")
		h.respond(writer, reqData.Body)
		return
	}

	reqData.Body.Extra["tenant"] = objCtx.TenantID
	reqData.Body.Extra["externalTenant"] = objCtx.ExternalTenantID
	reqData.Body.Extra["scope"] = objCtx.Scopes
	reqData.Body.Extra["consumerID"] = objCtx.ConsumerID
	reqData.Body.Extra["consumerType"] = objCtx.ConsumerType

	h.respond(writer, reqData.Body)
}

func (h *Handler) getObjectContext(ctx context.Context, reqData ReqData) (ObjectContext, error) {
	authID, authFlow, err := reqData.GetAuthID()
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while determining the auth ID from the request")
	}

	switch authFlow {
	case JWTAuthFlow:
		return h.mapperForUser.GetObjectContext(ctx, reqData, authID)
	case OAuth2Flow, CertificateFlow, OneTimeTokenFlow:
		return h.mapperForSystemAuth.GetObjectContext(ctx, reqData, authID, authFlow)
	}

	return ObjectContext{}, fmt.Errorf("unknown authentication flow (%s)", authFlow)
}

func (h *Handler) logError(err error, wrapperStr string) {
	wrappedErr := errors.Wrap(err, wrapperStr)
	h.logger.Error(wrappedErr)
}

func (h *Handler) respond(writer http.ResponseWriter, body ReqBody) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		h.logError(err, "while encoding data")
	}
}
