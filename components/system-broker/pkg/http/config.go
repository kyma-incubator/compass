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
	"fmt"
	"time"
)

type Config struct {
	Timeout               time.Duration `mapstructure:"timeout" description:"timeout specifies a time limit for the request. The timeout includes connection time, any redirects, and reading the response body"`
	TLSHandshakeTimeout   time.Duration `mapstructure:"tls_handshake_timeout"`
	IdleConnTimeout       time.Duration `mapstructure:"idle_conn_timeout"`
	ResponseHeaderTimeout time.Duration `mapstructure:"response_header_timeout"`
	DialTimeout           time.Duration `mapstructure:"dial_timeout"`
	ExpectContinueTimeout time.Duration `mapstructure:"expect_continue_timeout"`

	MaxIdleConns      int  `mapstructure:"max_idle_cons"`
	SkipSSLValidation bool `mapstructure:"skip_ssl_validation" description:"whether to skip ssl verification when making calls to external services"`

	ForwardHeaders []string `mapstructure:"forward_headers"`
}

func DefaultConfig() *Config {
	return &Config{
		Timeout:               time.Second * 15,
		TLSHandshakeTimeout:   time.Second * 10,
		IdleConnTimeout:       time.Second * 10,
		ResponseHeaderTimeout: time.Second * 10,
		DialTimeout:           time.Second * 10,
		ExpectContinueTimeout: time.Second * 2,
		MaxIdleConns:          90,
		SkipSSLValidation:     false,
		ForwardHeaders:        []string{},
	}
}

func (s *Config) Validate() error {
	if s.Timeout < 0 {
		return fmt.Errorf("validate httpclient settings: timeout should be >= 0")
	}
	if s.TLSHandshakeTimeout < 0 {
		return fmt.Errorf("validate httpclient settings: tls_handshake_timeout should be >= 0")
	}
	if s.IdleConnTimeout < 0 {
		return fmt.Errorf("validate httpclient settings: idle_conn_timeout should be >= 0")
	}
	if s.ResponseHeaderTimeout < 0 {
		return fmt.Errorf("validate httpclient settings: response_header_timeout should be >= 0")
	}
	if s.DialTimeout < 0 {
		return fmt.Errorf("validate httpclient settings: dial_timeout should be >= 0")
	}
	if s.ExpectContinueTimeout < 0 {
		return fmt.Errorf("validate httpclient settings: expect_continue_timeout should be >= 0")
	}
	if s.MaxIdleConns < 0 {
		return fmt.Errorf("validate httpclient settings: max_idle_cons should be >= 0")
	}
	return nil
}
