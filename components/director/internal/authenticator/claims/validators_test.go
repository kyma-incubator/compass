package claims_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/stretchr/testify/assert"
)

const (
	tenantID    = "af9f84a9-1d3a-4d9f-ae0c-94f883b33b6e"
	extTenantID = "bf9f84a9-1d3a-4d9f-ae0c-94f883b33b6e"
	consumerID  = "1e176e48-e258-4091-a584-feb1bf708b7e"
	scopes      = "application:read application:write"
)

func TestValidator_Validate(t *testing.T) {
	t.Run("Succeeds when all claims properties are present", func(t *testing.T) {
		v := claims.NewValidator()
		c := getClaims(tenantID, extTenantID, scopes)

		err := v.Validate(c)
		assert.NoError(t, err)
	})
	t.Run("Succeeds when no scopes are present", func(t *testing.T) {
		v := claims.NewValidator()
		c := getClaims(tenantID, extTenantID, "")

		err := v.Validate(c)
		assert.NoError(t, err)
	})
	t.Run("Succeeds when both internal and external tenant IDs are missing", func(t *testing.T) {
		v := claims.NewValidator()
		c := getClaims("", "", scopes)

		err := v.Validate(c)
		assert.NoError(t, err)
	})
	t.Run("Fails when internal tenant ID is missing", func(t *testing.T) {
		v := claims.NewValidator()
		c := getClaims("", extTenantID, "")

		err := v.Validate(c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Tenant not found")
	})
	t.Run("Fails when inner validation fails", func(t *testing.T) {
		v := claims.NewValidator()
		c := getClaims(tenantID, extTenantID, scopes)
		c.ExpiresAt = 1

		err := v.Validate(c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while validating claims")
	})
}

func TestScopesValidator_Validate(t *testing.T) {
	t.Run("Succeeds when all claims properties are present", func(t *testing.T) {
		v := claims.NewScopesValidator([]string{"application:read"})
		c := getClaims(tenantID, extTenantID, scopes)

		err := v.Validate(c)
		assert.NoError(t, err)
	})
	t.Run("Fails when no scopes are present", func(t *testing.T) {
		requiredScopes := []string{"application:read"}
		v := claims.NewScopesValidator(requiredScopes)
		c := getClaims(tenantID, extTenantID, "")

		err := v.Validate(c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("Not all required scopes %q were found in claim with scopes %q", requiredScopes, c.Scopes))
	})
	t.Run("Fails when inner validation fails", func(t *testing.T) {
		requiredScopes := []string{"application:read"}
		v := claims.NewScopesValidator(requiredScopes)
		c := getClaims(tenantID, extTenantID, scopes)
		c.ExpiresAt = 1

		err := v.Validate(c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while validating claims")
	})
}

func getClaims(intTenantID, extTenantID, scopes string) claims.Claims {
	return claims.Claims{
		Tenant:         intTenantID,
		ExternalTenant: extTenantID,
		Scopes:         scopes,
		ConsumerID:     consumerID,
		ConsumerType:   consumer.Runtime,
	}
}
