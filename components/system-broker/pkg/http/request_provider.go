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
	"bytes"
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"io"
	"net/http"
)


type RequestProvider interface {
	Provide(ctx context.Context, input RequestInput) (*http.Request, error)
}

func NewRequestProvider(uidsrv UUIDService) RequestProvider {
	return &RequestProviderImpl{
		uuidService: uidsrv,
	}
}

type UUIDService interface {
	Generate() string
}

type RequestInput struct {
	Method     string
	URL        string
	Parameters map[string]string
	Body       interface{}
	Headers    map[string]string
}

type RequestProviderImpl struct {
	uuidService UUIDService
}

func (rp *RequestProviderImpl) Provide(ctx context.Context, input RequestInput) (*http.Request, error) {
	var bodyReader io.Reader
	if input.Body != nil {
		bodyBytes, err := json.Marshal(input.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	request, err := http.NewRequest(input.Method, input.URL, bodyReader)
	if err != nil {
		return nil, err
	}

	if len(input.Headers) != 0 {
		for key, value := range input.Headers {
			request.Header.Add(key, value)
		}
	}

	if len(input.Parameters) != 0 {
		q := request.URL.Query()
		for k, v := range input.Parameters {
			q.Set(k, v)
		}
		request.URL.RawQuery = q.Encode()
	}

	//TODO probably not necessary as we have custom http client with transport that takes care of this in a central place
	//request = request.WithContext(ctx)
	logger := log.C(ctx)
	//correlationID, exists := logger.Data[log.FieldCorrelationID].(string)
	//if exists && correlationID != log.BootstrapCorrelationID {
	//	request.Header.Set(log.CorrelationIDHeaders[0], correlationID)
	//} else {
	//	request.Header.Set(log.CorrelationIDHeaders[0], rp.uuidService.Generate())
	//}

	logger.Debugf("Provided request %s %s", request.Method, request.URL)

	return request, nil
}
