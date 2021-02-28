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

package controllers_test

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// mockedLogger is a mock implementation of the Logger interface to fit the controller tests use case
type mockedLogger struct {
	log.NullLogger
	AssertErrorExpectations func(err error, msg string)
}

func (ml *mockedLogger) Info(_ string, _ ...interface{}) {}

func (ml *mockedLogger) Enabled() bool { return true }

func (ml *mockedLogger) V(_ int) logr.InfoLogger { return nil }

func (ml *mockedLogger) WithValues(_ ...interface{}) logr.Logger { return ml }

func (ml *mockedLogger) WithName(_ string) logr.Logger { return nil }

func (ml *mockedLogger) Error(err error, msg string, _ ...interface{}) {
	ml.AssertErrorExpectations(err, msg)
}
