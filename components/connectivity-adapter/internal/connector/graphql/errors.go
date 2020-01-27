package graphql

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"strings"
)

func toAppError(err error) apperrors.AppError {

	if err == nil {
		return nil
	}

	// Find some better way to distinguish error type
	if strings.Contains(err.Error(), "CSR: Invalid") {
		return apperrors.WrongInput(err.Error())
	}

	if strings.Contains(err.Error(), "Error while parsing base64 content") {
		return apperrors.WrongInput(err.Error())
	}

	return apperrors.Internal(err.Error())
}
