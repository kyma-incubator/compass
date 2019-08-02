package label

import (
	"regexp"

	"github.com/pkg/errors"
)

// ExtractValueFromJSONPath returns the value that is placed in the SQL/JSON path query
//
// For a given JSON path $[*] ? (@ == "dd") it extracts the actual value - dd
func ExtractValueFromJSONPath(jpq string) (*string, error) {
	re := regexp.MustCompile(`@\s*==\s*"(?P<value>.+)"`)

	res := re.FindAllStringSubmatch(jpq, -1)
	if res == nil {
		return nil, errors.New("value not found in the query parameter")
	}

	return &res[0][1], nil
}
