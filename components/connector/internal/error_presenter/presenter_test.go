package error_presenter_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"

	"github.com/kyma-incubator/compass/components/connector/internal/error_presenter"
	"github.com/kyma-incubator/compass/components/connector/internal/uid"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestPresenter_ErrorPresenter(t *testing.T) {
	//given
	errMsg := "testErr"
	uidSvc := uid.NewService()
	log, hook := test.NewNullLogger()
	presenter := error_presenter.NewPresenter(log, uidSvc)

	t.Run("Unknown error", func(t *testing.T) {
		//when
		err := presenter.Do(context.TODO(), errors.New(errMsg))

		//then
		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Equal(t, fmt.Sprintf("Unknown error: %s", errMsg), entry.Message)
		assert.Contains(t, err.Error(), "Internal Server Error")
		hook.Reset()
	})

	t.Run("Internal Error", func(t *testing.T) {
		//given
		customErr := apperrors.Internal(errMsg)

		//when
		err := presenter.Do(context.TODO(), customErr)

		//then
		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Equal(t, fmt.Sprintf("Internal Server Error: %s", errMsg), entry.Message)
		assert.Contains(t, err.Error(), "Internal Server Error")
		hook.Reset()
	})

	t.Run("Not Authenticated", func(t *testing.T) {
		//given
		customErr := apperrors.NotAuthenticated(errMsg)

		//when
		err := presenter.Do(context.TODO(), customErr)

		//then
		assert.EqualError(t, err, fmt.Sprint("input: ", errMsg))
		hook.Reset()
	})
}
