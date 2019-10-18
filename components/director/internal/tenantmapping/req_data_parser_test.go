package tenantmapping

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("returns ReqData based on request JSON", func(t *testing.T) {
		systemAuthID := uuid.New()
		username := "some-name"
		reqPayload := `{"Extra": {"client_id": "` + systemAuthID.String() + `", "name": "` + username + `"}, "Header": {"Client-Id-From-Certificate": ["` + systemAuthID.String() + `"]}}`
		req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", strings.NewReader(reqPayload))

		parser := NewReqDataParser()

		reqData, err := parser.Parse(req)

		require.NoError(t, err)
		require.Equal(t, systemAuthID.String(), reqData.Header.Get(ClientIDCertKey))
		require.Equal(t, systemAuthID.String(), reqData.Extra[ClientIDKey])
		require.Equal(t, username, reqData.Extra[UsernameKey])
	})

	t.Run("when request JSON does not contain Extra property the returned ReqData should have Extra property initialized", func(t *testing.T) {
		reqPayload := `{}`
		req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", strings.NewReader(reqPayload))

		parser := NewReqDataParser()

		reqData, err := parser.Parse(req)

		require.NoError(t, err)
		require.NotNil(t, reqData.Extra)
	})

	t.Run("returns error when request body is empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", bytes.NewReader(nil))

		parser := NewReqDataParser()

		_, err := parser.Parse(req)

		require.EqualError(t, err, "request body is empty")
	})

	t.Run("returns error when request body is empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", bytes.NewReader([]byte{1, 2, 3}))

		parser := NewReqDataParser()

		_, err := parser.Parse(req)

		require.EqualError(t, err, "while decoding request body: invalid character '\\x01' looking for beginning of value")
	})
}
