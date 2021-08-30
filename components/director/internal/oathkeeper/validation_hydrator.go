package oathkeeper

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	directorTime "github.com/kyma-incubator/compass/components/director/pkg/time"
	"github.com/pkg/errors"
)

type ValidationHydrator interface {
	ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request)
}

//go:generate mockery --name=Service --output=automock --outpkg=automock --case=underscore
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
	timeService            directorTime.Service
}

func NewValidationHydrator(tokenService Service, transact persistence.Transactioner, timeService directorTime.Service, csrTokenExpiration, appTokenExpiration, runtimeTokenExpiration time.Duration) ValidationHydrator {
	return &validationHydrator{
		csrTokenExpiration:     csrTokenExpiration,
		appTokenExpiration:     appTokenExpiration,
		runtimeTokenExpiration: runtimeTokenExpiration,
		tokenService:           tokenService,
		transact:               transact,
		timeService:            timeService,
	}
}

func (vh *validationHydrator) ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := vh.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to open db transaction: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("unexpected error occured while resolving one time token"))
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

	log.C(ctx).Info("Trying to decode and parse token...")

	decodedToken, err := base64.URLEncoding.DecodeString(connectorToken)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("While decoding token %s", connectorToken)
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	var tokenData model.TokenData
	if err := json.Unmarshal(decodedToken, &tokenData); err != nil {
		log.C(ctx).WithError(err).Errorf("while unmarshalling token %s", decodedToken)
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	log.C(ctx).Info("Trying to resolve token...")

	systemAuth, err := vh.tokenService.GetByToken(ctx, tokenData.Token)
	if err != nil {
		log.C(ctx).Infof("Invalid token provided: %s", err.Error())
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if systemAuth.Value == nil || systemAuth.Value.OneTimeToken == nil {
		log.C(ctx).Infof("Cannot get OneTimeToken from systemAuth with ID: %s", systemAuth.ID)
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	expirationTime, err := vh.getExpirationTimeForToken(systemAuth)
	if err != nil {
		log.C(ctx).Error(err)
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if systemAuth.Value.OneTimeToken.CreatedAt.Add(expirationTime).Before(vh.timeService.Now()) {
		log.C(ctx).Infof("One Time Token with validity %s for system auth with ID %q has expired", expirationTime.String(), systemAuth.ID)
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if authSession.Header == nil {
		authSession.Header = map[string][]string{}
	}

	authSession.Header.Add(oathkeeper.ClientIdFromTokenHeader, systemAuth.ID)

	if err := vh.tokenService.InvalidateToken(ctx, systemAuth); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to invalidate token: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("could not invalidate token"))
		return
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to commit db transaction: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("unexpected error occured while resolving one time token"))
		return
	}

	log.C(ctx).Infof("Token for %s resolved successfully", systemAuth.ID)
	respondWithAuthSession(ctx, w, authSession)
}

func (vh *validationHydrator) getExpirationTimeForToken(systemAuth *model.SystemAuth) (time.Duration, error) {
	switch systemAuth.Value.OneTimeToken.Type {
	case tokens.ApplicationToken:
		return vh.appTokenExpiration, nil
	case tokens.RuntimeToken:
		return vh.runtimeTokenExpiration, nil
	case tokens.CSRToken:
		return vh.csrTokenExpiration, nil
	default:
		return time.Duration(0), errors.Errorf("One Time Token for system auth id %s has no valid type", systemAuth.ID)
	}
}
func respondWithAuthSession(ctx context.Context, w http.ResponseWriter, authSession oathkeeper.AuthenticationSession) {
	httputils.RespondWithBody(ctx, w, http.StatusOK, authSession)
}
