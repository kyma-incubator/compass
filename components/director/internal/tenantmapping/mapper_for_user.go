package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

func NewMapperForUser(staticUserRepo StaticUserRepository, staticGroupRepo StaticGroupRepository, tenantRepo TenantRepository) *mapperForUser {
	return &mapperForUser{
		staticUserRepo:  staticUserRepo,
		staticGroupRepo: staticGroupRepo,
		tenantRepo:      tenantRepo,
	}
}

type mapperForUser struct {
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

	logger.Info("Getting scopes from groups")
	scopes = m.getScopesForUserGroups(reqData, logger)
	if !hasScopes(scopes) {
		logger.Info("No scopes found from groups, getting user data")

		staticUser, scopes, err = m.getUserData(reqData, username, logger)
		if err != nil {
			return ObjectContext{}, errors.Wrapf(err, "while getting user data for user: %s", username)
		}
	}

	externalTenantID, err = reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return ObjectContext{}, errors.Wrapf(err, "could not parse external ID for user: %s", username)
		}
		logger.Warningf("Could not get tenant external id, error: %s", err.Error())

		logger.Info("Could not create tenant context, returning empty context...")
		return NewObjectContext(TenantContext{}, scopes, username, consumer.User), nil
	}

	logger.Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			logger.Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())

			logger.Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(NewTenantContext(externalTenantID, ""), scopes, username, consumer.User), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	if staticUser != nil && !hasValidTenant(staticUser.Tenants, tenantMapping.ExternalTenant) {
		return ObjectContext{}, apperrors.NewInternalError(fmt.Sprintf("Static tenant with username: %s missmatch external tenant: %s", staticUser.Username, tenantMapping.ExternalTenant))
	}

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), scopes, username, consumer.User)
	logger.Infof("Successfully got object context: %+v", objCtx)

	return objCtx, nil
}

func (m *mapperForUser) getScopesForUserGroups(reqData oathkeeper.ReqData, logger *logrus.Entry) string {
	userGroups := reqData.GetUserGroups()
	if len(userGroups) == 0 {
		return ""
	}
	logger.Debugf("Found user groups: %s", strings.Join(userGroups, " "))

	staticGroups := m.staticGroupRepo.Get(userGroups, logger)
	if len(staticGroups) == 0 {
		return ""
	}

	scopes := staticGroups.GetGroupScopes()
	logger.Debugf("Found scopes: %s", scopes)

	return scopes
}

func (m *mapperForUser) getUserData(reqData oathkeeper.ReqData, username string, logger *logrus.Entry) (*StaticUser, string, error) {
	staticUser, err := m.staticUserRepo.Get(username)
	if err != nil {
		return nil, "", errors.Wrapf(err, "while searching for a static user with username %s", username)
	}
	logger.Debugf("Found static user with name %s and tenants: %s", staticUser.Username, staticUser.Tenants)

	scopes, err := reqData.GetScopes()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return nil, "", errors.Wrap(err, "while fetching scopes")
		}
		scopes = strings.Join(staticUser.Scopes, " ")
	}
	logger.Debugf("Found scopes: %s", scopes)

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
