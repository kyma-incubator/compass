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

func TestGqlFieldsProvider_OmitCombinedFieldsForApplication(t *testing.T) {
	type testCase struct {
		name               string
		fp                 *graphqlizer.GqlFieldsProvider
		omit               []string
		expectedProperties map[string]int
	}
	tests := []testCase{
		{
			name: "with omitted top level complex fields",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages", "webhooks"},
			expectedProperties: map[string]int{
				"id":                  2,
				"name":                1,
				"integrationSystemID": 1,
				"status":              1,
				"auths":               1,
				"packages":            0,
				"instanceAuths":       0,
				"webhooks":            0,
				"apiDefinitions":      0,
				"eventDefinitions":    0,
				"documents":           0,
				"fetchRequest":        0,
			},
		},
		{
			name: "with multiple omitted 'fetchRequest' fields",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.apiDefinitions.spec.fetchRequest", "packages.eventDefinitions.spec.fetchRequest", "packages.documents.fetchRequest"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"status":              2,
				"auths":               1,
				"packages":            1,
				"instanceAuths":       1,
				"webhooks":            1,
				"apiDefinitions":      1,
				"eventDefinitions":    1,
				"documents":           1,
				"fetchRequest":        0,
			},
		},
		{
			name: "with certain field omitted only in some nested complex fields",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.apiDefinitions.spec.fetchRequest", "packages.eventDefinitions.spec.fetchRequest"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"status":              3,
				"auths":               1,
				"packages":            1,
				"instanceAuths":       1,
				"webhooks":            1,
				"apiDefinitions":      1,
				"eventDefinitions":    1,
				"documents":           1,
				"fetchRequest":        1,
			},
		},
		{
			name: "with omitted nested fields",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"auths", "webhooks", "status", "packages.instanceAuths", "packages.documents", "packages.apiDefinitions.spec.fetchRequest", "packages.eventDefinitions.spec.fetchRequest"},
			expectedProperties: map[string]int{
				"id":                  4,
				"name":                4,
				"integrationSystemID": 1,
				"status":              0,
				"auths":               0,
				"packages":            1,
				"instanceAuths":       0,
				"webhooks":            0,
				"apiDefinitions":      1,
				"eventDefinitions":    1,
				"documents":           0,
				"fetchRequest":        0,
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

func TestGqlFieldsProvider_OmitSeparatelyFieldsForApplication(t *testing.T) {
	type testCase struct {
		name               string
		fp                 *graphqlizer.GqlFieldsProvider
		omit               []string
		expectedProperties map[string]int
	}
	tests := []testCase{
		{
			name: "with no omitted fields on 'applications' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{},
			expectedProperties: map[string]int{
				"id":           8,
				"name":         4,
				"fetchRequest": 3,
			},
		},
		{
			name: "with omitted simple field on 'applications' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"integrationSystemID"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 0,
				"packages":            1,
				"apiDefinitions":      1,
				"fetchRequest":        3,
			},
		},
		{
			name: "with omitted complex field 'webhooks' on 'applications' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"webhooks"},
			expectedProperties: map[string]int{
				"id":                  7,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            0,
				"apiDefinitions":      1,
				"fetchRequest":        3,
			},
		},
		{
			name: "with omitted simple field on 'webhooks' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"webhooks.id"},
			expectedProperties: map[string]int{
				"id":                  7,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"apiDefinitions":      1,
				"fetchRequest":        3,
			},
		},
		{
			name: "with omitted complex field 'packages' on 'applications' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages"},
			expectedProperties: map[string]int{
				"id":                  3,
				"name":                1,
				"integrationSystemID": 1,
				"packages":            0,
				"webhooks":            1,
				"apiDefinitions":      0,
				"fetchRequest":        0,
			},
		},
		{
			name: "with omitted simple field on 'packages' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.id"},
			expectedProperties: map[string]int{
				"id":                  7,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"apiDefinitions":      1,
				"fetchRequest":        3,
			},
		},
		{
			name: "with omitted simple field on 'instanceAuth' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.instanceAuths.id"},
			expectedProperties: map[string]int{
				"id":                  7,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"apiDefinitions":      1,
				"fetchRequest":        3,
				"reason":              1,
			},
		},
		{
			name: "with omitted simple field on 'instanceAuth.status' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.instanceAuths.status.reason"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"apiDefinitions":      1,
				"fetchRequest":        3,
				"reason":              0,
			},
		},
		{
			name: "with omitted simple field on 'apiDefinitions' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.apiDefinitions.id"},
			expectedProperties: map[string]int{
				"id":                  7,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"apiDefinitions":      1,
				"fetchRequest":        3,
				"filter":              3,
			},
		},
		{
			name: "with omitted 'fetchRequest' field on 'apiDefinitions.spec' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.apiDefinitions.spec.fetchRequest"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"apiDefinitions":      1,
				"fetchRequest":        2,
				"filter":              2,
			},
		},
		{
			name: "with omitted simple field on 'apiDefinitions.spec.fetchRequest' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.apiDefinitions.spec.fetchRequest.filter"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"apiDefinitions":      1,
				"fetchRequest":        3,
				"filter":              2,
				"forRemoval":          2,
			},
		},
		{
			name: "with omitted simple field on 'apiDefinitions.version' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.apiDefinitions.version.forRemoval"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"apiDefinitions":      1,
				"fetchRequest":        3,
				"forRemoval":          1,
			},
		},
		{
			name: "with omitted simple field on 'eventDefinitions' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.eventDefinitions.id"},
			expectedProperties: map[string]int{
				"id":                  7,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"eventDefinitions":    1,
				"fetchRequest":        3,
				"filter":              3,
			},
		},
		{
			name: "with omitted 'fetchRequest' field on 'eventDefinitions.spec' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.eventDefinitions.spec.fetchRequest"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"eventDefinitions":    1,
				"fetchRequest":        2,
				"filter":              2,
			},
		},
		{
			name: "with omitted simple field on 'eventDefinitions.spec.fetchRequest' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.eventDefinitions.spec.fetchRequest.filter"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"eventDefinitions":    1,
				"fetchRequest":        3,
				"filter":              2,
				"forRemoval":          2,
			},
		},
		{
			name: "with omitted simple field on 'eventDefinitions.version' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.eventDefinitions.version.forRemoval"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"eventDefinitions":    1,
				"fetchRequest":        3,
				"forRemoval":          1,
			},
		},
		{
			name: "with omitted simple field on 'documents' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.documents.id"},
			expectedProperties: map[string]int{
				"id":                  7,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"documents":           1,
				"fetchRequest":        3,
				"filter":              3,
			},
		},
		{
			name: "with omitted 'fetchRequest' field on 'documents' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.documents.fetchRequest"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"documents":           1,
				"fetchRequest":        2,
				"filter":              2,
			},
		},
		{
			name: "with omitted simple field on 'documents.fetchRequest' level",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.documents.fetchRequest.filter"},
			expectedProperties: map[string]int{
				"id":                  8,
				"name":                4,
				"integrationSystemID": 1,
				"packages":            1,
				"webhooks":            1,
				"documents":           1,
				"fetchRequest":        3,
				"filter":              2,
				"forRemoval":          2,
			},
		},
		{
			name: "with omitted non-existing fields",
			fp:   &graphqlizer.GqlFieldsProvider{},
			omit: []string{"packages.nonExisting", "idTypo", "packagesTypo.id", "packages.apiDefinitions.idTypo"},
			expectedProperties: map[string]int{
				"id":               8,
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
