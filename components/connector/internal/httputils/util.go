package httputils

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"io"
)

func Close(ctx context.Context, closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.C(ctx).Warnf("Warning: failed to close: %s", err.Error())
	}
}
