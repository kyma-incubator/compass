package str

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

func Unique(in []string) []string {
	set := SliceToMap(in)
	return MapToSlice(set)
}

func SliceToMap(in []string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, i := range in {
		set[i] = struct{}{}
	}

	return set
}

func MapToSlice(set map[string]struct{}) []string {
	var items []string
	for key := range set {
		items = append(items, key)
	}

	return items
}

func Ptr(s string) *string {
	return &s
}

func Cast(i interface{}) (string, error) {
	if s, ok := i.(string); ok {
		return s, nil
	}

	return "", apperrors.NewInternalError("unable to cast the value to a string type")
}

func PrefixStrings(in []string, prefix string) []string {
	var prefixedFieldNames []string
	for _, column := range in {
		prefixedFieldNames = append(prefixedFieldNames, fmt.Sprintf("%s%s", prefix, column))
	}

	return prefixedFieldNames
}

func Title(s string) string {
	return strings.Title(strings.ToLower(s))
}

func PtrStrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func InterfaceSliceToStringSlice(value []interface{}) ([]string, error) {
	var result []string
	for _, elem := range value {
		item, ok := elem.(string)
		if !ok {
			return nil, apperrors.NewInternalError("value is not a string")
		}

		result = append(result, item)
	}
	return result, nil
}

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

func SubstractSlice(s1, s2 []string) []string {
	s2Set := make(map[string]bool, 0)
	for _, elem := range s2 {
		s2Set[elem] = true
	}

	result := make([]string, 0)
	for _, elem := range s1 {
		if !s2Set[elem] {
			result = append(result, elem)
		}
	}
	return result
}

func IntersectSlice(s1, s2 []string) []string {
	s1Set := make(map[string]bool, 0)
	for _, elem := range s1 {
		s1Set[elem] = true
	}

	result := make([]string, 0)
	for _, elem := range s2 {
		if s1Set[elem] {
			result = append(result, elem)
		}
	}
	return result
func NewNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  len(s) != 0,
	}
}
