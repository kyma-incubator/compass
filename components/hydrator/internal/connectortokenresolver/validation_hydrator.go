package connectortokenresolver

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// ValidationHydrator missing godoc
type ValidationHydrator interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type validationHydrator struct {
	directorClient DirectorClient
}

// DirectorClient missing godoc
//go:generate mockery --name=DirectorClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type DirectorClient interface {
	GetSystemAuthByToken(ctx context.Context, token string) (*model.SystemAuth, error)
	InvalidateSystemAuthOneTimeToken(ctx context.Context, authID string) error
}

// NewValidationHydrator missing godoc
func NewValidationHydrator(clientProvider DirectorClient) ValidationHydrator {
	return &validationHydrator{
		directorClient: clientProvider,
	}
}

// ServeHTTP missing godoc
func (vh *validationHydrator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var authSession oathkeeper.AuthenticationSession
	if err := json.NewDecoder(r.Body).Decode(&authSession); err != nil {
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

	systemAuth, err := vh.directorClient.GetSystemAuthByToken(ctx, connectorToken)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Invalid token provided: %s", err.Error())
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if authSession.Header == nil {
		authSession.Header = map[string][]string{}
	}

	authSession.Header.Add(oathkeeper.ClientIdFromTokenHeader, systemAuth.ID)

	if err := vh.directorClient.InvalidateSystemAuthOneTimeToken(ctx, systemAuth.ID); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to invalidate token: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("could not invalidate token"))
		return
	}

	log.C(ctx).Infof("Token for %s resolved successfully", systemAuth.ID)
	respondWithAuthSession(ctx, w, authSession)
}

func respondWithAuthSession(ctx context.Context, w http.ResponseWriter, authSession oathkeeper.AuthenticationSession) {
	httputils.RespondWithBody(ctx, w, http.StatusOK, authSession)
}
