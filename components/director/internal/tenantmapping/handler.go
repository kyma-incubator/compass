package tenantmapping

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ScopesGetter -output=automock -outpkg=automock -case=underscore
type ScopesGetter interface {
	GetRequiredScopes(scopesDefinition string) ([]string, error)
}

//go:generate mockery -name=ReqDataParser -output=automock -outpkg=automock -case=underscore
type ReqDataParser interface {
	Parse(req *http.Request) (oathkeeper.ReqData, error)
}

//go:generate mockery -name=ObjectContextForUserProvider -output=automock -outpkg=automock -case=underscore
type ObjectContextForUserProvider interface {
	GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authID string) (ObjectContext, error)
}

//go:generate mockery -name=ObjectContextForSystemAuthProvider -output=automock -outpkg=automock -case=underscore
type ObjectContextForSystemAuthProvider interface {
	GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authID string, authFlow oathkeeper.AuthFlow) (ObjectContext, error)
}

//go:generate mockery -name=TenantRepository -output=automock -outpkg=automock -case=underscore
type TenantRepository interface {
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
}

type LoggerKey = struct{}

type Handler struct {
	reqDataParser       ReqDataParser
	transact            persistence.Transactioner
	mapperForUser       ObjectContextForUserProvider
	mapperForSystemAuth ObjectContextForSystemAuthProvider
	logger              *logrus.Entry
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
		err := fmt.Sprintf("Bad request method. Got %s, expected POST", req.Method)
		h.logger.Errorf(err)
		http.Error(writer, err, http.StatusBadRequest)
		return
	}

	reqData, err := h.reqDataParser.Parse(req)
	if err != nil {
		h.logger.Errorf("An error occurred while parsing the request: %s", err.Error())
		respond(writer, reqData.Body, h.logger)
		return
	}

	logger := configureLogger(h.logger, reqData)
	newCtx := saveLoggerToContext(req.Context(), logger)

	body := h.processRequest(newCtx, reqData)
	respond(writer, body, logger)
}

func (h Handler) processRequest(ctx context.Context, reqData oathkeeper.ReqData) oathkeeper.ReqBody {
	log := loggerFromContextOrDefault(ctx)

	tx, err := h.transact.Begin()
	if err != nil {
		log.Errorf("An error occurred while opening the db transaction: %s", err.Error())
		return reqData.Body
	}
	defer h.transact.RollbackUnlessCommitted(tx)

	newCtx := persistence.SaveToContext(ctx, tx)

	log.Debug("Getting object context")
	objCtx, err := h.getObjectContext(newCtx, reqData)
	if err != nil {
		log.Errorf("An error occurred while getting object context: %s", err.Error())
		return reqData.Body
	}

	if err := tx.Commit(); err != nil {
		log.Errorf("An error occurred while committing transaction: %s", err.Error())
		return reqData.Body
	}

	reqData.Body.Extra["tenant"] = objCtx.TenantID
	reqData.Body.Extra["externalTenant"] = objCtx.ExternalTenantID
	reqData.Body.Extra["scope"] = objCtx.Scopes
	reqData.Body.Extra["consumerID"] = objCtx.ConsumerID
	reqData.Body.Extra["consumerType"] = objCtx.ConsumerType

	return reqData.Body
}

func (h *Handler) getObjectContext(ctx context.Context, reqData oathkeeper.ReqData) (ObjectContext, error) {
	authID, authFlow, err := reqData.GetAuthID()
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while determining the auth ID from the request")
	}

	switch authFlow {
	case oathkeeper.JWTAuthFlow:
		return h.mapperForUser.GetObjectContext(ctx, reqData, authID)
	case oathkeeper.OAuth2Flow, oathkeeper.CertificateFlow, oathkeeper.OneTimeTokenFlow:
		return h.mapperForSystemAuth.GetObjectContext(ctx, reqData, authID, authFlow)
	}

	return ObjectContext{}, fmt.Errorf("unknown authentication flow (%s)", authFlow)
}

func respond(writer http.ResponseWriter, body oathkeeper.ReqBody, logger *logrus.Entry) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		logger.Errorf("An error occurred while encoding data: %s", err.Error())
	}
}

func saveLoggerToContext(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, LoggerKey{}, logger)
}

func loggerFromContextOrDefault(ctx context.Context) *logrus.Entry {
	log, ok := ctx.Value(LoggerKey{}).(*logrus.Entry)
	if !ok {
		return logrus.WithField("component", "tenant-mapping-handler")
	}
	return log
}

func configureLogger(logger *logrus.Entry, reqData oathkeeper.ReqData) *logrus.Entry {
	authID, authFlow, err := reqData.GetAuthID()
	if err != nil {
		return logger
	}

	return logger.WithFields(logrus.Fields{
		"authID":   authID,
		"authFlow": authFlow,
	})
}
