package formationconstraint

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	_ "github.com/itchyny/gojq"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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
	"copy": func(jsonstring, path string) string {
		return gjson.Get(jsonstring, path).String()
	},
	"updateAndCopy": func(fa model.FormationAssignment, path string, pathToValuePairs []string) string {
		configuration := str.StringifyJSONRawMessage(fa.Value)
		content := gjson.Get(configuration, path).String()
		var err error
		for _, pvPair := range pathToValuePairs {
			pvPairSplit := strings.Split(pvPair, "^")
			pathToUpdate := pvPairSplit[0]
			newValue := pvPairSplit[1]
			if json.Valid([]byte(newValue)) {
				content, err = sjson.SetRaw(content, pathToUpdate, newValue)
			} else {
				content, err = sjson.Set(content, pathToUpdate, newValue)
			}
			if err != nil {
				log.D().Errorf("Failed to update and copy configuration of formation assignment with ID %q from formation %q", fa.ID, fa.FormationID)
				return ""
			}
		}
		return content
	},
	"mkslice": func(args ...string) []string {
		return args
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
