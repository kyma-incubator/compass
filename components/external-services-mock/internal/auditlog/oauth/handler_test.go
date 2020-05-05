package oauth_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/auditlog/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Generate(t *testing.T) {
	//GIVEN
	secret, id := "secret", "id"
	req := httptest.NewRequest(http.MethodPost, "http://target.com/oauth/token", strings.NewReader(""))

	encodedAuthValue := base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", id, secret)))
	req.Header.Set("authorization", fmt.Sprintf("Basic %s", encodedAuthValue))

	h := oauth.NewHandler(secret, id)
	r := httptest.NewRecorder()

	//WHEN
	h.Generate(r, req)
	resp := r.Result()

	//THEN
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	var response oauth.TokenResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
}
