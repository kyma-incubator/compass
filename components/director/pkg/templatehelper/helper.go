package templatehelper

import (
	"encoding/json"
	"strconv"
	"strings"
	"text/template"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/Masterminds/sprig/v3"
)

// GetFuncMap returns a template function map with provided build-in functions
// and on top of them our additional custom functions
func GetFuncMap() template.FuncMap {
	fm := sprig.TxtFuncMap()
	fm["toString"] = toString
	fm["contains"] = contains
	fm["Join"] = joinStrings
	fm["copy"] = copy
	fm["updateAndCopy"] = updateAndCopy
	fm["mkslice"] = mkslice

	return fm
}

func toString(bytesData []byte) string {
	config := string(bytesData)
	if config == "" {
		config = "\"\""
	}

	return config
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

// copy copies the value that is found under "path" in "jsonString" and returns it
func copy(jsonstring, path string) string {
	return gjson.Get(jsonstring, path).String()
}

func updateAndCopy(input json.RawMessage, path string, entries []string) string {
	jsonstring := string(input)
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
}

func mkslice(args ...string) []string {
	return args
}
