package appcontext

import "context"

func NewAppContext() *AppContext {
	return &AppContext{}
}

type AppContext struct{}

func (c AppContext) WithValue(parent context.Context, key interface{}, val interface{}) context.Context {
	return context.WithValue(parent, key, val)
}
