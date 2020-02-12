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

func (m *mapperForUser) GetObjectContext(ctx context.Context, reqData ReqData, username string) (ObjectContext, error) {
	var externalTenantID, scopes string
	var staticUser StaticUser

	userGroup, err := reqData.GetGroup()
	log.Infof("GetGroup returned %s\n", userGroup)

	if err != nil {
		return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while getting group for a static user with username %s", username))
	}

	staticGroup, groupExistsError := m.staticGroupRepo.Get(userGroup)
	if groupExistsError != nil {
		// proceed with group scopes flow
		scopes = strings.Join(staticGroup.Scopes, " ")
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
		}
	}

	externalTenantID, err = reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return ObjectContext{}, errors.Wrap(err, "while fetching external tenant")
		}
		// TODO: co tu zrobic
		return NewObjectContext(TenantContext{}, scopes, staticUser.Username, consumer.User), nil
	}

	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	if userGroup == "" && !hasValidTenant(staticUser.Tenants, tenantMapping.ExternalTenant) {
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
