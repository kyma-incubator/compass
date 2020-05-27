package customerrors

import (
	"context"
	"errors"
	"fmt"
)

func RecoverFn(ctx context.Context, err interface{}) error {
	errText := fmt.Sprintf("%+v", err)

	return NewBuilder().InternalError("Panic Error").Wrap(errors.New(errText)).Build()
}
