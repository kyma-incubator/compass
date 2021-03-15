package plugins_test

import (
	"sort"
	"testing"

	"github.com/kyma-incubator/compass/components/director/hack/plugins"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestOrderedDefinitionList(t *testing.T) {
	// GIVEN
	definitions := plugins.OrderedDefinitionList{defMutation(), defQuery(), defObjectZ(), defObjectA(), defScalarB(), defScalarA(), defEnumB(), defEnumA()}
	// WHEN
	sort.Sort(definitions)
	// THEN
	require.Len(t, definitions, 8)
	assert.Equal(t, definitions[0], defScalarA())
	assert.Equal(t, definitions[1], defScalarB())
	assert.Equal(t, definitions[2], defEnumA())
	assert.Equal(t, definitions[3], defEnumB())
	assert.Equal(t, definitions[4], defObjectA())
	assert.Equal(t, definitions[5], defObjectZ())
	assert.Equal(t, definitions[6], defQuery())
	assert.Equal(t, definitions[7], defMutation())
}

func defScalarA() ast.Definition {
	return ast.Definition{
		Kind: ast.Scalar,
		Name: "A-scalar",
	}
}

func defScalarB() ast.Definition {
	return ast.Definition{
		Kind: ast.Scalar,
		Name: "B-scalar",
	}
}

func defEnumA() ast.Definition {
	return ast.Definition{
		Kind: ast.Enum,
		Name: "A-enum",
	}
}

func defEnumB() ast.Definition {
	return ast.Definition{
		Kind: ast.Enum,
		Name: "B-enum",
	}
}

func defObjectA() ast.Definition {
	return ast.Definition{
		Kind: ast.Object,
		Name: "A-object",
	}
}

func defMutation() ast.Definition {
	return ast.Definition{
		Kind: ast.Object,
		Name: "Mutation",
	}
}

func defQuery() ast.Definition {
	return ast.Definition{
		Kind: ast.Object,
		Name: "Query",
	}
}

func defObjectZ() ast.Definition {
	return ast.Definition{
		Kind: ast.Object,
		Name: "Z-object",
	}
}
