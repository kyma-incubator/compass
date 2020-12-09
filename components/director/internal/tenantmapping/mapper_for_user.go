package tenantmapping

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	"github.com/tidwall/gjson"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

func NewMapperForUser(authenticators []authenticator.Config, staticUserRepo StaticUserRepository, staticGroupRepo StaticGroupRepository, tenantRepo TenantRepository) *mapperForUser {
	return &mapperForUser{
		authenticators:  authenticators,
		staticUserRepo:  staticUserRepo,
		staticGroupRepo: staticGroupRepo,
		tenantRepo:      tenantRepo,
	}
}

type mapperForUser struct {
	authenticators  []authenticator.Config
	staticUserRepo  StaticUserRepository
	staticGroupRepo StaticGroupRepository
	tenantRepo      TenantRepository
}

func (m *mapperForUser) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, username string) (ObjectContext, error) {
	var externalTenantID, scopes string
	var staticUser *StaticUser
	var err error

	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumer.User,
	})

	ctx = log.ContextWithLogger(ctx, logger)

	extra, err := json.Marshal(reqData.Body.Extra)
	if err != nil {
		return ObjectContext{}, err
	}

	tknAttributes := string(extra)
	authenticatorMatched := false
	for _, authn := range m.authenticators {
		uniqueAttribute := gjson.Get(tknAttributes, authn.Attributes.UniqueAttribute.Key).String()
		if uniqueAttribute == "" || uniqueAttribute != authn.Attributes.UniqueAttribute.Value {
			log.C(ctx).Debugf("Request token does not match %q authenticator", authn.Name)
			continue
		}

		scopes = strings.Join(reqData.GetUserScopes(), " ")

		externalTenantID = gjson.Get(tknAttributes, authn.Attributes.TenantAttribute.Key).String()
		if externalTenantID == "" {
			return ObjectContext{}, errors.Errorf("tenant attribute %q missing from %s authenticator token", authn.Attributes.TenantAttribute.Key, authn.Name)
		}

		log.C(ctx).Infof("Matched %q authenticator with the incoming request data", authn.Name)
		authenticatorMatched = true
		break
	}

	if !authenticatorMatched {
		log.C(ctx).Info("Getting scopes from groups")
		scopes = m.getScopesForUserGroups(ctx, reqData)
		if !hasScopes(scopes) {
			log.C(ctx).Info("No scopes found from groups, getting user data")

			staticUser, scopes, err = m.getUserData(ctx, reqData, username)
			if err != nil {
				return ObjectContext{}, errors.Wrapf(err, "while getting user data for user: %s", username)
			}
		}

		externalTenantID, err = reqData.GetExternalTenantID()
		if err != nil {
			if !apperrors.IsKeyDoesNotExist(err) {
				return ObjectContext{}, errors.Wrapf(err, "could not parse external ID for user: %s", username)
			}
			log.C(ctx).Warningf("Could not get tenant external id, error: %s", err.Error())

			log.C(ctx).Info("Could not create tenant context, returning empty context...")
			return NewObjectContext(TenantContext{}, scopes, username, consumer.User), nil
		}
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(NewTenantContext(externalTenantID, ""), scopes, username, consumer.User), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	if staticUser != nil && !hasValidTenant(staticUser.Tenants, tenantMapping.ExternalTenant) {
		return ObjectContext{}, apperrors.NewInternalError(fmt.Sprintf("Static tenant with username: %s missmatch external tenant: %s", staticUser.Username, tenantMapping.ExternalTenant))
	}

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), scopes, username, consumer.User)
	log.C(ctx).Infof("Successfully got object context: %+v", objCtx)

	return objCtx, nil
}

func (m *mapperForUser) getScopesForUserGroups(ctx context.Context, reqData oathkeeper.ReqData) string {
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

func (m *mapperForUser) getUserData(ctx context.Context, reqData oathkeeper.ReqData, username string) (*StaticUser, string, error) {
	staticUser, err := m.staticUserRepo.Get(username)
	if err != nil {
		return nil, "", errors.Wrapf(err, "while searching for a static user with username %s", username)
	}
	log.C(ctx).Debugf("Found static user with name %s and tenants: %s", staticUser.Username, staticUser.Tenants)

	scopes, err := reqData.GetScopes()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return nil, "", errors.Wrap(err, "while fetching scopes")
		}
		scopes = strings.Join(staticUser.Scopes, " ")
	}
	log.C(ctx).Debugf("Found scopes: %s", scopes)

	return &staticUser, scopes, nil
}

func hasValidTenant(assignedTenants []string, tenant string) bool {
	for _, assignedTenant := range assignedTenants {
		if assignedTenant == tenant {
			return true
		}
	}

	return false
}

func hasScopes(scopes string) bool {
	return len(scopes) > 0
}
