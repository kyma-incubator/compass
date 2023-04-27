package tenantbusinesstype_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenantbusinesstype"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.TenantBusinessType
		Expected *graphql.TenantBusinessType
	}{
		{
			Name:     "All properties given",
			Input:    fixModelTenantBusinessType(tbtID, tbtCode, tbtName),
			Expected: fixGQLTenantBusinessType(tbtID, tbtCode, tbtName),
		},
		{
			Name:     "Empty",
			Input:    &model.TenantBusinessType{},
			Expected: &graphql.TenantBusinessType{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			converter := tenantbusinesstype.NewConverter()
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	conv := tenantbusinesstype.NewConverter()

	t.Run("All properties given", func(t *testing.T) {
		// GIVEN
		tbtModel := fixModelTenantBusinessType(tbtID, tbtCode, tbtName)

		// WHEN
		tbtEntity := conv.ToEntity(tbtModel)

		// THEN
		assertTenantBusinessTypeDefinition(t, tbtModel, tbtEntity)
	})

	t.Run("Nil", func(t *testing.T) {
		// WHEN
		tbtEntity := conv.ToEntity(nil)

		// THEN
		assert.Nil(t, tbtEntity)
	})

	t.Run("Empty", func(t *testing.T) {
		// GIVEN
		tbtModel := &model.TenantBusinessType{}

		// WHEN
		tbtEntity := conv.ToEntity(tbtModel)

		// THEN
		assertTenantBusinessTypeDefinition(t, tbtModel, tbtEntity)
	})
}

func TestConverter_FromEntity(t *testing.T) {
	conv := tenantbusinesstype.NewConverter()

	t.Run("All properties given", func(t *testing.T) {
		// GIVEN
		tbtEntity := fixEntityTenantBusinessType(tbtID, tbtCode, tbtName)

		// WHEN
		tbtModel := conv.FromEntity(tbtEntity)

		// THEN
		assertTenantBusinessTypeDefinition(t, tbtModel, tbtEntity)
	})

	t.Run("Nil", func(t *testing.T) {
		// WHEN
		tbtModel := conv.FromEntity(nil)

		// THEN
		assert.Nil(t, tbtModel)
	})

	t.Run("Empty", func(t *testing.T) {
		// GIVEN
		tbtEntity := &tenantbusinesstype.Entity{}

		// WHEN
		tbtModel := conv.FromEntity(tbtEntity)

		// THEN
		assertTenantBusinessTypeDefinition(t, tbtModel, tbtEntity)
	})
}

func assertTenantBusinessTypeDefinition(t *testing.T, tbtModel *model.TenantBusinessType, entity *tenantbusinesstype.Entity) {
	assert.Equal(t, tbtModel.ID, entity.ID)
	assert.Equal(t, tbtModel.Code, entity.Code)
	assert.Equal(t, tbtModel.Name, entity.Name)

}
