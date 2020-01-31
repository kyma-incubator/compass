package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

func NewMapperForUser(staticUserRepo StaticUserRepository, tenantRepo TenantRepository) *mapperForUser {
	return &mapperForUser{
		staticUserRepo: staticUserRepo,
		tenantRepo:     tenantRepo,
	}
}

type mapperForUser struct {
	staticUserRepo StaticUserRepository
	tenantRepo     TenantRepository
}

func (m *mapperForUser) GetObjectContext(ctx context.Context, reqData ReqData, username string) (ObjectContext, error) {
	var externalTenantID, scopes string

	staticUser, err := m.staticUserRepo.Get(username)
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

	externalTenantID, err = reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return ObjectContext{}, errors.Wrap(err, "while fetching external tenant")
		}

		return NewObjectContext(TenantContext{}, scopes, staticUser.Username, consumer.User), nil
	}

	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	if !hasValidTenant(staticUser.Tenants, tenantMapping.ExternalTenant) {
		return ObjectContext{}, errors.New("tenant mismatch")
	}

	return NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), scopes, staticUser.Username, consumer.User), nil
}

func hasValidTenant(assignedTenants []uuid.UUID, tenant string) bool {
	for _, assignedTenant := range assignedTenants {
		if assignedTenant.String() == tenant {
			return true
		}
	}

	return false
}
