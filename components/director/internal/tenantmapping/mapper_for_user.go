package tenantmapping

import (
	"fmt"
	"strings"

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

func (m *mapperForUser) GetTenantAndScopes(reqData ReqData, username string) (string, string, error) {
	var tenant string
	var scopes string

	hasScopes, hasTenant := true, true

	tenant, err := reqData.GetTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return "", "", errors.Wrap(err, "while fetching tenant")
		}

		hasTenant = false
	}

	scopes, err = reqData.GetScopes()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return "", "", errors.Wrap(err, "while fetching scopes")
		}

		hasScopes = false
	}

	if !hasScopes || !hasTenant {
		staticUser, err := m.staticUserRepo.Get(username)
		if err != nil {
			return "", "", errors.Wrap(err, fmt.Sprintf("while searching for a static user with username %s", username))
		}

		if !hasTenant {
			tenant = staticUser.Tenant.String()
		}

		if !hasScopes {
			scopes = strings.Join(staticUser.Scopes, " ")
		}
	}

	return tenant, scopes, nil
}
