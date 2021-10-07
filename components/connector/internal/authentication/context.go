package authentication

import (
	"context"
	"errors"
	"fmt"
)

type ContextKey string

const (
	ConnectorTokenKey          ContextKey = "ConnectorToken"
	TenantKey                  ContextKey = "TenantKey"
	ConsumerType               ContextKey = "ConsumerType"
	ClientIdFromTokenKey       ContextKey = "ClientIdFromToken"
	ClientIdFromCertificateKey ContextKey = "ClientIdFromCertificate"
	ClientCertificateHashKey   ContextKey = "ClientCertificateHash"
)

func GetStringFromContext(ctx context.Context, key ContextKey) (string, error) {
	value := ctx.Value(key)

	str, ok := value.(string)
	if !ok {
		return "", errors.New(fmt.Sprintf("Cannot read %s key from context", string(key)))
	}

	return str, nil
}

func PutIntoContext(ctx context.Context, key ContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}
