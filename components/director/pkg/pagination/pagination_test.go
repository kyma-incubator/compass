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
			InputCursor:     string(base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(offset)))),
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
			InputCursor:     string(base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(-offset)))),
			ExptectedOffset: 0,
			ExpectedErr:     errors.New("Cursor is not correct"),
		},
		{
			Name:            "Return error when input is not integer",
			InputCursor:     string(base64.StdEncoding.EncodeToString([]byte("foo-bar"))),
			ExptectedOffset: 0,
			ExpectedErr:     errors.New("Cursor is not correct"),
		},
		{
			Name:            "Return error when input is not valid BASE64 string",
			InputCursor:     "Zm9vLWJh-1cg==",
			ExptectedOffset: 0,
			ExpectedErr:     errors.New("Cursor is not correct"),
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
	nextPageCursor := EncodeOffsetCursor(pageSize, offset)

	// THEN
	require.Equal(t, "MTAw", nextPageCursor)
}
