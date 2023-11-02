package formationconstraint

import (
	"bytes"
	"encoding/json"
	"text/template"

	"github.com/kyma-incubator/compass/components/director/pkg/templatehelper"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

// ParseInputTemplate parses tmpl using data and stores the result in dest
func ParseInputTemplate(tmpl string, data interface{}, dest interface{}) error {
	t, err := template.New("").Option("missingkey=zero").Funcs(templatehelper.GetFuncMap()).Parse(tmpl)
	if err != nil {
		return err
	}

	res := new(bytes.Buffer)
	if err = t.Execute(res, data); err != nil {
		return err
	}

	// <nil> comes after parsing the template with a go field that is a nil pointer
	// As we are expecting the resulting object to be valid JSON object, the <nil> value on its own is misleading in the following contexts
	// If we are working with a *string value we add quotes in the template around it.
	// If the *string is nil, it would result in "<nil>", which is not what we want,
	// but rather an empty string as it is the default value for an empty string in JSON.
	// This is in order to remove the <nil> that comes in templates that are surrounded by quotes that come from the template.
	resBytes := bytes.ReplaceAll(res.Bytes(), []byte(`"<nil>"`), []byte(`""`))
	// In other cases, we do not add quotes around the template, in such cases the value should be null,
	// as it is the correct default value for null JSON objects
	resBytes = bytes.ReplaceAll(resBytes, []byte(`<nil>`), []byte(`null`))
	if err = json.Unmarshal(resBytes, dest); err != nil {
		return err
	}

	if validatable, ok := dest.(inputvalidation.Validatable); ok {
		return validatable.Validate()
	}

	return nil
}
