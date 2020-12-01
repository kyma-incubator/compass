package normalizer_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	"github.com/stretchr/testify/assert"
)

func TestNormalizer(t *testing.T) {
	const expectedName = "mp-test-application"

	// given
	testCases := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name:     "Contains upper case characters",
			Input:    "tEsT-ApPlIcAtIoN",
			Expected: expectedName,
		},
		{
			Name:     "Contains dot non-alphanumeric character in the middle",
			Input:    "test.application",
			Expected: expectedName,
		},
		{
			Name:     "Contains $ non-alphanumeric character in the middle",
			Input:    "test$application",
			Expected: expectedName,
		},
		{
			Name:     "Contains & non-alphanumeric character in the middle",
			Input:    "test&application",
			Expected: expectedName,
		},
		{
			Name:     "Contains @ non-alphanumeric character in the middle",
			Input:    "test@application",
			Expected: expectedName,
		},
		{
			Name:     "Contains a sequence of non-alphanumeric character in the middle",
			Input:    "test%.-&application",
			Expected: expectedName,
		},
		{
			Name:     "Contains non-alphanumeric character at the beginning",
			Input:    "!test-application",
			Expected: expectedName,
		},
		{
			Name:     "Contains sequence of non-alphanumeric character at the beginning",
			Input:    "!@#$%test-application",
			Expected: expectedName,
		},
		{
			Name:     "Contains non-alphanumeric character at the end",
			Input:    "test-application%",
			Expected: expectedName,
		},
		{
			Name:     "Contains sequence of non-alphanumeric character at the end",
			Input:    "test-application&*^%",
			Expected: expectedName,
		},
		{
			Name:     "Contains non-alphanumeric character at the beginning, end and in the middle",
			Input:    "[test@application]",
			Expected: expectedName,
		},
		{
			Name:     "Contains multiple dash characters in the middle",
			Input:    "test---application",
			Expected: expectedName,
		},
		{
			Name:     "Contains multiple dash characters at the beginning",
			Input:    "---test-application",
			Expected: expectedName,
		},
		{
			Name:     "Contains multiple dash characters at the end",
			Input:    "test-application---",
			Expected: expectedName,
		},
		{
			Name:     "Contains multiple dash characters at the beginning, end and in the middle",
			Input:    "---test---application---",
			Expected: expectedName,
		},
		{
			Name:     "Contains a single dash character at the end",
			Input:    "test-application-",
			Expected: expectedName,
		},
		{
			Name:     "Contains a combination of dashes and non-dash non-alphanumeric characters at the end",
			Input:    "test-application&--*^-%",
			Expected: expectedName,
		},
		{
			Name:     "Contains upper case characters, non-alphanumeric characters and multiple dashes",
			Input:    "---tEsT@ApPlIcAtIoN.!$",
			Expected: expectedName,
		},
		{
			Name:     "Doesn't contain upper case characters, non-alphanumeric characters or any redundant dashes",
			Input:    "test-application",
			Expected: expectedName,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			defaultNormalizer := normalizer.DefaultNormalizer

			// when
			normalizedResult := defaultNormalizer(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, normalizedResult)
		})
	}
}
