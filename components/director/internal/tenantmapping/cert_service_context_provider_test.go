package tenantmapping_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/require"
)

var internalTenant = "internalTenantID"

func TestCertServiceContextProvider(t *testing.T) {
	testError := errors.New("test error")
	notFoundErr := apperrors.NewNotFoundErrorWithType(resource.Tenant)
	emptyCtx := context.TODO()
	subaccount := uuid.New().String()
	authDetails := oathkeeper.AuthDetails{AuthID: subaccount, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ExternalIssuer}
	tenantRepo := &automock.TenantRepository{}
	scopesGetter := &automock.ScopesGetter{}
	consumerExistsCheckers := map[model.SystemAuthReferenceObjectType]func(ctx context.Context, id string) (bool, error){
		model.IntegrationSystemReference: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}
	provider := tenantmapping.NewCertServiceContextProvider(tenantRepo, scopesGetter, consumerExistsCheckers)
	scopes := []string{"runtime:read", "runtime:write", "tenant:read"}
	scopesString := "runtime:read runtime:write tenant:read"
	componentHeaderKey := "X-Component-Name"
	ordComponentHeader := map[string][]string{componentHeaderKey: {"ord"}}
	directorComponentHeader := map[string][]string{componentHeaderKey: {"director"}}
	reqData := oathkeeper.ReqData{Body: oathkeeper.ReqBody{Header: directorComponentHeader}}
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

	testTenant := &model.BusinessTenantMapping{
		ID:             internalTenant,
		Name:           "testTenant",
		ExternalTenant: "externalTestTenant",
		Type:           "subaccount",
	}

	t.Run("Error when there is no matching component name", func(t *testing.T) {
		headers := map[string][]string{"invalidKey": {""}}
		reqData = oathkeeper.ReqData{Body: oathkeeper.ReqBody{Header: headers}}

		objectCtx, err := provider.GetObjectContext(context.TODO(), reqData, authDetails)
		require.Error(t, err)
		require.Contains(t, err.Error(), "empty matched component header")
		require.Empty(t, objectCtx)

		tenantRepo.AssertExpectations(t)
	})

	t.Run("Success when component is director and cannot find internal tenant", func(t *testing.T) {
		reqData = oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: subaccount,
				},
				Header: map[string][]string{
					componentHeaderKey: {"director"},
				},
			}}

		tenantRepo.On("GetByExternalTenant", mock.Anything, subaccount).Return(nil, notFoundErr).Once()
		scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(scopes, nil)
		provider := tenantmapping.NewCertServiceContextProvider(tenantRepo, scopesGetter, consumerExistsCheckers)

		objectCtx, err := provider.GetObjectContext(emptyCtx, reqData, authDetails)
		require.NoError(t, err)
		require.Empty(t, objectCtx.TenantContext.TenantID)
		require.Equal(t, subaccount, objectCtx.TenantContext.ExternalTenantID)
		require.Equal(t, subaccount, objectCtx.ConsumerID)
		require.Equal(t, scopesString, objectCtx.Scopes)

		tenantRepo.AssertExpectations(t)
	})

	t.Run("Error when component is director and the error is different from not found", func(t *testing.T) {
		reqData = oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: subaccount,
				},
				Header: map[string][]string{
					componentHeaderKey: {"director"},
				},
			}}

		tenantRepo.On("GetByExternalTenant", mock.Anything, subaccount).Return(nil, testError).Once()
		scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(scopes, nil)
		provider := tenantmapping.NewCertServiceContextProvider(tenantRepo, scopesGetter, consumerExistsCheckers)

		objectCtx, err := provider.GetObjectContext(emptyCtx, reqData, authDetails)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("while getting external tenant mapping [ExternalTenantID=%s]", subaccount))
		require.Empty(t, objectCtx)

		tenantRepo.AssertExpectations(t)
	})

	t.Run("Success when component is director and internal tenant exists", func(t *testing.T) {
		reqData = oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: subaccount,
				},
				Header: map[string][]string{
					componentHeaderKey: {"director"},
				},
			}}

		tenantRepo.On("GetByExternalTenant", mock.Anything, subaccount).Return(testTenant, nil).Once()
		scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(scopes, nil)
		provider := tenantmapping.NewCertServiceContextProvider(tenantRepo, scopesGetter, consumerExistsCheckers)

		objectCtx, err := provider.GetObjectContext(emptyCtx, reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, consumer.Runtime, objectCtx.ConsumerType)
		require.Equal(t, subaccount, objectCtx.ConsumerID)
		require.Equal(t, internalTenant, objectCtx.TenantContext.TenantID)
		require.Equal(t, subaccount, objectCtx.TenantContext.ExternalTenantID)
		require.Equal(t, scopesString, objectCtx.Scopes)

		tenantRepo.AssertExpectations(t)
	})

	t.Run("Error when can't get required scopes", func(t *testing.T) {
		reqData = oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ExternalTenantKey: subaccount,
				},
				Header: map[string][]string{
					componentHeaderKey: {"director"},
				},
			}}
		tenantRepo := &automock.TenantRepository{}
		scopesGetter := &automock.ScopesGetter{}

		scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.runtime").Return(nil, testError)
		provider := tenantmapping.NewCertServiceContextProvider(tenantRepo, scopesGetter, consumerExistsCheckers)

		objectCtx, err := provider.GetObjectContext(emptyCtx, reqData, authDetails)

		require.Error(t, err)
		require.Equal(t, tenantmapping.ObjectContext{}, objectCtx)

		tenantRepo.AssertExpectations(t)
	})

	t.Run("Error when can't extract external tenant id", func(t *testing.T) {
		reqData = oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: map[string][]string{
					componentHeaderKey: {"director"},
				},
			}}

		scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.default").Return(nil, testError)
		provider := tenantmapping.NewCertServiceContextProvider(tenantRepo, scopesGetter, consumerExistsCheckers)

		objectCtx, err := provider.GetObjectContext(emptyCtx, reqData, authDetails)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to extract scopes")
		require.Equal(t, tenantmapping.ObjectContext{}, objectCtx)

		tenantRepo.AssertExpectations(t)
	})

	t.Run("Success when component is ord", func(t *testing.T) {
		reqData = oathkeeper.ReqData{Body: oathkeeper.ReqBody{Header: ordComponentHeader}}

		objectCtx, err := provider.GetObjectContext(emptyCtx, reqData, authDetails)

		require.NoError(t, err)
		require.Equal(t, consumer.Runtime, objectCtx.ConsumerType)
		require.Equal(t, subaccount, objectCtx.ConsumerID)
		require.Equal(t, subaccount, objectCtx.TenantContext.TenantID)
		require.Equal(t, subaccount, objectCtx.TenantContext.ExternalTenantID)
		require.Empty(t, objectCtx.Scopes)

		tenantRepo.AssertExpectations(t)
	})
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
