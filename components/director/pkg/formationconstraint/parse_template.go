package formationconstraint

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

var templateFuncMap = template.FuncMap{
	"toString": func(bytesData []byte) string {
		config := string(bytesData)
		if config == "" {
			config = "\"\""
		}

		return config
	},
}

// ParseInputTemplate parses tmpl using data and stores the result in dest
func ParseInputTemplate(tmpl string, data interface{}, dest interface{}) error {
	t, err := template.New("").Option("missingkey=zero").Funcs(templateFuncMap).Parse(tmpl)
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
