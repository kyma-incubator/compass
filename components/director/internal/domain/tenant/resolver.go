package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Resolver struct {
}

func (r *Resolver) Tenants(ctx context.Context) ([]*graphql.Tenant, error) {

	return fixTenantPage(), nil

}

func NewResolver() *Resolver {
	return &Resolver{}
}

func fixTenantPage() []*graphql.Tenant {
	return []*graphql.Tenant{
		{
			Tenant: "3e64ebae-38b5-46a0-b1ed-9ccee153a0ae",
			Name:   str.Ptr("default"),
		},
		{
			Tenant: "1eba80dd-8ff6-54ee-be4d-77944d17b10b",
			Name:   str.Ptr("foo"),
		},
		{
			Tenant: "9ca034f1-11ab-5b25-b76f-dc77106f571d",
			Name:   str.Ptr("bar"),
		},
		{
			Tenant: "1143ea4c-76da-472b-9e01-930f90639cdc",
			Name:   str.Ptr("generated"),
		}}

}
