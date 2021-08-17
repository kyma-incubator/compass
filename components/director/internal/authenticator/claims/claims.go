package claims

import (
	"context"
)

type Claims interface {
	ContextWithClaims(ctx context.Context) context.Context
}
