package tenantmapping_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCertServiceContextProvider(t *testing.T) {
	testError := errors.New("test error")
	notFoundErr := apperrors.NewNotFoundErrorWithType(resource.Tenant)
	emptyCtx := context.TODO()
	subaccount := uuid.New().String()
	authDetails := oathkeeper.AuthDetails{AuthID: subaccount, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ExternalIssuer}
	consumerExistsCheckers := map[model.SystemAuthReferenceObjectType]func(ctx context.Context, id string) (bool, error){
		model.IntegrationSystemReference: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}
	scopes := []string{"runtime:read", "runtime:write", "tenant:read"}
	scopesString := "runtime:read runtime:write tenant:read"
	componentHeaderKey := "X-Component-Name"
	ordComponentHeader := map[string][]string{componentHeaderKey: {"ord"}}
	directorComponentHeader := map[string][]string{componentHeaderKey: {"director"}}
	reqData := oathkeeper.ReqData{
		Body: oathkeeper.ReqBody{
			Extra: map[string]interface{}{
				oathkeeper.ExternalTenantKey: subaccount,
			},
			Header: directorComponentHeader,
		},
	}
	//reqDataWithExtra := oathkeeper.ReqData{
	//	Body: oathkeeper.ReqBody{
	//		Header: directorComponentHeader,
	//		Extra: map[string]interface{}{
	//			"tenant":                     subaccount,
	//			cert.ConsumerTypeExtraField:  consumer.Runtime,
	//			cert.InternalConsumerIDField: "test_internal_consumer_id",
	//			cert.AccessLevelExtraField:   "test_access_level",
	//		},
	//	},
	//}
	internalAccount := "internalAccountID"
	internalSubaccount := "internalSubaccountID"

	testSubaccount := &model.BusinessTenantMapping{
		ID:             internalSubaccount,
		Name:           "testSubaccount",
		ExternalTenant: "externalTestSubaccount",
		Type:           "subaccount",
	}

	testAccount := &model.BusinessTenantMapping{
		ID:             internalAccount,
		Name:           "testAccount",
		ExternalTenant: "externalTestAccount",
		Type:           "account",
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
				tenantRepo.On("GetByExternalTenant", mock.Anything, subaccount).Return(nil, notFoundErr).Once()
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
				tenantRepo.On("GetByExternalTenant", mock.Anything, subaccount).Return(nil, testError).Once()
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
				tenantRepo.On("GetByExternalTenant", mock.Anything, subaccount).Return(testSubaccount, nil).Once()
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
			Name:                   "Error when there is no matching component name",
			TenantRepoFn:           unusedTenantRepo,
			ScopesGetterFn:         unusedScopesGetter,
			ConsumerExistsCheckers: consumerExistsCheckers,
			ReqDataInput:           oathkeeper.ReqData{Body: oathkeeper.ReqBody{Header: map[string][]string{"invalidKey": {""}}}},
			AuthDetailsInput:       authDetails,
			ExpectedScopes:         "",
			ExpectedInternalID:     subaccount,
			ExpectedErr:            errors.New("empty matched component header"),
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
			Name:         "Error when can't extract external tenant id",
			TenantRepoFn: unusedTenantRepo,
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.default").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Header: map[string][]string{
						componentHeaderKey: {"director"},
					},
				}},
			ConsumerExistsCheckers: consumerExistsCheckers,
			AuthDetailsInput:       authDetails,
			ExpectedErr:            errors.New("failed to extract external tenant"),
		},
		{
			Name: "Error when consumer don't have access",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, subaccount).Return(testAccount, nil).Once()
				return tenantRepo
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(scopes, nil)
				return scopesGetter
			},
			ConsumerExistsCheckers: consumerExistsCheckers,
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Extra: map[string]interface{}{
						oathkeeper.ExternalTenantKey: subaccount,
						cert.AccessLevelExtraField:   "subaccount",
					},
					Header: directorComponentHeader,
				},
			},
			AuthDetailsInput: authDetails,
			ExpectedErr:      apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s has no access to tenant with ID %s", authDetails.AuthID, testAccount.ExternalTenant)),
		},
		{
			Name: "Error when consumer exists check function fails",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, subaccount).Return(testAccount, nil).Once()
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
						oathkeeper.ExternalTenantKey: subaccount,
						cert.InternalConsumerIDField: subaccount,
						cert.ConsumerTypeExtraField:  model.IntegrationSystemReference,
					},
					Header: directorComponentHeader,
				},
			},
			AuthDetailsInput: authDetails,
			ExpectedErr:      testError,
		},
		{
			Name: "Error when consumer exists check function can't find the consumer",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, subaccount).Return(testAccount, nil).Once()
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
						oathkeeper.ExternalTenantKey: subaccount,
						cert.InternalConsumerIDField: subaccount,
						cert.ConsumerTypeExtraField:  model.IntegrationSystemReference,
					},
					Header: directorComponentHeader,
				},
			},
			AuthDetailsInput: authDetails,
			ExpectedErr:      apperrors.NewUnauthorizedError(fmt.Sprintf("%s with ID %s does not exist", model.IntegrationSystemReference, subaccount)),
		},
		{
			Name:               "Success when component is ord",
			TenantRepoFn:       unusedTenantRepo,
			ScopesGetterFn:     unusedScopesGetter,
			ReqDataInput:       oathkeeper.ReqData{Body: oathkeeper.ReqBody{Header: ordComponentHeader}},
			AuthDetailsInput:   authDetails,
			ExpectedScopes:     "",
			ExpectedInternalID: subaccount,
			ExpectedErr:        nil,
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
				require.Equal(t, subaccount, objectCtx.ConsumerID)
				require.Equal(t, testCase.ExpectedInternalID, objectCtx.TenantContext.TenantID)
				require.Equal(t, subaccount, objectCtx.TenantContext.ExternalTenantID)
				require.Equal(t, testCase.ExpectedScopes, objectCtx.Scopes)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErr.Error())
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
