package middlewares

import (
	"context"
	"errors"
	"fmt"
)

type ContextKey string

const ClientIdKey ContextKey = "ClientIdWithContext"
const BaseURLsKey ContextKey = "BaseURLs"

type BaseURLs struct {
	ConnectivityAdapterBaseURL string
	EventServiceBaseURL        string
}

func GetStringFromContext(ctx context.Context, key ContextKey) (string, error) {
	value := ctx.Value(key)

	str, ok := value.(string)
	if !ok {
		return "", errors.New(fmt.Sprintf("Cannot read %s key from context", string(key)))
	}

	return str, nil
}

func GetBaseURLsFromContext(ctx context.Context, key ContextKey) (BaseURLs, error) {
	value := ctx.Value(key)

	baseURLs, ok := value.(BaseURLs)
	if !ok {
		return BaseURLs{}, errors.New(fmt.Sprintf("Cannot read %s key from context", string(key)))
	}

	return baseURLs, nil
}

func PutIntoContext(ctx context.Context, key ContextKey, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}
