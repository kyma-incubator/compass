package certsubjectmapping_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/certsubjmapping"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/hydrator/internal/certsubjectmapping"
	"github.com/kyma-incubator/compass/components/hydrator/internal/certsubjectmapping/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	validSubject         = "C=DE, OU=Compass Clients, OU=Region, OU=ed1f789b-1a85-4a63-b360-fac9d6484544, L=unit-tests, CN=unit-test-compass"
	validConsumerType    = consumer.Runtime
	validTntAccessLevels = []string{inputvalidation.GlobalAccessLevel}
)

func TestNewCertSubjectMappingLoader(t *testing.T) {
	testID := "testID"
	internalConsumerID := "internalConsumerID"
	endCursor := "endCursor"
	envMappings := "[{\"consumer_type\":\"Runtime\",\"tenant_access_levels\":[\"account\"],\"subject\":\"C=DE,O=SAP SE,OU=SAP Cloud Platform Clients,OU=unit-test-ou,L=unit-test-locality,CN=unit-test-cn\"}]"
	testErr := errors.New("cert-subject-mapping-test-error")

	cfg := certsubjectmapping.Config{
		ResyncInterval:      100 * time.Millisecond,
		EnvironmentMappings: envMappings,
	}

	certSubjectMapping := &graphql.CertificateSubjectMapping{
		ID:                 testID,
		Subject:            validSubject,
		ConsumerType:       string(validConsumerType),
		InternalConsumerID: &internalConsumerID,
		TenantAccessLevels: validTntAccessLevels,
	}

	certSubjectMappingPageWithoutNextPage := &graphql.CertificateSubjectMappingPage{
		Data: []*graphql.CertificateSubjectMapping{certSubjectMapping},
		PageInfo: &graphql.PageInfo{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
		TotalCount: 1,
	}

	certSubjectMappingPageWithoutPageInfo := &graphql.CertificateSubjectMappingPage{
		Data:       []*graphql.CertificateSubjectMapping{certSubjectMapping},
		TotalCount: 1,
	}

	certSubjectMappingWithNextPage := &graphql.CertificateSubjectMappingPage{
		Data: []*graphql.CertificateSubjectMapping{certSubjectMapping, certSubjectMapping},
		PageInfo: &graphql.PageInfo{
			StartCursor: "",
			EndCursor:   graphql.PageCursor(endCursor),
			HasNextPage: true,
		},
		TotalCount: 2,
	}

	testCases := []struct {
		name                            string
		certSubjectMappingCfg           certsubjectmapping.Config
		directorClientFn                func() *automock.DirectorClient
		eventualTickInterval            time.Duration
		expectedCertSubjectMappingCount int
	}{
		{
			name:                  "Successfully resync certificate subject mappings",
			certSubjectMappingCfg: cfg,
			directorClientFn: func() *automock.DirectorClient {
				directorClient := &automock.DirectorClient{}
				directorClient.On("ListCertificateSubjectMappings", ctxWithCorrelationIDMatcher(), "").Return(certSubjectMappingPageWithoutNextPage, nil).Twice()
				return directorClient
			},
			eventualTickInterval:            100 * time.Millisecond,
			expectedCertSubjectMappingCount: 2,
		},
		{
			name:                  "Successfully resync certificate subject mappings with paging",
			certSubjectMappingCfg: cfg,
			directorClientFn: func() *automock.DirectorClient {
				directorClient := &automock.DirectorClient{}
				directorClient.On("ListCertificateSubjectMappings", ctxWithCorrelationIDMatcher(), "").Return(certSubjectMappingWithNextPage, nil)
				directorClient.On("ListCertificateSubjectMappings", ctxWithCorrelationIDMatcher(), endCursor).Return(certSubjectMappingPageWithoutNextPage, nil)
				return directorClient
			},
			eventualTickInterval:            30 * time.Millisecond,
			expectedCertSubjectMappingCount: 4,
		},
		{
			name:                  "Error when the list of certificate subject mappings returns neither result nor error and the second one succeeds",
			certSubjectMappingCfg: cfg,
			directorClientFn: func() *automock.DirectorClient {
				directorClient := &automock.DirectorClient{}
				directorClient.On("ListCertificateSubjectMappings", ctxWithCorrelationIDMatcher(), "").Return(nil, nil).Once()
				directorClient.On("ListCertificateSubjectMappings", ctxWithCorrelationIDMatcher(), "").Return(certSubjectMappingPageWithoutNextPage, nil)
				return directorClient
			},
			eventualTickInterval:            100 * time.Millisecond,
			expectedCertSubjectMappingCount: 2,
		},
		{
			name:                  "Error when the list of certificate subject mappings response doesn't have page info and the second request succeeds",
			certSubjectMappingCfg: cfg,
			directorClientFn: func() *automock.DirectorClient {
				directorClient := &automock.DirectorClient{}
				directorClient.On("ListCertificateSubjectMappings", ctxWithCorrelationIDMatcher(), "").Return(certSubjectMappingPageWithoutPageInfo, nil).Once()
				directorClient.On("ListCertificateSubjectMappings", ctxWithCorrelationIDMatcher(), "").Return(certSubjectMappingPageWithoutNextPage, nil)
				return directorClient
			},
			eventualTickInterval:            100 * time.Millisecond,
			expectedCertSubjectMappingCount: 2,
		},
		{
			name:                  "Error when the first list of certificate subject mappings fails and the second one succeeds",
			certSubjectMappingCfg: cfg,
			directorClientFn: func() *automock.DirectorClient {
				directorClient := &automock.DirectorClient{}
				directorClient.On("ListCertificateSubjectMappings", ctxWithCorrelationIDMatcher(), "").Return(nil, testErr).Once()
				directorClient.On("ListCertificateSubjectMappings", ctxWithCorrelationIDMatcher(), "").Return(certSubjectMappingPageWithoutNextPage, nil)
				return directorClient
			},
			eventualTickInterval:            100 * time.Millisecond,
			expectedCertSubjectMappingCount: 2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// GIVEN
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			directorClient := testCase.directorClientFn()
			defer mock.AssertExpectationsForObjects(t, directorClient)

			// WHEN
			cache, _ := certsubjectmapping.StartCertSubjectMappingLoader(ctx, testCase.certSubjectMappingCfg, directorClient)

			// THEN
			assert.Eventually(t, func() bool {
				cacheMappings := cache.Get()
				if testCase.expectedCertSubjectMappingCount > 0 {
					assert.NotEmpty(t, cacheMappings)
				} else {
					assert.Empty(t, cacheMappings)
				}
				assert.Len(t, cacheMappings, testCase.expectedCertSubjectMappingCount)
				return true
			}, time.Second, testCase.eventualTickInterval)
			cancel()
			assert.Eventually(t, func() bool {
				<-ctx.Done()
				return true
			}, time.Second, 50*time.Millisecond)
		})
	}
}

