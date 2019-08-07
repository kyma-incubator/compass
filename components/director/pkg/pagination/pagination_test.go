package pagination

import (
	"encoding/base64"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestComputeOffset(t *testing.T) {
	//GIVEN
	offset := 2000
	testCases := []struct {
		Name            string
		InputCursor     string
		ExptectedOffset int
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			InputCursor:     convertIntToBase64String(offset),
			ExptectedOffset: offset,
			ExpectedErr:     nil,
		},
		{
			Name:            "Success with easter egg",
			InputCursor:     string(base64.StdEncoding.EncodeToString([]byte("DpKtJ4j9jDq" + strconv.Itoa(offset)))),
			ExptectedOffset: offset,
			ExpectedErr:     nil,
		},
		{
			Name:            "Success when page size is positive and offset is empty",
			InputCursor:     "",
			ExptectedOffset: 0,
			ExpectedErr:     nil,
		},
		{
			Name:            "Return error when cursor is negative",
			InputCursor:     convertIntToBase64String(-offset),
			ExptectedOffset: 0,
			ExpectedErr:     errors.New("cursor is not correct"),
		},
		{
			Name:            "Return error when input is not integer",
			InputCursor:     string(base64.StdEncoding.EncodeToString([]byte("foo-bar"))),
			ExptectedOffset: 0,
			ExpectedErr:     errors.New("cursor is not correct"),
		},
		{
			Name:            "Return error when input is not valid BASE64 string",
			InputCursor:     "Zm9vLWJh-1cg==",
			ExptectedOffset: 0,
			ExpectedErr:     errors.New("cursor is not correct"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//WHEN
			offset, err := DecodeOffsetCursor(testCase.InputCursor)

			//THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExptectedOffset, offset)
			}
		})
	}
}

func TestConvertOffsetToPageCursor(t *testing.T) {
	// GIVEN
	pageSize := 50
	offset := 50

	// WHEN
	nextPageCursor := EncodeNextOffsetCursor(pageSize, offset)

	// THEN
	require.Equal(t, "RHBLdEo0ajlqRHExMDA=", nextPageCursor)
}

func TestConvertOffsetLimitAndOrderedColumnToSQL(t *testing.T) {
	t.Run("Success converting Offset and Limit to SQL ", func(t *testing.T) {
		// WHEN
		sql, err := ConvertOffsetLimitAndOrderedColumnToSQL(5, 5, "id")

		//THEN
		require.NoError(t, err)
		assert.Equal(t, sql, ` ORDER BY "id" LIMIT 5 OFFSET 5`)
	})

	t.Run("Return error when column to order by is empty", func(t *testing.T) {
		// WHEN
		_, err := ConvertOffsetLimitAndOrderedColumnToSQL(5, 5, "")

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), `to use pagination you must provide column to order by`)
	})

	t.Run("Return error when page size is smaller than 1", func(t *testing.T) {
		// WHEN
		_, err := ConvertOffsetLimitAndOrderedColumnToSQL(-1, 5, "id")

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), `page size cannot be smaller than 1`)
	})

	t.Run("Return error when offset is smaller than 0", func(t *testing.T) {
		// WHEN
		_, err := ConvertOffsetLimitAndOrderedColumnToSQL(5, -1, "id")

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), `offset cannot be smaller than 0`)
	})
}

func TestDecodeAndEncodeCursorTogether(t *testing.T) {
	t.Run("Success encoding and then decoding cursor", func(t *testing.T) {
		//GIVEN
		offset := 4

		//WHEN
		cursor := EncodeNextOffsetCursor(offset, 0)
		decodedOffset, err := DecodeOffsetCursor(cursor)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, offset, decodedOffset)
	})

	t.Run("Success encoding and then decoding next page cursor", func(t *testing.T) {
		//GIVEN
		offset := 4
		pageSize := 5

		//WHEN
		cursor := EncodeNextOffsetCursor(offset, pageSize)
		decodedOffset, err := DecodeOffsetCursor(cursor)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, offset+pageSize, decodedOffset)
	})

	t.Run("Success decoding and then encoding cursor", func(t *testing.T) {
		//GIVEN
		cursor := "RHBLdEo0ajlqRHExMDA="

		//WHEN
		offset, err := DecodeOffsetCursor(cursor)
		encodedCusor := EncodeNextOffsetCursor(offset, 0)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, cursor, encodedCusor)
	})

	t.Run("Success decoding and then encoding incremented cursor", func(t *testing.T) {
		//GIVEN
		cursor := "RHBLdEo0ajlqRHExMDA="
		nextCursor := "RHBLdEo0ajlqRHExMDU="

		//WHEN
		offset, err := DecodeOffsetCursor(cursor)
		encodedCusor := EncodeNextOffsetCursor(offset, 5)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, nextCursor, encodedCusor)
	})
}

func convertIntToBase64String(number int) string {
	return string(base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(number))))
}
