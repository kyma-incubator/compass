package tenantmapping

import (
	"context"
	"fmt"

	cfg "github.com/kyma-incubator/compass/components/hydrator/internal/config"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	directorErrors "github.com/kyma-incubator/compass/components/hydrator/internal/director"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type userContextData struct {
	clientID         string
	externalTenantID string
	subdomain        string
}

type consumerContextProvider struct {
	directorClient           DirectorClient
	tenantKeys               KeysExtra
	consumerClaimsKeysConfig cfg.ConsumerClaimsKeysConfig
}

// NewConsumerContextProvider implements the ObjectContextProvider interface by looking for "user_context" header from the request.
func NewConsumerContextProvider(clientProvider DirectorClient, consumerClaimsKeysConfig cfg.ConsumerClaimsKeysConfig) *consumerContextProvider {
	return &consumerContextProvider{
		directorClient: clientProvider,
		tenantKeys: KeysExtra{
			TenantKey:         tenantmapping.ConsumerTenantKey,
			ExternalTenantKey: tenantmapping.ExternalTenantKey,
		},
		consumerClaimsKeysConfig: consumerClaimsKeysConfig,
	}
}

// GetObjectContext is the consumerContextProvider implementation of the ObjectContextProvider interface.
// From the information provided in the "user_context" header, it builds a ObjectContext with it.
// In that header we have claims from which we extract the necessary information, there is NO JWT token and signature validation.
func (c *consumerContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	userContextHeader := reqData.Header.Get(oathkeeper.UserContextKey)
	userCtxData, err := c.getUserContextData(userContextHeader)
	if err != nil {
		return ObjectContext{}, errors.Wrapf(err, "while getting user context data from %q header", oathkeeper.UserContextKey)
	}

	externalTenantID := userCtxData.externalTenantID
	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, region, err := getTenantWithRegion(ctx, c.directorClient, externalTenantID)
	if err != nil {
		if directorErrors.IsGQLNotFoundError(err) {
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(NewTenantContext(externalTenantID, ""), c.tenantKeys, "", mergeWithOtherScopes, "", userCtxData.clientID, authDetails.AuthID, authDetails.AuthFlow, consumer.User, tenantmapping.ConsumerProviderObjectContextProvider), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	authDetails.Region = region

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.InternalID), c.tenantKeys, "", mergeWithOtherScopes, authDetails.Region, userCtxData.clientID, authDetails.AuthID, authDetails.AuthFlow, consumer.User, tenantmapping.ConsumerProviderObjectContextProvider)
	log.C(ctx).Infof("Successfully got object context: %+v", objCtx)

	_, exists := tenantMapping.Labels["subdomain"]
	if !exists {
		log.C(ctx).Warningf("subdomain label not found for tenant with ID: %q", tenantMapping.ID)
		tenantToUpdate := schema.BusinessTenantMappingInput{
			Name:           *tenantMapping.Name,
			ExternalTenant: tenantMapping.ID,
			Type:           tenantMapping.Type,
			Parent:         &tenantMapping.ParentID,
			Region:         &region,
			Subdomain:      &userCtxData.subdomain,
			Provider:       tenantMapping.Provider,
		}

		if err := c.directorClient.WriteTenants(ctx, []schema.BusinessTenantMappingInput{tenantToUpdate}); err != nil {
			log.C(ctx).Errorf("an error occurred while write tenant with external ID: %q: %v", tenantToUpdate.ExternalTenant, err)
			return ObjectContext{}, errors.Wrapf(err, "an error occurred while write tenant with external ID: %q", tenantToUpdate.ExternalTenant)
		}
	}

	return objCtx, nil
}

// Match checks if there is "user_context" Header with non-empty value. If so AuthDetails object is build.
func (c *consumerContextProvider) Match(_ context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error) {
	userContextHeader := data.Header.Get(oathkeeper.UserContextKey)
	if userContextHeader == "" {
		return false, nil, apperrors.NewKeyDoesNotExistError(oathkeeper.UserContextKey)
	}

	idVal := data.Body.Header.Get(oathkeeper.ClientIDCertKey)
	certIssuer := data.Body.Header.Get(oathkeeper.ClientIDCertIssuer)

	if idVal == "" || certIssuer != oathkeeper.ExternalIssuer {
		return false, nil, nil
	}

	authID := gjson.Get(userContextHeader, c.consumerClaimsKeysConfig.UserNameKey).String()
	if authID == "" {
		return false, nil, apperrors.NewInvalidDataError(fmt.Sprintf("could not find %s property", c.consumerClaimsKeysConfig.UserNameKey))
	}

	return true, &oathkeeper.AuthDetails{AuthID: authID, AuthFlow: oathkeeper.ConsumerProviderFlow}, nil
}

func (c *consumerContextProvider) getUserContextData(userContextHeader string) (*userContextData, error) {
	clientID := gjson.Get(userContextHeader, c.consumerClaimsKeysConfig.ClientIDKey)
	if !clientID.Exists() {
		return &userContextData{}, apperrors.NewInvalidDataError(fmt.Sprintf("property %q is mandatory", c.consumerClaimsKeysConfig.ClientIDKey))
	}

	externalTenantID := gjson.Get(userContextHeader, c.consumerClaimsKeysConfig.TenantIDKey)
	if !externalTenantID.Exists() {
		return &userContextData{}, apperrors.NewInvalidDataError(fmt.Sprintf("property %q is mandatory", c.consumerClaimsKeysConfig.TenantIDKey))
	}

	subdomain := gjson.Get(userContextHeader, c.consumerClaimsKeysConfig.SubdomainKey)
	if !subdomain.Exists() {
		return &userContextData{}, apperrors.NewInvalidDataError(fmt.Sprintf("property %q is mandatory", c.consumerClaimsKeysConfig.SubdomainKey))
	}

	return &userContextData{
		clientID:         clientID.String(),
		externalTenantID: externalTenantID.String(),
		subdomain:        subdomain.String(),
	}, nil
}

func getTenantWithRegion(ctx context.Context, directorClient DirectorClient, externalTenantID string) (*schema.Tenant, string, error) {
	tenantMapping, err := directorClient.GetTenantByExternalID(ctx, externalTenantID)
	if err != nil {
		return nil, "", err
	}

	region, ok := tenantMapping.Labels["region"]
	if !ok {
		return nil, "", fmt.Errorf("region label not found for tenant with ID: %q", externalTenantID)
	}
	regionStr, ok := region.(string)
	if !ok {
		return nil, "", errors.New(fmt.Sprintf("unexpected region label type: %T, should be string", region))
	}

	return tenantMapping, regionStr, nil
}
