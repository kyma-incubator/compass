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

package log

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// CorrelationIDHeaders are the headers whose values will be taken as a correlation id for incoming requests
var CorrelationIDHeaders = []string{
	"X-Correlation-ID",
	"X-CorrelationID",
	"X-ForRequest-ID",
	"X-Request-ID",
	"X-Request-Id",
	"X-Vcap-Request-Id",
	"X-Broker-API-Request-Identity"}

type UUIDService interface {
	Generate() string
}

//RequestLogger returns middleware that setups request scoped logging
func RequestLogger(service UUIDService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			entry := C(ctx)
			if correlationID := CorrelationIDForRequest(r, service); correlationID != "" {
				entry = entry.WithField(FieldCorrelationID, correlationID)
			}
			ctx = ContextWithLogger(ctx, entry)
			r = r.WithContext(ctx)

			start := time.Now()

			remoteAddr := r.RemoteAddr
			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				remoteAddr = realIP
			}

			beforeLogger := entry.WithFields(logrus.Fields{
				"request": r.RequestURI,
				"method":  r.Method,
				"remote":  remoteAddr,
			})

			beforeLogger.Info("started handling request...")

			lrw := newLoggingResponseWriter(rw)
			next.ServeHTTP(lrw, r)

			duration := time.Since(start)

			afterLogger := entry.WithFields(logrus.Fields{
				"status_code": lrw.statusCode,
				"took":        duration,
			})

			afterLogger.Info("finished handling request...")
		})
	}
}

// CorrelationIDForRequest returns checks the http headers for any of the supported correlation id headers.
// The first that matches is taken as the correlation id. If none exists a new one is generated.
func CorrelationIDForRequest(request *http.Request, service UUIDService) string {
	ctx := request.Context()
	logger := C(ctx)
	for _, header := range CorrelationIDHeaders {
		headerValue := request.Header.Get(header)
		if headerValue != "" && headerValue != Configuration().BootstrapCorrelationID {
			return headerValue
		}
	}

	//this header setting is used both for http client outbound requests and for inbound broker api requests as middlewares.AddCorrelationIDToContext will look at headers
	correlationID, exists := logger.Data[FieldCorrelationID].(string)
	if exists && correlationID != Configuration().BootstrapCorrelationID {
		request.Header.Set(CorrelationIDHeaders[0], correlationID)
	} else {
		correlationID = service.Generate()
		request.Header.Set(CorrelationIDHeaders[0], correlationID)
	}

	return correlationID
}
