package tenant_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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

func TestConverter_TenantAccessInputFromGraphQL(t *testing.T) {
	testCases := []struct {
		Name             string
		Input            graphql.TenantAccessInput
		ExpectedErrorMsg string
		ExpectedOutput   *model.TenantAccess
	}{
		{
			Name: "Success for application",
			Input: graphql.TenantAccessInput{
				TenantID:     testExternal,
				ResourceType: graphql.TenantAccessObjectTypeApplication,
				ResourceID:   testID,
				Owner:        false,
			},
			ExpectedOutput: &model.TenantAccess{
				ExternalTenantID: testExternal,
				ResourceType:     resource.Application,
				ResourceID:       testID,
				Owner:            false,
			},
		},
		{
			Name: "Success for runtime",
			Input: graphql.TenantAccessInput{
				TenantID:     testExternal,
				ResourceType: graphql.TenantAccessObjectTypeRuntime,
				ResourceID:   testID,
				Owner:        false,
			},
			ExpectedOutput: &model.TenantAccess{
				ExternalTenantID: testExternal,
				ResourceType:     resource.Runtime,
				ResourceID:       testID,
				Owner:            false,
			},
		},
		{
			Name: "Success for runtime context",
			Input: graphql.TenantAccessInput{
				TenantID:     testExternal,
				ResourceType: graphql.TenantAccessObjectTypeRuntimeContext,
				ResourceID:   testID,
				Owner:        false,
			},
			ExpectedOutput: &model.TenantAccess{
				ExternalTenantID: testExternal,
				ResourceType:     resource.RuntimeContext,
				ResourceID:       testID,
				Owner:            false,
			},
		},
		{
			Name: "Error when converting resource type",
			Input: graphql.TenantAccessInput{
				TenantID:     testExternal,
				ResourceType: graphql.TenantAccessObjectType(resource.FormationConstraint),
				ResourceID:   testID,
				Owner:        false,
			},
			ExpectedErrorMsg: fmt.Sprintf("Unknown tenant access resource type %q", resource.FormationConstraint),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			c := tenant.NewConverter()
			output, err := c.TenantAccessInputFromGraphQL(testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Equal(t, testCase.ExpectedErrorMsg, err.Error())
				require.Nil(t, output)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedOutput, output)
			}
		})
	}
}

func TestConverter_TenantAccessToGraphQL(t *testing.T) {
	testCases := []struct {
		Name             string
		Input            *model.TenantAccess
		ExpectedErrorMsg string
		ExpectedOutput   *graphql.TenantAccess
	}{
		{
			Name:           "Success when nil input",
			Input:          nil,
			ExpectedOutput: nil,
		},
		{
			Name: "Success for application",
			Input: &model.TenantAccess{
				ExternalTenantID: testExternal,
				InternalTenantID: testInternal,
				ResourceType:     resource.Application,
				ResourceID:       testID,
				Owner:            false,
			},
			ExpectedOutput: &graphql.TenantAccess{
				TenantID:     testExternal,
				ResourceType: graphql.TenantAccessObjectTypeApplication,
				ResourceID:   testID,
				Owner:        false,
			},
		},
		{
			Name: "Success for runtime",
			Input: &model.TenantAccess{
				ExternalTenantID: testExternal,
				InternalTenantID: testInternal,
				ResourceType:     resource.Runtime,
				ResourceID:       testID,
				Owner:            false,
			},
			ExpectedOutput: &graphql.TenantAccess{
				TenantID:     testExternal,
				ResourceType: graphql.TenantAccessObjectTypeRuntime,
				ResourceID:   testID,
				Owner:        false,
			},
		},
		{
			Name: "Success for runtime context",
			Input: &model.TenantAccess{
				ExternalTenantID: testExternal,
				InternalTenantID: testInternal,
				ResourceType:     resource.RuntimeContext,
				ResourceID:       testID,
				Owner:            false,
			},
			ExpectedOutput: &graphql.TenantAccess{
				TenantID:     testExternal,
				ResourceType: graphql.TenantAccessObjectTypeRuntimeContext,
				ResourceID:   testID,
				Owner:        false,
			},
		},
		{
			Name: "Error when converting resource type",
			Input: &model.TenantAccess{
				ExternalTenantID: testExternal,
				InternalTenantID: testInternal,
				ResourceType:     resource.FormationConstraint,
				ResourceID:       testID,
				Owner:            false,
			},
			ExpectedErrorMsg: fmt.Sprintf("Unknown tenant access resource type %q", resource.FormationConstraint),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			c := tenant.NewConverter()
			output, err := c.TenantAccessToGraphQL(testCase.Input)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Equal(t, testCase.ExpectedErrorMsg, err.Error())
				require.Nil(t, output)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedOutput, output)
			}
		})
	}
}

func TestConverter_TenantAccessToEntity(t *testing.T) {
	t.Run("when input is nil", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		res := c.TenantAccessToEntity(nil)

		// THEN
		require.Nil(t, res)
	})
	t.Run("all fields", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		in := &model.TenantAccess{
			ExternalTenantID: testExternal,
			InternalTenantID: testInternal,
			ResourceType:     resource.Application,
			ResourceID:       testID,
		}
		res := c.TenantAccessToEntity(in)
		expected := &repo.TenantAccess{
			TenantID:   testInternal,
			ResourceID: in.ResourceID,
			Owner:      in.Owner,
		}

		// THEN
		require.Equal(t, expected, res)
	})
}

func TestConverter_TenantAccessFromEntity(t *testing.T) {
	t.Run("when input is nil", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		res := c.TenantAccessToEntity(nil)

		// THEN
		require.Nil(t, res)
	})
	t.Run("all fields", func(t *testing.T) {
		c := tenant.NewConverter()

		// WHEN
		in := &repo.TenantAccess{
			TenantID:   testInternal,
			ResourceID: testID,
			Owner:      false,
		}
		res := c.TenantAccessFromEntity(in)
		expected := &model.TenantAccess{
			InternalTenantID: testInternal,
			ResourceID:       testID,
			Owner:            false,
		}

		// THEN
		require.Equal(t, expected, res)
	})
}
