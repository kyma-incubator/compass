package request

import (
	"io/ioutil"
	"net/http"
	urlpkg "net/url"
	"strings"

	"github.com/stretchr/testify/require"
)

const acceptHeader = "Accept"

func MakeRequestWithHeadersAndStatusExpect(t require.TestingT, httpClient *http.Client, url string, headers map[string][]string, expectedHTTPStatus int, ordServiceDefaultResponseType string) string {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	for key, values := range headers {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}

	response, err := httpClient.Do(request)

	require.NoError(t, err)
	require.Equal(t, expectedHTTPStatus, response.StatusCode)

	parsedURL, err := urlpkg.Parse(url)
	require.NoError(t, err)

	if !strings.Contains(parsedURL.Path, "/specification") {
		formatParam := parsedURL.Query().Get("$format")
		acceptHeader, acceptHeaderProvided := headers[acceptHeader]

		contentType := response.Header.Get("Content-Type")
		if formatParam != "" {
			require.Contains(t, contentType, formatParam)
		} else if acceptHeaderProvided && acceptHeader[0] != "*/*" {
			require.Contains(t, contentType, acceptHeader[0])
		} else {
			require.Contains(t, contentType, ordServiceDefaultResponseType)
		}
	}

	body, err := ioutil.ReadAll(response.Body)
	require.NoError(t, err)

	return string(body)
}
