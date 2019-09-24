package scopesdecorator

import (
	"fmt"
	"os"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin"
	"github.com/vektah/gqlparser/ast"
)

type GraphqlOperationType string

const (
	Query          GraphqlOperationType = "query"
	Mutation       GraphqlOperationType = "mutation"
	directiveName                       = "hasScopes"
	directiveArg                        = "path"
	schemaFileName                      = "schema.graphql"
)

var _ plugin.ConfigMutator = &scopesDecoratorPlugin{}

func NewPlugin() *scopesDecoratorPlugin {
	return &scopesDecoratorPlugin{}
}

type scopesDecoratorPlugin struct {
	paths []string
}

func (m *scopesDecoratorPlugin) Name() string {
	return "scopes_decorator"
}

func (m *scopesDecoratorPlugin) MutateConfig(cfg *config.Config) error {
	fmt.Printf("[%s] Mutate Configuration", m.Name())
	if err := cfg.Check(); err != nil {
		return err
	}

	schema, _, err := cfg.LoadSchema()
	if err != nil {
		return err
	}

	for _, f := range schema.Query.Fields {
		if !m.hasDirective(*f) {
			f.Directives = append(f.Directives, &ast.Directive{
				Name:      directiveName,
				Arguments: m.getDirectiveArguments(Query, f.Name),
			})
		}
	}

	for _, f := range schema.Mutation.Fields {
		if !m.hasDirective(*f) {
			f.Directives = append(f.Directives, &ast.Directive{
				Name:      directiveName,
				Arguments: m.getDirectiveArguments(Mutation, f.Name),
			})
		}
	}
	
	if err := cfg.Check(); err != nil {
		return err
	}

	schemaFile, err := os.Create(schemaFileName)
	if err != nil {
		return err
	}

	f := NewFormatter(schemaFile)
	f.FormatSchema(schema)
	return schemaFile.Close()
}

func (m *scopesDecoratorPlugin) hasDirective(def ast.FieldDefinition) bool {
	for _, d := range def.Directives {
		if d.Name == directiveName {
			return true
		}
	}
	return false
}

func (m *scopesDecoratorPlugin) getDirectiveArguments(opType GraphqlOperationType, opName string) ast.ArgumentList {
	var args ast.ArgumentList
	path := fmt.Sprintf("%s.%s", opType, opName)
	args = append(args, &ast.Argument{Name: directiveArg, Value: &ast.Value{Raw: path, Kind: ast.StringValue}})
	return args
}

func (m *scopesDecoratorPlugin) registerDirectivePath(path string) {
	m.paths = append(m.paths, path)
}
