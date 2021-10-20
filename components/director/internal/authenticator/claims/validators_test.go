package claims_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/stretchr/testify/assert"
)

const (
	tenantID             = "tnt"
	extTenantID          = "ext-tnt"
	consumerID           = "consumerID"
	providerTenantID     = "provider-tnt"
	providerExtTenantID  = "provider-ext-tnt"
	onBehalfOfConsumerID = "onBehalfOfConsumer"
	region               = "region"
	clientID             = "client_id"
	scopes               = "application:read application:write"

	runtimeID  = "rt-id"
	runtime2ID = "rt-id2"
)

func TestValidator_Validate(t *testing.T) {
	providerLabelKey := "providerName"
	consumerIDsLabelKey := "consumerIDs"
	testErr := errors.New("test")

	runtimes := []*model.Runtime{
		{
			ID:   runtimeID,
			Name: "rt",
		},
		{
			ID:   runtime2ID,
			Name: "rt2",
		},
	}

	expectedFilters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(providerLabelKey, fmt.Sprintf("\"%s\"", clientID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", region)),
	}

	invalidLabel := &model.Label{
		ID:         "lbl-id",
		Tenant:     providerTenantID,
		Key:        consumerIDsLabelKey,
		Value:      []interface{}{"invalid-value"},
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	lbl := &model.Label{
		ID:         "lbl-id",
		Tenant:     providerTenantID,
		Key:        consumerIDsLabelKey,
		Value:      []interface{}{extTenantID},
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	testCases := []struct {
		Name            string
		RuntimeServicFn func() *automock.RuntimeService
		Claims          claims.Claims
		ExpectedErr     string
	}{
		{
			Name:   "Succeeds when all claims properties are present",
			Claims: getClaims(tenantID, extTenantID, scopes),
		},
		{
			Name:   "Succeeds when no scopes are present",
			Claims: getClaims(tenantID, extTenantID, ""),
		},
		{
			Name:   "Succeeds when both internal and external tenant IDs are missing",
			Claims: getClaims("", "", scopes),
		},
		{
			Name:        "Fails when internal tenant ID is missing",
			Claims:      getClaims("", extTenantID, ""),
			ExpectedErr: "Tenant not found",
		},
		{
			Name: "Fails when inner validation fails",
			Claims: claims.Claims{
				Tenant: map[string]string{
					"consumerTenant": tenantID,
					"externalTenant": extTenantID,
				},

				Scopes:       scopes,
				ConsumerID:   consumerID,
				ConsumerType: consumer.Runtime,
				StandardClaims: jwt.StandardClaims{
					ExpiresAt: 1,
				},
			},
			ExpectedErr: "while validating claims",
		},
		{
			Name:   "Success for consumer - provider flow does not proceed when rt with label is found",
			Claims: getClaimsForConsumerProviderFlow(tenantID, extTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServicFn: func() *automock.RuntimeService {
				rsvc := &automock.RuntimeService{}
				rsvc.On("ListByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(runtimes, nil).Once()
				rsvc.On("GetLabel", contextThatHasTenant(providerTenantID), runtimeID, consumerIDsLabelKey).Return(lbl, nil).Once()
				return rsvc
			},
		},
		{
			Name:   "Success for consumer - provider flow proceed when rt without label is found",
			Claims: getClaimsForConsumerProviderFlow(tenantID, extTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServicFn: func() *automock.RuntimeService {
				rsvc := &automock.RuntimeService{}
				rsvc.On("ListByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(runtimes, nil).Once()
				rsvc.On("GetLabel", contextThatHasTenant(providerTenantID), runtimeID, consumerIDsLabelKey).Return(invalidLabel, nil).Once()
				rsvc.On("GetLabel", contextThatHasTenant(providerTenantID), runtime2ID, consumerIDsLabelKey).Return(lbl, nil).Once()
				return rsvc
			},
		},
		{
			Name:   "Success for consumer - provider flow proceed when rt label is found",
			Claims: getClaimsForConsumerProviderFlow(tenantID, extTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServicFn: func() *automock.RuntimeService {
				rsvc := &automock.RuntimeService{}
				rsvc.On("ListByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(runtimes, nil).Once()
				rsvc.On("GetLabel", contextThatHasTenant(providerTenantID), runtimeID, consumerIDsLabelKey).Return(nil, apperrors.NewNotFoundError(resource.Label, "")).Once()
				rsvc.On("GetLabel", contextThatHasTenant(providerTenantID), runtime2ID, consumerIDsLabelKey).Return(lbl, nil).Once()
				return rsvc
			},
		},
		{
			Name:        "consumer-provider flow: error when token clientID missing",
			Claims:      getClaimsForConsumerProviderFlow(tenantID, extTenantID, providerTenantID, providerExtTenantID, scopes, region, ""),
			ExpectedErr: "could not find consumer token client ID",
		},
		{
			Name:        "consumer-provider flow: error when region missing",
			Claims:      getClaimsForConsumerProviderFlow(tenantID, extTenantID, providerTenantID, providerExtTenantID, scopes, "", clientID),
			ExpectedErr: "could not determine token's region",
		},
		{
			Name:   "consumer-provider flow: error while listing runtimes",
			Claims: getClaimsForConsumerProviderFlow(tenantID, extTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServicFn: func() *automock.RuntimeService {
				rsvc := &automock.RuntimeService{}
				rsvc.On("ListByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(nil, testErr).Once()
				return rsvc
			},
			ExpectedErr: testErr.Error(),
		},
		{
			Name:   "consumer-provider flow: error when runtime does not exists",
			Claims: getClaimsForConsumerProviderFlow(tenantID, extTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServicFn: func() *automock.RuntimeService {
				rsvc := &automock.RuntimeService{}
				rsvc.On("ListByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID)).Once()
				return rsvc
			},
			ExpectedErr: "Object not found",
		},
		{
			Name:   "consumer-provider flow: error while getting labels",
			Claims: getClaimsForConsumerProviderFlow(tenantID, extTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServicFn: func() *automock.RuntimeService {
				rsvc := &automock.RuntimeService{}
				rsvc.On("ListByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(runtimes, nil).Once()
				rsvc.On("GetLabel", contextThatHasTenant(providerTenantID), runtimeID, consumerIDsLabelKey).Return(nil, testErr).Once()
				return rsvc
			},
			ExpectedErr: testErr.Error(),
		},
		{
			Name:   "consumer-provider flow: error when no rt with the right consumer label is found",
			Claims: getClaimsForConsumerProviderFlow(tenantID, extTenantID, providerTenantID, providerExtTenantID, scopes, region, clientID),
			RuntimeServicFn: func() *automock.RuntimeService {
				rsvc := &automock.RuntimeService{}
				rsvc.On("ListByFilters", contextThatHasTenant(providerTenantID), expectedFilters).Return(runtimes, nil).Once()
				rsvc.On("GetLabel", contextThatHasTenant(providerTenantID), runtimeID, consumerIDsLabelKey).Return(invalidLabel, nil).Once()
				rsvc.On("GetLabel", contextThatHasTenant(providerTenantID), runtime2ID, consumerIDsLabelKey).Return(invalidLabel, nil).Once()
				return rsvc
			},
			ExpectedErr: fmt.Sprintf("Consumer's external tenant %s was not found in the %s label of any runtime in the provider tenant %s", extTenantID, consumerIDsLabelKey, providerTenantID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			runtimeSvc := &automock.RuntimeService{}
			if testCase.RuntimeServicFn != nil {
				runtimeSvc = testCase.RuntimeServicFn()
			}

			validator := claims.NewValidator(runtimeSvc, providerLabelKey, consumerIDsLabelKey)
			err := validator.Validate(context.TODO(), testCase.Claims)

			if len(testCase.ExpectedErr) > 0 {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr)
			} else {
				require.NoError(t, err)
			}

			runtimeSvc.AssertExpectations(t)
		})
	}
}

func TestScopesValidator_Validate(t *testing.T) {
	t.Run("Succeeds when all claims properties are present", func(t *testing.T) {
		v := claims.NewScopesValidator([]string{"application:read"})
		c := getClaims(tenantID, extTenantID, scopes)

		err := v.Validate(context.TODO(), c)
		assert.NoError(t, err)
	})
	t.Run("Fails when no scopes are present", func(t *testing.T) {
		requiredScopes := []string{"application:read"}
		v := claims.NewScopesValidator(requiredScopes)
		c := getClaims(tenantID, extTenantID, "")

		err := v.Validate(context.TODO(), c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("Not all required scopes %q were found in claim with scopes %q", requiredScopes, c.Scopes))
	})
	t.Run("Fails when inner validation fails", func(t *testing.T) {
		requiredScopes := []string{"application:read"}
		v := claims.NewScopesValidator(requiredScopes)
		c := getClaims(tenantID, extTenantID, scopes)
		c.ExpiresAt = 1

		err := v.Validate(context.TODO(), c)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "while validating claims")
	})
}

func getClaimsForConsumerProviderFlow(consumerTenant, consumerExternalTenant, providerTenant, providerExtTenant, scopes, region, clientID string) claims.Claims {
	return claims.Claims{
		Tenant: map[string]string{
			tenantmapping.ConsumerTenantKey:         consumerTenant,
			tenantmapping.ExternalTenantKey:         consumerExternalTenant,
			tenantmapping.ProviderTenantKey:         providerTenant,
			tenantmapping.ProviderExternalTenantKey: providerExtTenant,
		},
		Scopes:        scopes,
		ConsumerID:    consumerID,
		ConsumerType:  consumer.Runtime,
		OnBehalfOf:    onBehalfOfConsumerID,
		Region:        region,
		TokenClientID: clientID,
	}
}

func getClaims(intTenantID, extTenantID, scopes string) claims.Claims {
	return claims.Claims{
		Tenant: map[string]string{
			tenantmapping.ConsumerTenantKey: intTenantID,
			tenantmapping.ExternalTenantKey: extTenantID,
		},

		Scopes:       scopes,
		ConsumerID:   consumerID,
		ConsumerType: consumer.Runtime,
	}
}

func contextThatHasTenant(expectedTenant string) interface{} {
	return mock.MatchedBy(func(actual context.Context) bool {
		actualTenant, err := tenant.LoadFromContext(actual)
		if err != nil {
			return false
		}
		return actualTenant == expectedTenant
	})
}
