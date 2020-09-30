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
	"crypto/tls"
	"net/http"
	"sync"
	"time"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewHTTPTransport(config *Config) *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipSSLValidation,
		},
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		MaxIdleConns:          config.MaxIdleConns,
		IdleConnTimeout:       config.IdleConnTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
	}
}

func NewClient(timeout time.Duration, transport http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

func NewSecuredHTTPClient(timeout time.Duration, roundTripper HTTPRoundTripper, provider TokenProvider) (*http.Client, error) {
	transport := &SecuredTransport{
		roundTripper:  roundTripper,
		tokenProvider: provider,
		lock:          sync.Mutex{},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}
