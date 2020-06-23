package error_presenter_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/error_presenter"

	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/sirupsen/logrus/hooks/test"

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
		assert.Equal(t, fmt.Sprintf("Unknown error: %s\n", errMsg), entry.Message)
		assert.Contains(t, err.Error(), "Internal Server Error")
		hook.Reset()
	})

	t.Run("Internal Error", func(t *testing.T) {
		//given
		customErr := apperrors.NewInternalError(errMsg)

		//when
		err := presenter.Do(context.TODO(), customErr)

		//then
		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Equal(t, fmt.Sprintf("Internal Server Error: Internal Server Error: %s", errMsg), entry.Message)
		assert.Contains(t, err.Error(), "Internal Server Error")
		hook.Reset()
	})

	t.Run("Invalid Data error", func(t *testing.T) {
		//given
		customErr := apperrors.NewInvalidDataError(errMsg)

		//when
		err := presenter.Do(context.TODO(), customErr)

		//then
		assert.EqualError(t, err, fmt.Sprintf("input: Invalid data [reason=%s]", errMsg))
		hook.Reset()
	})
}
