package descriptionsdecorator

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"unicode"

	"github.com/kyma-incubator/compass/components/director/hack/plugins"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/99designs/gqlgen/plugin"
	"github.com/vektah/gqlparser/ast"
)

type GraphqlOperationType string

const (
	Query             GraphqlOperationType = "query"
	Mutation          GraphqlOperationType = "mutation"
	ExamplesDirectory                      = "../../examples"
	UnsanitizedAPI                         = "A-P-I"
)

var _ plugin.ConfigMutator = &descriptionsDecoratorPlugin{}

func NewPlugin(schemaFileName string) *descriptionsDecoratorPlugin {
	return &descriptionsDecoratorPlugin{schemaFileName: schemaFileName}
}

type descriptionsDecoratorPlugin struct {
	schemaFileName string
}

func (p *descriptionsDecoratorPlugin) Name() string {
	return "descriptions_decorator"
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

func (p *descriptionsDecoratorPlugin) ensureDescription(f *ast.FieldDefinition, opType GraphqlOperationType) {

	f.Description = deletePrevious(f.Description)
	dirs, err := ioutil.ReadDir(ExamplesDirectory)
	if err != nil {
		log.Printf("While reading the examples directory: %s", err.Error())
		return
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		if sanitizeName(f.Name, opType) != dir.Name() {
			continue
		}
		files, err := ioutil.ReadDir(path.Join(ExamplesDirectory, dir.Name()))
		if err != nil {
			log.Printf("While reading the examples subdirectory %s : %s", dir.Name(), err.Error())
			return
		}
		if len(f.Description) == 0 {
			f.Description += "**Examples**"
		} else {
			f.Description = fmt.Sprintf("%s\n\n%s", f.Description, "**Examples**")
		}
		for _, file := range files {
			withoutExt := strings.Replace(file.Name(), ".graphql", "", -1)
			withoutDash := strings.Replace(withoutExt, "-", " ", -1)
			f.Description = addExample(f.Description, withoutDash, dir.Name(), file.Name())
		}

	}
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

	index := strings.Index(description, "**Examples**")
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
