package templatehelper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"bytes"

	"github.com/Masterminds/sprig/v3"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
)

// getFuncMap returns a template function map with provided build-in functions
// and on top of them our additional custom functions
func getFuncMap() template.FuncMap {
	fm := sprig.TxtFuncMap()
	fm["toString"] = toString
	fm["contains"] = contains
	fm["Join"] = joinStrings
	fm["copy"] = copyFromJSON
	fm["updateAndCopy"] = updateAndCopy
	fm["mkslice"] = mkslice
	fm["stringify"] = stringify

	return fm
}

func ParseTemplate(tmpl *string, data interface{}, dest interface{}) error {
	t, err := template.New("").Funcs(getFuncMap()).Option("missingkey=zero").Parse(*tmpl)
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

func toString(bytesData []byte) *string {
	config := string(bytesData)
	if config == "" {
		return nil
	}

	return &config
}

func contains(faConfig json.RawMessage, str string) bool {
	return strings.Contains(string(faConfig), str)
}

func joinStrings(elems []string) string {
	if len(elems) == 0 {
		return ""
	}
	return `"` + strings.Join(elems, `", "`) + `"`
}

func copyFromJSON(input json.RawMessage, path string) string {
	jsonstring := string(input)
	var content string
	if path == "." || path == "" {
		content = jsonstring
	} else {
		content = gjson.Get(jsonstring, path).String()
	}
	return content
}

func updateAndCopy(input json.RawMessage, path string, entries []string) string {
	content := copyFromJSON(input, path)
	var err error
	for _, entry := range entries {
		key := gjson.Get(entry, "key").String()
		value := gjson.Get(entry, "value").String()
		if json.Valid([]byte(value)) {
			content, err = sjson.SetRaw(content, key, value)
		} else {
			content, err = sjson.Set(content, key, value)
		}
		if err != nil {
			log.D().Errorf("Failed to update and copy configuration")
			return ""
		}
	}
	return strconv.Quote(content)
}

func mkslice(args ...string) []string {
	return args
}

func stringify(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case *string:
		return *v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
