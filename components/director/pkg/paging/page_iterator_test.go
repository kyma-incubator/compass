package paging

import (
	"fmt"
	url_pkg "net/url"
	"strconv"
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"
)

const (
	testBaseURL       = "https://test.com"
	testSkipParam     = "$skip"
	testPageSizeParam = "$page"
)

func TestPageIterator(t *testing.T) {
	var (
		pageSize   = 3
		err        error
		i          = 0
		numPages   = 5
		additional = map[string]string{
			"$foo": "bar",
		}
		testErr = errors.New("test error")
	)
	tests := []struct {
		name          string
		pagingFunc    func(u string) (uint64, error)
		expectedError error
	}{
		{
			name: fmt.Sprintf("Successfully fetches all pages for %d pages", pageSize),
			pagingFunc: func(u string) (uint64, error) {
				defer func() { i++ }()

				url, err := url_pkg.ParseRequestURI(u)
				require.NoError(t, err)

				q := url.Query()
				skipValue, err := strconv.Atoi(q.Get("$skip"))
				require.NoError(t, err)
				require.Equal(t, i*pageSize, skipValue)

				pageSizeValue, err := strconv.Atoi(q.Get("$page"))
				require.NoError(t, err)
				require.Equal(t, pageSize, pageSizeValue)

				additionalValue := q.Get("$foo")
				require.Equal(t, "bar", additionalValue)

				if i < numPages {
					return uint64(pageSize), nil
				}
				return uint64(pageSize - 1), nil
			},
			expectedError: nil,
		},
		{
			name: "Returns an error when an error occurs during first paging func execution",
			pagingFunc: func(u string) (uint64, error) {
				return 0, testErr
			},
			expectedError: testErr,
		},
		{
			name: "Returns an error when an error occurs during other but the first paging func execution",
			pagingFunc: func(u string) (uint64, error) {
				url, err := url_pkg.ParseRequestURI(u)
				require.NoError(t, err)

				q := url.Query()
				skipValue, err := strconv.Atoi(q.Get("$skip"))
				require.NoError(t, err)
				if skipValue == 0 {
					return uint64(pageSize), nil
				}
				return 0, testErr
			},
			expectedError: testErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iterator := NewPageIterator(testBaseURL, testSkipParam, testPageSizeParam, additional, uint64(pageSize), test.pagingFunc)

			err = iterator.FetchAll()
			if test.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedError.Error())
			}
		})
	}
}
