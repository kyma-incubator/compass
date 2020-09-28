package authentication

import (
	"context"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
)

type ContextKey string

const (
	ConnectorTokenKey          ContextKey = "ConnectorToken"
	ClientIdFromTokenKey       ContextKey = "ClientIdFromToken"
	TokenTypeKey               ContextKey = "TokenType"
	ClientIdFromCertificateKey ContextKey = "ClientIdFromCertificate"
	ClientCertificateHashKey   ContextKey = "ClientCertificateHash"
)

func GetStringFromContext(ctx context.Context, key ContextKey) (string, apperrors.AppError) {
	value := ctx.Value(key)

	str, ok := value.(string)
	if !ok {
		return "", apperrors.NotAuthenticated("Cannot read %s key from context", string(key))
	}

	return str, nil
}

func PutIntoContext(ctx context.Context, key ContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}
