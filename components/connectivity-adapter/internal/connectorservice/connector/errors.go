package connector

import (
	"strings"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
)

func toAppError(err error) apperrors.AppError {

	if err == nil {
		return nil
	}

	// TODO: Find some better way to distinguish error type
	if strings.Contains(err.Error(), "CSR: Invalid") {
		return apperrors.WrongInput(err.Error())
	}

	if strings.Contains(err.Error(), "Error while parsing the base64 content") {
		return apperrors.WrongInput(err.Error())
	}

	return apperrors.Internal(err.Error())
}
