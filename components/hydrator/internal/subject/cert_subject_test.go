package subject_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/certsubjmapping"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/kyma-incubator/compass/components/hydrator/internal/certsubjectmapping/automock"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/hydrator/internal/subject"

	"github.com/stretchr/testify/require"
)

const (
	configTpl = `[{"consumer_type": "%s", "tenant_access_levels": ["%s"], "subject": "%s", "internal_consumer_id": "%s"}]`

	validConsumer             = "Integration System"
	validAccessLvl            = "account"
	validSubjectWithoutRegion = "C=DE, OU=Compass Clients, OU=ed1f789b-1a85-4a63-b360-fac9d6484544, L=validate, CN=test-compass-integration"
	validSubjectWithRegion    = "C=DE, OU=Compass Clients, OU=Region, OU=ed1f789b-1a85-4a63-b360-fac9d6484544, L=validate, CN=test-compass-integration"

	subjectMappingOUSeparatedWithPlus = "CN=test-compass-integration,OU=Compass Clients+OU=ed1f789b-1a85-4a63-b360-fac9d6484544,L=validate,C=DE"
	subjectMappingOUSeparatedWithComa = "CN=test-compass-integration,L=validate,OU=ed1f789b-1a85-4a63-b360-fac9d6484544,OU=Compass Clients,C=DE"
)

var validConfig = fmt.Sprintf(configTpl, validConsumer, validAccessLvl, validSubjectWithoutRegion, validInternalConsumerID)

var (
	ctx = context.Background()

	validSubject            = "C=DE, OU=Compass Clients, OU=Region, OU=ed1f789b-1a85-4a63-b360-fac9d6484544, L=validate, CN=test-compass-integration"
	invalidSubject          = "C=DE, OU=Compass Clients, OU=Region, OU=Random-OU, L=validate, CN=test-compass-integration"
	validConsumerType       = string(consumer.IntegrationSystem)
	validInternalConsumerID = "3bfbb60f-d67d-4657-8f9e-2d73a6b24a10"
	validTntAccessLevels    = []string{string(tenantEntity.Account)}

	validCertSubjectMappings                     = fixCertSubjectMappings(validSubject, validConsumerType, validInternalConsumerID, validTntAccessLevels)
	validCertSubjectMappingsMultipleOU           = fixCertSubjectMappings(subjectMappingOUSeparatedWithPlus, validConsumerType, validInternalConsumerID, validTntAccessLevels)
	validCertSubjectMappingsWithoutRegion        = fixCertSubjectMappings(validSubjectWithoutRegion, validConsumerType, validInternalConsumerID, validTntAccessLevels)
	certSubjectMappingsWithoutInternalConsumerID = fixCertSubjectMappings(validSubject, validConsumerType, "", validTntAccessLevels)
	certSubjectMappingWithNotMatchingSubject     = fixCertSubjectMappings(invalidSubject, validConsumerType, validInternalConsumerID, validTntAccessLevels)
	emptyMappings                                []certsubjmapping.SubjectConsumerTypeMapping
)

func TestNewProcessor(t *testing.T) {
	testCases := []struct {
		name                    string
		certSubjectMappingCache func() *automock.Cache
		expectedErrorMsg        string
	}{
		{
			name: "Success",
			certSubjectMappingCache: func() *automock.Cache {
				cache := &automock.Cache{}
				cache.On("Get").Return(validCertSubjectMappings).Once()
				return cache
			},
		},
	}

	for _, ts := range testCases {
		t.Run(ts.name, func(t *testing.T) {
			cache := ts.certSubjectMappingCache()
			defer mock.AssertExpectationsForObjects(t, cache)
			p := subject.NewProcessor(ctx, cache, "testOUPattern", "testOURegionPattern")

			if len(ts.expectedErrorMsg) > 0 {
				require.Nil(t, p)
			} else {
				require.NotEmpty(t, p)
			}
		})
	}
}

