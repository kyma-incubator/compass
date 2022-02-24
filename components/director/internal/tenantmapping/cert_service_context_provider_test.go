package tenantmapping_test

import (
	"context"
	"errors"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	oathkeeper2 "github.com/kyma-incubator/compass/components/director/pkg/oathkeeper"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCertServiceContextProvider(t *testing.T) {
	testError := errors.New("test error")
	notFoundErr := apperrors.NewNotFoundErrorWithType(resource.Tenant)

	emptyCtx := context.TODO()
	tenantID := uuid.New().String()
	authDetails := oathkeeper2.AuthDetails{AuthID: tenantID, AuthFlow: oathkeeper2.CertificateFlow, CertIssuer: oathkeeper2.ExternalIssuer}

	scopes := []string{"runtime:read", "runtime:write", "tenant:read"}
	scopesString := "runtime:read runtime:write tenant:read"

	reqData := oathkeeper2.ReqData{}

	internalConsumerID := "123"
	reqDataWithInternalConsumerID := oathkeeper2.ReqData{
		Body: oathkeeper2.ReqBody{Extra: map[string]interface{}{
			cert.InternalConsumerIDField: internalConsumerID,
		}},
	}

	internalSubaccount := "internalSubaccountID"

	testSubaccount := &model.BusinessTenantMapping{
		ID:             internalSubaccount,
		Name:           "testSubaccount",
		ExternalTenant: "externalTestSubaccount",
		Type:           tenantEntity.Subaccount,
	}

	testCases := []struct {
		Name               string
		TenantRepoFn       func() *automock.TenantRepository
		ScopesGetterFn     func() *automock.ScopesGetter
		ReqDataInput       oathkeeper2.ReqData
		AuthDetailsInput   oathkeeper2.AuthDetails
		ExpectedScopes     string
		ExpectedInternalID string
		ExpectedConsumerID string
		ExpectedErr        error
	}{
		{
			Name: "Success when cannot find internal tenant",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, tenantID).Return(nil, notFoundErr).Once()
				return tenantRepo
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:       reqData,
			AuthDetailsInput:   authDetails,
			ExpectedScopes:     scopesString,
			ExpectedInternalID: "",
			ExpectedErr:        nil,
		},
		{
			Name: "Error when the error from getting the internal tenant is different from not found",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, tenantID).Return(nil, testError).Once()
				return tenantRepo
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:       reqData,
			AuthDetailsInput:   authDetails,
			ExpectedScopes:     "",
			ExpectedInternalID: "",
			ExpectedErr:        testError,
		},
		{
			Name: "Success when internal tenant exists",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, tenantID).Return(testSubaccount, nil).Once()
				return tenantRepo
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:       reqData,
			AuthDetailsInput:   authDetails,
			ExpectedScopes:     scopesString,
			ExpectedInternalID: internalSubaccount,
			ExpectedErr:        nil,
		},
		{
			Name: "Success when internal consumer ID is provided",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, tenantID).Return(testSubaccount, nil).Once()
				return tenantRepo
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:       reqDataWithInternalConsumerID,
			AuthDetailsInput:   authDetails,
			ExpectedScopes:     scopesString,
			ExpectedInternalID: internalSubaccount,
			ExpectedConsumerID: internalConsumerID,
			ExpectedErr:        nil,
		},
		{
			Name:         "Error when can't get required scopes",
			TenantRepoFn: unusedTenantRepo,
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(nil, testError)
				return scopesGetter
			},
			ReqDataInput:     reqData,
			AuthDetailsInput: authDetails,
			ExpectedErr:      errors.New("failed to extract scopes"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			tenantRepo := testCase.TenantRepoFn()
			scopesGetter := testCase.ScopesGetterFn()
			provider := tenantmapping.NewCertServiceContextProvider(tenantRepo, scopesGetter)
			if testCase.ExpectedConsumerID == "" {
				testCase.ExpectedConsumerID = tenantID
			}
			// WHEN
			objectCtx, err := provider.GetObjectContext(emptyCtx, testCase.ReqDataInput, testCase.AuthDetailsInput)

			// THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, consumer.Runtime, objectCtx.ConsumerType)
				require.Equal(t, testCase.ExpectedConsumerID, objectCtx.ConsumerID)
				require.Equal(t, testCase.ExpectedInternalID, objectCtx.TenantContext.TenantID)
				require.Equal(t, tenantID, objectCtx.TenantContext.ExternalTenantID)
				require.Equal(t, testCase.ExpectedScopes, objectCtx.Scopes)
			} else {
				require.Error(t, err)
				require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(testCase.ExpectedErr.Error()))
				require.Empty(t, objectCtx)
			}
			mock.AssertExpectationsForObjects(t, scopesGetter, tenantRepo)
		})
	}
}

func TestCertServiceContextProviderMatch(t *testing.T) {
	t.Run("returns ID string and CertificateFlow when a client-id-from-certificate is specified in the Header map of request body", func(t *testing.T) {
		clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
		provider := tenantmapping.NewCertServiceContextProvider(nil, nil)

		reqData := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertKey):    []string{clientID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertIssuer): []string{oathkeeper2.ExternalIssuer},
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper2.CertificateFlow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})

	t.Run("returns nil when does not match", func(t *testing.T) {
		provider := tenantmapping.NewCertServiceContextProvider(nil, nil)

		reqData := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Header: http.Header{},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.NoError(t, err)
	})
}

func unusedTenantRepo() *automock.TenantRepository {
	return &automock.TenantRepository{}
}

func unusedScopesGetter() *automock.ScopesGetter {
	return &automock.ScopesGetter{}
}
