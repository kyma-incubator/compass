package descriptionsdecorator

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/kyma-incubator/compass/components/director/hack/plugins"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin"
	"github.com/vektah/gqlparser/ast"
)

type GraphqlOperationType string

const (
	Query    GraphqlOperationType = "query"
	Mutation GraphqlOperationType = "mutation"
)

var _ plugin.ConfigMutator = &descriptionsDecoratorPlugin{}

func NewSDescriptionsDecoratorPlugin(schemaFileName string) *descriptionsDecoratorPlugin {
	return &descriptionsDecoratorPlugin{schemaFileName: schemaFileName}
}

type descriptionsDecoratorPlugin struct {
	schemaFileName string
}

func (p *descriptionsDecoratorPlugin) Name() string {
	return "descriptions_decorator"
}

func (p *descriptionsDecoratorPlugin) ensureDescription(f *ast.FieldDefinition, opType GraphqlOperationType) {
	f.Description = addExample(f.Description, f.Name, opType)
}

func (p *descriptionsDecoratorPlugin) MutateConfig(cfg *config.Config) error {
	fmt.Printf("[%s] Mutate Configuration\n", p.Name())
	if err := cfg.Check(); err != nil {
		return err
	}

	schema, _, err := cfg.LoadSchema()
	if err != nil {
		return err
	}

	for _, f := range schema.Query.Fields {
		p.ensureDescription(f, Query)
	}

	for _, f := range schema.Mutation.Fields {
		p.ensureDescription(f, Mutation)
	}

	if err := cfg.Check(); err != nil {
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

func addExample(description string, name string, opType GraphqlOperationType) string {
	counter := 0
	for index, letter := range name {
		if unicode.IsUpper(letter) {
			name = fmt.Sprintf("%s-%s", name[:index+counter], name[index+counter:])
			counter++
		}
	}
	if strings.Contains(name, "A-P-I") {
		name = strings.ReplaceAll(name, "A-P-I", "api")
	}

	if opType == Query {
		return strings.ToLower(fmt.Sprintf("%s see example [here](query-%s.graphql)", description, name))

	}
	return strings.ToLower(fmt.Sprintf("%s see example [here](%s.graphql)", description, name))
}
