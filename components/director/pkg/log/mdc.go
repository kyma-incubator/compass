package log

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
)

type mdcKey string

const contextKeyMdc mdcKey = "mapped-diagnostic-context"

// ContextWithMdc returns a new context with attached MDC
func ContextWithMdc(ctx context.Context) context.Context {
	if nil == MdcFromContext(ctx) {
		return context.WithValue(ctx, contextKeyMdc, NewMappedDiagnosticContext())
	}

	return ctx
}

// MdcFromContext returns the attached *MDC or nil if there is none
func MdcFromContext(ctx context.Context) *MDC {
	ptr, ok := ctx.Value(contextKeyMdc).(*MDC)
	if !ok {
		return nil
	}
	return ptr
}

// MDC provides a mechanism to enrich the log messages with
// information, which might not be available at the time/scope
// where the logging actually occurs.
type MDC struct {
	mdc  map[string]interface{}
	lock sync.Mutex
}

// NewMappedDiagnosticContext creates a new MDC instance
func NewMappedDiagnosticContext() *MDC {
	return &MDC{
		mdc:  make(map[string]interface{}),
		lock: sync.Mutex{},
	}
}

// Set adds or overwrites a key-value pair in the MDC
func (mdc *MDC) Set(key string, value interface{}) {
	mdc.lock.Lock()
	defer mdc.lock.Unlock()

	mdc.mdc[key] = value
}

// SetIfNotEmpty adds or overwrites a key-value pair in the MDC
// but only if the value is not empty
func (mdc *MDC) SetIfNotEmpty(key string, value string) {
	if value != "" {
		mdc.Set(key, value)
	}
}

func (mdc *MDC) appendFields(entry *logrus.Entry) *logrus.Entry {
	if len(mdc.mdc) == 0 {
		return entry
	}

	return entry.WithFields(mdc.mdc)
}
