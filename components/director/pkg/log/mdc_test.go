package log

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
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
			require.Equal(t, inner, &mdc.mdc)
		}
	})
}

func TestAppendMdcToLogEntry(t *testing.T) {
	t.Run("do not modify log entry on empty MDC", func(t *testing.T) {
		mdc := NewMappedDiagnosticContext()
		entry := logrus.WithFields(make(map[string]interface{}))
		newEntry := mdc.appendFields(entry)

		require.Equal(t, entry, newEntry)
	})

	t.Run("append MDC content to a log entry", func(t *testing.T) {
		const testValue = "test-value"
		const testKey = "test"

		mdc := NewMappedDiagnosticContext()
		mdc.Set(testKey, testValue)

		entry := logrus.WithFields(make(map[string]interface{}))
		newEntry := mdc.appendFields(entry)

		require.NotEqual(t, entry, newEntry)

		value, ok := newEntry.Data[testKey]
		require.Truef(t, ok, "Missing key: %v", testKey)
		require.Equal(t, testValue, value)
	})
}

func TestMdcSetValues(t *testing.T) {
	t.Run("add value to mdc", func(t *testing.T) {
		const testKey = "test"
		const testValue = "test-val"

		mdc := NewMappedDiagnosticContext()
		mdc.SetIfNotEmpty(testKey, testValue)

		_, ok := mdc.mdc[testKey]
		require.True(t, ok, "Missing key: %v:%v", testKey, testValue)
	})

	t.Run("do not add empty value to mdc", func(t *testing.T) {
		const testKey = "test"

		mdc := NewMappedDiagnosticContext()
		mdc.SetIfNotEmpty(testKey, "")

		val, ok := mdc.mdc[testKey]
		require.False(t, ok, "Unexpected value: %v", val)
	})
}
