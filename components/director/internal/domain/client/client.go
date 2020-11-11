package client

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type key int

const (
	ClientUserContextKey key = iota
)

func LoadFromContext(ctx context.Context) (string, error) {
	clientID, ok := ctx.Value(ClientUserContextKey).(string)

	if !ok {
		return "", apperrors.NewCannotReadClientUserError()
	}

	return clientID, nil
}

func SaveToContext(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, ClientUserContextKey, clientID)
}
