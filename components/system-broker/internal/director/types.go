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

type RequestPackageInstanceCredentialsInput struct {
	PackageID   string `valid:"required"`
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

type RequestPackageInstanceCredentialsOutput struct {
	InstanceAuth *schema.PackageInstanceAuth `json:"result"`
}

type FindPackageInstanceCredentialsByContextInput struct {
	ApplicationID string `valid:"required"`
	PackageID     string `valid:"required"`
	Context       map[string]string
}

type FindPackageInstanceCredentialsOutput struct {
	InstanceAuths []*schema.PackageInstanceAuth
}

type FindPackageInstanceCredentialInput struct {
	PackageID      string `valid:"required"`
	ApplicationID  string `valid:"required"`
	InstanceAuthID string `valid:"required"`
}

type FindPackageInstanceCredentialOutput struct {
	InstanceAuth *schema.PackageInstanceAuth `json:"result"`
}

type RequestPackageInstanceAuthDeletionInput struct {
	InstanceAuthID string `valid:"required"`
}

type RequestPackageInstanceAuthDeletionOutput struct {
	ID     string                           `json:"id"`
	Status schema.PackageInstanceAuthStatus `json:"status"`
}
