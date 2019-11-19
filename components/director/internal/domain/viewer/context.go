package viewer

import "context"

type key int

const (
	keyID   key = iota
	keyType key = iota
)

func SaveIDToContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyID, id)
}

func SaveTypeToContext(ctx context.Context, objType string) context.Context {
	return context.WithValue(ctx, keyType, objType)
}

func LoadIDFromContext(ctx context.Context) string {
	str, ok := (ctx.Value(keyID)).(string)
	if !ok {
		return ""
	}
	return str
}

func LoadTypeFromContext(ctx context.Context) string {
	str, ok := (ctx.Value(keyType)).(string)
	if !ok {
		return ""
	}
	return str
}
