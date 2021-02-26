package oauth

import (
	"context"
	"net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/pkg/errors"
)

const AuthzHeader = "Authorization"

type TokenAuthorizationProviderFromHeader struct {
	targetURL *url.URL
}

func NewTokenAuthorizationProviderFromHeader(targetURL string) (*TokenAuthorizationProviderFromHeader, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	return &TokenAuthorizationProviderFromHeader{
		targetURL: parsedURL,
	}, nil
}

func (c *TokenAuthorizationProviderFromHeader) Name() string {
	return "TokenAuthorizationProviderFromHeader"
}

func (c *TokenAuthorizationProviderFromHeader) Matches(ctx context.Context) bool {
	if _, err := getBearerAuthorizationValue(ctx); err != nil {
		log.C(ctx).WithError(err).Warn("while obtaining authorization from header")
		return false
	}

	return true
}

func (c *TokenAuthorizationProviderFromHeader) TargetURL() *url.URL {
	return c.targetURL
}

func (c *TokenAuthorizationProviderFromHeader) GetAuthorization(ctx context.Context) (string, error) {
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
