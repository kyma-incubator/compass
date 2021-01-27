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

type Scheduler interface {
	Schedule(operation Operation) (string, error)
	Fetch(operationID string) (Operation, error)
	FetchStatus(resourceID, resourceType string) (string, error)
	Watch() error
}

type DefaultScheduler struct{}

func (d DefaultScheduler) Schedule(_ Operation) (string, error) {
	return "", nil
}

func (d DefaultScheduler) Fetch(_ string) (Operation, error) {
	return Operation{}, nil
}

func (d DefaultScheduler) FetchStatus(_, _ string) (string, error) {
	return "", nil
}

func (d DefaultScheduler) Watch() error {
	return nil
}
