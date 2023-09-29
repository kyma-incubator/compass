package templatehelper

import (
	"encoding/json"
	"github.com/Masterminds/sprig/v3"
	"strings"
	"text/template"
)

func GetFuncMap() template.FuncMap {
	fm := sprig.TxtFuncMap()
	fm["toString"] = toString
	fm["contains"] = contains
	fm["Join"] = joinStrings

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
