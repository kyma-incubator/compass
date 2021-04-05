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
