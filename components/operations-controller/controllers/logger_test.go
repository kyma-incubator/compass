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

package controllers

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type mockedLogger struct {
	log.NullLogger
	RecordedError error
}

func (ml *mockedLogger) Info(_ string, _ ...interface{}) {}

func (ml *mockedLogger) Enabled() bool { return true }

func (ml *mockedLogger) V(level int) logr.InfoLogger { return nil }

func (ml *mockedLogger) WithValues(keysAndValues ...interface{}) logr.Logger { return ml }

func (ml *mockedLogger) WithName(name string) logr.Logger { return nil }

func (ml *mockedLogger) Error(err error, _ string, _ ...interface{}) { ml.RecordedError = err }
