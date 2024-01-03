package testingx

import (
	"os"
	"regexp"
	"testing"
)

const skipRegexEnvVar = "SKIP_TESTS_REGEX"

type T struct {
	*testing.T
}

func NewT(t *testing.T) *T {
	return &T{
		T: t,
	}
}

type Logger interface {
	Log(args ...any)
	Logf(format string, args ...any)
	Errorf(format string, args ...interface{})
	FailNow()
}

// OnceLogger is a logger that logs only once.
// It is useful in the cases that use WithEventually and also want to log something
// As WithEventually will print the log a lot of times while trying over and over,
// with the OnceLogger it will print it just once and if the WithEventually
// reaches another log, it will print it as well just once, making the logs more readable.
type OnceLogger struct {
	logger        Logger
	alreadyLogged map[string]bool
}

// NewOnceLogger returns a new OnceLogger that logs each log only once
func NewOnceLogger(logger Logger) OnceLogger {
	return OnceLogger{
		logger:        logger,
		alreadyLogged: map[string]bool{},
	}
}

func (l *OnceLogger) Log(log string, args ...any) {
	if l.alreadyLogged[log] {
		return
	}
	l.alreadyLogged[log] = true
	l.logger.Log(append([]any{log}, args...)...)
}

func (l *OnceLogger) Logf(format string, args ...any) {
	if l.alreadyLogged[format] {
		return
	}
	l.alreadyLogged[format] = true
	l.logger.Logf(format, args...)
}

func (l OnceLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args)
}
func (l OnceLogger) FailNow() {
	l.logger.FailNow()
}

func (t *T) Run(name string, f func(t *testing.T)) bool {
	newF := f

	pattern := os.Getenv(skipRegexEnvVar)
	if len(pattern) > 0 {
		newF = func(t *testing.T) {
			match, err := regexp.MatchString(pattern, t.Name())
			if err != nil {
				t.Fatalf("An error occured while parsing skip regex: %s", pattern)
			}
			if match {
				t.Skipf("Skipping test... Reason: test name %s matches skip pattern %s", t.Name(), pattern)
			}
			f(t)
		}
	}

	return t.T.Run(name, newF)
}
