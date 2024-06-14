package str

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/text/language"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"golang.org/x/text/cases"
)

// Unique missing godoc
func Unique(in []string) []string {
	set := SliceToMap(in)
	return MapToSlice(set)
}

// SliceToMap missing godoc
func SliceToMap(in []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, i := range in {
		set[i] = struct{}{}
	}

	return set
}

// MapToSlice missing godoc
func MapToSlice(set map[string]struct{}) []string {
	items := make([]string, 0, len(set))
	for key := range set {
		items = append(items, key)
	}

	return items
}

// Ptr missing godoc
func Ptr(s string) *string {
	return &s
}

// Cast missing godoc
func Cast(i interface{}) (string, error) {
	if s, ok := i.(string); ok {
		return s, nil
	}

	return "", apperrors.NewInternalError("unable to cast the value to a string type")
}

// PrefixStrings missing godoc
func PrefixStrings(in []string, prefix string) []string {
	prefixedFieldNames := make([]string, 0, len(in))
	for _, column := range in {
		prefixedFieldNames = append(prefixedFieldNames, fmt.Sprintf("%s%s", prefix, column))
	}

	return prefixedFieldNames
}

// Title missing godoc
func Title(s string) string {
	return cases.Title(language.Und).String(strings.ToLower(s))
}

// PtrStrToStr missing godoc
func PtrStrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Matches missing godoc
func Matches(actual []string, required []string) bool {
	actMap := make(map[string]interface{})

	for _, a := range actual {
		actMap[a] = struct{}{}
	}
	for _, r := range required {
		_, ex := actMap[r]
		if !ex {
			return false
		}
	}
	return true
}

// NewNullString missing godoc
func NewNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  len(s) != 0,
	}
}

// CastOrEmpty casts the given value to string and returns the empty string otherwise
func CastOrEmpty(i interface{}) string {
	if s, ok := i.(string); ok {
		return s
	}
	return ""
}

// ContainsInSlice checks if a string value is present in a given slice
func ContainsInSlice(s []string, str string) bool {
	for _, val := range s {
		if val == str {
			return true
		}
	}
	return false
}

// CastToBool casts the given value to string and then parsers it to a bool value
func CastToBool(i interface{}) (bool, error) {
	str := CastOrEmpty(i)
	return strconv.ParseBool(str)
}

// StringifyJSONRawMessage if the rawMessage contains value returns it, otherwise return nil
func StringifyJSONRawMessage(rawMessage json.RawMessage) *string {
	str := string(rawMessage)
	if str == "" {
		return nil
	}
	return &str
}

// ValueIn checks if the provided value is part of the provided string slice
func ValueIn(value string, in []string) bool {
	for _, v := range in {
		if value == v {
			return true
		}
	}
	return false
}

// ConvertToStringArray converts array of interfaces to string array
func ConvertToStringArray(arr []interface{}) ([]string, error) {
	stringArr := make([]string, 0, len(arr))
	for _, value := range arr {
		str, ok := value.(string)
		if !ok {
			return nil, errors.New("cannot convert interface value into a string")
		}

		stringArr = append(stringArr, str)
	}

	return stringArr, nil
}

func MergeWithoutDuplicates(a, b []string) []string {
	occurrence := make(map[string]bool)
	var result []string

	// Merge both slices
	for _, v := range append(a, b...) {
		// Check if the value is already present in the map
		if !occurrence[v] {
			result = append(result, v) // If not present, append to result slice and mark as present
			occurrence[v] = true
		}
	}

	return result
}
