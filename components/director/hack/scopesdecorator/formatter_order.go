package scopesdecorator

import (
	"strings"

	"github.com/vektah/gqlparser/ast"
)

const QueryTypeName = "Query"
const MutationTypeName = "Mutation"

type DefinitionList []ast.Definition

func (d DefinitionList) Len() int {
	return len(d)
}

func (d DefinitionList) Swap(i, j int) {
	tmp := d[i]
	d[i] = d[j]
	d[j] = tmp
}

func (d DefinitionList) Less(i, j int) bool {
	id := d[i]
	jd := d[j]

	if d.typeMapping(id.Kind) < d.typeMapping(jd.Kind) {

		return true
	}
	if d.typeMapping(id.Kind) > d.typeMapping(jd.Kind) {
		return false

	}

	if id.Kind == ast.Object {
		if id.Name == MutationTypeName {
			return false
		}
		if jd.Name == MutationTypeName {
			return true
		}
		if id.Name == QueryTypeName {
			return false
		}
		if jd.Name == QueryTypeName {
			return true
		}
	}
	return strings.Compare(id.Name, jd.Name) < 0
}

func (d DefinitionList) typeMapping(kind ast.DefinitionKind) int {
	switch kind {
	case ast.Scalar:
		return 1
	case ast.Enum:
		return 2
	case ast.Interface:
		return 3
	case ast.Union:
		return 4
	case ast.InputObject:
		return 5
	case ast.Object:
		return 6
	default:
		return 0
	}

}
