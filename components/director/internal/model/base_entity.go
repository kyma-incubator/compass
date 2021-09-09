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

package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity denotes a model-layer entity which can be timestamped with created_at, updated_at, deleted_at and ready values
type Entity interface {
	GetID() string
	GetType() resource.Type

	GetReady() bool
	SetReady(ready bool)

	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)

	GetUpdatedAt() time.Time
	SetUpdatedAt(t time.Time)

	GetDeletedAt() time.Time
	SetDeletedAt(t time.Time)

	GetError() *string
	SetError(err string)
}

// BaseEntity missing godoc
type BaseEntity struct {
	ID        string
	Ready     bool
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
	Error     *string
}

// GetID missing godoc
func (e *BaseEntity) GetID() string {
	return e.ID
}

// GetType missing godoc
func (e *BaseEntity) GetType() resource.Type {
	return ""
}

// GetReady missing godoc
func (e *BaseEntity) GetReady() bool {
	return e.Ready
}

// SetReady missing godoc
func (e *BaseEntity) SetReady(ready bool) {
	e.Ready = ready
}

// GetCreatedAt missing godoc
func (e *BaseEntity) GetCreatedAt() time.Time {
	if e.CreatedAt == nil {
		return time.Time{}
	}
	return *e.CreatedAt
}

// SetCreatedAt missing godoc
func (e *BaseEntity) SetCreatedAt(t time.Time) {
	e.CreatedAt = &t
}

// GetUpdatedAt missing godoc
func (e *BaseEntity) GetUpdatedAt() time.Time {
	if e.UpdatedAt == nil {
		return time.Time{}
	}
	return *e.UpdatedAt
}

// SetUpdatedAt missing godoc
func (e *BaseEntity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = &t
}

// GetDeletedAt missing godoc
func (e *BaseEntity) GetDeletedAt() time.Time {
	if e.DeletedAt == nil {
		return time.Time{}
	}
	return *e.DeletedAt
}

// SetDeletedAt missing godoc
func (e *BaseEntity) SetDeletedAt(t time.Time) {
	e.DeletedAt = &t
}

// GetError missing godoc
func (e *BaseEntity) GetError() *string {
	return e.Error
}

// SetError missing godoc
func (e *BaseEntity) SetError(err string) {
	if err == "" {
		e.Error = nil
	} else {
		e.Error = &err
	}
}
