package label

import (
	"regexp"

	"github.com/pkg/errors"
)

// ExtractValueFromJSONPath returns the value that is placed in the SQL/JSON path query
//
// For a given JSON path $[*] ? (@ == "dd") it extracts the actual value - dd
// For a given JSON path $[*] ? (@ == "dd" || @ == "ww") it extracts the actual values - dd and ww
func ExtractValueFromJSONPath(jpq string) ([]string, error) {
	re := regexp.MustCompile(`^\$\[\*\]\s*\?\s*\(\s*@\s*==\s*"(?P<value>[a-zA-Z0-9\-\_\s]+)"\s*|\|\|\s*@\s*==\s*"(?P<value>[a-zA-Z0-9\-\_\s]+)"`)
	res := re.FindAllStringSubmatch(jpq, -1)
	if res == nil {
		return nil, errors.New("value not found in the query parameter")
	}

	extractedValues := make([]string, len(res))
	for idx, r := range res {
		if idx == 0 {
			extractedValues[idx] = r[1]
			continue
		}

		extractedValues[idx] = r[len(r)-1]
	}

	return extractedValues, nil
}
