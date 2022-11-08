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

// mockedLogSink is a mock implementation of the LogSink interface to fit the controller tests use case
type mockedLogSink struct {
	log.NullLogSink
	AssertErrorExpectations func(err error, msg string)
}

func (ml *mockedLogSink) Info(level int, msg string, keysAndValues ...interface{}) {}

func (ml *mockedLogSink) Enabled(level int) bool { return true }

func (ml *mockedLogSink) V(_ int) logr.LogSink { return nil }

func (ml *mockedLogSink) WithValues(_ ...interface{}) logr.LogSink { return ml }

func (ml *mockedLogSink) WithName(_ string) logr.LogSink { return nil }

func (ml *mockedLogSink) Error(err error, msg string, _ ...interface{}) {
	ml.AssertErrorExpectations(err, msg)
}
