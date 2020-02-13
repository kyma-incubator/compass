package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

const tenantGroupPrefix = "tenantID="

// GetGroupScopesAndTenant get all scopes from group array, without duplicates
func GetGroupScopesAndTenant(groups []StaticGroup) (string, string) {
	var filteredScopesArray []string
	var tenantID string

	for _, group := range groups {
		if strings.HasPrefix(group.GroupName, tenantGroupPrefix) {
			// group name is tenantID
			tenantID = strings.Replace(group.GroupName, tenantGroupPrefix, "", 1)
		} else {
			// group name is
			for _, scope := range group.Scopes {

				if !stringInSlice(scope, filteredScopesArray) {
					filteredScopesArray = append(filteredScopesArray, scope)
				}
			}
		}
	}

	return strings.Join(filteredScopesArray, " "), tenantID
}

func (m *mapperForUser) GetObjectContext(ctx context.Context, reqData ReqData, username string) (ObjectContext, error) {
	var externalTenantID, scopes string
	var staticUser StaticUser

	userGroups, err := reqData.GetGroups()
	log.Infof("GetGroups returned %s\n", userGroups)

	if err != nil {
		return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while getting groups for a static user with username %s", username))
	}

	staticGroups := m.staticGroupRepo.Get(userGroups)

	if len(staticGroups) > 0 {
		// proceed with group scopes flow
		scopes, externalTenantID = GetGroupScopesAndTenant(staticGroups)
		log.Infof("Decided to use groups copes %s\n", scopes)
	} else {
		// proceed with staticUser (and his scopes) flow
		staticUser, err = m.staticUserRepo.Get(username)
		if err != nil {
			return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while searching for a static user with username %s", username))
		}

		scopes, err = reqData.GetScopes()
		if err != nil {
			if !apperrors.IsKeyDoesNotExist(err) {
				return ObjectContext{}, errors.Wrap(err, "while fetching scopes")
			}

			scopes = strings.Join(staticUser.Scopes, " ")
			log.Infof("Decided to use staticUser copes %s\n", scopes)
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

	if len(userGroups) <= 0 && !hasValidTenant(staticUser.Tenants, tenantMapping.ExternalTenant) {
		return ObjectContext{}, errors.New("tenant mismatch")
	}

	return NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), scopes, staticUser.Username, consumer.User), nil
}

func hasValidTenant(assignedTenants []string, tenant string) bool {
	for _, assignedTenant := range assignedTenants {
		if assignedTenant == tenant {
			return true
		}
	}

	return false
}
