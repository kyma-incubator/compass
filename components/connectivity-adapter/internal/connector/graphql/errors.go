package graphql

import "github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"

func toAppError(err error) apperrors.AppError {

	if err == nil {
		return nil
	}

	return apperrors.Internal(err.Error())
}
