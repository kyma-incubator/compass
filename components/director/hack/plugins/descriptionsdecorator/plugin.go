package descriptionsdecorator

import (
	"fmt"
	"os"
	"path"
	"strings"
	"unicode"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin"
	"github.com/kyma-incubator/compass/components/director/hack/plugins"
	"github.com/pkg/errors"
	"github.com/vektah/gqlparser/v2/ast"
)

// GraphqlOperationType missing godoc
type GraphqlOperationType string

const (
	// Query missing godoc
	Query GraphqlOperationType = "query"
	// Mutation missing godoc
	Mutation GraphqlOperationType = "mutation"
	// UnsanitizedAPI missing godoc
	UnsanitizedAPI = "A-P-I"
	// ExamplePrefix missing godoc
	ExamplePrefix = "**Examples**"
)

var _ plugin.ConfigMutator = &descriptionsDecoratorPlugin{}

// NewPlugin missing godoc
func NewPlugin(schemaFileName string, examplesDirectory string) *descriptionsDecoratorPlugin {
	return &descriptionsDecoratorPlugin{schemaFileName: schemaFileName, examplesDirectory: examplesDirectory}
}

type descriptionsDecoratorPlugin struct {
	schemaFileName    string
	examplesDirectory string
}

// Name missing godoc
func (p *descriptionsDecoratorPlugin) Name() string {
	return "descriptions_decorator"
}

// MutateConfig missing godoc
func (p *descriptionsDecoratorPlugin) MutateConfig(cfg *config.Config) error {
	log.D().Infof("[%s] Mutate Configuration\n", p.Name())

	if err := cfg.Init(); err != nil {
		return err
	}

	schema := cfg.Schema
	if schema.Query != nil {
		for _, f := range schema.Query.Fields {
			err := p.ensureDescription(f, Query)
			if err != nil {
				return err
			}
		}
	}

	if schema.Mutation != nil {
		for _, f := range schema.Mutation.Fields {
			err := p.ensureDescription(f, Mutation)
			if err != nil {
				return err
			}
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

func (p *descriptionsDecoratorPlugin) ensureDescription(f *ast.FieldDefinition, opType GraphqlOperationType) error {
	f.Description = deletePrevious(f.Description)
	dirs, err := os.ReadDir(p.examplesDirectory)
	if err != nil {
		log.D().Infof("no examples under %s directory, skipping adding description", p.examplesDirectory)
		//lint:ignore nilerr can proceed
		return nil
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		if sanitizeName(f.Name, opType) != dir.Name() {
			continue
		}
		files, err := os.ReadDir(path.Join(p.examplesDirectory, dir.Name()))
		if err != nil {
			return errors.Wrap(err, "while reading the examples subdirectory")
		}
		if len(f.Description) == 0 {
			f.Description += ExamplePrefix
		} else {
			f.Description = fmt.Sprintf("%s\n\n%s", f.Description, ExamplePrefix)
		}
		for _, file := range files {
			withoutExt := strings.ReplaceAll(file.Name(), ".graphql", "")
			withoutDash := strings.ReplaceAll(withoutExt, "-", " ")
			f.Description = addExample(f.Description, withoutDash, dir.Name(), file.Name())
		}
	}
	return nil
}

func sanitizeName(name string, opType GraphqlOperationType) string {
	counter := 0
	for index, letter := range name {
		if unicode.IsUpper(letter) {
			if index == 0 {
				continue
			}
			name = fmt.Sprintf("%s-%s", name[:index+counter], name[index+counter:])
			counter++
		}
	}
	if strings.Contains(name, UnsanitizedAPI) {
		name = strings.ReplaceAll(name, UnsanitizedAPI, "api")
	}

	if opType == Query {
		return strings.ToLower(fmt.Sprintf("query-%s", name))
	}
	return strings.ToLower(name)
}

func deletePrevious(description string) string {
	if len(description) == 0 {
		return ""
	}

	index := strings.Index(description, ExamplePrefix)
	if index == -1 {
		return description
	}
	if index == 0 {
		return ""
	}
	return description[:index-2]
}

func addExample(description string, name, dirName string, fileName string) string {
	return fmt.Sprintf("%s\n- [%s](examples/%s/%s)", description, name, dirName, fileName)
}
