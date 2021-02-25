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

package director

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/types"
	"net/http"
)

// client implements the DirectorClient interface
type client struct {
	types.ApplicationLister
	HTTPClient  http.Client
	DirectorURL string
}

type Request struct {
	OperationType graphql.OperationType `json:"operation_type,omitempty"`
	ResourceType  resource.Type         `json:"resource_type"`
	ResourceID    string                `json:"resource_id"`
	Error         string                `json:"error"`
}

// NewClient constructs a default implementation of the Client interface
func NewClient(directorURL string, httpClient http.Client, appLister types.ApplicationLister) *client {
	return &client{
		ApplicationLister: appLister,
		HTTPClient:        httpClient,
		DirectorURL:       directorURL,
	}
}

// UpdateOperation makes an http request to the Director to notify about any operation state changes
func (c *client) UpdateOperation(ctx context.Context, request *Request) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.DirectorURL+"/operations", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code when notifying director for application state: %d", resp.StatusCode)
	}

	return nil
}
