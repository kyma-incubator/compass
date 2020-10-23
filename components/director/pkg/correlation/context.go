package correlation

import (
	"context"

	"github.com/google/uuid"
)

func IDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(ContextField).(string); ok {
		return id
	}

	return ""
}

func newID() string {
	return uuid.New().String()
}