func ctxWithCorrelationIDMatcher() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		requestID, ok := log.C(ctx).Data[log.FieldRequestID]
		return ok == true && (requestID == certsubjectmapping.CertSubjectMappingLoaderCorrelationID || requestID == certsubjectmapping.CertSubjectMappingInitialLoaderCorrelationID)
	})
}

func TestSubjectConsumerTypeMapping_Validate(t *testing.T) {
	invalidConsumerType := "invalidConsumerType"
	invalidTntAccessLevels := []string{"invalidAccessLevel"}

	testCases := []struct {
		name           string
		input          certsubjmapping.SubjectConsumerTypeMapping
		expectedErrMsg string
	}{
		{
			name: "Success",
			input: certsubjmapping.SubjectConsumerTypeMapping{
				Subject:            validSubject,
				ConsumerType:       string(validConsumerType),
				TenantAccessLevels: validTntAccessLevels,
			},
		},
		{
			name:           "Error when the subject is invalid",
			input:          certsubjmapping.SubjectConsumerTypeMapping{},
			expectedErrMsg: "subject is not provided",
		},
		{
			name: "Error when the consumer type is unsupported",
			input: certsubjmapping.SubjectConsumerTypeMapping{
				Subject:      validSubject,
				ConsumerType: invalidConsumerType,
			},
			expectedErrMsg: fmt.Sprintf("consumer type %s is not valid", invalidConsumerType),
		},
		{
			name: "Error when the tenant access levels are unsupported",
			input: certsubjmapping.SubjectConsumerTypeMapping{
				Subject:            validSubject,
				ConsumerType:       string(validConsumerType),
				TenantAccessLevels: invalidTntAccessLevels,
			},
			expectedErrMsg: fmt.Sprintf("tenant access level %s is not valid", invalidTntAccessLevels[0]),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.input.Validate()
			if testCase.expectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