func TestAuthIDFromSubjectFunc(t *testing.T) {
	expectedID := "ed1f789b-1a85-4a63-b360-fac9d6484544"

	t.Run("Success when internal consumer id is provided", func(t *testing.T) {
		cache := &automock.Cache{}
		cache.On("Get").Return(validCertSubjectMappings).Twice()
		defer mock.AssertExpectationsForObjects(t, cache)

		p := subject.NewProcessor(ctx, cache, "", "")

		res := p.AuthIDFromSubjectFunc(ctx)(validSubject)
		require.Equal(t, validInternalConsumerID, res)
	})

	t.Run("Success when internal consumer id is not provided", func(t *testing.T) {
		cache := &automock.Cache{}
		cache.On("Get").Return(certSubjectMappingsWithoutInternalConsumerID).Twice()
		defer mock.AssertExpectationsForObjects(t, cache)

		p := subject.NewProcessor(ctx, cache, "Compass Clients", "")

		res := p.AuthIDFromSubjectFunc(ctx)(validSubjectWithoutRegion)
		require.Equal(t, expectedID, res)
	})

	t.Run("Success getting authID from mapping", func(t *testing.T) {
		cache := &automock.Cache{}
		cache.On("Get").Return(emptyMappings).Twice()
		defer mock.AssertExpectationsForObjects(t, cache)

		p := subject.NewProcessor(ctx, cache, "Compass Clients", "")

		res := p.AuthIDFromSubjectFunc(ctx)(validSubjectWithoutRegion)
		require.Equal(t, expectedID, res)
	})

	t.Run("Success getting authID from OUs when region is missing", func(t *testing.T) {
		cache := &automock.Cache{}
		cache.On("Get").Return(certSubjectMappingWithNotMatchingSubject).Twice()
		defer mock.AssertExpectationsForObjects(t, cache)

		p := subject.NewProcessor(ctx, cache, "Compass Clients", "")

		res := p.AuthIDFromSubjectFunc(ctx)(validSubjectWithoutRegion)
		require.Equal(t, expectedID, res)
	})

	t.Run("Success getting authID from OUs when region exists in subject", func(t *testing.T) {
		cache := &automock.Cache{}
		cache.On("Get").Return(certSubjectMappingWithNotMatchingSubject).Twice()
		defer mock.AssertExpectationsForObjects(t, cache)

		p := subject.NewProcessor(ctx, cache, "Compass Clients", "Region")

		res := p.AuthIDFromSubjectFunc(ctx)(validSubjectWithRegion)
		require.Equal(t, expectedID, res)
	})
}

func TestAuthSessionExtraFromSubjectFunc(t *testing.T) {
	testCases := []struct {
		Name          string
		CertCacheFn   func() *automock.Cache
		Subject       string
		ExpectedExtra map[string]interface{}
	}{
		{
			Name: "Success getting auth session extra",
			CertCacheFn: func() *automock.Cache {
				cache := &automock.Cache{}
				cache.On("Get").Return(validCertSubjectMappingsWithoutRegion).Twice()
				return cache
			},
			Subject: validSubjectWithoutRegion,
			ExpectedExtra: map[string]interface{}{
				"consumer_type":        validConsumer,
				"tenant_access_levels": []string{validAccessLvl},
				"internal_consumer_id": validInternalConsumerID,
			},
		},
		{
			Name: "Success getting auth session extra when OUs are separated with different separators",
			CertCacheFn: func() *automock.Cache {
				cache := &automock.Cache{}
				cache.On("Get").Return(validCertSubjectMappingsMultipleOU).Twice()
				return cache
			},
			Subject: subjectMappingOUSeparatedWithComa,
			ExpectedExtra: map[string]interface{}{
				"consumer_type":        validConsumer,
				"tenant_access_levels": []string{validAccessLvl},
				"internal_consumer_id": validInternalConsumerID,
			},
		},
		{
			Name: "Returns nil when can't match subjects components",
			CertCacheFn: func() *automock.Cache {
				cache := &automock.Cache{}
				cache.On("Get").Return(validCertSubjectMappings).Twice()
				return cache
			},
			Subject:       "C=DE, OU=Compass Clients, OU=Random OU, L=validate, CN=test-compass-integration",
			ExpectedExtra: nil,
		},
		{
			Name: "Returns nil when can't match number of subjects components",
			CertCacheFn: func() *automock.Cache {
				cache := &automock.Cache{}
				cache.On("Get").Return(validCertSubjectMappings).Twice()
				return cache
			},
			Subject:       "C=DE, OU=Compass Clients, L=validate, CN=test-compass-integration",
			ExpectedExtra: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			cache := testCase.CertCacheFn()

			p := subject.NewProcessor(ctx, cache, "", "")

			extra := p.AuthSessionExtraFromSubjectFunc()(ctx, testCase.Subject)
			if testCase.ExpectedExtra != nil {
				require.Equal(t, testCase.ExpectedExtra["consumer_type"], extra["consumer_type"])
				require.Equal(t, testCase.ExpectedExtra["tenant_access_levels"], extra["tenant_access_levels"])
				require.Equal(t, testCase.ExpectedExtra["internal_consumer_id"], extra["internal_consumer_id"])
			} else {
				require.Nil(t, extra)
			}

			mock.AssertExpectationsForObjects(t, cache)
		})
	}
}

func fixCertSubjectMappings(subject, consumerType, internalConsumerID string, tenantAccessLevels []string) []certsubjmapping.SubjectConsumerTypeMapping {
	return []certsubjmapping.SubjectConsumerTypeMapping{
		{
			Subject:            subject,
			ConsumerType:       consumerType,
			InternalConsumerID: internalConsumerID,
			TenantAccessLevels: tenantAccessLevels,
		},
	}
}
