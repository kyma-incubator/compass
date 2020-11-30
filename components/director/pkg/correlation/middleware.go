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

	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

const (
	HeadersContextKey  = "CorrelationHeaders"
	RequestIDHeaderKey = "x-request-id"
)

// headerKeys are the expected headers that are used for distributed tracing.
var headerKeys = []string{"x-request-id", "x-b3-traceid", "x-b3-spanid", "x-b3-parentspanid", "x-b3-sampled", "x-b3-flags", "b3"}

type Headers map[string]string

//AttachCorrelationIDToContext returns middleware that attaches all headers used for tracing in the current request.
func AttachCorrelationIDToContext() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if correlationHeaders := HeadersForRequest(r); len(correlationHeaders) != 0 {
				ctx = context.WithValue(ctx, HeadersContextKey, correlationHeaders)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(rw, r)
		})
	}
}

// HeadersForRequest returns all http headers used for tracing of the passed request.
// If the request headers are not set, but are part of the context, they're set as headers as well.
// If the x-request-id header does not exists a new one is generated, and set as a header.
func HeadersForRequest(request *http.Request) Headers {
	reqHeaders := make(map[string]string)
	headersFromCtx := headersFromContext(request.Context())

	for _, headerKey := range headerKeys {
		headerValue := request.Header.Get(headerKey)
		if headerValue != "" {
			reqHeaders[headerKey] = headerValue
			continue
		}

		if headerValue, ok := headersFromCtx[headerKey]; ok {
			request.Header.Set(headerKey, headerValue)
			reqHeaders[headerKey] = headerValue
		}
	}

	if _, ok := reqHeaders[RequestIDHeaderKey]; !ok {
		newRequestID := uuid.New().String()
		reqHeaders[RequestIDHeaderKey] = newRequestID
		request.Header.Set(RequestIDHeaderKey, newRequestID)
	}

	return reqHeaders
}

func headersFromContext(ctx context.Context) Headers {
	var headersFromCtx map[string]string
	if ctx.Value(HeadersContextKey) != nil {
		var ok bool
		headersFromCtx, ok = ctx.Value(HeadersContextKey).(Headers)
		if !ok {
			logrus.Errorf("unexpected type of %s: %T, should be %T", HeadersContextKey, headersFromCtx, Headers{})
		}
	}

	return headersFromCtx
}
