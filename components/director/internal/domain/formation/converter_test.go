package formation_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

const testFormation = "test-formation"

func TestFromGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		sut := formation.NewConverter()

		// WHEN
		actual := sut.FromGraphQL(graphql.FormationInput{Name: testFormation})

		// THEN
		assert.Equal(t, testFormation, actual.Name)
	})
}

func TestToGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		sut := formation.NewConverter()
		// WHEN

		actual := sut.ToGraphQL(&model.Formation{Name: testFormation})

		// THEN
		assert.Equal(t, testFormation, actual.Name)
	})
}
