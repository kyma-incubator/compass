package formation_constraint_input

import (
	"bytes"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"text/template"
)

func ParseInputTemplate(tmpl string, data interface{}, dest interface{}) error {
	t, err := template.New("").Option("missingkey=zero").Parse(tmpl)
	if err != nil {
		return err
	}

	res := new(bytes.Buffer)
	if err = t.Execute(res, data); err != nil {
		return err
	}
	if err = json.Unmarshal(res.Bytes(), dest); err != nil {
		return err
	}

	if validatable, ok := dest.(inputvalidation.Validatable); ok {
		return validatable.Validate()
	}

	return nil
}
