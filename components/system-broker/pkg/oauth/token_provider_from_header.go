package oauth

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/pkg/errors"
)

const AuthzHeader = "Authorization"

type tokenAuthorizationProviderFromHeader struct{}

func NewTokenAuthorizationProviderFromHeader() *tokenAuthorizationProviderFromHeader {
	return &tokenAuthorizationProviderFromHeader{}
}

func (c *tokenAuthorizationProviderFromHeader) Name() string {
	return "TokenAuthorizationProviderFromHeader"
}

func (c *tokenAuthorizationProviderFromHeader) Matches(ctx context.Context) bool {
	if _, err := getBearerAuthorizationValue(ctx); err != nil {
		log.C(ctx).WithError(err).Warn("while obtaining authorization from header")
		return false
	}

	return true
}

func (c *tokenAuthorizationProviderFromHeader) GetAuthorization(ctx context.Context) (string, error) {
	authorization, err := getBearerAuthorizationValue(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while obtaining authorization from header %s", AuthzHeader)
	}

	log.C(ctx).Info("Successfully unmarshal response authorization for accessing Director")
	return authorization, nil
}

func getBearerAuthorizationValue(ctx context.Context) (string, error) {
	headers, err := httputils.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Errorf("cannot read headers from context: %s", err.Error())
	}
	reqAuth, ok := headers[AuthzHeader]
	if !ok {
		return "", errors.Errorf("cannot read header %s from context", AuthzHeader)
	}

	if reqAuth == "" {
		return "", apperrors.NewUnauthorizedError("missing bearer token")
	}

	if !strings.HasPrefix(strings.ToLower(reqAuth), "bearer ") {
		return "", apperrors.NewUnauthorizedError("invalid bearer token prefix")
	}

	return reqAuth, nil
}
