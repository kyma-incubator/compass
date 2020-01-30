package middlewares

import (
	"context"
	"errors"
	"fmt"
)

type ContextKey string
type AuthorizationHeaders map[string]string

const AuthorizationHeadersKey ContextKey = "ClientIdWithContext"
const BaseURLsKey ContextKey = "BaseURLs"

type BaseURLs struct {
	ConnectivityAdapterBaseURL     string
	ConnectivityAdapterMTLSBaseURL string
	EventServiceBaseURL            string
}

func GetAuthHeadersFromContext(ctx context.Context, key ContextKey) (AuthorizationHeaders, error) {
	value := ctx.Value(key)

	headers, ok := value.(AuthorizationHeaders)
	if !ok {
		return map[string]string{}, errors.New(fmt.Sprintf("Cannot read %s key from context", string(key)))
	}

	return headers, nil
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
