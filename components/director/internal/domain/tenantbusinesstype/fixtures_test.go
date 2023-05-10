package tenantbusinesstype_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenantbusinesstype"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var (
	tbtID   = "bd0646fa-3c30-4255-84f8-182f57742aa1"
	tbtCode = "test-code"
	tbtName = "test-name"
)

func fixTbtColumns() []string {
	return []string{"id", "code", "name"}
}

func fixModelTenantBusinessType(id, code, name string) *model.TenantBusinessType {
	return &model.TenantBusinessType{
		ID:   id,
		Code: code,
		Name: name,
	}
}

func fixGQLTenantBusinessType(id, code, name string) *graphql.TenantBusinessType {
	return &graphql.TenantBusinessType{
		ID:   id,
		Name: name,
		Code: code,
	}
}

func fixEntityTenantBusinessType(id, code, name string) *tenantbusinesstype.Entity {
	return &tenantbusinesstype.Entity{
		ID:   id,
		Code: code,
		Name: name,
	}
}

func fixModelTenantBusinessTypeInput(code, name string) *model.TenantBusinessTypeInput {
	return &model.TenantBusinessTypeInput{
		Code: code,
		Name: name,
	}
}
