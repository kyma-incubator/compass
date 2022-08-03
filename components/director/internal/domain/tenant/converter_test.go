package tenant_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	tnt "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ids             = []string{"id1", "id2"}
	names           = []string{"name1", "name2"}
	externalTenants = []string{"external1", "external2"}
	parent          = ""
	subdomain       = "subdomain"
	region          = "region"
)

func TestConverter(t *testing.T) {
	// GIVEN
	id := "foo"
	name := "bar"

	t.Run("all fields", func(t *testing.T) {
		c := tenant.NewConverter()

		// When
		input := newModelBusinessTenantMapping(id, name)
		entity := c.ToEntity(input)
		outputModel := c.FromEntity(entity)

		// THEN
		assert.Equal(t, input, outputModel)
	})

	t.Run("initialized from entity", func(t *testing.T) {
		c := tenant.NewConverter()
		initialized := true
		input := newEntityBusinessTenantMapping(id, name)
		input.Initialized = &initialized

		// When
		outputModel := c.FromEntity(input)

		// Then
		assert.Equal(t, input.Initialized, outputModel.Initialized)
	})

	t.Run("nil model", func(t *testing.T) {
		c := tenant.NewConverter()

		// When
		entity := c.ToEntity(nil)

		// Then
		require.Nil(t, entity)
	})

	t.Run("nil entity", func(t *testing.T) {
		c := tenant.NewConverter()

		// When
		tenantModel := c.FromEntity(nil)

		// Then
		require.Nil(t, tenantModel)
	})
}

func TestConverter_ToGraphQL(t *testing.T) {
	t.Run("when input is nil", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		res := c.ToGraphQL(nil)

		// THEN
		require.Nil(t, res)
	})

	t.Run("all fields", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		in := newModelBusinessTenantMapping(ids[0], names[0])
		res := c.ToGraphQL(in)
		expected := &graphql.Tenant{
			ID:          testExternal,
			InternalID:  ids[0],
			Name:        &names[0],
			Type:        string(tnt.Account),
			ParentID:    "",
			Initialized: nil,
			Labels:      nil,
			Provider:    testProvider,
		}

		// THEN
		require.Equal(t, expected, res)
	})
}

func TestConverter_ToGraphQLInput(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		in := newModelBusinessTenantMappingInput(names[0], subdomain, region)
		res := c.ToGraphQLInput(in)
		expected := graphql.BusinessTenantMappingInput{
			Name:           names[0],
			ExternalTenant: testExternal,
			Parent:         str.Ptr(parent),
			Subdomain:      str.Ptr(subdomain),
			Region:         str.Ptr(region),
			Type:           string(tnt.Account),
			Provider:       testProvider,
		}

		// THEN
		require.Equal(t, expected, res)
	})
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		in := []*graphql.BusinessTenantMappingInput{
			{
				Name:           names[0],
				ExternalTenant: externalTenants[0],
				Parent:         str.Ptr(parent),
				Subdomain:      str.Ptr(subdomain),
				Region:         str.Ptr(region),
				Type:           string(tnt.Account),
				Provider:       testProvider,
			},
			{
				Name:           names[1],
				ExternalTenant: externalTenants[1],
				Parent:         str.Ptr(parent),
				Subdomain:      str.Ptr(subdomain),
				Region:         str.Ptr(region),
				Type:           string(tnt.Account),
				Provider:       testProvider,
			}}
		res := c.MultipleInputFromGraphQL(in)
		expected := []model.BusinessTenantMappingInput{
			{
				Name:           names[0],
				ExternalTenant: externalTenants[0],
				Parent:         parent,
				Subdomain:      subdomain,
				Region:         region,
				Type:           string(tnt.Account),
				Provider:       testProvider,
			},
			{
				Name:           names[1],
				ExternalTenant: externalTenants[1],
				Parent:         parent,
				Subdomain:      subdomain,
				Region:         region,
				Type:           string(tnt.Account),
				Provider:       testProvider,
			},
		}

		// THEN
		require.Equal(t, len(expected), len(res))
		require.Equal(t, expected, res)
	})
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	t.Run("when one of the tenants is nil", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		in := []*model.BusinessTenantMapping{
			{
				ID:             ids[0],
				Name:           names[0],
				ExternalTenant: externalTenants[0],
				Parent:         parent,
				Type:           tnt.Account,
				Provider:       testProvider,
				Status:         "",
				Initialized:    nil,
			},
			nil,
		}
		res := c.MultipleToGraphQL(in)
		expected := []*graphql.Tenant{
			{
				ID:          externalTenants[0],
				InternalID:  ids[0],
				Name:        &names[0],
				Type:        string(tnt.Account),
				ParentID:    "",
				Initialized: nil,
				Labels:      nil,
				Provider:    testProvider,
			},
		}

		// THEN
		require.Equal(t, len(expected), len(res))
		require.Equal(t, expected, res)
	})
	t.Run("all fields", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		in := []*model.BusinessTenantMapping{
			{
				ID:             ids[0],
				Name:           names[0],
				ExternalTenant: externalTenants[0],
				Parent:         parent,
				Type:           tnt.Account,
				Provider:       testProvider,
				Status:         "",
				Initialized:    nil,
			},
			{
				ID:             ids[1],
				Name:           names[1],
				ExternalTenant: externalTenants[1],
				Parent:         parent,
				Type:           tnt.Account,
				Provider:       testProvider,
				Status:         "",
				Initialized:    nil,
			},
		}
		res := c.MultipleToGraphQL(in)
		expected := []*graphql.Tenant{
			{
				ID:          externalTenants[0],
				InternalID:  ids[0],
				Name:        &names[0],
				Type:        string(tnt.Account),
				ParentID:    parent,
				Initialized: nil,
				Labels:      nil,
				Provider:    testProvider,
			},
			{
				ID:          externalTenants[1],
				InternalID:  ids[1],
				Name:        &names[1],
				Type:        string(tnt.Account),
				ParentID:    parent,
				Initialized: nil,
				Labels:      nil,
				Provider:    testProvider,
			},
		}

		// THEN
		require.Equal(t, len(expected), len(res))
		require.Equal(t, expected, res)
	})
}

func TestConverter_MultipleInputToGraphQLInputL(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		in := []model.BusinessTenantMappingInput{
			{
				Name:           names[0],
				ExternalTenant: externalTenants[0],
				Parent:         parent,
				Subdomain:      subdomain,
				Region:         region,
				Type:           string(tnt.Account),
				Provider:       testProvider,
			},
			{
				Name:           names[1],
				ExternalTenant: externalTenants[1],
				Parent:         parent,
				Subdomain:      subdomain,
				Region:         region,
				Type:           string(tnt.Account),
				Provider:       testProvider,
			},
		}
		res := c.MultipleInputToGraphQLInput(in)
		expected := []graphql.BusinessTenantMappingInput{
			{
				Name:           names[0],
				ExternalTenant: externalTenants[0],
				Parent:         str.Ptr(parent),
				Subdomain:      str.Ptr(subdomain),
				Region:         str.Ptr(region),
				Type:           string(tnt.Account),
				Provider:       testProvider,
			},
			{
				Name:           names[1],
				ExternalTenant: externalTenants[1],
				Parent:         str.Ptr(parent),
				Subdomain:      str.Ptr(subdomain),
				Region:         str.Ptr(region),
				Type:           string(tnt.Account),
				Provider:       testProvider,
			},
		}

		// THEN
		require.Equal(t, len(expected), len(res))
		require.Equal(t, expected, res)
	})
}
