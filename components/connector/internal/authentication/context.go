package authentication

import (
	"context"
	"fmt"
)

type ContextKey string

const (
	ConnectorTokenKey          ContextKey = "ConnectorToken"
	TenantKey                  ContextKey = "TenantKey"
	ConsumerType               ContextKey = "ConsumerType"
	ServiceAccountFile         ContextKey = "ServiceAccountFile"
	ClientIdFromTokenKey       ContextKey = "ClientIdFromToken"
	ClientIdFromCertificateKey ContextKey = "ClientIdFromCertificate"
	ClientCertificateHashKey   ContextKey = "ClientCertificateHash"
)

func GetStringFromContext(ctx context.Context, key ContextKey) (string, error) {
	value := ctx.Value(key)

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("cannot read %s key from context", string(key))
	}

	return str, nil
}

func PutIntoContext(ctx context.Context, key ContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}
