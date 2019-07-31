package label

import "regexp"

// ExtractValueFromJSONPath returns the value that is placed in the SQL/JSON path query
func ExtractValueFromJSONPath(jpq string) *string {
	re := regexp.MustCompile(`@\s*==\s*"(?P<value>.+)"`)

	res := re.FindAllStringSubmatch(jpq, -1)
	if res == nil {
		return nil
	}

	return &res[0][1]
}
