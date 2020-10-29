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

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"strconv"
)

type ApplicationsOutput []schema.ApplicationExt

//go:generate paginator ApplicationResponse ApplicationsOutput ".Result"
type ApplicationResponse struct {
	Result struct {
		Data ApplicationsOutput `json:"data"`
		Page graphql.PageInfo   `json:"pageInfo"`
	} `json:"result"`
}

type PackagessOutput []*schema.PackageExt

//go:generate paginator PackagesResponse PackagessOutput ".Result.Packages"
type PackagesResponse struct {
	Result struct {
		Packages struct {
			Data PackagessOutput  `json:"data"`
			Page graphql.PageInfo `json:"pageInfo"`
		} `json:"packages"`
	} `json:"result"`
}

type ApiDefinitionsOutput []*schema.APIDefinitionExt

//go:generate paginator ApiDefinitionsResponse ApiDefinitionsOutput ".Result.Package.ApiDefinitions"
type ApiDefinitionsResponse struct {
	Result struct {
		Package struct {
			ApiDefinitions struct {
				Data ApiDefinitionsOutput `json:"data"`
				Page graphql.PageInfo     `json:"pageInfo"`
			} `json:"apiDefinitions"`
		} `json:"package"`
	} `json:"result"`
}

type EventDefinitionsOutput []*schema.EventAPIDefinitionExt

//go:generate paginator EventDefinitionsResponse EventDefinitionsOutput ".Result.Package.EventDefinitions"
type EventDefinitionsResponse struct {
	Result struct {
		Package struct {
			EventDefinitions struct {
				Data EventDefinitionsOutput `json:"data"`
				Page graphql.PageInfo       `json:"pageInfo"`
			} `json:"eventDefinitions"`
		} `json:"package"`
	} `json:"result"`
}

type DocumentsOutput []*schema.DocumentExt

//go:generate paginator DocumentsResponse DocumentsOutput ".Result.Package.Documents"
type DocumentsResponse struct {
	Result struct {
		Package struct {
			Documents struct {
				Data DocumentsOutput  `json:"data"`
				Page graphql.PageInfo `json:"pageInfo"`
			} `json:"documents"`
		} `json:"package"`
	} `json:"result"`
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
	TargetURLs    map[string]string
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

type FindPackageSpecificationInput struct {
	ApplicationID string `valid:"required"`
	PackageID     string `valid:"required"`
	DefinitionID  string `valid:"required"`
}

type FindPackageSpecificationOutput struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`

	Data    *schema.CLOB      `json:"data,omitempty"`
	Format  schema.SpecFormat `json:"format"`
	Type    string            `json:"type"`
	Version *schema.Version   `json:"version,omitempty"`
}
