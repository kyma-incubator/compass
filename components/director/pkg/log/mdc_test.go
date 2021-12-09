package log

import (
	"context"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestContextWithMdc(t *testing.T) {
	t.Run("add MDC to context without MDC", func(t2 *testing.T) {
		ctx := context.Background()
		if mdc := MdcFromContext(ctx); nil != mdc {
			t.Fatalf("[pre-condition] There should not have been any MDC in the provided context!")
		}

		ctx = ContextWithMdc(ctx)
		if mdc := MdcFromContext(ctx); nil == mdc {
			t.Fatalf("An MDC instance should have been attached to the context")
		}
	})

	t.Run("do not replace the MDC instance in a context", func(t2 *testing.T) {
		ctx := ContextWithMdc(context.Background())

		var inner *map[string]interface{}
		if mdc := MdcFromContext(ctx); nil == mdc {
			t.Fatalf("[pre-condition] There should have been MDC in the provided context!")
		} else {
			inner = &mdc.mdc
		}

		ctx = ContextWithMdc(ctx)
		if mdc := MdcFromContext(ctx); nil == mdc {
			t.Fatalf("An MDC instance should have been attached to the context")
		} else {
			if inner != &mdc.mdc {
				t.Fatalf("The MDC instance has been replaced")
			}
		}

	})
}

func TestAppendMdcToLogEntry(t *testing.T) {
	t.Run("do not modify log entry on empty MDC", func(t *testing.T) {
		mdc := NewMappedDiagnosticContext()
		entry := logrus.WithFields(make(map[string]interface{}))
		newEntry := mdc.appendFields(entry)

		if entry != newEntry {
			t.Fatalf("The log entry should not have been modified")
		}
	})

	t.Run("append MDC content to a log entry", func(t *testing.T) {
		const testValue = "test-value"
		const testKey = "test"

		mdc := NewMappedDiagnosticContext()
		mdc.Set(testKey, testValue)

		entry := logrus.WithFields(make(map[string]interface{}))
		newEntry := mdc.appendFields(entry)

		if entry == newEntry {
			t.Fatalf("The log entry should have been modified")
		}

		value, ok := newEntry.Data[testKey]
		if !ok {
			t.Fatalf("The expected key was not found in the log entry's data")
		}
		if value != testValue {
			t.Fatalf("The filed value does not match. Expected '%v', but got '%v'", testValue, value)
		}

	})
}
