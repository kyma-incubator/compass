package formationconstraint_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
	"testing"
)

var converter = formationconstraint.NewConverter()

func TestToGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// WHEN
		actual := converter.ToGraphQL(formationConstraintModel)

		// THEN
		require.Equal(t, gqlFormationConstraint, actual)
	})
	t.Run("Nil input", func(t *testing.T) {
		// WHEN
		actual := converter.ToGraphQL(nil)

		// THEN
		require.Nil(t, actual)
	})
}

func TestMultipleToGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// WHEN
		actual := converter.MultipleToGraphQL([]*model.FormationConstraint{formationConstraintModel, formationConstraintModel2, nil})

		// THEN
		require.Equal(t, []*graphql.FormationConstraint{gqlFormationConstraint, gqlFormationConstraint2}, actual)
	})
	t.Run("Nil input", func(t *testing.T) {
		// WHEN
		actual := converter.MultipleToGraphQL(nil)

		// THEN
		require.Nil(t, actual)
	})
}

func TestToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// WHEN
		actual := converter.ToEntity(formationConstraintModel)

		// THEN
		require.Equal(t, &entity, actual)
	})
	t.Run("Nil input", func(t *testing.T) {
		// WHEN
		actual := converter.ToEntity(nil)

		// THEN
		require.Nil(t, actual)
	})
}

func TestFromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// WHEN
		actual := converter.FromEntity(&entity)

		// THEN
		require.Equal(t, formationConstraintModel, actual)
	})
	t.Run("Nil input", func(t *testing.T) {
		// WHEN
		actual := converter.FromEntity(nil)

		// THEN
		require.Nil(t, actual)
	})
}

func TestFromInputGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// WHEN
		actual := converter.FromInputGraphQL(gqlInput)

		// THEN
		require.Equal(t, modelInput, actual)
	})
}

func TestFromModelInputToModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// WHEN
		actual := converter.FromModelInputToModel(modelInput, testID)

		// THEN
		require.Equal(t, modelFromInput, actual)
	})
}
