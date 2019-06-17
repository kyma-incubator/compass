package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTenant_UnmarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errmsg   string
		expected Tenant
	}{
		//given
		"correct input": {
			input:    "tenant",
			err:      false,
			expected: Tenant("tenant"),
		},
		"error: input is nil": {
			input:  nil,
			err:    true,
			errmsg: "input should not be nil",
		},
		"error: invalid input": {
			input:  123,
			err:    true,
			errmsg: "unexpected input type: int, should be string",
		},
	} {
		t.Run(name, func(t *testing.T) {
			//when
			var tenant Tenant
			err := tenant.UnmarshalGQL(tc.input)

			//then
			if tc.err {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errmsg)
				assert.Empty(t, tenant)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, tenant)
			}
		})
	}
}

func TestTenant_MarshalGQL(t *testing.T) {
	//given
	fixTenant := Tenant("tenant1")
	expectedTenant := `"tenant1"`
	buf := bytes.Buffer{}

	//when
	fixTenant.MarshalGQL(&buf)

	//then
	assert.Equal(t, expectedTenant, buf.String())
}
