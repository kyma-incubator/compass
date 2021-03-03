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
	"encoding/json"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"strconv"
)

type ApplicationsOutput struct {
	Result *schema.ApplicationPageExt `json:"result"`
}

type ApplicationOutput struct {
	Result *schema.ApplicationExt `json:"result"`
}

type BundleInstanceCredentialsInput struct {
	BundleID    string `valid:"required"`
	AuthID      string `valid:"required"`
	Context     Values
	InputSchema Values
}

type Values map[string]interface{}

func (r *Values) MarshalToQGLJSON() (string, error) {
	input, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return strconv.Quote(string(input)), nil
}

type BundleInstanceCredentialsOutput struct {
	InstanceAuth *schema.BundleInstanceAuth
	TargetURLs   map[string]string
}

type BundleInstanceInput struct {
	InstanceAuthID string `valid:"required"`
	Context        map[string]string
}

type BundleInstanceAuthOutput struct {
	InstanceAuth *schema.BundleInstanceAuth `json:"result"`
}

type BundleInstanceAuthDeletionInput struct {
	InstanceAuthID string `valid:"required"`
}

type BundleInstanceAuthDeletionOutput struct {
	ID     string                          `json:"id"`
	Status schema.BundleInstanceAuthStatus `json:"status"`
}

type BundleSpecificationInput struct {
	ApplicationID string `valid:"required"`
	BundleID      string `valid:"required"`
	DefinitionID  string `valid:"required"`
}

type BundleSpecificationOutput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`

	Data    *schema.CLOB      `json:"data,omitempty"`
	Format  schema.SpecFormat `json:"format"`
	Type    string            `json:"type"`
	Version *schema.Version   `json:"version,omitempty"`
}
