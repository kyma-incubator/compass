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
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/onrik/logrus/filename"

	"github.com/sirupsen/logrus"
)

type logKey struct{}

const (
	// FieldRequestID missing godoc
	FieldRequestID     = "x-request-id"
	fieldComponentName = "component"
	jsonFormatterKey   = "json"
	textFormatterKey   = "text"
)

var (
	defaultEntry = logrus.NewEntry(logrus.StandardLogger())

	supportedFormatters = map[string]logrus.Formatter{
		jsonFormatterKey: &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		},
		textFormatterKey: &logrus.TextFormatter{},
	}

	supportedOutputs = map[string]io.Writer{
		os.Stdout.Name(): os.Stdout,
		os.Stderr.Name(): os.Stderr,
	}
	mutex           = sync.RWMutex{}
	currentSettings = DefaultConfig()

	// C missing godoc
	C = LoggerFromContext
	// D missing godoc
	D = DefaultLogger
)

func init() {
	// Configure default logger in init so we can log even before actual logging settings are loaded
	hook := prepareLogSourceHook()
	AddHook(hook)
	AddHook(&ErrorLocationHook{})
	defaultEntry = defaultEntry.WithField(FieldRequestID, currentSettings.BootstrapCorrelationID)

	_, err := Configure(context.Background(), currentSettings)
	if err != nil {
		panic(err)
	}
}

// Configure creates a new context with a logger using the provided settings.
func Configure(ctx context.Context, config *Config) (context.Context, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if err := config.Validate(); err != nil {
		return nil, err
	}

	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, err
	}
	formatter := supportedFormatters[config.Format]
	output := supportedOutputs[config.Output]

	currentSettings = config

	entry := ctx.Value(logKey{})
	if entry == nil {
		entry = defaultEntry
	} else {
		defaultEntry.Logger.SetOutput(output)
		defaultEntry.Logger.SetFormatter(formatter)
		defaultEntry.Logger.SetLevel(level)
	}
	e := copyEntry(entry.(*logrus.Entry))
	e.Logger.SetLevel(level)
	e.Logger.SetFormatter(formatter)
	e.Logger.SetOutput(output)

	return ContextWithLogger(ctx, e), nil
}

// Configuration returns the logger settings
func Configuration() Config {
	mutex.RLock()
	defer mutex.RUnlock()

	return *currentSettings
}

// ContextWithLogger returns a new context with the provided logger.
func ContextWithLogger(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, logKey{}, entry)
}

// LoggerFromContext retrieves the current logger from the context.
func LoggerFromContext(ctx context.Context) *logrus.Entry {
	mutex.RLock()
	defer mutex.RUnlock()
	entry := ctx.Value(logKey{})
	if entry == nil {
		entry = defaultEntry
	}
	return copyEntry(entry.(*logrus.Entry))
}

// DefaultLogger returns the default logger
func DefaultLogger() *logrus.Entry {
	return LoggerFromContext(context.Background())
}

// RegisterFormatter registers a new logrus Formatter with the given name.
// Returns an error if there is a formatter with the same name.
func RegisterFormatter(name string, formatter logrus.Formatter) error {
	if _, exists := supportedFormatters[name]; exists {
		return fmt.Errorf("formatter with name %s is already registered", name)
	}
	supportedFormatters[name] = formatter
	return nil
}

// AddHook adds a hook to all loggers
func AddHook(hook logrus.Hook) {
	defaultEntry.Logger.AddHook(hook)
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

func prepareLogSourceHook() *filename.Hook {
	hook := filename.NewHook()
	hook.Field = fieldComponentName
	hook.SkipPrefixes = append(hook.SkipPrefixes, "log/")
	hook.Formatter = func(file, function string, line int) string {
		split := strings.Split(function, string(filepath.Separator))
		return fmt.Sprintf("%s:%d:%s", file, line, split[len(split)-1])
	}

	return hook
}
