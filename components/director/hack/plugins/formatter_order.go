package plugins

import (
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// QueryTypeName missing godoc
const QueryTypeName = "Query"

// MutationTypeName missing godoc
const MutationTypeName = "Mutation"

// OrderedDefinitionList missing godoc
type OrderedDefinitionList []ast.Definition

// Len missing godoc
func (d OrderedDefinitionList) Len() int {
	return len(d)
}

// Swap missing godoc
func (d OrderedDefinitionList) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// Less missing godoc
func (d OrderedDefinitionList) Less(i, j int) bool {
	first := d[i]
	second := d[j]

	typeComparison := d.typeMapping(first.Kind) - d.typeMapping(second.Kind)
	if typeComparison < 0 {
		return true
	} else if typeComparison > 0 {
		return false
	}

	if first.Kind == ast.Object {
		// query and mutations should be at the end of the file
		if first.Name == MutationTypeName {
			return false
		}
		if second.Name == MutationTypeName {
			return true
		}
		if first.Name == QueryTypeName {
			return false
		}
		if second.Name == QueryTypeName {
			return true
		}
	}
	return strings.Compare(first.Name, second.Name) < 0
}

func (d OrderedDefinitionList) typeMapping(kind ast.DefinitionKind) int {
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
