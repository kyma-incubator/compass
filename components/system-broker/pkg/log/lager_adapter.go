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

package log

import (
	. "code.cloudfoundry.org/lager"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pivotal-cf/brokerapi/v7/middlewares"
	"github.com/sirupsen/logrus"
)

const (
	// FieldComponentName is the key of the component field in the log message.
	FieldComponentName = "component"
	FieldCorrelationID = "x-request-id"
)

// LagerAdapter is used to adapt log.Logger interface with logrus logger
type LagerAdapter struct {
	entry *logrus.Entry
}

// NewDefaultLagerAdapter returns new LagerAdapter from logrus entity
func NewDefaultLagerAdapter() *LagerAdapter {
	return &LagerAdapter{entry: log.D()}
}

// RegisterSink is not applicable for LagerAdapter, implemented to satisfy log.Logger interface
func (l *LagerAdapter) RegisterSink(Sink) {
	if l.entry != nil {
		l.entry.Error("LagerAdapter does not work with sinks.")
	}
}

// Session returns new Logger with provided data
func (l *LagerAdapter) Session(session string, data ...Data) Logger {
	logger := l.newWithData(data...)
	logger.entry.Data[FieldComponentName] = session
	return logger
}

// SessionName returns session name
func (l *LagerAdapter) SessionName() string {
	session, exists := l.entry.Data[FieldComponentName]
	if !exists {
		return ""
	}
	return session.(string)
}

// Debug logs message on debug level
func (l *LagerAdapter) Debug(action string, data ...Data) {
	appendData(l.entry, data...)
	l.entry.Debug(action)
}

// Info logs message on info level
func (l *LagerAdapter) Info(action string, data ...Data) {
	appendData(l.entry, data...)
	l.entry.Info(action)
}

// Error logs message on error level
func (l *LagerAdapter) Error(action string, err error, data ...Data) {
	appendData(l.entry, data...)
	l.entry.Data[logrus.ErrorKey] = err
	l.entry.Error(action)
}

// Fatal logs message on fatal level
func (l *LagerAdapter) Fatal(action string, err error, data ...Data) {
	appendData(l.entry, data...)
	l.entry.Data[logrus.ErrorKey] = err
	l.entry.Fatal(action)
}

// WithData returns new Logger with provided data
func (l *LagerAdapter) WithData(data Data) Logger {
	return l.newWithData(data)
}

func (l *LagerAdapter) newWithData(data ...Data) *LagerAdapter {
	if l.entry == nil {
		l.entry = log.D()
	}

	e := copyEntry(l.entry)
	appendData(e, data...)
	return &LagerAdapter{entry: e}
}

func appendData(entry *logrus.Entry, data ...Data) {
	for _, d := range data {
		for k, v := range d {
			// "brokerapi" uses "correlation-id" as a key but we need "correlation_id"
			if k == middlewares.CorrelationIDKey {
				entry.Data[FieldCorrelationID] = v
			} else {
				entry.Data[k] = v
			}
		}
	}
}

func copyEntry(entry *logrus.Entry) *logrus.Entry {
	entryData := make(logrus.Fields, len(entry.Data))
	for k, v := range entry.Data {
		entryData[k] = v
	}

	newEntry := logrus.NewEntry(entry.Logger)
	newEntry.Level = entry.Level
	newEntry.Data = entryData
	newEntry.Time = entry.Time
	newEntry.Message = entry.Message
	newEntry.Buffer = entry.Buffer

	return newEntry
}
