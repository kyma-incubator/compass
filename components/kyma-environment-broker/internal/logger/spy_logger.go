// Package copied from https://github.com/kyma-project/kyma/blob/1.11.0/components/service-binding-usage-controller/internal/platform/logger/spy/logger.go
// Only Reset() method was added.
package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

// LogSpy is a helper construct for testing logging in unit tests.
// Beware: all methods are working on copies of of original messages buffer and are safe for multiple uses.
type LogSpy struct {
	buffer    *bytes.Buffer
	RawLogger *logrus.Logger
	Logger    *logrus.Entry
}

// NewLogSpy is a factory for LogSpy
func NewLogSpy() *LogSpy {
	buffer := bytes.NewBuffer([]byte(""))

	rawLgr := &logrus.Logger{
		Out: buffer,
		// standard json formatter is used to ease testing
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	lgr := rawLgr.WithField("testing", true)

	return &LogSpy{
		buffer:    buffer,
		RawLogger: rawLgr,
		Logger:    lgr,
	}
}

// AssertErrorLogged checks whatever a specific string was logged as error.
//
// Compared elements: level, message
//
// Wrapped errors are supported as long as original error message ends up in resulting one.
func (s *LogSpy) AssertErrorLogged(t *testing.T, errorExpected error) {
	if !s.wasLogged(t, logrus.ErrorLevel, errorExpected.Error()) {
		t.Errorf("error was not logged, expected: \"%s\"", errorExpected.Error())
	}
}

// AssertLogged checks whatever a specific string was logged at a specific level.
//
// Compared elements: level, message
//
// Beware: we are checking for sub-strings and not for the exact match.
func (s *LogSpy) AssertLogged(t *testing.T, level logrus.Level, message string) {
	if !s.wasLogged(t, level, message) {
		t.Errorf("message was not logged, message: \"%s\", level: %s", message, level)
	}
}

// AssertNotLogged checks whatever a specific string was not logged at a specific level.
//
// Compared elements: level, message
//
// Beware: we are checking for sub-strings and not for the exact match.
func (s *LogSpy) AssertNotLogged(t *testing.T, level logrus.Level, message string) {
	if s.wasLogged(t, level, message) {
		t.Errorf("message was logged, message: \"%s\", level: %s", message, level)
	}
}

// wasLogged checks whatever a message was logged.
//
// Compared elements: level, message
func (s *LogSpy) wasLogged(t *testing.T, level logrus.Level, message string) bool {
	// new reader is created so we are safe for multiple reads
	buf := bytes.NewReader(s.buffer.Bytes())
	scanner := bufio.NewScanner(buf)
	var entryPartial struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}

	for scanner.Scan() {
		line := scanner.Text()

		err := json.Unmarshal([]byte(line), &entryPartial)
		if err != nil {
			t.Fatalf("unexpected error on log line unmarshalling, line: %s", line)
		}

		levelMatches := entryPartial.Level == level.String()

		// We are looking only if expected is contained (as opposed to exact match check),
		// so that e.g. errors wrapping is supported.
		containsMessage := strings.Contains(entryPartial.Msg, message)

		if levelMatches && containsMessage {
			return true
		}
	}

	return false
}

// DumpAll returns all logged messages.
func (s *LogSpy) DumpAll() []string {
	// new reader is created so we are safe for multiple reads
	buf := bytes.NewReader(s.buffer.Bytes())
	scanner := bufio.NewScanner(buf)

	out := []string{}
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	return out
}

// Reset removes logged messages.
func (s *LogSpy) Reset() {
	s.buffer.Reset()
}

// NewLogDummy returns dummy logger which discards logged messages on the fly.
// Useful when logger is required as dependency in unit testing.
func NewLogDummy() *logrus.Entry {
	rawLgr := logrus.New()
	rawLgr.Out = ioutil.Discard
	lgr := rawLgr.WithField("testing", true)

	return lgr
}
