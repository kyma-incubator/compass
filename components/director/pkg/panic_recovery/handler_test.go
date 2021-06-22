package panic_recovery_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/panic_recovery"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestPanicRecoveryMiddleware(t *testing.T) {
	middleware := panic_recovery.NewPanicRecoveryMiddleware()
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()

	middleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		panic(errors.New("test"))
	})).ServeHTTP(recorder, req)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Unrecovered panic")
}
