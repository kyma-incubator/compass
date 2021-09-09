package apperrors_test

import (
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	errors1 "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternalErrorFrom(t *testing.T) {
	//GIVEN
	err := errors.New("Error")

	//WHEN
	err = apperrors.InternalErrorFrom(err, "Very bad error")
	//THEN
	require.Error(t, err)
	assert.EqualError(t, err, "Internal Server Error: Very bad error: Error")
}

func TestErrorMessage(t *testing.T) {
	t.Run("Single field", func(t *testing.T) {
		//GIVEN
		validationErrors := map[string]error{
			"field1": errors.New("field1 is invalid"),
		}

		//WHEN
		err := apperrors.NewInvalidDataErrorWithFields(validationErrors, "testObject")

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Invalid data testObject [field1=field1 is invalid]")
	})
	t.Run("Multiple fields", func(t *testing.T) {
		//GIVEN
		validationErrors := map[string]error{
			"field2": errors.New("field2 is invalid"),
			"field3": errors.New("field3 is invalid"),
			"field1": errors.New("field1 is invalid"),
		}

		//WHEN
		err := apperrors.NewInvalidDataErrorWithFields(validationErrors, "testObject")

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Invalid data testObject [field1=field1 is invalid; field2=field2 is invalid; field3=field3 is invalid]")
	})
}

func TestError_Is(t *testing.T) {
	t.Run("The same error", func(t *testing.T) {
		//GIVEN
		err, ok := apperrors.NewInternalError("test error").(apperrors.Error)
		require.True(t, ok)

		err1 := apperrors.NewInternalError("test error1")
		//WHEN
		equal := err.Is(err1)

		//THEN
		require.True(t, equal)
	})

	t.Run("Not the same error", func(t *testing.T) {
		//GIVEN
		err, ok := apperrors.NewInternalError("test error").(apperrors.Error)
		require.True(t, ok)

		err1 := apperrors.NewNotFoundError(resource.Application, "test error1")
		//WHEN
		equal := err.Is(err1)

		//THEN
		require.False(t, equal)
	})

	t.Run("Basic error", func(t *testing.T) {
		//GIVEN
		err, ok := apperrors.NewInternalError("test error").(apperrors.Error)
		require.True(t, ok)

		err1 := errors.New("very bad error")
		//WHEN
		equal := err.Is(err1)

		//THEN
		require.False(t, equal)
	})
}

func TestErrors_As(t *testing.T) {
	//GIVEN

	internalErr := apperrors.NewInternalError("very bad error")
	chainedErr := errors1.Wrap(internalErr, "wrap no 1")
	chainedErr = errors1.Wrap(chainedErr, "wrap no 2")
	chainedErr = errors1.Wrap(chainedErr, "wrap no 3")

	//WHEN
	err := apperrors.Error{}
	errors.As(chainedErr, &err)

	//THEN
	assert.Equal(t, internalErr, err)
}

func TestErrors_Is(t *testing.T) {
	type testCase struct {
		name           string
		input          error
		testFunc       func(err error) bool
		expectedResult bool
	}

	t.Run("Errors Match", func(t *testing.T) {
		testCases := []testCase{
			{
				name:           "TenantRequired",
				input:          apperrors.NewTenantRequiredError(),
				testFunc:       apperrors.IsTenantRequired,
				expectedResult: true,
			},
			{
				name:           "KeyDoesNotExist",
				input:          apperrors.NewKeyDoesNotExistError("magic key"),
				testFunc:       apperrors.IsKeyDoesNotExist,
				expectedResult: true,
			},
			{
				name:           "CannotReadTenant",
				input:          apperrors.NewCannotReadTenantError(),
				testFunc:       apperrors.IsCannotReadTenant,
				expectedResult: true,
			},
			{
				name:           "ValueNotFoundInConfiguration",
				input:          apperrors.NewValueNotFoundInConfigurationError(),
				testFunc:       apperrors.IsValueNotFoundInConfiguration,
				expectedResult: true,
			},
			{
				name:           "IsNotUniqueError",
				input:          apperrors.NewNotUniqueError(resource.AutomaticScenarioAssigment),
				testFunc:       apperrors.IsNotUniqueError,
				expectedResult: true,
			},
			{
				name:           "IsTenantNotFoundError",
				input:          apperrors.NewTenantNotFoundError("external-tenant"),
				testFunc:       apperrors.IsTenantNotFoundError,
				expectedResult: true,
			},
			{
				name:           "IsTenantRequired",
				input:          apperrors.NewTenantRequiredError(),
				testFunc:       apperrors.IsTenantRequired,
				expectedResult: true,
			},
			{
				name:           "IsNotFoundError",
				input:          apperrors.NewNotFoundError(resource.Label, "scenarios"),
				testFunc:       apperrors.IsNotFoundError,
				expectedResult: true,
			},
			{
				name:           "IsInvalidStatusCondition",
				input:          apperrors.NewInvalidStatusCondition(resource.Application),
				testFunc:       apperrors.IsInvalidStatusCondition,
				expectedResult: true,
			},
			{
				name:           "IsCannotUpdateObjectInManyBundlesError",
				input:          apperrors.NewCannotUpdateObjectInManyBundles(),
				testFunc:       apperrors.IsCannotUpdateObjectInManyBundlesError,
				expectedResult: true,
			},
		}
		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				//WHEN
				output := testCase.testFunc(testCase.input)

				//THEN
				assert.Equal(t, testCase.expectedResult, output)
			})
		}
	})

	t.Run("Errors not match", func(t *testing.T) {
		err := apperrors.NewInternalError("very bad error")
		testCases := []testCase{
			{
				name:           "TenantRequired",
				input:          err,
				testFunc:       apperrors.IsTenantRequired,
				expectedResult: false,
			},
			{
				name:           "KeyDoesNotExist",
				input:          err,
				testFunc:       apperrors.IsKeyDoesNotExist,
				expectedResult: false,
			},
			{
				name:           "KeyDoesNotExist - basic error",
				input:          errors.New("test error"),
				testFunc:       apperrors.IsKeyDoesNotExist,
				expectedResult: false,
			},
			{
				name:           "CannotReadTenant",
				input:          err,
				testFunc:       apperrors.IsCannotReadTenant,
				expectedResult: false,
			},
			{
				name:           "ValueNotFoundInConfiguration",
				input:          err,
				testFunc:       apperrors.IsValueNotFoundInConfiguration,
				expectedResult: false,
			},
			{
				name:           "IsNotUniqueError",
				input:          err,
				testFunc:       apperrors.IsNotUniqueError,
				expectedResult: false,
			},
			{
				name:           "IsTenantNotFoundError",
				input:          err,
				testFunc:       apperrors.IsTenantNotFoundError,
				expectedResult: false,
			},
			{
				name:           "IsTenantRequired",
				input:          err,
				testFunc:       apperrors.IsTenantRequired,
				expectedResult: false,
			},
			{
				name:           "IsNotFoundError",
				input:          err,
				testFunc:       apperrors.IsNotFoundError,
				expectedResult: false,
			},
			{
				name:           "IsInvalidStatusCondition",
				input:          err,
				testFunc:       apperrors.IsInvalidStatusCondition,
				expectedResult: false,
			},
			{
				name:           "IsCannotUpdateObjectInManyBundles",
				input:          err,
				testFunc:       apperrors.IsCannotUpdateObjectInManyBundlesError,
				expectedResult: false,
			},
		}
		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				//WHEN
				output := testCase.testFunc(testCase.input)

				//THEN
				assert.Equal(t, testCase.expectedResult, output)
			})
		}
	})
}
