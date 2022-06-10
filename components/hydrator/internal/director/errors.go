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
	"strings"

	"github.com/pkg/errors"
)

func IsGQLNotFoundError(err error) bool {
	err = errors.Cause(err)

	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "Object not found")
}

type NotFoundError struct {
}

func (fe *NotFoundError) Error() string {
	return "NotFound"
}

func (fe *NotFoundError) NotFound() bool {
	return true
}
