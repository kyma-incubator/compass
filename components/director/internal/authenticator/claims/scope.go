package claims

import (
	"context"
	"fmt"
	"strings"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
)

type scopeBasedClaims struct {
	Scopes []string `json:"scope"`
	ZID    string   `json:"zid"`
	jwt.StandardClaims
}

func (c *scopeBasedClaims) ContextWithClaims(ctx context.Context) context.Context {
	return scope.SaveToContext(ctx, c.Scopes)
}

type scopeBasedClaimsParser struct {
	zoneID                     string
	trustedClaimPrefixes       []string
	subscriptionCallbacksScope string
}

func NewScopeBasedClaimsParser(zoneID string, trustedClaimPrefixes []string, subscriptionCallbacksScope string) *scopeBasedClaimsParser {
	return &scopeBasedClaimsParser{
		zoneID:                     zoneID,
		trustedClaimPrefixes:       trustedClaimPrefixes,
		subscriptionCallbacksScope: subscriptionCallbacksScope,
	}
}

func (p *scopeBasedClaimsParser) ParseClaims(ctx context.Context, bearerToken string, keyfunc jwt.Keyfunc) (Claims, error) {
	parsed := scopeBasedClaims{}
	_, err := jwt.ParseWithClaims(bearerToken, &parsed, keyfunc)
	if err != nil {
		return nil, err
	}

	if err := p.validateClaims(ctx, parsed); err != nil {
		return nil, err
	}

	return &parsed, nil
}

func (p *scopeBasedClaimsParser) validateClaims(ctx context.Context, claims scopeBasedClaims) error {
	if err := claims.Valid(); err != nil {
		return err
	}

	if claims.ZID != p.zoneID {
		log.C(ctx).Errorf("Zone id %q from user token does not match the trusted zone %s", claims.ZID, p.zoneID)
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Zone id %q from user token is not trusted", claims.ZID))
	}

	scopes := prefixScopes(p.trustedClaimPrefixes, p.subscriptionCallbacksScope)
	log.C(ctx).Infof("SCOPES WHILE VALIDATING: %s", strings.Join(scopes, " "))
	if !stringsAnyEquals(scopes, strings.Join(claims.Scopes, " ")) {
		log.C(ctx).Errorf(`Scope "%s" from user token does not match the trusted scopes`, claims.Scopes)
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Scope %q is not trusted", claims.Scopes))
	}
	return nil
}

func prefixScopes(prefixes []string, callbackScope string) []string {
	prefixedScopes := make([]string, 0, len(prefixes))
	for _, scope := range prefixes {
		prefixedScopes = append(prefixedScopes, fmt.Sprintf("%s%s", scope, callbackScope))
	}
	return prefixedScopes
}

func stringsAnyEquals(stringSlice []string, str string) bool {
	for _, v := range stringSlice {
		if strings.Contains(str, v) {
			return true
		}
	}
	return false
}
