package oathkeeper

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

type ValidationHydrator interface {
	ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request)
}

type Service interface {
	GetByToken(ctx context.Context, token string) (*model.SystemAuth, error)
	InvalidateToken(ctx context.Context, item *model.SystemAuth) error
}

type validationHydrator struct {
	tokenService           Service
	transact               persistence.Transactioner
	csrTokenExpiration     time.Duration
	appTokenExpiration     time.Duration
	runtimeTokenExpiration time.Duration
}

func NewValidationHydrator(tokenService Service, transact persistence.Transactioner, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration time.Duration) ValidationHydrator {
	return &validationHydrator{
		csrTokenExpiration:     csrTokenExpiration,
		appTokenExpiration:     appTokenExpiration,
		runtimeTokenExpiration: runtimeTokenExpiration,
		tokenService:           tokenService,
		transact:               transact,
	}
}

func (tvh *validationHydrator) ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := tvh.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to open db transaction")
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("unexpected error occured while resolving one time token"))
		return
	}
	defer tvh.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var authSession oathkeeper.AuthenticationSession
	err = json.NewDecoder(r.Body).Decode(&authSession)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to decode request body")
		httputils.RespondWithError(ctx, w, http.StatusBadRequest, errors.Wrap(err, "failed to decode Authentication Session from body"))
		return
	}
	defer httputils.Close(ctx, r.Body)

	connectorToken := r.Header.Get(oathkeeper.ConnectorTokenHeader)
	if connectorToken == "" {
		connectorToken = r.URL.Query().Get(oathkeeper.ConnectorTokenQueryParam)
	}

	if connectorToken == "" {
		log.C(ctx).Info("Token not provided")
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	log.C(ctx).Info("Trying to resolve token...")

	systemAuth, err := tvh.tokenService.GetByToken(ctx, connectorToken)
	if err != nil {
		log.C(ctx).Infof("Invalid token provided: %s", err.Error())
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	var expirationTime time.Duration
	switch systemAuth.Value.OneTimeToken.Type {
	case tokens.ApplicationToken:
		expirationTime = tvh.appTokenExpiration
	case tokens.RuntimeToken:
		expirationTime = tvh.runtimeTokenExpiration
	case tokens.CSRToken:
		expirationTime = tvh.csrTokenExpiration
	default:
		log.C(ctx).Errorf("One Time Token for system auth id %s has no valid type", systemAuth.ID)
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	log.C(ctx).Infof("Validity time to check for is %s...", expirationTime.String())

	if systemAuth.Value.OneTimeToken.CreatedAt.Add(expirationTime).Before(time.Now()) {
		log.C(ctx).Infof("One Time Token for system auth id %s has expired", systemAuth.ID)
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if authSession.Header == nil {
		authSession.Header = map[string][]string{}
	}

	authSession.Header.Add(oathkeeper.ClientIdFromTokenHeader, systemAuth.ID)

	if err := tvh.tokenService.InvalidateToken(ctx, systemAuth); err != nil {
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("could not invalidate token"))
		return
	}

	err = tx.Commit()
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to commit db transaction")
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("unexpected error occured while resolving one time token"))
		return
	}

	log.C(ctx).Infof("Token for %s resolved successfully", systemAuth.ID)
	respondWithAuthSession(ctx, w, authSession)
}

func respondWithAuthSession(ctx context.Context, w http.ResponseWriter, authSession oathkeeper.AuthenticationSession) {
	httputils.RespondWithBody(ctx, w, http.StatusOK, authSession)
}
