package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
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

// getGroupScopes get all scopes from group array, without duplicates
func (groups StaticGroups) getGroupScopes() string {
	scopeMap := make(map[string]bool)
	filteredScopes := []string{}

	for _, group := range groups {
		for _, scope := range group.Scopes {
			_, ok := scopeMap[scope]
			if !ok {
				filteredScopes = append(filteredScopes, scope)
				scopeMap[scope] = true
			}
		}
	}

	return strings.Join(filteredScopes, " ")
}

// getGroupData get scopes and username from group
// func (m *mapperForUser) getGroupData(reqData ReqData) (scopes string, proceedWithUser bool) {
// 	userGroups := reqData.GetUserGroups()

// 	if len(userGroups) == 0 {
// 		return "", true
// 	}

// 	staticGroups := m.staticGroupRepo.Get(userGroups)

// 	if len(staticGroups) == 0 {
// 		return "", true
// 	}

// 	scopes = staticGroups.getGroupScopes()

// 	return scopes, false
// }

// getUserData get all scopes, tenants and username from user
// func (m *mapperForUser) getUserData(reqData ReqData, username string) (scopes string, tenants []string, user string, err error) {
// 	staticUser, err := m.staticUserRepo.Get(username)
// 	if err != nil {
// 		return "", []string{}, "", errors.Wrap(err, fmt.Sprintf("while searching for a static user with username %s", username))
// 	}
// 	scopes, err = reqData.GetScopes()
// 	if err != nil {
// 		if !apperrors.IsKeyDoesNotExist(err) {
// 			return "", []string{}, "", errors.Wrap(err, "while fetching scopes")
// 		}
// 		scopes = strings.Join(staticUser.Scopes, " ")
// 	}

// 	return scopes, staticUser.Tenants, staticUser.Username, nil
// }

// func (m *mapperForUser) GetObjectContext(ctx context.Context, reqData ReqData, username string) (ObjectContext, error) {
// 	var externalTenantID, scopes string
// 	var tenants []string
// 	var err error
// 	proceedWithUser := false
// 	consumerType := consumer.Group
// 	finalUserName := username

// 	scopes, proceedWithUser = m.getGroupData(reqData)

// 	if proceedWithUser {
// 		scopes, tenants, finalUserName, err = m.getUserData(reqData, username)
// 		if err != nil {
// 			return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while getting user data"))
// 		}
// 		consumerType = consumer.User
// 	}

// 	externalTenantID, err = reqData.GetExternalTenantID()

// 	if err != nil {
// 		if !apperrors.IsKeyDoesNotExist(err) {
// 			return ObjectContext{}, errors.Wrap(err, "while fetching external tenant")
// 		}
// 		return NewObjectContext(TenantContext{}, scopes, finalUserName, consumer.User), nil
// 	}

// 	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
// 	if err != nil {
// 		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
// 	}

// 	if proceedWithUser && !hasValidTenant(tenants, tenantMapping.ExternalTenant) {
// 		return ObjectContext{}, errors.New("tenant mismatch")
// 	}

// 	return NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), scopes, finalUserName, consumerType), nil
// }

func (m *mapperForUser) getScopesForUserGroups(reqData ReqData) string {
	userGroups := reqData.GetUserGroups()
	if len(userGroups) == 0 {
		return ""
	}

	staticGroups := m.staticGroupRepo.Get(userGroups)
	if len(staticGroups) == 0 {
		return ""
	}

	return staticGroups.getGroupScopes()
}

func (m *mapperForUser) getUserData(reqData ReqData, username string) (*StaticUser, string, error) {
	staticUser, err := m.staticUserRepo.Get(username)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("while searching for a static user with username %s", username))
	}

	scopes, err := reqData.GetScopes()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return nil, "", errors.Wrap(err, "while fetching scopes")
		}
		scopes = strings.Join(staticUser.Scopes, " ")
	}

	return &staticUser, scopes, nil
}

func (m *mapperForUser) GetObjectContext(ctx context.Context, reqData ReqData, username string) (ObjectContext, error) {
	var externalTenantID, scopes string
	var staticUser *StaticUser
	consumerType := consumer.Group
	var err error

	scopes = m.getScopesForUserGroups(reqData)
	if !hasScopes(scopes) {
		staticUser, scopes, err = m.getUserData(reqData, username)
		if err != nil {
			return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while getting user data"))
		}

		consumerType = consumer.User
	}

	externalTenantID, err = reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return ObjectContext{}, errors.Wrap(err, "while fetching external tenant")
		}
		return NewObjectContext(TenantContext{}, scopes, username, consumerType), nil
	}

	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	if staticUser != nil && !hasValidTenant(staticUser.Tenants, tenantMapping.ExternalTenant) {
		return ObjectContext{}, errors.New("tenant mismatch")
	}

	return NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), scopes, username, consumerType), nil
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
