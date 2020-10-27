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

package correlation

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	ContextField       = "CorrelationID"
	RequestIDHeaderKey = "x-request-id"
)

// CorrelationIDHeaders are the headers whose values will be taken as a correlation id for incoming requests
var CorrelationIDHeaders = []string{
	RequestIDHeaderKey,
	"X-Request-ID",
	"X-Request-Id",
	"X-Correlation-ID",
	"X-CorrelationID",
	"X-ForRequest-ID"}

//AttachCorrelationIDToContext returns middleware that attaches a correlation ID to the current request
func AttachCorrelationIDToContext() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if correlationID := CorrelationIDForRequest(r); correlationID != "" {
				ctx = context.WithValue(ctx, ContextField, correlationID)
				r.Header.Set(RequestIDHeaderKey, correlationID)
			}

			next.ServeHTTP(rw, r)
		})
	}
}

// CorrelationIDForRequest checks the http headers for any of the supported correlation id headers.
// The first that matches is taken as the correlation id. If none exists a new one is generated and set as a header.
func CorrelationIDForRequest(request *http.Request) string {
	for _, header := range CorrelationIDHeaders {
		headerValue := request.Header.Get(header)
		if headerValue != "" {
			return headerValue
		}
	}

	correlationID, ok := request.Context().Value(ContextField).(string)
	if !ok || correlationID == "" {
		correlationID = uuid.New().String()
	}

	request.Header.Set(CorrelationIDHeaders[0], correlationID)
	return correlationID
}
