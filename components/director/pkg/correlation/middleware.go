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

type contextKey string

// HeadersContextKey missing godoc
const HeadersContextKey contextKey = "CorrelationHeaders"

const (
	// RequestIDHeaderKey missing godoc
	RequestIDHeaderKey = "x-request-id"

	// TraceIDHeaderKey missing godoc
	TraceIDHeaderKey = "x-b3-traceid"

	// SpanIDHeaderKey missing godoc
	SpanIDHeaderKey = "x-b3-spanid"

	// ParentSpanIDHeaderKey missing godoc
	ParentSpanIDHeaderKey = "x-b3-parentspanid"
)

// headerKeys are the expected headers that are used for distributed tracing.
var headerKeys = []string{"x-request-id", "x-b3-traceid", "x-b3-spanid", "x-b3-parentspanid", "x-b3-sampled", "x-b3-flags", "b3"}

// Headers missing godoc
type Headers map[string]string

// CorrelationIDForRequest returns the correlation ID for the current request
func CorrelationIDForRequest(request *http.Request) string {
	return HeadersForRequest(request)[RequestIDHeaderKey]
}

// TraceIDForRequest returns the trace ID for the current request
func TraceIDForRequest(request *http.Request) string {
	return HeadersForRequest(request)[TraceIDHeaderKey]
}

// SpanIDForRequest returns the span ID for the current request
func SpanIDForRequest(request *http.Request) string {
	return HeadersForRequest(request)[SpanIDHeaderKey]
}

// ParentSpanIDForRequest returns the parent span ID for the current request
func ParentSpanIDForRequest(request *http.Request) string {
	return HeadersForRequest(request)[ParentSpanIDHeaderKey]
}

// AttachCorrelationIDToContext returns middleware that attaches all headers used for tracing in the current request.
func AttachCorrelationIDToContext() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if correlationHeaders := HeadersForRequest(r); len(correlationHeaders) != 0 {
				ctx = SaveToContext(ctx, correlationHeaders)
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
	headersFromCtx := HeadersFromContext(request.Context())

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

	// Context might have been enriched with additional headers (outside of those among the well known header keys array)
	// which should be attached as well
	for headerKey, headerValue := range headersFromCtx {
		if _, ok := reqHeaders[headerKey]; !ok {
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

// HeadersFromContext returns the headers for the provided context
func HeadersFromContext(ctx context.Context) Headers {
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

// CorrelationIDFromContext returns correlation ID from the given context
func CorrelationIDFromContext(ctx context.Context) string {
	return HeadersFromContext(ctx)[RequestIDHeaderKey]
}

// TraceIDFromContext returns trace ID from the given context
func TraceIDFromContext(ctx context.Context) string {
	return HeadersFromContext(ctx)[TraceIDHeaderKey]
}

// SpanIDFromContext returns span ID from the given context
func SpanIDFromContext(ctx context.Context) string {
	return HeadersFromContext(ctx)[SpanIDHeaderKey]
}

// ParentSpanIDFromContext returns parent span ID from the given context
func ParentSpanIDFromContext(ctx context.Context) string {
	return HeadersFromContext(ctx)[ParentSpanIDHeaderKey]
}

// SaveToContext saves the provided headers as correlation ID headers in the specified context
func SaveToContext(ctx context.Context, headers Headers) context.Context {
	return context.WithValue(ctx, HeadersContextKey, headers)
}

// SaveCorrelationKeyValuePairToContext saves the provided correlation key/value pair in the specified context
func SaveCorrelationKeyValuePairToContext(ctx context.Context, key, value string) context.Context {
	if key == "" || value == "" {
		return ctx
	}

	headers := HeadersFromContext(ctx)
	if headers == nil {
		headers = make(map[string]string)
	}

	headers[key] = value

	return SaveToContext(ctx, headers)
}

// AddCorrelationIDsToContext add all correlation IDs to the context
func AddCorrelationIDsToContext(ctx context.Context, correlationID, traceID, spanID, parentSpanID string) context.Context {
	ctx = SaveCorrelationKeyValuePairToContext(ctx, RequestIDHeaderKey, correlationID)
	ctx = SaveCorrelationKeyValuePairToContext(ctx, TraceIDHeaderKey, traceID)
	ctx = SaveCorrelationKeyValuePairToContext(ctx, SpanIDHeaderKey, spanID)
	ctx = SaveCorrelationKeyValuePairToContext(ctx, ParentSpanIDHeaderKey, parentSpanID)

	return ctx
}
