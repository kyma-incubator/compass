package tenantmapping

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

func NewMapperForUser(staticUserRepo StaticUserRepository) *mapperForUser {
	return &mapperForUser{
		staticUserRepo: staticUserRepo,
	}
}

type mapperForUser struct {
	staticUserRepo StaticUserRepository
}

func (m *mapperForUser) GetObjectContext(reqData ReqData, username string) (ObjectContext, error) {
	var tenant, scopes string

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

	tenant, err = reqData.GetTenantID()
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while fetching tenant")
	}

	if !hasValidTenant(staticUser.Tenants, tenant) {
		return ObjectContext{}, errors.New("tenant missmatch")
	}

	return NewObjectContext(scopes, tenant, staticUser.Username, "Static User"), nil
}

func hasValidTenant(assignedTenants []uuid.UUID, tenant string) bool {
	for _, assignedTenant := range assignedTenants {
		if assignedTenant.String() == tenant {
			return true
		}
	}

	return false
}
