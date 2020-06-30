package apperrors

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sirupsen/logrus/hooks/test"

	"github.com/stretchr/testify/assert"
)

func TestPresenter_ErrorPresenter(t *testing.T) {
	//given
	errMsg := "testErr"
	log, hook := test.NewNullLogger()
	presenter := NewPresenter(log)

	t.Run("Unknown error", func(t *testing.T) {
		//when
		err := presenter.Do(context.TODO(), errors.New(errMsg))

		//then
		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Equal(t, fmt.Sprintf("Unknown error: %s\n", errMsg), entry.Message)
		assert.Contains(t, err.Error(), "testErr")
		hook.Reset()
	})

	t.Run("Internal Error", func(t *testing.T) {
		//given
		customErr := Internal(errMsg)

		//when
		err := presenter.Do(context.TODO(), customErr)

		//then
		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Equal(t, fmt.Sprintf("Internal Server Error: %s", errMsg), entry.Message)
		assert.Equal(t, customErr.Code(), err.Extensions["error_code"])
		assert.Contains(t, err.Error(), "testErr")
		hook.Reset()
	})
}
