package str

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
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
	return strings.Title(strings.ToLower(s))
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
