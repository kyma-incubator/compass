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
	"net/http"
)

// HTTPRoundTripper missing godoc
//go:generate mockery --name=HTTPRoundTripper --output=automock --outpkg=automock --case=underscore --disable-version-string
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . HTTPRoundTripper
type HTTPRoundTripper interface {
	RoundTrip(*http.Request) (*http.Response, error)
	Clone() HTTPRoundTripper
	GetTransport() *http.Transport
}

type httpTransportWrapper struct {
	tr *http.Transport
}

// NewHTTPTransportWrapper wraps http transport
func NewHTTPTransportWrapper(tr *http.Transport) HTTPRoundTripper {
	return &httpTransportWrapper{
		tr: tr,
	}
}

func (h *httpTransportWrapper) RoundTrip(request *http.Request) (*http.Response, error) {
	return h.tr.RoundTrip(request)
}

func (h *httpTransportWrapper) Clone() HTTPRoundTripper {
	return &httpTransportWrapper{
		tr: h.tr.Clone(),
	}
}

func (h *httpTransportWrapper) GetTransport() *http.Transport {
	return h.tr
}
