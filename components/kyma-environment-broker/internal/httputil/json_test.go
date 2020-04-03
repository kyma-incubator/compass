package httputil_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/httputil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeSuccess(t *testing.T) {
	// given
	rw := httptest.NewRecorder()

	// when
	err := httputil.JSONEncode(rw, fixData())

	// then
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
	assert.JSONEq(t, `{
		  "field_int": 1,
		  "field_string": "andrzej"
		}`, rw.Body.String())
}

func TestEncodeWithCodeSuccess(t *testing.T) {
	// given
	rw := httptest.NewRecorder()

	// when
	err := httputil.JSONEncodeWithCode(rw, fixData(), http.StatusCreated)

	// then
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rw.Code)
	assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
	assert.JSONEq(t, `{
		  "field_int": 1,
		  "field_string": "andrzej"
		}`, rw.Body.String())

}

type dataToEncode struct {
	FieldInt    int    `json:"field_int"`
	FieldString string `json:"field_string,omitempty"`
}

func fixData() dataToEncode {
	return dataToEncode{
		FieldInt:    1,
		FieldString: "andrzej",
	}
}
