package oathkeeper

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/systemauth"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// ValidationHydrator missing godoc
type ValidationHydrator interface {
	ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request)
}

// SystemAuthService missing godoc
//go:generate mockery --name=SystemAuthService --output=automock --outpkg=automock --case=underscore
type SystemAuthService interface {
	GetByToken(ctx context.Context, token string) (*systemauth.SystemAuth, error)
	InvalidateToken(ctx context.Context, id string) (*systemauth.SystemAuth, error)
}

// OneTimeTokenService missing godoc
//go:generate mockery --name=OneTimeTokenService --output=automock --outpkg=automock --case=underscore
type OneTimeTokenService interface {
	IsTokenValid(systemAuth *systemauth.SystemAuth) (bool, error)
}

type validationHydrator struct {
	systemAuthService   SystemAuthService
	transact            persistence.Transactioner
	oneTimeTokenService OneTimeTokenService
}

// NewValidationHydrator missing godoc
func NewValidationHydrator(systemAuthService SystemAuthService, transact persistence.Transactioner, oneTimeTokenService OneTimeTokenService) ValidationHydrator {
	return &validationHydrator{
		systemAuthService:   systemAuthService,
		transact:            transact,
		oneTimeTokenService: oneTimeTokenService,
	}
}

// ResolveConnectorTokenHeader missing godoc
func (vh *validationHydrator) ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := vh.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to open db transaction: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("unexpected error occurred while resolving one time token"))
		return
	}
	defer vh.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var authSession oathkeeper.AuthenticationSession
	if err = json.NewDecoder(r.Body).Decode(&authSession); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to decode request body: %v", err)
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

	systemAuth, err := vh.systemAuthService.GetByToken(ctx, connectorToken)
	if err != nil {
		log.C(ctx).Infof("Invalid token provided: %s", err.Error())
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if _, err := vh.oneTimeTokenService.IsTokenValid(systemAuth); err != nil {
		log.C(ctx).Error(err)
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if authSession.Header == nil {
		authSession.Header = map[string][]string{}
	}

	authSession.Header.Add(oathkeeper.ClientIdFromTokenHeader, systemAuth.ID)

	if _, err := vh.systemAuthService.InvalidateToken(ctx, systemAuth.ID); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to invalidate token: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("could not invalidate token"))
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to commit db transaction: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("unexpected error occurred while resolving one time token"))
		return
	}

	log.C(ctx).Infof("Token for %s resolved successfully", systemAuth.ID)
	respondWithAuthSession(ctx, w, authSession)
}

func respondWithAuthSession(ctx context.Context, w http.ResponseWriter, authSession oathkeeper.AuthenticationSession) {
	httputils.RespondWithBody(ctx, w, http.StatusOK, authSession)
}
