package tenantmapping

import (
	"context"
	"fmt"

	cfg "github.com/kyma-incubator/compass/components/hydrator/internal/config"

	"github.com/davecgh/go-spew/spew"
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
	authID           string
	subdomain        string
	scopes           string
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
	userCtxData, err := c.getUserContextData(ctx, userContextHeader)
	if err != nil {
		return ObjectContext{}, errors.Wrapf(err, "while getting user context data from %q header", oathkeeper.UserContextKey)
	}

	externalTenantID := userCtxData.externalTenantID
	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := c.directorClient.GetTenantByExternalID(ctx, externalTenantID)
	if err != nil {
		if directorErrors.IsGQLNotFoundError(err) {
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(NewTenantContext(externalTenantID, ""), c.tenantKeys, userCtxData.scopes, mergeWithOtherScopes, "", userCtxData.clientID, authDetails.AuthID, authDetails.AuthFlow, consumer.User, tenantmapping.ConsumerProviderObjectContextProvider), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	region, ok := tenantMapping.Labels["region"]
	if !ok {
		return ObjectContext{}, errors.New(fmt.Sprintf("region label not found for tenant with ID: %q", tenantMapping.ID))
	}
	regionStr, ok := region.(string)
	if !ok {
		return ObjectContext{}, errors.New(fmt.Sprintf("unexpected region label type: %T, should be string", region))
	}
	authDetails.Region = regionStr

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.InternalID), c.tenantKeys, userCtxData.scopes, mergeWithOtherScopes, authDetails.Region, userCtxData.clientID, authDetails.AuthID, authDetails.AuthFlow, consumer.User, tenantmapping.ConsumerProviderObjectContextProvider)
	log.C(ctx).Infof("Successfully got object context: %+v", objCtx)

	_, exists := tenantMapping.Labels["subdomain"]
	if !exists {
		log.C(ctx).Warningf("subdomain label not found for tenant with ID: %q", tenantMapping.ID)
		tenantToUpdate := schema.BusinessTenantMappingInput{
			Name:           *tenantMapping.Name,
			ExternalTenant: tenantMapping.ID,
			Type:           tenantMapping.Type,
			Parent:         &tenantMapping.ParentID,
			Region:         &regionStr,
			Subdomain:      &userCtxData.subdomain,
			Provider:       tenantMapping.Provider,
		}

		if err := c.directorClient.WriteTenants(ctx, []schema.BusinessTenantMappingInput{tenantToUpdate}); err != nil {
			log.C(ctx).Errorf("an error occurred while write tenant with external ID: %q: %v", tenantToUpdate.ExternalTenant, err)
			return ObjectContext{}, err
		}
	}

	return objCtx, nil
}

// Match checks if there is "user_context" Header with non-empty value. If so AuthDetails object is build.
func (c *consumerContextProvider) Match(_ context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error) {
	spew.Dump(data) // todo::: remove
	userContextHeader := data.Header.Get(oathkeeper.UserContextKey)
	if userContextHeader == "" {
		return false, nil, apperrors.NewKeyDoesNotExistError(oathkeeper.UserContextKey)
	}

	fmt.Printf("pptt - userContextHeader in Match: %s\n\n", userContextHeader) // todo::: delete

	// todo::: cert check?
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

func (c *consumerContextProvider) getUserContextData(ctx context.Context, userContextHeader string) (*userContextData, error) {
	fmt.Printf("pptt - userContextHeader in getUserContextData: %s\n\n", userContextHeader) // todo::: delete
	clientID := gjson.Get(userContextHeader, c.consumerClaimsKeysConfig.ClientIDKey).String()
	if clientID == "" {
		return &userContextData{}, apperrors.NewInvalidDataError(fmt.Sprintf("property %q could not be empty", c.consumerClaimsKeysConfig.ClientIDKey))
	}

	externalTenantID := gjson.Get(userContextHeader, c.consumerClaimsKeysConfig.TenantIDKey).String()
	if externalTenantID == "" {
		return &userContextData{}, apperrors.NewInvalidDataError(fmt.Sprintf("property %q could not be empty", c.consumerClaimsKeysConfig.TenantIDKey))
	}

	authID := gjson.Get(userContextHeader, c.consumerClaimsKeysConfig.UserNameKey).String()
	if authID == "" {
		return &userContextData{}, apperrors.NewInvalidDataError(fmt.Sprintf("property %q could not be empty", c.consumerClaimsKeysConfig.UserNameKey))
	}

	subdomain := gjson.Get(userContextHeader, c.consumerClaimsKeysConfig.SubdomainKey).String()
	if subdomain == "" {
		return &userContextData{}, apperrors.NewInvalidDataError(fmt.Sprintf("property %q could not be empty", c.consumerClaimsKeysConfig.SubdomainKey))
	}

	scopes := gjson.Get(userContextHeader, c.consumerClaimsKeysConfig.ScopesKey).String()
	if scopes == "" {
		log.C(ctx).Warningf("property %q is empty", c.consumerClaimsKeysConfig.ScopesKey)
	}

	return &userContextData{
		clientID:         clientID,
		externalTenantID: externalTenantID,
		authID:           authID,
		subdomain:        subdomain,
		scopes:           scopes,
	}, nil
}
