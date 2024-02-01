package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
)

type AddTenantAccess struct {
	tenantID     string
	resourceType graphql.TenantAccessObjectType
	resourceID   string
	isOwner      bool
	asserters    []asserters.Asserter
}

func NewAddTenantAccess() *AddTenantAccess {
	return &AddTenantAccess{}
}

func (o *AddTenantAccess) WithTenantID(tenantID string) *AddTenantAccess {
	o.tenantID = tenantID
	return o
}
func (o *AddTenantAccess) WithResourceType(resourceType graphql.TenantAccessObjectType) *AddTenantAccess {
	o.resourceType = resourceType
	return o
}
func (o *AddTenantAccess) WithResourceID(resourceID string) *AddTenantAccess {
	o.resourceID = resourceID
	return o
}
func (o *AddTenantAccess) WithOwnership(isOwner bool) *AddTenantAccess {
	o.isOwner = isOwner
	return o
}

func (o *AddTenantAccess) WithAsserters(asserters ...asserters.Asserter) *AddTenantAccess {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *AddTenantAccess) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.AddTenantAccessForResource(t, ctx, gqlClient, o.tenantID, o.resourceID, o.resourceType, o.isOwner)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *AddTenantAccess) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
}

func (o *AddTenantAccess) Operation() Operation {
	return o
}
