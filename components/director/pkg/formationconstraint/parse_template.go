package formationconstraint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"strconv"
	"text/template"

	_ "github.com/itchyny/gojq"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

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
	// copy copies the value that is found under "path" in "jsonString" and returns it
	"copy": func(input json.RawMessage, path string) string {
		jsonstring := string(input)
		spew.Dump("HEHEEHHEHEHEHE", jsonstring)
		var content string
		if path == "." || path == "" {
			content = jsonstring
		} else {
			content = gjson.Get(jsonstring, path).String()
		}
		return content
	},
	"updateAndCopy": func(input json.RawMessage, path string, entries []string) string {
		jsonstring := string(input)
		spew.Dump("NONONONONONO", jsonstring)
		var content string
		if path == "." || path == "" {
			content = jsonstring
		} else {
			content = gjson.Get(jsonstring, path).String()
		}
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
	},
	"mkslice": func(args ...string) []string {
		return args
	},
	"stringify": func(v interface{}) string {
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
