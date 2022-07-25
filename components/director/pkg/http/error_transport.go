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

package http

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// NewErrorHandlerTransport missing godoc
func NewErrorHandlerTransport(roundTripper HTTPRoundTripper) *ErrorHandlerTransport {
	return &ErrorHandlerTransport{
		roundTripper: roundTripper,
	}
}

// ErrorHandlerTransport missing godoc
type ErrorHandlerTransport struct {
	roundTripper HTTPRoundTripper
}

// RoundTrip missing godoc
func (c *ErrorHandlerTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := c.roundTripper.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= http.StatusBadRequest {
		return nil, handleResponseError(request.Context(), response)
	}

	return response, nil
}

// Clone clones the underlying transport.
func (c *ErrorHandlerTransport) Clone() HTTPRoundTripper {
	return &ErrorHandlerTransport{
		roundTripper: c.roundTripper.Clone(),
	}
}

func (c *ErrorHandlerTransport) GetTransport() *http.Transport {
	return c.roundTripper.GetTransport()
}

func handleResponseError(ctx context.Context, response *http.Response) error {
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.C(ctx).Errorf("ReadCloser couldn't be closed: %v", err)
		}
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrap(err, "while reading response body bytes")
	}

	err = errors.Errorf("statusCode: %d Body: %s", response.StatusCode, body)
	if response.Request != nil {
		return errors.WithMessagef(err, "%s %s", response.Request.Method, response.Request.URL)
	}

	return errors.WithMessagef(err, "request failed")
}
