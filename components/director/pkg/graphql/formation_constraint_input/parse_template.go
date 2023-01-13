package formation_constraint_input

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"text/template"
)

func ParseInputTemplate(tmpl string, data interface{}, dest interface{}) error {
	t, err := template.New("").Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		fmt.Println("LOL")
		return err
	}

	res := new(bytes.Buffer)
	if err = t.Execute(res, data); err != nil {
		fmt.Println("NAH")
		return err
	}
	fmt.Println("HERE")
	fmt.Println(res.String())
	if err = json.Unmarshal(res.Bytes(), dest); err != nil {
		spew.Dump("ERROR")
		spew.Dump(err)
		return err
	}

	if validatable, ok := dest.(inputvalidation.Validatable); ok {
		fmt.Println("YOOOOO")
		return validatable.Validate()
	}

	return nil
}
