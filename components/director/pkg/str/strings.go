package str

import (
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

	return "", apperrors.NewInvalidStringCastError()
}
