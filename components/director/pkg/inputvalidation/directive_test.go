package inputvalidation

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirective_Validate(t *testing.T) {
	validatorDirective := NewDirective()
	testErr := errors.New("testError")
	ctx := context.TODO()
	t.Run("Success", func(t *testing.T) {
		//TODO:
	})

	t.Run("object is not validatable", func(t *testing.T) {
		_, err := validatorDirective.Validate(ctx, nil, next("test", nil))
		//THEN
		require.Error(t, err)
	})

	t.Run("next resolver return err", func(t *testing.T) {
		//WHEN
		_, err := validatorDirective.Validate(ctx, nil, next(nil, testErr))
		//THEN
		require.Error(t, err)
	})
}

func next(obj interface{}, testErr error) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (i interface{}, e error) {
		return obj, testErr
	}
}
