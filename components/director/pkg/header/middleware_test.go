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

package header_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/header"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/stretchr/testify/assert"
)

const expectedRequestID = "123"

func TestAttachHeadersToContext(t *testing.T) {
	// given
	handler := header.AttachHeadersToContext()

	t.Run("request headers are persisted in context", func(t *testing.T) {

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headersFromContext, ok := r.Context().Value(header.ContextKey).(http.Header)
			assert.True(t, ok)

			actual := headersFromContext.Get(correlation.RequestIDHeaderKey)
			assert.Equal(t, actual, expectedRequestID)

			headerFromRequest := r.Header.Get(correlation.RequestIDHeaderKey)
			assert.Equal(t, headerFromRequest, expectedRequestID)
		})

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(correlation.RequestIDHeaderKey, expectedRequestID)

		handler(nextHandler).ServeHTTP(httptest.NewRecorder(), req)
	})

}
