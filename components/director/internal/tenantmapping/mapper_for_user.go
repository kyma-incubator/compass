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

func NewMapperForUser(staticUserRepo StaticUserRepository, tenantStorageService TenantStorageService) *mapperForUser {
	return &mapperForUser{
		staticUserRepo:       staticUserRepo,
		tenantStorageService: tenantStorageService,
	}
}

type mapperForUser struct {
	staticUserRepo       StaticUserRepository
	tenantStorageService TenantStorageService
}

func (m *mapperForUser) GetObjectContext(ctx context.Context, reqData ReqData, username string) (ObjectContext, error) {
	var externalTenant, scopes string

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

	externalTenant, err = reqData.GetExternalTenantID()
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while fetching tenant")
	}

	if !hasValidTenant(staticUser.Tenants, externalTenant) {
		return ObjectContext{}, errors.New("tenant mismatch")
	}

	internalTenant, err := m.tenantStorageService.GetInternalTenant(ctx, externalTenant)
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while mapping external to internal tenant")
	}

	return NewObjectContext(scopes, internalTenant, staticUser.Username, consumer.User), nil
}

func hasValidTenant(assignedTenants []uuid.UUID, tenant string) bool {
	for _, assignedTenant := range assignedTenants {
		if assignedTenant.String() == tenant {
			return true
		}
	}

	return false
}
