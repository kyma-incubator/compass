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

// type mapperForGroup struct {
// 	staticGroupRepository StaticGroupRepository
// 	tenantRepo     TenantRepository
// }

// func (m *mapperForGroup) GetObjectContext(ctx context.Context, reqData ReqData, userGroup string) (ObjectContext, error) {

// 	scopes, err = reqData.GetScopes()
// 	if err != nil {
// 		if !apperrors.IsKeyDoesNotExist(err) {
// 			return ObjectContext{}, errors.Wrap(err, "while fetching scopes")
// 		}

// 		scopes = strings.Join([], " ")
// 	}

// 	externalTenantID, err = reqData.GetExternalTenantID()
// 	if err != nil {
// 		if !apperrors.IsKeyDoesNotExist(err) {
// 			return ObjectContext{}, errors.Wrap(err, "while fetching external tenant")
// 		}

// 		// do zmiany
// 		return NewObjectContext(TenantContext{}, scopes, staticUser.Username, consumer.User), nil
// 	}

// 	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
// 	if err != nil {
// 		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
// 	}

// 	return NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), scopes, staticUser.Username, consumer.User), nil

// }

func (m *mapperForUser) GetObjectContext(ctx context.Context, reqData ReqData, username string) (ObjectContext, error) {
	var externalTenantID, scopes string
	var staticUser StaticUser
	// sprawdzanie grupy
	userGroup, err := reqData.GetGroup()
	log.Infof("GetGroup returned %s\n", userGroup)
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while getting group for a static user with username %s", username))
	}

	if userGroup != "" {
		// proceed with group scopes

		staticGroup, err := m.staticGroupRepo.Get(userGroup)
		if err != nil {
			return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while searching for a static group with userGroup %s", userGroup))
		}

		scopes = strings.Join(staticGroup.Scopes, " ")

	} else {
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

		// do zmiany
		return NewObjectContext(TenantContext{}, scopes, staticUser.Username, consumer.User), nil
	}

	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	// do wyjebania
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
