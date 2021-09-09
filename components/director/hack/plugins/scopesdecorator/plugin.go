package scopesdecorator

import (
	"fmt"
	"os"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/hack/plugins"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin"
	"github.com/vektah/gqlparser/v2/ast"
)

// GraphqlOperationType missing godoc
type GraphqlOperationType string

const (
	// Query missing godoc
	Query GraphqlOperationType = "query"
	// Mutation missing godoc
	Mutation GraphqlOperationType = "mutation"
)

const (
	directiveArgumentPrefix = "graphql"
	directiveName           = "hasScopes"
	directiveArg            = "path"
)

var _ plugin.ConfigMutator = &scopesDecoratorPlugin{}

// NewPlugin missing godoc
func NewPlugin(schemaFileName string) *scopesDecoratorPlugin {
	return &scopesDecoratorPlugin{schemaFileName: schemaFileName}
}

type scopesDecoratorPlugin struct {
	schemaFileName string
}

// Name missing godoc
func (p *scopesDecoratorPlugin) Name() string {
	return "scopes_decorator"
}

// MutateConfig missing godoc
func (p *scopesDecoratorPlugin) MutateConfig(cfg *config.Config) error {
	log.D().Infof("[%s] Mutate Configuration\n", p.Name())
	if err := cfg.Init(); err != nil {
		return err
	}

	schema := cfg.Schema
	if schema.Query != nil {
		for _, f := range schema.Query.Fields {
			p.ensureDirective(f, Query)
		}
	}
	if schema.Mutation != nil {
		for _, f := range schema.Mutation.Fields {
			p.ensureDirective(f, Mutation)
		}
	}
	if err := cfg.LoadSchema(); err != nil {
		return err
	}

	schemaFile, err := os.Create(p.schemaFileName)
	if err != nil {
		return err
	}

	f := plugins.NewFormatter(schemaFile)
	f.FormatSchema(schema)
	return schemaFile.Close()
}

func (p *scopesDecoratorPlugin) ensureDirective(f *ast.FieldDefinition, opType GraphqlOperationType) {
	if d := p.getDirective(f); d == nil {
		f.Directives = append(f.Directives, &ast.Directive{
			Name:      directiveName,
			Arguments: p.getDirectiveArguments(opType, f.Name),
		})
	} else {
		d.Name = directiveName
		d.Arguments = p.getDirectiveArguments(opType, f.Name)
	}
}

func (p *scopesDecoratorPlugin) getDirective(def *ast.FieldDefinition) *ast.Directive {
	for _, d := range def.Directives {
		if d.Name == directiveName {
			return d
		}
	}
	return nil
}

func (p *scopesDecoratorPlugin) getDirectiveArguments(opType GraphqlOperationType, opName string) ast.ArgumentList {
	var args ast.ArgumentList
	path := fmt.Sprintf("%s.%s.%s", directiveArgumentPrefix, opType, opName)
	args = append(args, &ast.Argument{Name: directiveArg, Value: &ast.Value{Raw: path, Kind: ast.StringValue}})
	return args
}
