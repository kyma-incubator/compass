package str

import (
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
	var scenariosString []string
	for _, scenario := range value {
		item, ok := scenario.(string)
		if !ok {
			return nil, apperrors.NewInternalError("value is not a string")
		}
		scenariosString = append(scenariosString, item)
	}
	return scenariosString, nil
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
	for _, scenario := range s2 {
		s2Set[scenario] = true
	}

	result := make([]string, 0)
	for _, scenario := range s1 {
		if _, ok := s2Set[scenario]; !ok {
			result = append(result, scenario)
		}
	}
	return result
}

func IntersectSlice(s1, s2 []string) []string {
	existingScenarioMap := make(map[string]bool, 0)
	for _, scenario := range s1 {
		existingScenarioMap[scenario] = true
	}

	result := make([]string, 0)
	for _, scenario := range s2 {
		if _, ok := existingScenarioMap[scenario]; ok {
			result = append(result, scenario)
		}
	}
	return result
}
