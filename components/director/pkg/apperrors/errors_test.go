package apperrors

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var isErrorFunctionsTests = []struct {
	name string
	err  error
	fn   func(err error) bool
}{
	{name: "IsNotFoundError", err: NewNotFoundError("test"), fn: IsNotFoundError},
	{name: "IsNotUnique", err: NewNotUniqueError("test"), fn: IsNotUnique},
	{name: "IsKeyDoesNotExist", err: NewKeyDoesNotExistError("test"), fn: IsKeyDoesNotExist},
	{name: "IsInvalidCast", err: NewInvalidStringCastError(), fn: IsInvalidCast},
	{name: "IsConstraintViolation", err: NewConstraintViolationError("test"), fn: IsConstraintViolation},
	{name: "IsValueNotFoundInConfiguration", err: NewValueNotFoundInConfigurationError(), fn: IsValueNotFoundInConfiguration},
	{name: "IsNoScopesInContext", err: NewNoScopesInContextError(), fn: IsNoScopesInContext},
	{name: "IsRequiredScopesNotDefined", err: NewRequiredScopesNotDefinedError(), fn: IsRequiredScopesNotDefined},
	{name: "IsInsufficientScopes", err: NewInsufficientScopesError([]string{"test"}, []string{"test"}), fn: IsInsufficientScopes},
	{name: "IsNoTenant", err: NewNoTenantError(), fn: IsNoTenant},
	{name: "IsEmptyTenant", err: NewEmptyTenantError(), fn: IsEmptyTenant},
}

func TestIsErrorFunctions(t *testing.T) {
	// GIVEN
	differentError := errors.New("test")

	for _, isFnTest := range isErrorFunctionsTests {
		t.Run(isFnTest.name, func(t *testing.T) {
			testedError := isFnTest.err
			wrappedError := errors.Wrap(testedError, "wrapped text")
			multiWrappedError := errors.Wrap(wrappedError, "multi wrapped")

			testCases := []struct {
				Name           string
				Error          error
				expectedResult bool
			}{
				{
					Name:           "Unwrapped error",
					Error:          testedError,
					expectedResult: true,
				},
				{
					Name:           "Wrapped error",
					Error:          wrappedError,
					expectedResult: true,
				},
				{
					Name:           "Multi wrapped error",
					Error:          multiWrappedError,
					expectedResult: true,
				},
				{
					Name:           "Different error",
					Error:          differentError,
					expectedResult: false,
				},
			}

			for _, testCase := range testCases {
				t.Run(testCase.Name, func(t *testing.T) {
					// WHEN
					result := isFnTest.fn(testCase.Error)

					// THEN
					assert.Equal(t, testCase.expectedResult, result)
				})
			}
		})
	}
}
