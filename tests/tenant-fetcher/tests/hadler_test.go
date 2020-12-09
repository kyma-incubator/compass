/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	tenantFetcherURL = "TENANT_FETCHER_URL"
	handlerEndpoint  = "HANDLER_ENDPOINT"
	tenantPathParam  = "TENANT_PATH_PARAM"
)

func TestOnboardingHandler(t *testing.T) {
	// GIVEN
	tenantFetcherURL := os.Getenv(tenantFetcherURL)
	handlerEndpoint := os.Getenv(handlerEndpoint)
	tenantPathParam := os.Getenv(tenantPathParam)

	require.NotEmpty(t, tenantFetcherURL)
	require.NotEmpty(t, handlerEndpoint)
	require.NotEmpty(t, tenantPathParam)

	// WHEN
	tenantID := "ad0bb8f2-7b44-4dd2-bce1-fa0c19169b72"
	endpoint := strings.Replace(handlerEndpoint, fmt.Sprintf("{%s}", tenantPathParam), tenantID, 1)
	url := tenantFetcherURL + endpoint

	request, err := http.NewRequest(http.MethodPut, url, nil)
	require.NoError(t, err)

	response, err := http.DefaultClient.Do(request)

	// THEN
	require.NoError(t, err)
	require.Equal(t, response.StatusCode, http.StatusOK)
}

func TestDecommissioningHandler(t *testing.T) {
	// GIVEN
	tenantFetcherURL := os.Getenv(tenantFetcherURL)
	handlerEndpoint := os.Getenv(handlerEndpoint)
	tenantPathParam := os.Getenv(tenantPathParam)

	require.NotEmpty(t, tenantFetcherURL)
	require.NotEmpty(t, handlerEndpoint)
	require.NotEmpty(t, tenantPathParam)

	// WHEN
	tenantID := "ad0bb8f2-7b44-4dd2-bce1-fa0c19169b72"
	endpoint := strings.Replace(handlerEndpoint, fmt.Sprintf("{%s}", tenantPathParam), tenantID, 1)
	url := tenantFetcherURL + endpoint

	request, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)

	response, err := http.DefaultClient.Do(request)

	// THEN
	require.NoError(t, err)
	require.Equal(t, response.StatusCode, http.StatusOK)
}
