/*
 * Copyright 2018 The Service Manager Authors
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

package main

const PageTypeTemplate = `// GENERATED. DO NOT MODIFY!

package {{.PackageName}}

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)


func (p *{{.Type}}) PageInfo() *graphql.PageInfo {
	return &p{{.DataPath}}.Page
}

func (p *{{.Type}}) ListAll(ctx context.Context, pager *Paginator) ({{.OutputType}}, error) {
	pageResult := {{.OutputType}}{}

	for {
		items := &{{.Type}}{}

		hasNext, err := pager.Next(ctx, items)
		if err != nil {
			return nil, err
		}

		pageResult = append(pageResult, items{{.DataPath}}.Data...)
		if !hasNext {
			return pageResult, nil
		}
	}
}
`
