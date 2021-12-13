package tenantmapping_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
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
	internalConsumerID := uuid.New().String()
	authDetails := oathkeeper.AuthDetails{AuthID: tenantID, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ExternalIssuer}
	consumerExistsCheckers := map[model.SystemAuthReferenceObjectType]func(ctx context.Context, id string) (bool, error){
		model.IntegrationSystemReference: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}
	scopes := []string{"runtime:read", "runtime:write", "tenant:read"}
	scopesString := "runtime:read runtime:write tenant:read"
	componentHeaderKey := "X-Component-Name"
	directorComponentHeader := map[string][]string{componentHeaderKey: {"director"}}
	directorHeadersWithTenant := map[string][]string{componentHeaderKey: {"director"}, textproto.CanonicalMIMEHeaderKey(oathkeeper.ExternalTenantKey): {tenantID}}

	reqData := oathkeeper.ReqData{
		Body: oathkeeper.ReqBody{Header: directorComponentHeader},
	}
	internalAccount := "internalAccountID"
	internalSubaccount := "internalSubaccountID"

	testSubaccount := &model.BusinessTenantMapping{
		ID:             internalSubaccount,
		Name:           "testSubaccount",
		ExternalTenant: "externalTestSubaccount",
		Type:           tenantEntity.Subaccount,
	}

	testAccount := &model.BusinessTenantMapping{
		ID:             internalAccount,
		Name:           "testAccount",
		ExternalTenant: "externalTestAccount",
		Type:           tenantEntity.Account,
	}

	testCases := []struct {
		Name                   string
		TenantRepoFn           func() *automock.TenantRepository
		ScopesGetterFn         func() *automock.ScopesGetter
		ConsumerExistsCheckers map[model.SystemAuthReferenceObjectType]func(ctx context.Context, id string) (bool, error)
		ReqDataInput           oathkeeper.ReqData
		AuthDetailsInput       oathkeeper.AuthDetails
		ExpectedScopes         string
		ExpectedInternalID     string
		ExpectedErr            error
	}{
		{
			Name: "Success when component is director and cannot find internal tenant",
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
			ConsumerExistsCheckers: consumerExistsCheckers,
			ReqDataInput:           reqData,
			AuthDetailsInput:       authDetails,
			ExpectedScopes:         scopesString,
			ExpectedInternalID:     "",
			ExpectedErr:            nil,
		},
		{
			Name: "Error when component is director and the error is different from not found",
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
			ConsumerExistsCheckers: consumerExistsCheckers,
			ReqDataInput:           reqData,
			AuthDetailsInput:       authDetails,
			ExpectedScopes:         "",
			ExpectedInternalID:     "",
			ExpectedErr:            testError,
		},
		{
			Name: "Success when component is director and internal tenant exists",
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
			ConsumerExistsCheckers: consumerExistsCheckers,
			ReqDataInput:           reqData,
			AuthDetailsInput:       authDetails,
			ExpectedScopes:         scopesString,
			ExpectedInternalID:     internalSubaccount,
			ExpectedErr:            nil,
		},
		{
			Name:         "Error when can't get required scopes",
			TenantRepoFn: unusedTenantRepo,
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(nil, testError)
				return scopesGetter
			},
			ConsumerExistsCheckers: consumerExistsCheckers,
			ReqDataInput:           reqData,
			AuthDetailsInput:       authDetails,
			ExpectedErr:            errors.New("failed to extract scopes"),
		},
		{
			Name: "Error when consumer don't have access",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, tenantID).Return(testAccount, nil).Once()
				return tenantRepo
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil)
				return scopesGetter
			},
			ConsumerExistsCheckers: consumerExistsCheckers,
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Extra: map[string]interface{}{
						cert.AccessLevelsExtraField: []interface{}{tenantEntity.Subaccount},
						cert.ConsumerTypeExtraField: model.IntegrationSystemReference,
					},
					Header: directorHeadersWithTenant,
				},
			},
			AuthDetailsInput: authDetails,
			ExpectedErr:      apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s has no access to %s tenant with ID %s", authDetails.AuthID, testAccount.Type, testAccount.ExternalTenant)),
		},
		{
			Name: "Error when consumer exists check function fails",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, tenantID).Return(testAccount, nil).Once()
				return tenantRepo
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil)
				return scopesGetter
			},
			ConsumerExistsCheckers: map[model.SystemAuthReferenceObjectType]func(ctx context.Context, id string) (bool, error){
				model.IntegrationSystemReference: func(ctx context.Context, id string) (bool, error) {
					return false, testError
				},
			},
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Extra: map[string]interface{}{
						cert.AccessLevelsExtraField:  []interface{}{tenantEntity.Account},
						cert.InternalConsumerIDField: internalConsumerID,
						cert.ConsumerTypeExtraField:  model.IntegrationSystemReference,
					},
					Header: directorHeadersWithTenant,
				},
			},
			AuthDetailsInput: authDetails,
			ExpectedErr:      testError,
		},
		{
			Name: "Error when consumer exists check function can't find the consumer",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, tenantID).Return(testAccount, nil).Once()
				return tenantRepo
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil)
				return scopesGetter
			},
			ConsumerExistsCheckers: map[model.SystemAuthReferenceObjectType]func(ctx context.Context, id string) (bool, error){
				model.IntegrationSystemReference: func(ctx context.Context, id string) (bool, error) {
					return false, nil
				},
			},
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Extra: map[string]interface{}{
						cert.AccessLevelsExtraField:  []interface{}{tenantEntity.Account},
						cert.InternalConsumerIDField: internalConsumerID,
						cert.ConsumerTypeExtraField:  model.IntegrationSystemReference,
					},
					Header: directorHeadersWithTenant,
				},
			},
			AuthDetailsInput: authDetails,
			ExpectedErr:      apperrors.NewUnauthorizedError(fmt.Sprintf("%s with ID %s does not exist", model.IntegrationSystemReference, internalConsumerID)),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			tenantRepo := testCase.TenantRepoFn()
			scopesGetter := testCase.ScopesGetterFn()
			provider := tenantmapping.NewCertServiceContextProvider(tenantRepo, scopesGetter, testCase.ConsumerExistsCheckers)

			// WHEN
			objectCtx, err := provider.GetObjectContext(emptyCtx, testCase.ReqDataInput, testCase.AuthDetailsInput)

			// THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, consumer.Runtime, objectCtx.ConsumerType)
				require.Equal(t, tenantID, objectCtx.ConsumerID)
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
		provider := tenantmapping.NewCertServiceContextProvider(nil, nil, nil)

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{clientID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.CertificateFlow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})

	t.Run("returns nil when does not match", func(t *testing.T) {
		provider := tenantmapping.NewCertServiceContextProvider(nil, nil, nil)

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
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
