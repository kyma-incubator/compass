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

package graphql

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

type Entity interface {
	GetID() string
	GetType() resource.Type
}

type BaseEntity struct {
	ID        string     `json:"id"`
	Ready     bool       `json:"ready"`
	CreatedAt *Timestamp `json:"createdAt"`
	UpdatedAt *Timestamp `json:"updatedAt"`
	DeletedAt *Timestamp `json:"deletedAt"`
	Error     *string    `json:"error"`
}

func (e *BaseEntity) GetID() string {
	return e.ID
}

func (e *BaseEntity) GetType() resource.Type {
	return ""
}
