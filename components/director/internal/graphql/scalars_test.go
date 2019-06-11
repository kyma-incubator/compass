package graphql

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimestamp_UnmarshalGQL(t *testing.T) {
	//given
	var timestamp Timestamp
	fixTime := "2002-10-02T10:00:00-05:00"
	parsedTime, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	assert.NoError(t, err)
	expectedTimestamp := Timestamp(parsedTime)

	//when
	err = timestamp.UnmarshalGQL(fixTime)

	//then
	require.NoError(t, err)
	assert.Equal(t, expectedTimestamp, timestamp)
}

func TestTenant_UnmarshalGQL(t *testing.T) {
	//given
	var tenant Tenant
	fixTenant := "tenant1"
	expectedTenant := Tenant("tenant1")

	//when
	err := tenant.UnmarshalGQL(fixTenant)

	//then
	require.NoError(t, err)
	assert.Equal(t, expectedTenant, tenant)
}

func TestTenant_MarshalGQL(t *testing.T) {
	//given
	fixTenant := Tenant("tenant1")
	expectedTenant := `{"tenant":"tenant1"}`
	buf := bytes.Buffer{}
	//when
	fixTenant.MarshalGQL(&buf)

	//then
	assert.Equal(t, expectedTenant, buf.String())
}

func TestLabels_UnmarshalGQL(t *testing.T) {
	//given
	l := Labels{}
	fixLabels := map[string]interface{}{
		"label1": []string{"val1", "val2"},
	}
	expectedLabels := Labels{
		"label1": []string{"val1", "val2"},
	}

	//when
	err := l.UnmarshalGQL(fixLabels)

	//then
	require.NoError(t, err)
	assert.Equal(t, l, expectedLabels)
}

func TestLabels_MarshalGQL(t *testing.T) {
	//given
	fixLabels := Labels{
		"label1": []string{"val1", "val2"},
	}
	expectedLabels := `{"label1":["val1","val2"]}`
	buf := bytes.Buffer{}

	//when
	fixLabels.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, buf.String(), expectedLabels)
}

func TestAnnotations_UnmarshalGQL(t *testing.T) {
	//given
	a := Annotations{}
	fixAnnotations := map[string]interface{}{
		"annotation1": "val1",
	}
	expectedAnnotations := Annotations{
		"annotation1": "val1",
	}

	//when
	err := a.UnmarshalGQL(fixAnnotations)

	//then
	require.NoError(t, err)
	assert.Equal(t, a, expectedAnnotations)
}

func TestAnnotations_MarshalGQL(t *testing.T) {
	//given
	fixAnnotations1 := Annotations{
		"annotation1": 123,
	}
	fixAnnotations2 := Annotations{
		"annotation2": []string{"val1", "val2"},
	}

	expectedAnnotations1 := `{"annotation1":123}`
	expectedAnnotations2 := `{"annotation2":["val1","val2"]}`
	buf1 := bytes.Buffer{}
	buf2 := bytes.Buffer{}
	//when
	fixAnnotations1.MarshalGQL(&buf1)
	fixAnnotations2.MarshalGQL(&buf2)

	//then
	assert.NotNil(t, buf1)
	assert.NotNil(t, buf2)
	assert.Equal(t, expectedAnnotations1, buf1.String())
	assert.Equal(t, expectedAnnotations2, buf2.String())
}

func TestHttpHeaders_UnmarshalGQL(t *testing.T) {
	//given
	h := HttpHeaders{}
	fixHeaders := map[string]interface{}{
		"header1": []string{"val1", "val2"},
	}
	expectedHeaders := HttpHeaders{
		"header1": []string{"val1", "val2"},
	}

	//when
	err := h.UnmarshalGQL(fixHeaders)

	//then
	require.NoError(t, err)
	assert.Equal(t, h, expectedHeaders)
}

func TestHttpHeaders_MarshalGQL(t *testing.T) {
	//given
	fixHeaders := HttpHeaders{
		"header": []string{"val1", "val2"},
	}

	expectedHeaders := `{"header":["val1","val2"]}`
	buf := bytes.Buffer{}
	//when
	fixHeaders.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedHeaders, buf.String())
}

func TestQueryParams_UnmarshalGQL(t *testing.T) {
	//given
	p := QueryParams{}
	fixParams := map[string]interface{}{
		"param1": []string{"val1", "val2"},
	}
	expectedParams := QueryParams{
		"param1": []string{"val1", "val2"},
	}

	//when
	err := p.UnmarshalGQL(fixParams)

	//then
	require.NoError(t, err)
	assert.Equal(t, p, expectedParams)
}

func TestQueryParams_MarshalGQL(t *testing.T) {
	//given
	fixParams := QueryParams{
		"param": []string{"val1", "val2"},
	}

	expectedParams := `{"param":["val1","val2"]}`
	buf := bytes.Buffer{}
	//when
	fixParams.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedParams, buf.String())
}

func TestClob_UnmarshalGQL(t *testing.T) {
	//given
	var clob CLOB
	fixClob := "very_big_clob"
	expectedClob := CLOB("very_big_clob")

	//when
	err := clob.UnmarshalGQL(fixClob)

	//then
	require.NoError(t, err)
	assert.Equal(t, clob, expectedClob)
}

func TestCLOB_MarshalGQL(t *testing.T) {
	//given
	fixClob := CLOB("very_big_clob")

	expectedClob := `{"CLOB":"very_big_clob"}`
	buf := bytes.Buffer{}
	//when
	fixClob.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedClob, buf.String())
}

func TestPageCursor_UnmarshalGQL(t *testing.T) {
	//given
	var pageCursor PageCursor
	fixCursor := "cursor"
	expectedCursor := PageCursor("cursor")

	//when
	err := pageCursor.UnmarshalGQL(fixCursor)

	//then
	require.NoError(t, err)
	assert.Equal(t, pageCursor, expectedCursor)
}

func TestPageCursor_MarshalGQL(t *testing.T) {
	//given
	fixCursor := PageCursor("cursor")

	expectedCursor := `{"pageCursor":"cursor"}`
	buf := bytes.Buffer{}
	//when
	fixCursor.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedCursor, buf.String())
}
