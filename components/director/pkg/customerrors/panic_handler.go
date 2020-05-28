package customerrors

import (
	"context"
	"errors"
	"fmt"
)

func RecoverFn(ctx context.Context, err interface{}) error {
	errText := fmt.Sprintf("%+v", err)

	return newBuilder().withStatusCode(InternalError).wrap(errors.New(errText)).build()
}
