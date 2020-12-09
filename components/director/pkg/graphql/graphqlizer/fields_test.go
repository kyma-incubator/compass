package graphqlizer_test

import (
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/stretchr/testify/assert"
)

func TestGqlFieldsProvider_Page(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fp := &graphqlizer.GqlFieldsProvider{}
		expected := "data {\n\t\tproperty\n\t}\n\tpageInfo {startCursor\n\t\tendCursor\n\t\thasNextPage}\n\ttotalCount\n\t"
		actual := fp.Page("property")
		assert.Equal(t, expected, actual)
	})
}

func TestGqlFieldsProvider_OmitForApplication(t *testing.T) {
	type testCase struct {
		name               string
		fp                 *graphqlizer.GqlFieldsProvider
		omit               []string
		expectedProperties map[string]int
	}
	tests := []testCase{
		{
			name: "with no omitted fields",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{},
			expectedProperties: map[string]int{
				"id":           7,
				"name":         4,
				"fetchRequest": 3,
			},
		},
		{
			name: "with omitted top level simple field",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"integrationSystemID"},
			expectedProperties: map[string]int{
				"id":                  7,
				"name":                4,
				"packages":            1,
				"apiDefinitions":      1,
				"integrationSystemID": 0,
			},
		},
		{
			name: "with omitted top level complex field",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages"},
			expectedProperties: map[string]int{
				"id":                  3,
				"name":                1,
				"integrationSystemID": 1,
				"packages":            0,
				"apiDefinitions":      0,
				"eventDefinitions":    0,
				"targetURL":           0,
			},
		},
		{
			name: "with omitted nested fields",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.apiDefinitions.spec.fetchRequest", "packages.eventDefinitions.spec.fetchRequest", "packages.documents.fetchRequest"},
			expectedProperties: map[string]int{
				"packages":         1,
				"apiDefinitions":   1,
				"eventDefinitions": 1,
				"fetchRequest":     0,
			},
		},
		{
			name: "with certain field omitted only in some nested complex fields",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.apiDefinitions.spec.fetchRequest", "packages.eventDefinitions.spec.fetchRequest"},
			expectedProperties: map[string]int{
				"apiDefinitions":   1,
				"eventDefinitions": 1,
				"documents":        1,
				"fetchRequest":     1,
			},
		},
		{
			name: "with omit for non-existing field",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.nonExisting"},
			expectedProperties: map[string]int{
				"id":               7,
				"name":             4,
				"packages":         1,
				"fetchRequest":     3,
				"apiDefinitions":   1,
				"eventDefinitions": 1,
				"documents":        1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.fp.OmitForApplication(tt.omit)
			for expectedProp, expectedCount := range tt.expectedProperties {
				fieldRegex := regexp.MustCompile(`\b` + expectedProp + `\b`)

				matches := fieldRegex.FindAllStringIndex(actual, -1)
				actualCount := len(matches)

				assert.Equal(t, expectedCount, actualCount, expectedProp)
			}
		})
	}
}
