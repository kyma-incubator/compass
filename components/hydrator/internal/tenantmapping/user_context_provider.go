package tenantmapping

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	directorErrors "github.com/kyma-incubator/compass/components/hydrator/internal/director"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	"github.com/pkg/errors"
)

// NewUserContextProvider missing godoc
func NewUserContextProvider(clientProvider DirectorClient, staticGroupRepo StaticGroupRepository) *userContextProvider {
	return &userContextProvider{
		directorClient:  clientProvider,
		staticGroupRepo: staticGroupRepo,
		tenantKeys: KeysExtra{
			TenantKey:         tenantmapping.ConsumerTenantKey,
			ExternalTenantKey: tenantmapping.ExternalTenantKey,
		},
	}
}

type userContextProvider struct {
	directorClient  DirectorClient
	staticGroupRepo StaticGroupRepository
	tenantKeys      KeysExtra
}

// GetObjectContext missing godoc
func (m *userContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	var externalTenantID string
	var err error

	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumer.User,
	})

	ctx = log.ContextWithLogger(ctx, logger)
	log.C(ctx).Info("Getting scopes from groups")

	scopes := m.getScopesForUserGroups(ctx, reqData)

	externalTenantID, err = reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return ObjectContext{}, errors.Wrapf(err, "could not parse external ID for user: %s", authDetails.AuthID)
		}
		log.C(ctx).Warningf("Could not get tenant external id, error: %s", err.Error())

		log.C(ctx).Info("Could not create tenant context, returning empty context...")
		return NewObjectContext(TenantContext{}, m.tenantKeys, scopes, intersectWithOtherScopes, authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.User, tenantmapping.UserObjectContextProvider), nil
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := m.directorClient.GetTenantByExternalID(ctx, externalTenantID)
	if err != nil {
		if directorErrors.IsGQLNotFoundError(err) {
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(NewTenantContext(externalTenantID, ""), m.tenantKeys, scopes, intersectWithOtherScopes, authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.User, tenantmapping.UserObjectContextProvider), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.InternalID), m.tenantKeys, scopes, intersectWithOtherScopes, authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.User, tenantmapping.UserObjectContextProvider)
	log.C(ctx).Infof("Successfully got object context: %+v", objCtx)

	return objCtx, nil
}

func (m *userContextProvider) Match(_ context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error) {
	if usernameVal, ok := data.Body.Extra[oathkeeper.UsernameKey]; ok {
		username, err := str.Cast(usernameVal)
		if err != nil {
			return false, nil, errors.Wrapf(err, "while parsing the value for %s", oathkeeper.UsernameKey)
		}
		return true, &oathkeeper.AuthDetails{AuthID: username, AuthFlow: oathkeeper.JWTAuthFlow}, nil
	}

	return false, nil, nil
}

func (m *userContextProvider) getScopesForUserGroups(ctx context.Context, reqData oathkeeper.ReqData) string {
	userGroups := reqData.GetUserGroups()
	if len(userGroups) == 0 {
		return ""
	}
	log.C(ctx).Debugf("Found user groups: %s", strings.Join(userGroups, " "))

	staticGroups := m.staticGroupRepo.Get(ctx, userGroups)
	if len(staticGroups) == 0 {
		return ""
	}

	scopes := staticGroups.GetGroupScopes()
	log.C(ctx).Debugf("Found scopes: %s", scopes)

	return scopes
}
