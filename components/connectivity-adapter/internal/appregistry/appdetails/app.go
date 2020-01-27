package appdetails

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

const AppDetailsContextKey = "compass/app-details"

var NoAppDetailsError = errors.New("cannot read Application details from context")

func LoadFromContext(ctx context.Context) (graphql.ApplicationExt, error) {
	if ctx == nil {
		return graphql.ApplicationExt{}, NoAppDetailsError
	}

	value := ctx.Value(AppDetailsContextKey)

	appDetails, ok := value.(graphql.ApplicationExt)

	if !ok {
		return graphql.ApplicationExt{}, NoAppDetailsError
	}

	return appDetails, nil
}

func SaveToContext(ctx context.Context, appDetails graphql.ApplicationExt) context.Context {
	return context.WithValue(ctx, AppDetailsContextKey, appDetails)
}

