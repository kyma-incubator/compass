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
	rootAPI          = "ROOT_API"
	handlerEndpoint  = "HANDLER_ENDPOINT"
	tenantPathParam  = "TENANT_PATH_PARAM"
)

type config struct {
	TenantFetcherURL string
	RootAPI          string
	HandlerEndpoint  string
	TenantPathParam  string
}

func TestOnboardingHandler(t *testing.T) {
	// GIVEN
	config := loadConfig(t)

	// WHEN
	tenantID := "ad0bb8f2-7b44-4dd2-bce1-fa0c19169b72"
	endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), tenantID, 1)
	url := config.TenantFetcherURL + config.RootAPI + endpoint

	request, err := http.NewRequest(http.MethodPut, url, nil)
	require.NoError(t, err)

	response, err := http.DefaultClient.Do(request)

	// THEN
	require.NoError(t, err)
	require.Equal(t, response.StatusCode, http.StatusOK)
}

func TestDecommissioningHandler(t *testing.T) {
	// GIVEN
	config := loadConfig(t)

	// WHEN
	tenantID := "ad0bb8f2-7b44-4dd2-bce1-fa0c19169b72"
	endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), tenantID, 1)
	url := config.TenantFetcherURL + config.RootAPI + endpoint

	request, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)

	response, err := http.DefaultClient.Do(request)

	// THEN
	require.NoError(t, err)
	require.Equal(t, response.StatusCode, http.StatusOK)
}

func loadConfig(t *testing.T) config {
	config := config{
		TenantFetcherURL: os.Getenv(tenantFetcherURL),
		RootAPI:          os.Getenv(rootAPI),
		HandlerEndpoint:  os.Getenv(handlerEndpoint),
		TenantPathParam:  os.Getenv(tenantPathParam),
	}

	require.NotEmpty(t, config.TenantFetcherURL)
	require.NotEmpty(t, config.RootAPI)
	require.NotEmpty(t, config.HandlerEndpoint)
	require.NotEmpty(t, config.TenantPathParam)

	return config
}
