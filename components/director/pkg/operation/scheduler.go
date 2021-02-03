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

package operation

// Scheduler is responsible for scheduling any provided Operation entity for later processing
//go:generate mockery -name=Scheduler -output=automock -outpkg=automock -case=underscore
type Scheduler interface {
	Schedule(operation Operation) (string, error)
}

// DefaultScheduler defines an initial Scheduler implementation
type DefaultScheduler struct{}

// Schedule is responsible for scheduling any provided Operation
func (d DefaultScheduler) Schedule(_ Operation) (string, error) {
	return "", nil
}
