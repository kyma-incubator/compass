package inputvalidation

import (
	"context"
	"errors"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	returnErr error
}

func (ts testStruct) Validate() error {
	if ts.returnErr != nil {
		return validation.Errors{"field": ts.returnErr}
	}
	return nil
}

func TestDirective_Validate(t *testing.T) {
	validatorDirective := NewDirective()
	testErr := errors.New("testError")
	ctx := context.TODO()
	t.Run("success", func(t *testing.T) {
		//GIVEN
		ts := testStruct{}
		// WHEN
		result, err := validatorDirective.Validate(ctx, ts, next(ts, nil))
		// THEN
		require.NoError(t, err)
		require.Equal(t, ts, result)
	})

	t.Run("object is invalid", func(t *testing.T) {
		//GIVEN
		ts := testStruct{returnErr: testErr}
		// WHEN
		result, err := validatorDirective.Validate(ctx, ts, next(ts, nil))
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "field=testError")
		assert.Contains(t, err.Error(), "Invalid data testStruct")
		require.Equal(t, ts, result)
	})

	t.Run("object is not validatable", func(t *testing.T) {
		_, err := validatorDirective.Validate(ctx, nil, next("test", nil))
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "object is not validatable")
	})

	t.Run("next resolver return err", func(t *testing.T) {
		// WHEN
		_, err := validatorDirective.Validate(ctx, nil, next(nil, testErr))
		// THEN
		require.Error(t, err)
	})
}

func next(obj interface{}, testErr error) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (i interface{}, e error) {
		return obj, testErr
	}
}
