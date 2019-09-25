package tenantmapping_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	t.Run("for request with certificate", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", strings.NewReader("{}"))
		req.Header.Set(tenantmapping.ClientIdFromCertificateHeader, "something")
		w := httptest.NewRecorder()

		sut := tenantmapping.NewHandler()
		sut.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		d := tenantmapping.Data{}
		err = json.Unmarshal(body, &d)
		require.NoError(t, err)

		extraMap, ok := d.Extra.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "9ac609e1-7487-4aa6-b600-0904b272b11f", extraMap["tenant"])
		assert.Equal(t, []interface{}{
			"application_read",
			"application_write",
			"runtime_read", "runtime_write",
			"label_definition_read",
			"label_definition_write",
			"health_checks_read"}, extraMap["scope"])
	})
}
