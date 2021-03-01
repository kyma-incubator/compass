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

package director_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql"

	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestClient_UpdateStatus_WhenNewRequestWithContextFails_ShouldReturnError(t *testing.T) {
	client, err := director.NewClient("", &graphql.Config{}, &http.Client{})
	require.NoError(t, err)

	err = client.UpdateOperation(nil, &director.Request{})
	require.Error(t, err)
}

func TestClient_UpdateStatus_WhenClientDoFails_ShouldReturnError(t *testing.T) {
	mockedErr := errors.New("mocked error")
	httpClient := http.Client{
		Transport: mockedTransport{resp: nil, err: mockedErr},
	}

	client, err := director.NewClient("", &graphql.Config{}, &httpClient)
	require.NoError(t, err)

	err = client.UpdateOperation(context.Background(), &director.Request{})

	require.Error(t, err)
	require.Contains(t, err.Error(), mockedErr.Error())
}

func TestClient_UpdateStatus_WhenResponseStatusCodeIsNotOK_ShouldReturnError(t *testing.T) {
	httpClient := http.Client{
		Transport: mockedTransport{resp: &http.Response{
			StatusCode: http.StatusBadRequest,
		}, err: nil},
	}

	client, err := director.NewClient("", &graphql.Config{}, &httpClient)
	require.NoError(t, err)

	err = client.UpdateOperation(context.Background(), &director.Request{})

	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected status code")
}

func TestClient_UpdateStatus_WhenRequestIsSuccessful_ShouldNotReturnError(t *testing.T) {
	httpClient := http.Client{
		Transport: mockedTransport{resp: &http.Response{
			StatusCode: http.StatusOK,
		}, err: nil},
	}

	client, err := director.NewClient("", &graphql.Config{}, &httpClient)
	require.NoError(t, err)

	err = client.UpdateOperation(context.Background(), &director.Request{})
	require.NoError(t, err)
}

type mockedTransport struct {
	resp *http.Response
	err  error
}

func (m mockedTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return m.resp, m.err
}
