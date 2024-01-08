package util

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestStringToInt64(t *testing.T) {
	uuidString := "3a31599c-7a86-455d-83db-0014a7d459e8"
	uuidExpected := int64(491666746554389322)
	name := "some name here"
	nameExpected := int64(-6756220559490625949)

	t.Run("Success with UUID as string to StringToInt64", func(t *testing.T) {
		result, err := StringToInt64(uuidString)
		require.NoError(t, err)
		require.Equal(t, uuidExpected, result)
	})

	t.Run("Success with regular string to StringToInt64", func(t *testing.T) {
		result, err := StringToInt64(name)
		require.NoError(t, err)
		require.Equal(t, nameExpected, result)
	})

	t.Run("Fail when empty string send to StringToInt64", func(t *testing.T) {
		result, err := StringToInt64("")
		require.ErrorContains(t, err, "input cannot be empty")
		require.Equal(t, int64(0), result)
	})

	t.Run("Check conversiopn for uniqunes", func(t *testing.T) {
		results := map[int64]string{}
		for i := 0; i < 1000000; i++ {
			uuidIDSring := uuid.New().String()
			result, err := StringToInt64(uuidIDSring)
			require.NoError(t, err)
			results[result] = uuidIDSring
		}
		require.Equal(t, 1000000, len(results))
	})
}
