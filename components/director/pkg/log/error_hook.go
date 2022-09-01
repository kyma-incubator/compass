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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	errorSourceField            = "error_source"
	errorSourceFieldUnavailable = "check component log field"
)

// ErrorLocationHook provides an implementation of the sirupsen/logrus/Hook interface.
// Attaches error location information to log entries if an error is being logged and it has stack-trace information
// (i.e. if it originates from or is wrapped by github.com/pkg/errors).
type ErrorLocationHook struct {
}

// Levels missing godoc
func (h *ErrorLocationHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire missing godoc
func (h *ErrorLocationHook) Fire(entry *logrus.Entry) error {
	var (
		errObj interface{}
		exists bool
	)

	if errObj, exists = entry.Data[logrus.ErrorKey]; !exists {
		return nil
	}

	err, ok := errObj.(error)
	if !ok {
		return errors.New("object logged as error does not satisfy error interface")
	}

	stackErr := getInnermostStackTrace(err)

	if stackErr != nil {
		stackTrace := stackErr.StackTrace()
		bndl := getPkgName(stackTrace)
		errSource := fmt.Sprintf("%s/%s:%d:%n", bndl, stackTrace[0], stackTrace[0], stackTrace[0])
		entry.Data[errorSourceField] = errSource
	} else {
		entry.Data[errorSourceField] = errorSourceFieldUnavailable
	}

	return nil
}

func getPkgName(trace errors.StackTrace) string {
	formattedTrace := fmt.Sprintf("%+s", trace[0])
	split := strings.Split(formattedTrace, "/") // Even on Windows this should be "/"

	return split[len(split)-2]
}

type stackTracer interface {
	error
	StackTrace() errors.StackTrace
}

type causer interface {
	Cause() error
}

type unwrapper interface {
	Unwrap() error
}

// getInnermostStackTrace drills down into all possible error wrappings looking for the inner-most error that is annotated with a stack trace
func getInnermostStackTrace(err error) stackTracer {
	var tracer stackTracer

	for {
		t, isTracer := err.(stackTracer)
		if isTracer {
			tracer = t
		}

		c, isCauser := err.(causer)
		if isCauser {
			err = c.Cause()
			continue
		}

		u, inUnwrappable := err.(unwrapper)
		if inUnwrappable {
			err = u.Unwrap()
			continue
		}

		return tracer
	}
}
