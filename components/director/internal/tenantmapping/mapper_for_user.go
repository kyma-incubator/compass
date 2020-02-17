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

const tenantGroupPrefix = "tenantID="

// GetGroupScopesAndTenant get all scopes from group array, without duplicates
func GetGroupScopesAndTenant(groups []StaticGroup) (string, string) {
	var scopeMap map[string]bool
	var tenantID string
	filteredScopes := []string{}

	for _, group := range groups {
		if strings.HasPrefix(group.GroupName, tenantGroupPrefix) {
			tenantID = strings.Replace(group.GroupName, tenantGroupPrefix, "", 1)
		} else {
			for _, scope := range group.Scopes {
				_, ok := scopeMap[scope]
				if !ok {
					filteredScopes = append(filteredScopes, scope)
					scopeMap[scope] = true
				}

			}
		}
	}

	return strings.Join(filteredScopes, " "), tenantID
}

func (m *mapperForUser) GetObjectContext(ctx context.Context, reqData ReqData, username string) (ObjectContext, error) {
	var externalTenantID, scopes, finalUserName string
	var staticUser StaticUser
	var staticGroups []StaticGroup
	var assignedTenants []string

	userGroups, err := reqData.GetGroups()

	if err != nil {
		return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while getting groups for a static user with username %s", username))
	}

	if len(userGroups) > 0 {
		staticGroups = m.staticGroupRepo.Get(userGroups)
	}

	if len(staticGroups) > 0 {
		// proceed with group scopes flow
		scopes, externalTenantID = GetGroupScopesAndTenant(staticGroups)
		for _, group := range staticGroups {
			assignedTenants = append(assignedTenants, group.Tenants...)
		}
		finalUserName = username

	} else {
		// proceed with staticUser (and his scopes) flow
		staticUser, err = m.staticUserRepo.Get(username)
		if err != nil {
			return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while searching for a static user with username %s", username))
		}
		assignedTenants = staticUser.Tenants
		finalUserName = staticUser.Username

		scopes, err = reqData.GetScopes()
		if err != nil {
			if !apperrors.IsKeyDoesNotExist(err) {
				return ObjectContext{}, errors.Wrap(err, "while fetching scopes")
			}

			scopes = strings.Join(staticUser.Scopes, " ")

		}
		externalTenantID, err = reqData.GetExternalTenantID()
		if err != nil {
			if !apperrors.IsKeyDoesNotExist(err) {
				return ObjectContext{}, errors.Wrap(err, "while fetching external tenant")
			}
			return NewObjectContext(TenantContext{}, scopes, staticUser.Username, consumer.User), nil
		}
	}

	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	if !hasValidTenant(assignedTenants, tenantMapping.ExternalTenant) {
		return ObjectContext{}, errors.New("tenant mismatch")
	}

	return NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), scopes, finalUserName, consumer.User), nil
}

func hasValidTenant(assignedTenants []string, tenant string) bool {
	for _, assignedTenant := range assignedTenants {
		if assignedTenant == tenant {
			return true
		}
	}

	return false
}
